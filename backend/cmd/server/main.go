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
	"report-scheduler/backend/internal/generator"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/scheduler"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
	"report-scheduler/backend/internal/worker"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// newProcessFunc 建立一個包含其依賴項的 Worker 處理函式。
func newProcessFunc(s store.Store, genFactory *generator.Factory) worker.ProcessFunc {
	return func(task *queue.Task) error {
		startTime := time.Now()
		log.Printf("任務 %s: 開始處理 (來自排程 %s)", task.ID, task.ScheduleID)

		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return fmt.Errorf("處理任務 %s 時找不到對應的排程 %s", task.ID, task.ScheduleID)
		}

		var lastErr error
		var reportURLs []string
		// 迭代處理任務中包含的每個報表 ID
		for _, reportID := range task.ReportIDs {
			reportDef, err := s.GetReportDefinitionByID(context.Background(), reportID)
			if err != nil || reportDef == nil {
				log.Printf("任務 %s: 錯誤：找不到報表定義 %s，跳過", task.ID, reportID)
				lastErr = err
				continue
			}

			dataSource, err := s.GetDataSourceByID(context.Background(), reportDef.DataSourceID)
			if err != nil || dataSource == nil {
				log.Printf("任務 %s: 錯誤：找不到報表 '%s' 的資料來源 %s，跳過", task.ID, reportDef.Name, reportDef.DataSourceID)
				lastErr = err
				continue
			}

			gen, err := genFactory.GetGenerator(dataSource.Type)
			if err != nil {
				log.Printf("任務 %s: 錯誤：找不到報表 '%s' 的產生器，跳過", task.ID, reportDef.Name)
				lastErr = err
				continue
			}

			result, err := gen.Generate(task, dataSource, reportDef)
			if err != nil {
				log.Printf("任務 %s: 錯誤：產生報表 '%s' 失敗: %v", task.ID, reportDef.Name, err)
				lastErr = err
				continue
			}

			log.Printf("任務 %s: 報表 '%s' 產生成功，檔案位於 %s", task.ID, reportDef.Name, result.FilePath)
			reportURLs = append(reportURLs, result.FilePath)
		}

		duration := time.Since(startTime)
		logEntry := &models.HistoryLog{
			ScheduleID:        task.ScheduleID,
			ScheduleName:      schedule.Name,
			TriggerTime:       task.CreatedAt,
			ExecutionDuration: duration.Milliseconds(),
			Recipients:        schedule.Recipients,
		}

		if lastErr != nil {
			logEntry.Status = models.LogStatusFailed
			logEntry.ErrorMessage = lastErr.Error()
		} else {
			logEntry.Status = models.LogStatusSuccess
			// For now, we'll just join the paths. A better solution might be a JSON array.
			logEntry.ReportURL = strings.Join(reportURLs, ", ")
		}

		if err := s.CreateHistoryLog(context.Background(), logEntry); err != nil {
			log.Printf("錯誤：無法為任務 %s 建立歷史紀錄: %v", task.ID, err)
		}

		log.Printf("任務 %s: 處理完畢，並已建立歷史紀錄 %s", task.ID, logEntry.ID)
		return nil
	}
}

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("無法載入設定: %v", err)
	}

	dbStore, err := store.NewStore(cfg)
	if err != nil {
		log.Fatalf("無法連線到資料庫: %v", err)
	}

	secretsManager := secrets.NewMockSecretsManager()
	taskQueue := queue.NewInMemoryQueue(100)
	genFactory := generator.NewFactory(dbStore, secretsManager)

	appScheduler := scheduler.NewScheduler(dbStore, taskQueue)
	processFunc := newProcessFunc(dbStore, genFactory)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	apiHandler := api.NewAPIHandler(dbStore, secretsManager, taskQueue)

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

	appWorker.Start()
	go func() {
		if err := appScheduler.Start(); err != nil {
			log.Fatalf("排程器啟動失敗: %v", err)
		}
	}()

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		log.Println("伺服器啟動於 http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 伺服器監聽失敗: %s\n", err)
		}
	}()

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
