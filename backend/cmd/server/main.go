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
		for _, reportID := range task.ReportIDs {
			reportDef, _ := s.GetReportDefinitionByID(context.Background(), reportID)
			dataSource, _ := s.GetDataSourceByID(context.Background(), reportDef.DataSourceID)
			gen, _ := genFactory.GetGenerator(dataSource.Type)
			result, err := gen.Generate(task, dataSource, reportDef)
			if err != nil {
				lastErr = err
				continue
			}
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
			logEntry.ReportURL = strings.Join(reportURLs, ", ")
		}
		return s.CreateHistoryLog(context.Background(), logEntry)
	}
}

func main() {
	cfg, _ := config.LoadConfig(".")
	dbStore, _ := store.NewStore(cfg)
	secretsManager := secrets.NewMockSecretsManager()
	taskQueue := queue.NewInMemoryQueue(100)
	genFactory := generator.NewFactory(dbStore, secretsManager)
	appScheduler := scheduler.NewScheduler(dbStore, taskQueue)
	processFunc := newProcessFunc(dbStore, genFactory)
	appWorker := worker.NewWorker(taskQueue, processFunc)
	apiHandler := api.NewAPIHandler(dbStore, secretsManager, taskQueue)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	// API Routes
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

	// Frontend static file serving
	// Note: This path is relative to the 'backend' directory,
	// where the 'go run' command is executed.
	filesDir := http.Dir("../frontend/dist")
	FileServer(r, "/", filesDir)

	// Start background services
	appWorker.Start()
	go func() {
		if err := appScheduler.Start(); err != nil {
			log.Fatalf("Scheduler failed to start: %v", err)
		}
	}()

	// Start HTTP server
	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	schedulerCtx := appScheduler.Stop()
	<-schedulerCtx.Done()

	appWorker.Stop()
	taskQueue.Close()
	dbStore.Close()

	log.Println("Server shut down gracefully")
}

// FileServer sets up a http.FileServer handler to serve static files.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}
	fs := http.StripPrefix(path, http.FileServer(root))
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	r.Get(path+"*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}
