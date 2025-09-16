package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupTestRouter 建立一個包含我們所有 API 路由的測試路由器
// 這讓我們可以測試完整的請求 -> 路由 -> 處理器流程
func setupTestRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/datasources", func(r chi.Router) {
		r.Get("/", GetDataSources)
		r.Post("/", CreateDataSource)
		r.Route("/{datasourceID}", func(r chi.Router) {
			r.Get("/", GetDataSourceByID)
			r.Put("/", UpdateDataSource)
			r.Delete("/", DeleteDataSource)
		})
	})
	return r
}

// TestDatasourceAPI 是一個整合測試，驗證所有 datasource 相關的端點
func TestDatasourceAPI(t *testing.T) {
	// 建立一個包含完整路由的測試伺服器
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// 定義測試案例
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		verify     func(t *testing.T, resp *http.Response)
	}{
		{
			name:       "GET /api/v1/datasources (取得所有資源)",
			method:     http.MethodGet,
			path:       "/api/v1/datasources",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, resp *http.Response) {
				var ds []models.DataSource
				if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
					t.Fatalf("無法解析 JSON: %v", err)
				}
				if len(ds) < 1 {
					t.Error("預期至少回傳一個資料來源")
				}
			},
		},
		{
			name:       "GET /api/v1/datasources/{id} (取得單一資源)",
			method:     http.MethodGet,
			path:       "/api/v1/datasources/test-id-123",
			wantStatus: http.StatusOK,
			verify: func(t *testing.T, resp *http.Response) {
				var ds models.DataSource
				if err := json.NewDecoder(resp.Body).Decode(&ds); err != nil {
					t.Fatalf("無法解析 JSON: %v", err)
				}
				if ds.ID != "test-id-123" {
					t.Errorf("預期 ID 為 'test-id-123', 得到 '%s'", ds.ID)
				}
			},
		},
		{
			name:       "POST /api/v1/datasources (建立資源)",
			method:     http.MethodPost,
			path:       "/api/v1/datasources",
			wantStatus: http.StatusCreated,
			verify:     nil, // 只檢查狀態碼
		},
		{
			name:       "PUT /api/v1/datasources/{id} (更新資源)",
			method:     http.MethodPut,
			path:       "/api/v1/datasources/update-id-456",
			wantStatus: http.StatusOK,
			verify:     nil, // 只檢查狀態碼
		},
		{
			name:       "DELETE /api/v1/datasources/{id} (刪除資源)",
			method:     http.MethodDelete,
			path:       "/api/v1/datasources/delete-id-789",
			wantStatus: http.StatusOK,
			verify:     nil, // 只檢查狀態碼
		},
		{
			name:       "GET /api/v1/datasources/bad/path (不存在的路徑)",
			method:     http.MethodGet,
			path:       "/api/v1/datasources/bad/path",
			wantStatus: http.StatusNotFound,
			verify:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, server.URL+tt.path, nil)
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
