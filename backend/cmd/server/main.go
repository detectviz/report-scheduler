package main

import (
	"log"
	"net/http"
	"report-scheduler/backend/internal/api"
	"report-scheduler/backend/internal/config" // 引入 config
	"report-scheduler/backend/internal/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 載入設定
	cfg, err := config.LoadConfig(".") // 從當前目錄讀取 config.yaml
	if err != nil {
		log.Fatalf("無法載入設定: %v", err)
	}

	// 建立資料層的實例
	// 根據 "Factory Provider" 模式，呼叫工廠函式來建立 Store 實例
	dbStore, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("無法連線到資料庫: %v", err)
	}

	// 建立 API 處理器的實例，並注入 store
	apiHandler := api.NewAPIHandler(dbStore)

	// 建立一個新的 chi 路由器
	r := chi.NewRouter()

	// 使用中介軟體
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// --- API 路由 ---
	// 將 API 路由封裝在一個群組中，方便管理
	r.Route("/api/v1", func(r chi.Router) {
		// Datasources 路由
		r.Route("/datasources", func(r chi.Router) {
			r.Get("/", apiHandler.GetDataSources)
			r.Post("/", apiHandler.CreateDataSource)
			r.Route("/{datasourceID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetDataSourceByID)
				r.Put("/", apiHandler.UpdateDataSource)
				r.Delete("/", apiHandler.DeleteDataSource)
			})
		})

		// Report Definitions 路由
		r.Route("/reports", func(r chi.Router) {
			r.Get("/", apiHandler.GetReportDefinitions)
			r.Post("/", apiHandler.CreateReportDefinition)
			r.Route("/{reportID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetReportDefinitionByID)
				r.Put("/", apiHandler.UpdateReportDefinition)
				r.Delete("/", apiHandler.DeleteReportDefinition)
			})
		})
	})

	// 啟動 HTTP 伺服器
	log.Println("伺服器啟動於 http://localhost:8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("伺服器啟動失敗: %v", err)
	}
}
