package queue

import (
	"context"
	"time"
)

// Task 代表一個需要被 Worker 執行的報表產生任務
type Task struct {
	ID         string    `json:"id"`
	ScheduleID string    `json:"schedule_id"`
	// 雖然 Schedule 中有 ReportIDs，但在 Task 中也放一份可以讓 Task 本身更獨立，
	// 未來如果支援手動觸發單一報表，這樣的設計會更有彈性。
	ReportIDs []string `json:"report_ids"`
	CreatedAt time.Time `json:"created_at"`
}

// Queue 是任務佇列的介面，定義了排程器和工作者如何與佇列互動。
// 這種設計符合 Factory Provider 模式，允許我們未來輕易地從 InMemoryQueue 切換到 RedisQueue。
type Queue interface {
	// Enqueue 將一個任務加入到佇列中
	Enqueue(ctx context.Context, task *Task) error
	// Dequeue 從佇列中取出一個任務。如果佇列是空的，這個方法應該會阻塞直到有新任務可用或 context 被取消。
	Dequeue(ctx context.Context) (*Task, error)
	// Close 優雅地關閉佇列
	Close()
}
