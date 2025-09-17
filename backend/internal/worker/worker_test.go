package worker

import (
	"context"
	"report-scheduler/backend/internal/queue"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWorker_ProcessesTaskFromQueue(t *testing.T) {
	// 1. 準備依賴
	testQueue := queue.NewInMemoryQueue(10)
	defer testQueue.Close()

	processedTaskID := ""
	var mu sync.Mutex // Mutex to protect processedTaskID from concurrent access

	// 建立一個假的處理函式，它會記錄被處理的任務 ID
	testProcessFunc := func(task *queue.Task) error {
		mu.Lock()
		processedTaskID = task.ID
		mu.Unlock()
		return nil
	}

	// 2. 建立並啟動 Worker
	worker := NewWorker(testQueue, testProcessFunc)
	worker.Start()
	defer worker.Stop()

	// 3. 將一個任務推入佇列
	task := &queue.Task{ID: "task-to-process"}
	err := testQueue.Enqueue(context.Background(), task)
	require.NoError(t, err)

	// 4. 驗證結果
	// 使用 require.Eventually 來取代固定的 time.Sleep，這讓測試更可靠
	// 它會在一段時間內（這裡設定為 1 秒）反覆檢查條件是否為真
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return processedTaskID == "task-to-process"
	}, 1*time.Second, 10*time.Millisecond, "預期 Worker 應在 1 秒內處理完任務")
}
