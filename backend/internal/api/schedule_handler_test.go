package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScheduleAPI_WithRealDB(t *testing.T) {
	handler, cleanup := newTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(handler)
	defer server.Close()

	// --- 準備前置資料：建立 Datasource 和 ReportDefinition ---
	// (在真實測試中，我們可能需要先建立這些依賴項，但為簡化，此處省略)
	// 假設 report ID "rep-1" 和 "rep-2" 是有效的

	// --- 開始測試 Schedule API ---

	// 1. 開始時，GET all schedules 應該是空的
	t.Run("get all schedules from empty table", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/schedules")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var schedules []models.Schedule
		err = json.NewDecoder(resp.Body).Decode(&schedules)
		require.NoError(t, err)
		require.Len(t, schedules, 0)
	})

	// 2. 建立一個新的 schedule
	var createdSchedule models.Schedule
	t.Run("create schedule", func(t *testing.T) {
		recipients := models.Recipients{To: []string{"test@example.com"}}
		recipientsJSON, _ := json.Marshal(recipients)
		reportIDs := models.ReportIDList{"rep-1", "rep-2"}
		reportIDsJSON, _ := json.Marshal(reportIDs)

		scheduleJSON := []byte(`{
			"name": "Weekly Summary",
			"cron_spec": "0 0 * * 1",
			"timezone": "Asia/Taipei",
			"recipients": ` + string(recipientsJSON) + `,
			"email_subject": "Weekly Report",
			"email_body": "Here is your report.",
			"report_ids": ` + string(reportIDsJSON) + `,
			"is_enabled": true
		}`)

		resp, err := http.Post(server.URL+"/api/v1/schedules", "application/json", bytes.NewBuffer(scheduleJSON))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&createdSchedule)
		require.NoError(t, err)
		require.NotEmpty(t, createdSchedule.ID)
		require.Equal(t, "Weekly Summary", createdSchedule.Name)
		require.Len(t, createdSchedule.ReportIDs, 2)
		require.Equal(t, "test@example.com", createdSchedule.Recipients.To[0])
	})

	// 3. 透過 ID 取得剛剛建立的 schedule
	t.Run("get created schedule by id", func(t *testing.T) {
		require.NotEmpty(t, createdSchedule.ID)
		resp, err := http.Get(server.URL + "/api/v1/schedules/" + createdSchedule.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var fetchedSchedule models.Schedule
		err = json.NewDecoder(resp.Body).Decode(&fetchedSchedule)
		require.NoError(t, err)
		require.Equal(t, createdSchedule.ID, fetchedSchedule.ID)
	})

	// 4. 刪除剛剛建立的 schedule
	t.Run("delete schedule", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/schedules/"+createdSchedule.ID, nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. 再次透過 ID 取得，應該會回傳 404 Not Found
	t.Run("get deleted schedule by id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/schedules/" + createdSchedule.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
