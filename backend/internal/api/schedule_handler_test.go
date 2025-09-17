package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/store"
	"report-scheduler/backend/internal/worker"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestProcessFunc is a helper for tests, mimicking the one in main.go
func newTestProcessFunc(s store.Store) worker.ProcessFunc {
	return func(task *queue.Task) error {
		startTime := time.Now()
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return err
		}

		// Simulate fetching reports and datasources
		for _, reportID := range task.ReportIDs {
			_, _ = s.GetReportDefinitionByID(context.Background(), reportID)
		}

		duration := time.Since(startTime)
		logEntry := &models.HistoryLog{
			ScheduleID:        task.ScheduleID,
			ScheduleName:      schedule.Name,
			TriggerTime:       task.CreatedAt,
			ExecutionDuration: duration.Milliseconds(),
			Status:            models.LogStatusSuccess,
			Recipients:        schedule.Recipients,
		}
		return s.CreateHistoryLog(context.Background(), logEntry)
	}
}

func TestEndToEndTrigger(t *testing.T) {
	handler, dbStore, taskQueue, cleanup := newTestHandler(t)
	defer cleanup()

	processFunc := newTestProcessFunc(dbStore)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	appWorker.Start()
	defer appWorker.Stop()

	server := httptest.NewServer(handler)
	defer server.Close()

	// 1. Create a DataSource
	dsJSON := []byte(`{"name": "E2E DS", "type": "kibana", "url": "http://e2e.test", "auth_type": "none", "status": "verified"}`)
	resp, err := http.Post(server.URL+"/api/v1/datasources", "application/json", bytes.NewBuffer(dsJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdDS models.DataSource
	json.NewDecoder(resp.Body).Decode(&createdDS)
	resp.Body.Close()

	// 2. Create a ReportDefinition
	reportJSON := []byte(`{"name": "E2E Report", "datasource_id": "` + createdDS.ID + `"}`)
	resp, err = http.Post(server.URL+"/api/v1/reports", "application/json", bytes.NewBuffer(reportJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdReport models.ReportDefinition
	json.NewDecoder(resp.Body).Decode(&createdReport)
	resp.Body.Close()

	// 3. Create a Schedule
	scheduleJSON := []byte(`{"name": "E2E Schedule", "report_ids": ["` + createdReport.ID + `"]}`)
	resp, err = http.Post(server.URL+"/api/v1/schedules", "application/json", bytes.NewBuffer(scheduleJSON))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdSchedule models.Schedule
	json.NewDecoder(resp.Body).Decode(&createdSchedule)
	resp.Body.Close()

	// 4. Trigger the Schedule
	triggerURL := server.URL + "/api/v1/schedules/" + createdSchedule.ID + "/trigger"
	resp, err = http.Post(triggerURL, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)

	// 5. Verify that a history log was created
	require.Eventually(t, func() bool {
		historyURL := server.URL + "/api/v1/history?schedule_id=" + createdSchedule.ID
		resp, err := http.Get(historyURL)
		if err != nil { return false }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK { return false }
		var logs []models.HistoryLog
		if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil { return false }
		if len(logs) != 1 { return false }
		require.Equal(t, "E2E Schedule", logs[0].ScheduleName)
		return true
	}, 3*time.Second, 100*time.Millisecond, "expected worker to create a history log for the triggered schedule")
}
