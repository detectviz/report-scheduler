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

func TestReportAPI_WithRealDB(t *testing.T) {
	handler, _, _, cleanup := newTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(handler)
	defer server.Close()

	// --- 準備前置資料：建立一個 Datasource ---
	dsJSON := `{"name": "Prerequisite DS", "type": "kibana", "url": "http://ds.test", "auth_type": "none", "status": "verified"}`
	resp, err := http.Post(server.URL+"/api/v1/datasources", "application/json", bytes.NewBuffer([]byte(dsJSON)))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var createdDS models.DataSource
	err = json.NewDecoder(resp.Body).Decode(&createdDS)
	require.NoError(t, err)
	resp.Body.Close()

	// --- 開始測試 Report API ---

	// 1. 開始時，GET all reports 應該只有一筆種子資料
	t.Run("get all reports from seeded table", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/reports")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var reports []models.ReportDefinition
		err = json.NewDecoder(resp.Body).Decode(&reports)
		require.NoError(t, err)
		require.Len(t, reports, 1)
		require.Equal(t, "report-1", reports[0].ID)
	})

	// 2. 建立一個新的 report definition
	var createdReport models.ReportDefinition
	t.Run("create report definition", func(t *testing.T) {
		elements := models.ReportElements{
			{ID: "viz-1", Type: models.VisualizationType, Title: "My Viz"},
		}
		elementsJSON, _ := json.Marshal(elements)
		reportJSON := []byte(`{
			"name": "Daily Sales Report",
			"datasource_id": "` + createdDS.ID + `",
			"time_range": "now-24h",
			"elements": ` + string(elementsJSON) + `
		}`)

		resp, err := http.Post(server.URL+"/api/v1/reports", "application/json", bytes.NewBuffer(reportJSON))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&createdReport)
		require.NoError(t, err)
		require.NotEmpty(t, createdReport.ID)
		require.Equal(t, "Daily Sales Report", createdReport.Name)
		require.Len(t, createdReport.Elements, 1)
		require.Equal(t, "viz-1", createdReport.Elements[0].ID)
	})

	// 3. 透過 ID 取得剛剛建立的 report
	t.Run("get created report by id", func(t *testing.T) {
		require.NotEmpty(t, createdReport.ID, "created report ID should not be empty")
		resp, err := http.Get(server.URL + "/api/v1/reports/" + createdReport.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var fetchedReport models.ReportDefinition
		err = json.NewDecoder(resp.Body).Decode(&fetchedReport)
		require.NoError(t, err)
		require.Equal(t, createdReport.ID, fetchedReport.ID)
	})

	// 4. 刪除剛剛建立的 report
	t.Run("delete report", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/reports/"+createdReport.ID, nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. 再次透過 ID 取得，應該會回傳 404 Not Found
	t.Run("get deleted report by id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/reports/" + createdReport.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
