package scheduler

import (
	"context"
	"log"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/store"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Scheduler 管理所有排程任務
type Scheduler struct {
	Store store.Store
	Queue queue.Queue
	cron  *cron.Cron
}

// NewScheduler 建立一個新的 Scheduler 實例
func NewScheduler(s store.Store, q queue.Queue) *Scheduler {
	return &Scheduler{
		Store: s,
		Queue: q,
		cron:  cron.New(cron.WithSeconds()),
	}
}

// Start 開始執行排程器
func (s *Scheduler) Start() error {
	log.Println("啟動排程器服務...")

	// 從資料庫獲取所有已啟用的排程
	schedules, err := s.Store.GetSchedules(context.Background())
	if err != nil {
		log.Printf("錯誤：啟動排程器時無法從資料庫獲取排程: %v", err)
		return err
	}

	log.Printf("找到 %d 個排程準備加入...", len(schedules))
	for _, schedule := range schedules {
		if schedule.IsEnabled {
			sch := schedule // 使用閉包捕獲 schedule 的副本
			entryID, err := s.cron.AddFunc(sch.CronSpec, func() {
				// 當 cron 任務觸發時，建立一個 Task 並將其推入佇列
				task := &queue.Task{
					ID:         uuid.New().String(),
					ScheduleID: sch.ID,
					ReportIDs:  sch.ReportIDs,
					CreatedAt:  time.Now(),
				}
				log.Printf("觸發排程: %s (ID: %s), 正在將任務 %s 推入佇列...", sch.Name, sch.ID, task.ID)
				if err := s.Queue.Enqueue(context.Background(), task); err != nil {
					log.Printf("錯誤：無法將任務 %s 推入佇列: %v", task.ID, err)
				}
			})
			if err != nil {
				log.Printf("錯誤：無法新增排程 '%s' (ID: %s): %v", sch.Name, sch.ID, err)
			} else {
				log.Printf("成功新增排程: %s (ID: %s), Cron規格: '%s', Cron Entry ID: %d", sch.Name, sch.ID, sch.CronSpec, entryID)
			}
		}
	}

	s.cron.Start()
	log.Printf("排程器服務已成功啟動，共執行 %d 個任務", len(s.cron.Entries()))
	return nil
}

// Stop 停止排程器，並等待所有執行中的任務完成
func (s *Scheduler) Stop() context.Context {
	log.Println("正在停止排程器服務...")
	ctx := s.cron.Stop()
	log.Println("排程器服務已停止")
	return ctx
}
