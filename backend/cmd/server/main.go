package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"report-scheduler/backend/internal/api"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/scheduler"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
	"report-scheduler/backend/internal/worker"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// newProcessFunc 建立一個包含其依賴項（如 store）的 Worker 處理函式。
// 透過閉包，我們可以將應用程式等級的依賴注入到 Worker 的處理邏輯中。
func newProcessFunc(s store.Store) worker.ProcessFunc {
	return func(task *queue.Task) error {
		startTime := time.Now()
		log.Printf("開始處理報表任務 %s, 來自排程 %s", task.ID, task.ScheduleID)

		// 為了記錄完整的歷史，我們需要從資料庫獲取排程的詳細資訊
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil {
			// 如果連排程都找不到，這是一個嚴重的問題，但我們還是要記錄下來
			log.Printf("錯誤：處理任務 %s 時找不到對應的排程 %s: %v", task.ID, task.ScheduleID, err)
			// 也可以在這裡建立一筆失敗的歷史紀錄
			return err
		}
		if schedule == nil {
			log.Printf("錯誤：處理任務 %s 時找不到對應的排程 %s", task.ID, task.ScheduleID)
			return nil
		}

		// TODO: 在這裡實作真正的報表產生邏輯
		time.Sleep(2 * time.Second) // 模擬耗時的工作

		duration := time.Since(startTime)

		// 建立歷史紀錄
		logEntry := &models.HistoryLog{
			ScheduleID:        task.ScheduleID,
			ScheduleName:      schedule.Name,
			TriggerTime:       task.CreatedAt,
			ExecutionDuration: duration.Milliseconds(),
			Status:            models.LogStatusSuccess,
			Recipients:        schedule.Recipients,
			// ReportURL: "...", // 報表產生後會有 URL
		}

		if err := s.CreateHistoryLog(context.Background(), logEntry); err != nil {
			log.Printf("錯誤：無法為任務 %s 建立歷史紀錄: %v", task.ID, err)
			return err
		}

		log.Printf("成功處理完報表任務 %s，並已建立歷史紀錄 %s", task.ID, logEntry.ID)
		return nil
	}
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

	// 3. 建立憑證管理器
	secretsManager := secrets.NewMockSecretsManager()

	// 4. 建立任務佇列
	taskQueue := queue.NewInMemoryQueue(100)

	// 5. 建立排程器服務
	appScheduler := scheduler.NewScheduler(dbStore, taskQueue)

	// 6. 建立 Worker 服務，並傳入真正的處理函式
	processFunc := newProcessFunc(dbStore)
	appWorker := worker.NewWorker(taskQueue, processFunc)

	// 7. 建立 API 處理器
	apiHandler := api.NewAPIHandler(dbStore, secretsManager, taskQueue)

	// 8. 建立 HTTP 路由器
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
				r.Post("/validate", apiHandler.ValidateDataSource)
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
				r.Post("/trigger", apiHandler.TriggerSchedule)
			})
		})
		r.Route("/history", func(r chi.Router) {
			r.Get("/", apiHandler.GetHistory)
		})
	})

	// 9. 啟動背景服務
	appWorker.Start()
	go func() {
		if err := appScheduler.Start(); err != nil {
			log.Fatalf("排程器啟動失敗: %v", err)
		}
	}()

	// 10. 啟動 HTTP 伺服器
	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		log.Println("伺服器啟動於 http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 伺服器監聽失敗: %s\n", err)
		}
	}()

	// 11. 處理優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("收到關閉訊號，正在進行優雅關閉...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("伺服器關閉失敗: %v", err)
	} else {
		log.Println("伺服器已優雅關閉")
	}

	schedulerCtx := appScheduler.Stop()
	<-schedulerCtx.Done()

	appWorker.Stop()

	taskQueue.Close()

	log.Println("所有服務已優雅關閉")
}
