package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/store"

	"github.com/go-chi/chi/v5"
)

// APIHandler 是一個包含應用程式依賴（如資料庫 store）的結構
type APIHandler struct {
	Store store.Store
}

// NewAPIHandler 建立並回傳一個新的 APIHandler
func NewAPIHandler(s store.Store) *APIHandler {
	return &APIHandler{
		Store: s,
	}
}

// --- Utility Functions ---

// respondWithError 是一個輔助函式，用於發送統一格式的 JSON 錯誤訊息
func (h *APIHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON 是一個輔助函式，用於將 payload 編碼為 JSON 並發送
func (h *APIHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// 如果連錯誤訊息本身都無法序列化，就只能回傳一個基本的伺服器錯誤
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- Handler Methods ---

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
