package worker

import (
	"context"
	"log"
	"report-scheduler/backend/internal/queue"
	"sync"
	"time"
)

// ProcessFunc is a function that processes a task.
type ProcessFunc func(task *queue.Task) error

// Worker pulls tasks from a queue and executes them.
type Worker struct {
	Queue       queue.Queue
	ProcessFunc ProcessFunc

	wg         sync.WaitGroup
	stop       chan struct{}
	cancelFunc context.CancelFunc // To cancel operations like Dequeue
}

// NewWorker creates a new Worker instance.
func NewWorker(q queue.Queue, fn ProcessFunc) *Worker {
	return &Worker{
		Queue:       q,
		ProcessFunc: fn,
		stop:        make(chan struct{}),
	}
}

// Start begins the worker's main loop in a new goroutine.
func (w *Worker) Start() {
	log.Println("啟動 Worker 服務...")

	// Create a context that can be cancelled by the Stop method.
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelFunc = cancel

	w.wg.Add(1)
	go w.run(ctx)
}

// run is the main loop for the worker.
func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()
	log.Println("Worker run 迴圈已啟動")
	for {
		// Prioritize the stop signal.
		select {
		case <-w.stop:
			log.Println("Worker 收到停止信號，正在退出 run 迴圈...")
			return
		default:
			// Non-blocking check for the stop signal.
		}

		// Now, attempt to dequeue. This will block, but it's cancellable via the context.
		task, err := w.Queue.Dequeue(ctx)
		if err != nil {
			// If the context was cancelled, it's part of a graceful shutdown.
			if err == context.Canceled || err == queue.ErrQueueClosed {
				log.Println("Worker Dequeue 被中斷或佇列已關閉，正在停止...")
				return
			}
			log.Printf("錯誤：Worker 無法從佇列中取出任務: %v", err)
			time.Sleep(1 * time.Second) // Avoid fast spinning on other errors.
			continue
		}

		// Process the task.
		log.Printf("Worker 開始處理任務: %s (來自排程 ID: %s)", task.ID, task.ScheduleID)
		if err := w.ProcessFunc(task); err != nil {
			log.Printf("錯誤：處理任務 %s 失敗: %v", task.ID, err)
		} else {
			log.Printf("Worker 完成處理任務: %s", task.ID)
		}
	}
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	log.Println("正在發送停止信號給 Worker...")

	// Signal the run loop to stop trying to dequeue more tasks.
	close(w.stop)

	// Cancel any blocking operations (like Dequeue).
	if w.cancelFunc != nil {
		w.cancelFunc()
	}

	// Wait for the run goroutine to finish.
	w.wg.Wait()
	log.Println("Worker 服務已優雅停止")
}
