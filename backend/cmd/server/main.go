package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"report-scheduler/backend/internal/api"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/scheduler"
	"report-scheduler/backend/internal/store"
	"report-scheduler/backend/internal/worker"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// handleReportTask 是 Worker 用來處理報表產生任務的實際函式
func handleReportTask(task *queue.Task) error {
	log.Printf("開始處理報表任務 %s, 來自排程 %s", task.ID, task.ScheduleID)
	// TODO: 在這裡實作真正的報表產生邏輯
	// 1. 根據 task.ReportIDs 從資料庫獲取報表定義
	// 2. 使用 puppeteer 或 API 擷取報表內容
	// 3. 格式化為 PDF/CSV
	// 4. 寄送郵件
	time.Sleep(5 * time.Second) // 模擬耗時的工作
	log.Printf("成功處理完報表任務 %s", task.ID)
	return nil
}

func main() {
	// 1. 載入設定
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("無法載入設定: %v", err)
	}

	// 2. 建立資料層實例
	dbStore, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("無法連線到資料庫: %v", err)
	}

	// 3. 建立任務佇列
	taskQueue := queue.NewInMemoryQueue(100)

	// 4. 建立排程器服務
	appScheduler := scheduler.NewScheduler(dbStore, taskQueue)

	// 5. 建立 Worker 服務，並傳入真正的處理函式
	appWorker := worker.NewWorker(taskQueue, handleReportTask)

	// 6. 建立 API 處理器
	apiHandler := api.NewAPIHandler(dbStore)

	// 7. 建立 HTTP 路由器
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/datasources", func(r chi.Router) {
			r.Get("/", apiHandler.GetDataSources)
			r.Post("/", apiHandler.CreateDataSource)
			r.Route("/{datasourceID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetDataSourceByID)
				r.Put("/", apiHandler.UpdateDataSource)
				r.Delete("/", apiHandler.DeleteDataSource)
			})
		})
		r.Route("/reports", func(r chi.Router) {
			r.Get("/", apiHandler.GetReportDefinitions)
			r.Post("/", apiHandler.CreateReportDefinition)
			r.Route("/{reportID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetReportDefinitionByID)
				r.Put("/", apiHandler.UpdateReportDefinition)
				r.Delete("/", apiHandler.DeleteReportDefinition)
			})
		})
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", apiHandler.GetSchedules)
			r.Post("/", apiHandler.CreateSchedule)
			r.Route("/{scheduleID}", func(r chi.Router) {
				r.Get("/", apiHandler.GetScheduleByID)
				r.Put("/", apiHandler.UpdateSchedule)
				r.Delete("/", apiHandler.DeleteSchedule)
			})
		})
	})

	// 8. 啟動背景服務
	appWorker.Start()
	go func() {
		if err := appScheduler.Start(); err != nil {
			log.Fatalf("排程器啟動失敗: %v", err)
		}
	}()

	// 9. 啟動 HTTP 伺服器
	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		log.Println("伺服器啟動於 http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 伺服器監聽失敗: %s\n", err)
		}
	}()

	// 10. 處理優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("收到關閉訊號，正在進行優雅關閉...")

	// 關閉 HTTP 伺服器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("伺服器關閉失敗: %v", err)
	} else {
		log.Println("伺服器已優雅關閉")
	}

	// 停止排程器
	schedulerCtx := appScheduler.Stop()
	<-schedulerCtx.Done()

	// 停止 Worker
	appWorker.Stop()

	// 關閉佇列
	taskQueue.Close()

	log.Println("所有服務已優雅關閉")
}
