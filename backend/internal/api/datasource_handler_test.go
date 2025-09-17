package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/store"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// newTestHandler 建立一個使用真實 SqliteStore 的測試路由器。
// 它會回傳一個 http.Handler 和一個用於清理暫存資料庫的函式。
func newTestHandler(t *testing.T) (http.Handler, func()) {
	tempDir, err := ioutil.TempDir("", "test-db-")
	require.NoError(t, err)

	dbPath := tempDir + "/test.db"

	testCfg := config.Config{
		Database: config.DBConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}

	dbStore, err := store.NewStore(testCfg)
	require.NoError(t, err)

	apiHandler := NewAPIHandler(dbStore)
	r := chi.NewRouter()

	// 路由設定必須跟 main.go 完全一樣
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/datasources", func(r chi.Router) {
			r.Get("/", apiHandler.GetDataSources)
			r.Post("/", apiHandler.CreateDataSource)
			r.Route("/{datasourceID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetDataSourceByID)
				r.Put("/", apiHandler.UpdateDataSource)
				r.Delete("/", apiHandler.DeleteDataSource)
			})
		})
		r.Route("/reports", func(r chi.Router) {
			r.Get("/", apiHandler.GetReportDefinitions)
			r.Post("/", apiHandler.CreateReportDefinition)
			r.Route("/{reportID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetReportDefinitionByID)
				r.Put("/", apiHandler.UpdateReportDefinition)
				r.Delete("/", apiHandler.DeleteReportDefinition)
			})
		})
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", apiHandler.GetSchedules)
			r.Post("/", apiHandler.CreateSchedule)
			r.Route("/{scheduleID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetScheduleByID)
				r.Put("/", apiHandler.UpdateSchedule)
				r.Delete("/", apiHandler.DeleteSchedule)
			})
		})
	})

	cleanup := func() {
		dbStore.Close()
		os.RemoveAll(tempDir)
	}

	return r, cleanup
}

func TestDatasourceAPI_WithRealDB(t *testing.T) {
	handler, cleanup := newTestHandler(t)
	defer cleanup()

	server := httptest.NewServer(handler)
	defer server.Close()

	// 1. 開始時，GET all 應該是空的
	t.Run("get all from empty db", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/datasources")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var ds []models.DataSource
		err = json.NewDecoder(resp.Body).Decode(&ds)
		require.NoError(t, err)
		require.Len(t, ds, 0)
	})

	// 2. 建立一個新的 datasource
	var createdDS models.DataSource
	t.Run("create datasource", func(t *testing.T) {
		dsJSON := `{"name": "Test Kibana", "type": "kibana", "url": "http://k.test", "auth_type": "api_token", "status": "unverified"}`
		resp, err := http.Post(server.URL+"/api/v1/datasources", "application/json", bytes.NewBuffer([]byte(dsJSON)))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&createdDS)
		require.NoError(t, err)
		require.NotEmpty(t, createdDS.ID)
		require.Equal(t, "Test Kibana", createdDS.Name)
	})

	// 3. 透過 ID 取得剛剛建立的 datasource
	t.Run("get created datasource by id", func(t *testing.T) {
		require.NotEmpty(t, createdDS.ID, "created datasource ID should not be empty")
		resp, err := http.Get(server.URL + "/api/v1/datasources/" + createdDS.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var fetchedDS models.DataSource
		err = json.NewDecoder(resp.Body).Decode(&fetchedDS)
		require.NoError(t, err)
		require.Equal(t, createdDS.ID, fetchedDS.ID)
		require.Equal(t, createdDS.Name, fetchedDS.Name)
	})

	// 4. 再次 GET all，應該會有一筆資料
	t.Run("get all with one item", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/datasources")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var ds []models.DataSource
		err = json.NewDecoder(resp.Body).Decode(&ds)
		require.NoError(t, err)
		require.Len(t, ds, 1)
		require.Equal(t, createdDS.ID, ds[0].ID)
	})

	// 5. 刪除剛剛建立的 datasource
	t.Run("delete datasource", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/v1/datasources/"+createdDS.ID, nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 6. 再次透過 ID 取得，應該會回傳 404 Not Found
	t.Run("get deleted datasource by id", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/datasources/" + createdDS.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
