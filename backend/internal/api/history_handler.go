package api

import (
	"net/http"
)

// GetHistory 處理獲取執行歷史紀錄的請求
// 它會根據查詢參數中的 `schedule_id` 來過濾結果
func (h *APIHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	if scheduleID == "" {
		h.respondWithError(w, http.StatusBadRequest, "缺少 'schedule_id' 查詢參數")
		return
	}

	logs, err := h.Store.GetHistoryLogs(r.Context(), scheduleID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "無法獲取歷史紀錄")
		return
	}

	// 即使沒有紀錄，也回傳一個空的陣列而不是 null
	if logs == nil {
		h.respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	h.respondWithJSON(w, http.StatusOK, logs)
}
