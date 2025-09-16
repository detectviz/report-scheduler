package main

import (
	"log"
	"net/http"
	"report-scheduler/backend/internal/api"
	"report-scheduler/backend/internal/store" // 新增 store import

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 建立資料層的實例
	// 根據 "Factory Provider" 模式，未來這裡可以根據設定檔來決定要建立 MockStore 還是 PostgresStore
	dbStore := store.NewMockStore()

	// 建立 API 處理器的實例，並注入 store
	apiHandler := api.NewAPIHandler(dbStore)

	// 建立一個新的 chi 路由器
	r := chi.NewRouter()

	// 使用中介軟體
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 定義 API 路由群組，並將請求導向到 handler 的方法
	r.Route("/api/v1/datasources", func(r chi.Router) {
		r.Get("/", apiHandler.GetDataSources)
		r.Post("/", apiHandler.CreateDataSource)

		r.Route("/{datasourceID}", func(r chi.Router) {
			r.Get("/", apiHandler.GetDataSourceByID)
			r.Put("/", apiHandler.UpdateDataSource)
			r.Delete("/", apiHandler.DeleteDataSource)
		})
	})

	// 啟動 HTTP 伺服器
	log.Println("伺服器啟動於 http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
