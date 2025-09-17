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

// newProcessFunc is a helper for tests, mimicking the one in main.go
func newProcessFunc(s store.Store) worker.ProcessFunc {
	return func(task *queue.Task) error {
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return err
		}
		logEntry := &models.HistoryLog{
			ScheduleID:   task.ScheduleID,
			ScheduleName: schedule.Name,
			TriggerTime:  task.CreatedAt,
			Status:       models.LogStatusSuccess,
			Recipients:   schedule.Recipients,
		}
		return s.CreateHistoryLog(context.Background(), logEntry)
	}
}

func TestScheduleAPI_WithRealDB(t *testing.T) {
	handler, dbStore, taskQueue, cleanup := newTestHandler(t)
	defer cleanup()

	// For the trigger test, we need a real worker running in the background.
	processFunc := newProcessFunc(dbStore)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	appWorker.Start()
	defer appWorker.Stop()

	server := httptest.NewServer(handler)
	defer server.Close()

	var createdSchedule models.Schedule

	t.Run("create schedule for trigger test", func(t *testing.T) {
		scheduleJSON := []byte(`{"name": "Test For Trigger", "cron_spec": "0 0 1 1 *", "is_enabled": true}`)
		resp, err := http.Post(server.URL+"/api/v1/schedules", "application/json", bytes.NewBuffer(scheduleJSON))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		err = json.NewDecoder(resp.Body).Decode(&createdSchedule)
		require.NoError(t, err)
	})

	t.Run("trigger schedule manually and check history", func(t *testing.T) {
		require.NotEmpty(t, createdSchedule.ID)
		triggerURL := server.URL + "/api/v1/schedules/" + createdSchedule.ID + "/trigger"
		resp, err := http.Post(triggerURL, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		// Verify that the worker processed the task and created a history log.
		require.Eventually(t, func() bool {
			historyURL := server.URL + "/api/v1/history?schedule_id=" + createdSchedule.ID
			resp, err := http.Get(historyURL)
			if err != nil {
				return false
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return false
			}

			var logs []models.HistoryLog
			if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
				return false
			}
			return len(logs) == 1
		}, 3*time.Second, 50*time.Millisecond, "expected worker to create a history log")
	})
}
