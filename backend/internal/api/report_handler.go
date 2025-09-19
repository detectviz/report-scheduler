package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"report-scheduler/backend/internal/generator"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"time"

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

// GenerateReport 處理同步產生單一報表的請求
func (h *APIHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reportID := chi.URLParam(r, "reportID")

	reportDef, err := h.Store.GetReportDefinitionByID(ctx, reportID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取報表定義: "+err.Error())
		return
	}
	if reportDef == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的報表定義")
		return
	}

	dataSource, err := h.Store.GetDataSourceByID(ctx, reportDef.DataSourceID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取資料來源: "+err.Error())
		return
	}
	if dataSource == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到報表定義對應的資料來源")
		return
	}

	genFactory := generator.NewFactory(h.Store, h.Secrets)
	gen, err := genFactory.GetGenerator(dataSource.Type)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法建立報表產生器: "+err.Error())
		return
	}

	fakeTask := &queue.Task{
		ID:        "sync-generate-" + reportID,
		CreatedAt: time.Now(),
	}

	result, err := gen.Generate(fakeTask, dataSource, reportDef)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "產生報表失敗: "+err.Error())
		return
	}

	previewURL := "/api/v1/files/" + filepath.Base(result.FilePath)
	h.respondWithJSON(w, http.StatusOK, map[string]string{"preview_url": previewURL})
}
