package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
)

// APIHandler 是一個包含所有應用程式依賴的結構
type APIHandler struct {
	Store   store.Store
	Secrets secrets.SecretsManager
	Queue   queue.Queue
}

// NewAPIHandler 建立並回傳一個新的 APIHandler
func NewAPIHandler(s store.Store, sm secrets.SecretsManager, q queue.Queue) *APIHandler {
	return &APIHandler{
		Store:   s,
		Secrets: sm,
		Queue:   q,
	}
}

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
