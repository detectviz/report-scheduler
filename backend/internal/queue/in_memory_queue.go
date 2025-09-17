package queue

import (
	"context"
	"errors"
)

var ErrQueueClosed = errors.New("queue is closed")

// InMemoryQueue 是使用緩衝 channel 實作的記憶體佇列
type InMemoryQueue struct {
	tasks chan *Task
	done  chan struct{} // 用於發送關閉信號
}

// NewInMemoryQueue 建立一個新的記憶體佇列
func NewInMemoryQueue(size int) *InMemoryQueue {
	return &InMemoryQueue{
		tasks: make(chan *Task, size),
		done:  make(chan struct{}),
	}
}

// Enqueue 將任務加入佇列。如果佇列已關閉，則回傳錯誤。
func (q *InMemoryQueue) Enqueue(ctx context.Context, task *Task) error {
	select {
	case q.tasks <- task:
		return nil
	case <-q.done:
		return ErrQueueClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Dequeue 從佇列中取出任務。如果佇列已關閉且為空，則回傳錯誤。
func (q *InMemoryQueue) Dequeue(ctx context.Context) (*Task, error) {
	select {
	case task := <-q.tasks:
		return task, nil
	case <-q.done:
		// 關閉後，再嘗試清空剩餘的任務
		select {
		case task := <-q.tasks:
			return task, nil
		default:
			return nil, ErrQueueClosed
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close 關閉佇列，不再接受新的任務。
func (q *InMemoryQueue) Close() {
	// 使用 select 避免重複關閉 channel 導致 panic
	select {
	case <-q.done:
		// 已經關閉
		return
	default:
		// 安全地關閉
		close(q.done)
	}
}
