package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHistoryAPI_WithRealDB(t *testing.T) {
	handler, dbStore, cleanup := newTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(handler)
	defer server.Close()

	// --- 準備前置資料 ---
	// 1. 建立一個 Schedule，這樣我們才有 schedule ID 可以用
	schedule := &models.Schedule{Name: "History Test Schedule", CronSpec: "* * * * *"}
	err := dbStore.CreateSchedule(context.Background(), schedule)
	require.NoError(t, err)

	// 2. 為這個 schedule 手動建立兩筆歷史紀錄
	logEntry1 := &models.HistoryLog{
		ScheduleID:   schedule.ID,
		ScheduleName: schedule.Name,
		TriggerTime:  time.Now().Add(-1 * time.Hour),
		Status:       models.LogStatusSuccess,
	}
	err = dbStore.CreateHistoryLog(context.Background(), logEntry1)
	require.NoError(t, err)

	logEntry2 := &models.HistoryLog{
		ScheduleID:   schedule.ID,
		ScheduleName: schedule.Name,
		TriggerTime:  time.Now(),
		Status:       models.LogStatusFailed,
		ErrorMessage: "something went wrong",
	}
	err = dbStore.CreateHistoryLog(context.Background(), logEntry2)
	require.NoError(t, err)

	// 3. 建立另一個不相關的 schedule 和它的歷史紀錄，以確保我們的查詢有正確過濾
	otherSchedule := &models.Schedule{Name: "Other Schedule", CronSpec: "* * * * *"}
	err = dbStore.CreateSchedule(context.Background(), otherSchedule)
	require.NoError(t, err)
	otherLog := &models.HistoryLog{ScheduleID: otherSchedule.ID, ScheduleName: otherSchedule.Name, TriggerTime: time.Now(), Status: models.LogStatusSuccess}
	err = dbStore.CreateHistoryLog(context.Background(), otherLog)
	require.NoError(t, err)

	// --- 開始測試 History API ---
	t.Run("get history by schedule id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/history?schedule_id=" + schedule.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var logs []models.HistoryLog
		err = json.NewDecoder(resp.Body).Decode(&logs)
		require.NoError(t, err)

		// 驗證只回傳了正確的 schedule ID 的紀錄，且數量為 2
		require.Len(t, logs, 2)
		// 驗證回傳的順序是依照 trigger_time 降序排列
		require.Equal(t, logEntry2.ID, logs[0].ID)
		require.Equal(t, logEntry1.ID, logs[1].ID)
	})

	t.Run("get history for non-existent schedule id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/history?schedule_id=non-existent-id")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var logs []models.HistoryLog
		err = json.NewDecoder(resp.Body).Decode(&logs)
		require.NoError(t, err)
		require.Len(t, logs, 0) // 應該回傳空的陣列
	})

	t.Run("get history without schedule id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/history")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode) // 應該回傳錯誤，因為 schedule_id 是必要的
	})
}
