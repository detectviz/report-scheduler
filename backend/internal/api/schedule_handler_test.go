package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return fmt.Errorf("test worker could not find schedule %s", task.ScheduleID)
		}

		var lastErr error
		var reportURLs []string
		for _, reportID := range task.ReportIDs {
			reportDef, _ := s.GetReportDefinitionByID(context.Background(), reportID)
			dataSource, _ := s.GetDataSourceByID(context.Background(), reportDef.DataSourceID)
			gen, _ := genFactory.GetGenerator(dataSource.Type)
			result, err := gen.Generate(task, dataSource, reportDef)
			if err != nil {
				lastErr = err
				continue
			}
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

func TestEndToEndTrigger(t *testing.T) {
	handler, dbStore, taskQueue, cleanup := newTestHandler(t)
	defer cleanup()

	secretsManager := secrets.NewMockSecretsManager()
	genFactory := generator.NewFactory(dbStore, secretsManager)
	processFunc := newTestProcessFunc(dbStore, genFactory)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	appWorker.Start()
	defer appWorker.Stop()

	server := httptest.NewServer(handler)
	defer server.Close()

	// --- Test Data Setup ---
	var createdDS models.DataSource
	dsJSON := []byte(`{"name": "E2E DS", "type": "kibana", "url": "http://e2e.test", "auth_type": "none", "status": "verified"}`)
	resp, err := http.Post(server.URL+"/api/v1/datasources", "application/json", bytes.NewBuffer(dsJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&createdDS)
	resp.Body.Close()

	var createdReport models.ReportDefinition
	reportJSON := []byte(`{"name": "E2E Report", "datasource_id": "` + createdDS.ID + `"}`)
	resp, err = http.Post(server.URL+"/api/v1/reports", "application/json", bytes.NewBuffer(reportJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&createdReport)
	resp.Body.Close()

	var createdSchedule models.Schedule
	scheduleJSON := []byte(`{"name": "E2E Schedule", "report_ids": ["` + createdReport.ID + `"]}`)
	resp, err = http.Post(server.URL+"/api/v1/schedules", "application/json", bytes.NewBuffer(scheduleJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&createdSchedule)
	resp.Body.Close()

	// --- Test Trigger ---
	t.Run("trigger schedule manually and check history for report url", func(t *testing.T) {
		triggerURL := server.URL + "/api/v1/schedules/" + createdSchedule.ID + "/trigger"
		resp, err := http.Post(triggerURL, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		// Verify that the worker created a history log with a report URL
		require.Eventually(t, func() bool {
			historyURL := server.URL + "/api/v1/history?schedule_id=" + createdSchedule.ID
			resp, err := http.Get(historyURL)
			if err != nil { return false }
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK { return false }
			var logs []models.HistoryLog
			if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil { return false }
			if len(logs) != 1 { return false }

			log.Printf("Found history log with ReportURL: %s", logs[0].ReportURL)
			return strings.Contains(logs[0].ReportURL, ".pdf")
		}, 3*time.Second, 100*time.Millisecond, "expected worker to create a history log with a report URL")
	})
}
