package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// --- ReportDefinition Handler Methods ---

// GetReportDefinitions 處理獲取所有報表定義的請求
func (h *APIHandler) GetReportDefinitions(w http.ResponseWriter, r *http.Request) {
	reports, err := h.Store.GetReportDefinitions(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取報表定義")
		return
	}
	h.respondWithJSON(w, http.StatusOK, reports)
}

// CreateReportDefinition 處理新增報表定義的請求
func (h *APIHandler) CreateReportDefinition(w http.ResponseWriter, r *http.Request) {
	var rd models.ReportDefinition
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.CreateReportDefinition(r.Context(), &rd); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法建立報表定義")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, rd)
}

// GetReportDefinitionByID 處理根據 ID 獲取單一報表定義的請求
func (h *APIHandler) GetReportDefinitionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "reportID")
	rd, err := h.Store.GetReportDefinitionByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取報表定義: "+err.Error())
		return
	}
	if rd == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的報表定義")
		return
	}
	h.respondWithJSON(w, http.StatusOK, rd)
}

// UpdateReportDefinition 處理更新報表定義的請求
func (h *APIHandler) UpdateReportDefinition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "reportID")
	var rd models.ReportDefinition
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.UpdateReportDefinition(r.Context(), id, &rd); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法更新報表定義")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "報表定義 " + id + " 已成功更新"})
}

// DeleteReportDefinition 處理刪除報表定義的請求
func (h *APIHandler) DeleteReportDefinition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "reportID")
	if err := h.Store.DeleteReportDefinition(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法刪除報表定義")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "報表定義 " + id + " 已成功刪除"})
}
