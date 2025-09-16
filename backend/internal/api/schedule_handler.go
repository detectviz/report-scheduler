package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// --- Schedule Handler Methods ---

// GetSchedules 處理獲取所有排程的請求
func (h *APIHandler) GetSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.Store.GetSchedules(r.Context())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取排程")
		return
	}
	h.respondWithJSON(w, http.StatusOK, schedules)
}

// CreateSchedule 處理新增排程的請求
func (h *APIHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var s models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.CreateSchedule(r.Context(), &s); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法建立排程")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, s)
}

// GetScheduleByID 處理根據 ID 獲取單一排程的請求
func (h *APIHandler) GetScheduleByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scheduleID")
	s, err := h.Store.GetScheduleByID(r.Context(), id)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取排程: "+err.Error())
		return
	}
	if s == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的排程")
		return
	}
	h.respondWithJSON(w, http.StatusOK, s)
}

// UpdateSchedule 處理更新排程的請求
func (h *APIHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scheduleID")
	var s models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "無效的請求內容")
		return
	}
	defer r.Body.Close()

	if err := h.Store.UpdateSchedule(r.Context(), id, &s); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法更新排程")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "排程 " + id + " 已成功更新"})
}

// DeleteSchedule 處理刪除排程的請求
func (h *APIHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "scheduleID")
	if err := h.Store.DeleteSchedule(r.Context(), id); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法刪除排程")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "排程 " + id + " 已成功刪除"})
}
