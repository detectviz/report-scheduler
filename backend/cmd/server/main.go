package main

import (
	"log"
	"net/http"
	"report-scheduler/backend/internal/api"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 建立一個新的 chi 路由器
	r := chi.NewRouter()

	// 使用一些推薦的 chi 中介軟體 (middleware)
	// RequestID 為每個請求加上一個唯一的 ID
	// RealIP 從標頭中找出真實的 IP 位址
	// Logger 記錄每個 HTTP 請求的資訊
	// Recoverer 從 panic 中恢復，並回傳 500 錯誤
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 定義 API 路由群組
	r.Route("/api/v1/datasources", func(r chi.Router) {
		r.Get("/", api.GetDataSources)          // GET /api/v1/datasources
		r.Post("/", api.CreateDataSource)       // POST /api/v1/datasources

		// 處理帶有 ID 的子路由
		// chi 會自動將 {datasourceID} 解析出來
		r.Route("/{datasourceID}", func(r chi.Router) {
			r.Get("/", api.GetDataSourceByID)    // GET /api/v1/datasources/{datasourceID}
			r.Put("/", api.UpdateDataSource)    // PUT /api/v1/datasources/{datasourceID}
			r.Delete("/", api.DeleteDataSource) // DELETE /api/v1/datasources/{datasourceID}
		})
	})

	// 啟動 HTTP 伺服器
	log.Println("伺服器啟動於 http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
