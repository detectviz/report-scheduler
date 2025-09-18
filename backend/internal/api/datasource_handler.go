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

	// 在真實應用中，這裡可能還需要驗證 ds 的內容
	if err := h.Store.CreateDataSource(r.Context(), &ds); err != nil {
		log.Printf("Error creating data source: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "無法建立資料來源")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, ds) // 回傳建立後包含 ID 的物件
}

// GetDataSourceByID 處理根據 ID 獲取單一資料來源的請求
func (h *APIHandler) GetDataSourceByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	ds, err := h.Store.GetDataSourceByID(r.Context(), id)
	if err != nil {
		// 這邊可以更細緻地處理 not found 錯誤，但暫時先用 500
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

	// 1. 從資料庫獲取資料來源
	ds, err := h.Store.GetDataSourceByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取資料來源: "+err.Error())
		return
	}
	if ds == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的資料來源")
		return
	}

	// 2. 從憑證管理器獲取憑證 (此處為模擬)
	// 在真實世界中，ds.CredentialsRef 將會被傳入
	creds, err := h.Secrets.GetCredentials("kv/report-scheduler/kibana-prod")
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取憑證: "+err.Error())
		return
	}

	// 3. 對外部服務發起測試請求
	req, err := http.NewRequestWithContext(r.Context(), "GET", ds.URL, nil)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法建立驗證請求")
		return
	}

	// 在真實世界中，我們會根據 ds.AuthType 和 creds 來設定認證標頭
	// 例如: req.Header.Set("Authorization", "Bearer "+creds.Token)
	log.Printf("正在使用 Token '%s' 驗證資料來源 %s...", creds.Token, ds.URL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	// 4. 根據驗證結果更新資料庫中的狀態
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
