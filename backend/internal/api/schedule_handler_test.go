package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"report-scheduler/backend/internal/generator"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
	"report-scheduler/backend/internal/worker"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestProcessFunc is a helper for tests, mimicking the one in main.go
func newTestProcessFunc(s store.Store, genFactory *generator.Factory) worker.ProcessFunc {
	return func(task *queue.Task) error {
		startTime := time.Now()
		log.Printf("測試 Worker: 開始處理任務 %s", task.ID)

		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return fmt.Errorf("test worker could not find schedule %s", task.ScheduleID)
		}

		var lastErr error
		var reportURLs []string
		for _, reportID := range task.ReportIDs {
			reportDef, err := s.GetReportDefinitionByID(context.Background(), reportID)
			if err != nil || reportDef == nil {
				log.Printf("任務 %s: 錯誤：找不到報表定義 %s，跳過", task.ID, reportID)
				lastErr = err
				continue
			}

			dataSource, err := s.GetDataSourceByID(context.Background(), reportDef.DataSourceID)
			if err != nil || dataSource == nil {
				log.Printf("任務 %s: 錯誤：找不到報表 '%s' 的資料來源 %s，跳過", task.ID, reportDef.Name, reportDef.DataSourceID)
				lastErr = err
				continue
			}

			gen, err := genFactory.GetGenerator(dataSource.Type)
			if err != nil {
				log.Printf("任務 %s: 錯誤：找不到報表 '%s' 的產生器，跳過", task.ID, reportDef.Name)
				lastErr = err
				continue
			}

			result, err := gen.Generate(task, dataSource, reportDef)
			if err != nil {
				log.Printf("任務 %s: 錯誤：產生報表 '%s' 失敗: %v", task.ID, reportDef.Name, err)
				lastErr = err
				continue
			}

			log.Printf("任務 %s: 報表 '%s' 產生成功，檔案位於 %s", task.ID, reportDef.Name, result.FilePath)
			reportURLs = append(reportURLs, result.FilePath)
		}

		duration := time.Since(startTime)
		logEntry := &models.HistoryLog{
			ScheduleID:        task.ScheduleID,
			ScheduleName:      schedule.Name,
			TriggerTime:       task.CreatedAt,
			ExecutionDuration: duration.Milliseconds(),
			Recipients:        schedule.Recipients,
		}

		if lastErr != nil {
			logEntry.Status = models.LogStatusFailed
			logEntry.ErrorMessage = lastErr.Error()
		} else {
			logEntry.Status = models.LogStatusSuccess
			logEntry.ReportURL = strings.Join(reportURLs, ", ")
		}
		return s.CreateHistoryLog(context.Background(), logEntry)
	}
}

func TestEndToEndTriggerWithReportGeneration(t *testing.T) {
	handler, dbStore, taskQueue, cleanup := newTestHandler(t)
	defer cleanup()

	// 1. 建立一個模擬外部服務的 http server (Mock Kibana)
	mockKibana := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "ApiKey mock-api-token-12345", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("dummy-pdf-content"))
	}))
	defer mockKibana.Close()

	// 2. 建立依賴項，並啟動 Worker
	secretsManager := secrets.NewMockSecretsManager()
	genFactory := generator.NewFactory(dbStore, secretsManager)
	processFunc := newTestProcessFunc(dbStore, genFactory)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	appWorker.Start()
	defer appWorker.Stop()

	// 3. 建立 API 伺服器
	server := httptest.NewServer(handler)
	defer server.Close()

	// 4. 準備測試資料
	// a. 建立指向 Mock Kibana 的資料來源
	dsJSON := []byte(fmt.Sprintf(`{"name": "E2E DS", "type": "kibana", "url": "%s", "auth_type": "api_token", "credentials_ref": "kv/report-scheduler/kibana-prod"}`, mockKibana.URL))
	resp, err := http.Post(server.URL+"/api/v1/datasources", "application/json", bytes.NewBuffer(dsJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdDS models.DataSource
	json.NewDecoder(resp.Body).Decode(&createdDS)
	resp.Body.Close()

	// b. 建立報表定義
	reportJSON := []byte(fmt.Sprintf(`{"name": "E2E Report", "datasource_id": "%s", "elements": [{"id": "my-dashboard"}]}`, createdDS.ID))
	resp, err = http.Post(server.URL+"/api/v1/reports", "application/json", bytes.NewBuffer(reportJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdReport models.ReportDefinition
	json.NewDecoder(resp.Body).Decode(&createdReport)
	resp.Body.Close()

	// c. 建立排程
	scheduleJSON := []byte(fmt.Sprintf(`{"name": "E2E Schedule", "report_ids": ["%s"]}`, createdReport.ID))
	resp, err = http.Post(server.URL+"/api/v1/schedules", "application/json", bytes.NewBuffer(scheduleJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdSchedule models.Schedule
	json.NewDecoder(resp.Body).Decode(&createdSchedule)
	resp.Body.Close()

	// 5. 執行觸發
	triggerURL := server.URL + "/api/v1/schedules/" + createdSchedule.ID + "/trigger"
	resp, err = http.Post(triggerURL, "application/json", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	resp.Body.Close()

	// 6. 驗證結果
	require.Eventually(t, func() bool {
		historyURL := server.URL + "/api/v1/history?schedule_id=" + createdSchedule.ID
		resp, err := http.Get(historyURL)
		if err != nil { return false }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK { return false }
		var logs []models.HistoryLog
		if json.NewDecoder(resp.Body).Decode(&logs) != nil || len(logs) != 1 {
			return false
		}

		logEntry := logs[0]
		require.Equal(t, models.LogStatusSuccess, logEntry.Status)
		require.NotEmpty(t, logEntry.ReportURL, "ReportURL should be populated")

		// 驗證檔案內容
		fileContent, err := ioutil.ReadFile(logEntry.ReportURL)
		require.NoError(t, err)
		require.Equal(t, "dummy-pdf-content", string(fileContent))

		// 清理暫存檔案
		os.Remove(logEntry.ReportURL)
		return true
	}, 5*time.Second, 100*time.Millisecond, "expected worker to create a history log with a valid report file")
}
