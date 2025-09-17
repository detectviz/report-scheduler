package main

import (
	"context"
	"fmt"
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
func newProcessFunc(s store.Store) worker.ProcessFunc {
	return func(task *queue.Task) error {
		startTime := time.Now()
		log.Printf("任務 %s: 開始處理 (來自排程 %s)", task.ID, task.ScheduleID)

		// 1. 獲取排程詳細資訊
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil {
			// 如果連排程都找不到，這是一個嚴重的問題，但我們還是要記錄下來
			// TODO: 建立一筆失敗的歷史紀錄
			return fmt.Errorf("處理任務 %s 時找不到對應的排程 %s: %w", task.ID, task.ScheduleID, err)
		}
		if schedule == nil {
			return fmt.Errorf("處理任務 %s 時找不到對應的排程 %s", task.ID, task.ScheduleID)
		}

		// 2. 迭代處理任務中包含的每個報表 ID
		for _, reportID := range task.ReportIDs {
			log.Printf("任務 %s: 正在處理報表 ID: %s", task.ID, reportID)

			// 3. 獲取報表定義
			reportDef, err := s.GetReportDefinitionByID(context.Background(), reportID)
			if err != nil || reportDef == nil {
				log.Printf("任務 %s: 錯誤：找不到報表定義 %s，跳過此報表", task.ID, reportID)
				continue // 跳過這個報表，繼續處理下一個
			}

			// 4. 獲取資料來源定義
			dataSource, err := s.GetDataSourceByID(context.Background(), reportDef.DataSourceID)
			if err != nil || dataSource == nil {
				log.Printf("任務 %s: 錯誤：找不到報表 '%s' 的資料來源 %s，跳過此報表", task.ID, reportDef.Name, reportDef.DataSourceID)
				continue
			}

			log.Printf("任務 %s: 成功載入報表 '%s' 和資料來源 '%s'。準備進行擷取...", task.ID, reportDef.Name, dataSource.Name)
			// TODO: 在此處實作與外部服務的互動 (擷取、寄送等)
		}

		// 3. 模擬耗時的工作
		time.Sleep(1 * time.Second)
		duration := time.Since(startTime)

		// 4. 建立歷史紀錄
		logEntry := &models.HistoryLog{
			ScheduleID:        task.ScheduleID,
			ScheduleName:      schedule.Name,
			TriggerTime:       task.CreatedAt,
			ExecutionDuration: duration.Milliseconds(),
			Status:            models.LogStatusSuccess,
			Recipients:        schedule.Recipients,
		}

		if err := s.CreateHistoryLog(context.Background(), logEntry); err != nil {
			// 即使歷史紀錄建立失敗，也不應該讓整個任務失敗，僅記錄錯誤
			log.Printf("錯誤：無法為任務 %s 建立歷史紀錄: %v", task.ID, err)
		}

		log.Printf("任務 %s: 成功處理完畢，並已建立歷史紀錄 %s", task.ID, logEntry.ID)
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

	// 6. 建立 Worker 服務
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
