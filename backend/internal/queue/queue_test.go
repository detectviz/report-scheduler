package queue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInMemoryQueue(t *testing.T) {
	t.Run("enqueue and dequeue a single task", func(t *testing.T) {
		q := NewInMemoryQueue(10)
		defer q.Close()

		taskIn := &Task{ID: "task-1"}
		err := q.Enqueue(context.Background(), taskIn)
		require.NoError(t, err)

		taskOut, err := q.Dequeue(context.Background())
		require.NoError(t, err)
		require.NotNil(t, taskOut)
		require.Equal(t, "task-1", taskOut.ID)
	})

	t.Run("dequeue blocks until a task is available", func(t *testing.T) {
		q := NewInMemoryQueue(10)
		defer q.Close()

		go func() {
			time.Sleep(20 * time.Millisecond)
			err := q.Enqueue(context.Background(), &Task{ID: "task-2"})
			require.NoError(t, err)
		}()

		// Dequeue should block for at least a moment
		startTime := time.Now()
		taskOut, err := q.Dequeue(context.Background())
		duration := time.Since(startTime)

		require.NoError(t, err)
		require.NotNil(t, taskOut)
		require.Equal(t, "task-2", taskOut.ID)
		require.GreaterOrEqual(t, duration, 15*time.Millisecond, "Dequeue should have blocked")
	})

	t.Run("close unblocks dequeue and returns error for empty queue", func(t *testing.T) {
		q := NewInMemoryQueue(10)

		go func() {
			time.Sleep(20 * time.Millisecond)
			q.Close()
		}()

		taskOut, err := q.Dequeue(context.Background())
		require.Error(t, err)
		require.Equal(t, ErrQueueClosed, err)
		require.Nil(t, taskOut)
	})

	t.Run("enqueue on a closed queue returns error", func(t *testing.T) {
		q := NewInMemoryQueue(10)
		q.Close() // Close it immediately

		err := q.Enqueue(context.Background(), &Task{ID: "task-3"})
		require.Error(t, err)
		require.Equal(t, ErrQueueClosed, err)
	})

	t.Run("close is idempotent", func(t *testing.T) {
		q := NewInMemoryQueue(10)
		q.Close()
		require.NotPanics(t, func() { q.Close() }, "closing an already closed queue should not panic")
	})
}
