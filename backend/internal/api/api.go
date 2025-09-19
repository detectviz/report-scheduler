package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"

	"github.com/go-chi/chi/v5"
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

// ServeFile 處理提供暫存檔案的請求
func (h *APIHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		h.respondWithError(w, http.StatusBadRequest, "缺少檔案名稱")
		return
	}

	// 建立指向暫存目錄中檔案的完整路徑
	// 警告：這是一個簡化的實作。在正式產品中，需要更嚴格的路徑清理和安全檢查，
	// 以防止目錄遍歷攻擊 (directory traversal)。
	filePath := filepath.Join(os.TempDir(), filename)

	// 檢查檔案是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.respondWithError(w, http.StatusNotFound, "找不到指定的檔案")
		return
	}

	// 提供檔案下載
	http.ServeFile(w, r, filePath)
}
