package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
)

// APIHandler 是一個包含所有應用程式依賴的結構，例如資料庫 store 和憑證管理器。
// 所有的 HTTP 處理器都將是這個結構的方法。
type APIHandler struct {
	Store   store.Store
	Secrets secrets.SecretsManager
}

// NewAPIHandler 建立並回傳一個新的 APIHandler
func NewAPIHandler(s store.Store, sm secrets.SecretsManager) *APIHandler {
	return &APIHandler{
		Store:   s,
		Secrets: sm,
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
