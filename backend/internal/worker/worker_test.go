package worker

import (
	"context"
	"io/ioutil"
	"os"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWorker_CreatesHistoryLog(t *testing.T) {
	// 1. 建立一個真實的、暫時的資料庫
	tempDir, err := ioutil.TempDir("", "test-worker-db-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := tempDir + "/test.db"
	testCfg := config.Config{
		Database: config.DBConfig{Type: "sqlite", Path: dbPath},
	}
	dbStore, err := store.NewStore(testCfg)
	require.NoError(t, err)
	defer dbStore.Close()

	// 2. 準備前置資料：建立一個 schedule
	testSchedule := &models.Schedule{
		Name:      "Test Schedule For Worker",
		CronSpec:  "* * * * * *",
		IsEnabled: true,
	}
	err = dbStore.CreateSchedule(context.Background(), testSchedule)
	require.NoError(t, err)

	// 3. 建立佇列和處理函式
	testQueue := queue.NewInMemoryQueue(10)
	defer testQueue.Close()

	// 建立一個包含真實 store 的處理函式
	processFunc := newProcessFunc(dbStore)

	// 4. 建立並啟動 Worker
	worker := NewWorker(testQueue, processFunc)
	worker.Start()
	defer worker.Stop()

	// 5. 將一個任務推入佇列
	task := &queue.Task{
		ID:         "task-for-history",
		ScheduleID: testSchedule.ID,
		CreatedAt:  time.Now(),
	}
	err = testQueue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// 6. 驗證結果
	// 等待一下，讓 worker 有時間處理任務並寫入資料庫
	time.Sleep(100 * time.Millisecond) // 給予足夠的時間讓 worker 處理

	// 7. 直接從資料庫查詢，確認 HistoryLog 已建立
	logs, err := dbStore.GetHistoryLogs(context.Background(), testSchedule.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1, "預期 worker 應該已經建立了一筆歷史紀錄")
	require.Equal(t, testSchedule.Name, logs[0].ScheduleName)
	require.Equal(t, models.LogStatusSuccess, logs[0].Status)
}

// newProcessFunc 是一個輔助函式，它模仿 main.go 中的版本，但使用更短的延遲
func newProcessFunc(s store.Store) ProcessFunc {
	return func(task *queue.Task) error {
		schedule, err := s.GetScheduleByID(context.Background(), task.ScheduleID)
		if err != nil || schedule == nil {
			return err
		}
		logEntry := &models.HistoryLog{
			ScheduleID:   task.ScheduleID,
			ScheduleName: schedule.Name,
			TriggerTime:  task.CreatedAt,
			Status:       models.LogStatusSuccess,
			Recipients:   schedule.Recipients,
		}
		return s.CreateHistoryLog(context.Background(), logEntry)
	}
}
