package api

import (
	"encoding/json"
	"log"
	"net/http"
	"report-scheduler/backend/internal/models"
	"time"

	"github.com/go-chi/chi/v5"
)

// --- DataSource Handler Methods ---

// GetDataSources 處理獲取所有資料來源的請求
func (h *APIHandler) GetDataSources(w http.ResponseWriter, r *http.Request) {
	dataSources, err := h.Store.GetDataSources(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusOK, dataSources)
}

// CreateDataSource 處理新增資料來源的請求
func (h *APIHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	var ds models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&ds); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.CreateDataSource(r.Context(), &ds); err != nil {
		log.Printf("Error creating data source: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "無法建立資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, ds)
}

// GetDataSourceByID 處理根據 ID 獲取單一資料來源的請求
func (h *APIHandler) GetDataSourceByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	ds, err := h.Store.GetDataSourceByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取資料來源: "+err.Error())
		return
	}
	if ds == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusOK, ds)
}

// UpdateDataSource 處理更新資料來源的請求
func (h *APIHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	var ds models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&ds); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.UpdateDataSource(r.Context(), id, &ds); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法更新資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "資料來源 " + id + " 已成功更新"})
}

// DeleteDataSource 處理刪除資料來源的請求
func (h *APIHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	if err := h.Store.DeleteDataSource(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法刪除資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "資料來源 " + id + " 已成功刪除"})
}

// ValidateDataSource 處理驗證資料來源連線的請求
func (h *APIHandler) ValidateDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	ds, err := h.Store.GetDataSourceByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取資料來源: "+err.Error())
		return
	}
	if ds == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的資料來源")
		return
	}

	creds, err := h.Secrets.GetCredentials("kv/report-scheduler/kibana-prod")
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取憑證: "+err.Error())
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), "GET", ds.URL, nil)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法建立驗證請求")
		return
	}

	log.Printf("正在使用 Token '%s' 驗證資料來源 %s...", creds.Token, ds.URL)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		ds.Status = models.Error
		if err := h.Store.UpdateDataSource(r.Context(), ds.ID, ds); err != nil {
			h.respondWithError(w, http.StatusInternalServerError, "更新資料來源狀態失敗: "+err.Error())
		} else {
			h.respondWithError(w, http.StatusBadGateway, "驗證失敗：無法連線到資料來源")
		}
		return
	}
	resp.Body.Close()

	ds.Status = models.Verified
	if err := h.Store.UpdateDataSource(r.Context(), ds.ID, ds); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "更新資料來源狀態失敗: "+err.Error())
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "verified", "message": "資料來源連線驗證成功"})
}

// GetDataSourceElements 處理獲取資料來源下可用元素的請求 (模擬)
func (h *APIHandler) GetDataSourceElements(w http.ResponseWriter, r *http.Request) {
	dataSourceID := chi.URLParam(r, "datasourceID")
	log.Printf("正在為資料來源 %s 獲取可用元素...", dataSourceID)

	if dataSourceID == "ds-4" {
		demoElements := []models.AvailableElement{
			{ID: "security-detection-rule-monitoring-default", Type: "dashboard", Title: "[Demo] Security Detection Rule Monitoring"},
			{ID: "kubernetes-f4dc26db-1b53-4ea2-a78b-1bfab8ea267c", Type: "dashboard", Title: "[Demo] Kubernetes Overview"},
			{ID: "elastic_agent-0600ffa0-6b5e-11ed-98de-67bdecd21824", Type: "dashboard", Title: "[Demo] Elastic Agent Overview"},
		}
		h.respondWithJSON(w, http.StatusOK, demoElements)
		return
	}

	defaultMockElements := []models.AvailableElement{
		{ID: "kibana:dashboard:722b74f0-b882-11e8-a6d9-e546fe2bba5f", Type: "dashboard", Title: "[eCommerce] Revenue Dashboard"},
		{ID: "kibana:visualization:89382180-b883-11e8-a6d9-e546fe2bba5f", Type: "visualization", Title: "[Flights] Flight Count and Average Ticket Price"},
		{ID: "kibana:dashboard:a5419300-b883-11e8-a6d9-e546fe2bba5f", Type: "dashboard", Title: "[Logs] Web Traffic"},
	}
	h.respondWithJSON(w, http.StatusOK, defaultMockElements)
}
