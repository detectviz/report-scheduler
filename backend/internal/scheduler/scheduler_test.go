package scheduler

import (
	"context"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScheduler_EnqueuesTasks(t *testing.T) {
	// 1. 準備依賴
	mockStore := store.NewMockStore()
	testQueue := queue.NewInMemoryQueue(10)

	// 準備一個會在一秒後觸發的排程
	mockStore.SchedulesToReturn = []models.Schedule{
		{
			ID:        "sch-1",
			Name:      "Test Schedule",
			CronSpec:  "@every 1s",
			IsEnabled: true,
			ReportIDs: []string{"rep-1"},
		},
		{
			ID:        "sch-2",
			Name:      "Disabled Schedule",
			CronSpec:  "@every 1s",
			IsEnabled: false,
		},
	}

	// 2. 建立 Scheduler 實例
	scheduler := NewScheduler(mockStore, testQueue)

	// 3. 啟動 Scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer func() {
		ctx := scheduler.Stop()
		<-ctx.Done()
		testQueue.Close()
	}()

	// 4. 驗證結果
	// 從佇列中取出任務，並設定超時
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 應該只會取出一個任務，因為 sch-2 是停用的
	task, err := testQueue.Dequeue(ctx)
	require.NoError(t, err, "預期在 2 秒內能從佇列中取出一個任務")

	// 驗證任務內容
	require.NotNil(t, task)
	require.Equal(t, "sch-1", task.ScheduleID)
	require.Len(t, task.ReportIDs, 1)
	require.Equal(t, "rep-1", task.ReportIDs[0])

	// 嘗試再取一個，應該會因為超時而失敗，證明只有一個任務被加入
	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()
	_, err = testQueue.Dequeue(ctx2)
	require.Error(t, err, "預期佇列中沒有第二個任務")
	require.Equal(t, context.DeadlineExceeded, err)
}
