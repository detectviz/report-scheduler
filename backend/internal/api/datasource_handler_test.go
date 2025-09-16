package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/store" // 引入 store package
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupTestRouter 現在會建立一個 MockStore 並將其注入到 APIHandler
func setupTestRouter() http.Handler {
	mockStore := store.NewMockStore()
	apiHandler := NewAPIHandler(mockStore)

	r := chi.NewRouter()
	r.Route("/api/v1/datasources", func(r chi.Router) {
		r.Get("/", apiHandler.GetDataSources)
		r.Post("/", apiHandler.CreateDataSource)
		r.Route("/{datasourceID}", func(r chi.Router) {
			r.Get("/", apiHandler.GetDataSourceByID)
			r.Put("/", apiHandler.UpdateDataSource)
			r.Delete("/", apiHandler.DeleteDataSource)
		})
	})
	return r
}

// TestDatasourceAPI 是一個整合測試，驗證所有 datasource 相關的端點
func TestDatasourceAPI(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// 測試案例
	tests := []struct {
		name       string
		method     string
		path       string
		body       []byte
		wantStatus int
		verify     func(t *testing.T, resp *http.Response)
	}{
		{
			name:       "GET /api/v1/datasources",
			method:     http.MethodGet,
			path:       "/api/v1/datasources",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, resp *http.Response) {
				var ds []models.DataSource
				if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
					t.Fatalf("無法解析 JSON: %v", err)
				}
				if len(ds) == 0 || ds[0].ID != "mock-ds-1" {
					t.Errorf("回傳的資料不符合預期")
				}
			},
		},
		{
			name:       "GET /api/v1/datasources/{id}",
			method:     http.MethodGet,
			path:       "/api/v1/datasources/get-id-123",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, resp *http.Response) {
				var ds models.DataSource
				if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
					t.Fatalf("無法解析 JSON: %v", err)
				}
				if ds.ID != "get-id-123" {
					t.Errorf("預期 ID 為 'get-id-123', 得到 '%s'", ds.ID)
				}
			},
		},
		{
			name:       "POST /api/v1/datasources",
			method:     http.MethodPost,
			path:       "/api/v1/datasources",
			body:       []byte(`{"name": "new ds"}`),
			wantStatus: http.StatusCreated,
			verify: func(t *testing.T, resp *http.Response) {
				var ds models.DataSource
				if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
					t.Fatalf("無法解析 JSON: %v", err)
				}
				if ds.ID != "new-mock-id" {
					t.Errorf("預期建立後的 ID 為 'new-mock-id', 得到 '%s'", ds.ID)
				}
			},
		},
		{
			name:       "PUT /api/v1/datasources/{id}",
			method:     http.MethodPut,
			path:       "/api/v1/datasources/update-id-456",
			body:       []byte(`{"name": "updated ds"}`),
			wantStatus: http.StatusOK,
		},
		{
			name:       "DELETE /api/v1/datasources/{id}",
			method:     http.MethodDelete,
			path:       "/api/v1/datasources/delete-id-789",
			wantStatus: http.StatusOK,
		},
		{
			name:       "GET /api/v1/datasources/not/found",
			method:     http.MethodGet,
			path:       "/api/v1/datasources/not/found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewBuffer(tt.body))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("請求失敗: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("預期狀態碼 %d, 得到 %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}
