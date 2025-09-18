package api

import (
	"fmt"
	"net/http"
	"report-scheduler/backend/internal/queue"
	"time"

	"github.com/go-chi/chi/v5"
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

// ResendHistoryLog 處理重寄特定歷史紀錄的請求
func (h *APIHandler) ResendHistoryLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logID := chi.URLParam(r, "log_id")
	if logID == "" {
		h.respondWithError(w, http.StatusBadRequest, "缺少 log_id")
		return
	}

	// 1. 根據 logID 獲取歷史紀錄
	logEntry, err := h.Store.GetHistoryLogByID(ctx, logID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("獲取歷史紀錄失敗: %v", err))
		return
	}
	if logEntry == nil {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的歷史紀錄")
		return
	}

	// 2. 獲取原始的排程，以取得報表 ID 列表
	schedule, err := h.Store.GetScheduleByID(ctx, logEntry.ScheduleID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("獲取原始排程失敗: %v", err))
		return
	}
	if schedule == nil {
		// 如果原始排程被刪除，就無法重寄
		h.respondWithError(w, http.StatusNotFound, "找不到此紀錄對應的原始排程，可能已被刪除")
		return
	}

	// 3. 建立一個新的任務並推入佇列
	// 規格要求使用原始的 TriggerTime 作為時間基準，但在此 MVP 中，我們重新建立一個新的任務
	task := &queue.Task{
		ID:         fmt.Sprintf("resend-%s-%d", logEntry.ID, time.Now().Unix()),
		ScheduleID: schedule.ID,
		ReportIDs:  schedule.ReportIDs,
		CreatedAt:  time.Now(), // 使用當前時間作為新的觸發時間
	}

	if err := h.Queue.Enqueue(ctx, task); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("無法將重寄任務加入佇列: %v", err))
		return
	}

	h.respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "已成功將重寄任務加入佇列"})
}
