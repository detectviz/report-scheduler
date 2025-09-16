package main

import (
	"log"
	"net/http"
	"report-scheduler/backend/internal/api"
	"strings"
)

func main() {
	// 建立一個新的路由多工器
	mux := http.NewServeMux()

	// 註冊資料來源相關的 API 端點
	// 為了更精確地匹配 /api/v1/datasources 和 /api/v1/datasources/{id}
	// 我們建立一個更明確的處理器
	datasourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		// 移除路徑前綴以判斷是否有 ID
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/datasources/")

		if id != "" { // 路徑包含 ID, e.g., /api/v1/datasources/some-id
			switch r.Method {
			case http.MethodGet:
				api.GetDataSourceByID(w, r)
			case http.MethodPut:
				api.UpdateDataSource(w, r)
			case http.MethodDelete:
				api.DeleteDataSource(w, r)
			default:
				http.Error(w, "不支援的請求方法", http.StatusMethodNotAllowed)
			}
		} else { // 路徑不含 ID, e.g., /api/v1/datasources
			switch r.Method {
			case http.MethodGet:
				api.GetDataSources(w, r)
			case http.MethodPost:
				api.CreateDataSource(w, r)
			default:
				http.Error(w, "不支援的請求方法", http.StatusMethodNotAllowed)
			}
		}
	}

	mux.HandleFunc("/api/v1/datasources/", datasourcesHandler)

	// 啟動 HTTP 伺服器
	log.Println("伺服器啟動於 http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
