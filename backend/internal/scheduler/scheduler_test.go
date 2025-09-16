package scheduler

import (
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/store"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScheduler_Start(t *testing.T) {
	// 1. 準備測試資料
	mockStore := store.NewMockStore()
	mockStore.SchedulesToReturn = []models.Schedule{
		{
			ID:        "sch-1",
			Name:      "Enabled Schedule",
			CronSpec:  "@every 1s", // A spec that fires frequently for tests
			IsEnabled: true,
		},
		{
			ID:        "sch-2",
			Name:      "Disabled Schedule",
			CronSpec:  "@every 1s",
			IsEnabled: false,
		},
		{
			ID:        "sch-3",
			Name:      "Another Enabled Schedule",
			CronSpec:  "@every 1s",
			IsEnabled: true,
		},
		{
			ID:        "sch-4",
			Name:      "Invalid Spec Schedule",
			CronSpec:  "not a valid cron spec",
			IsEnabled: true,
		},
	}

	// 2. 建立 Scheduler 實例
	scheduler := NewScheduler(mockStore)

	// 3. 啟動 Scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	defer func() {
		ctx := scheduler.Stop()
		<-ctx.Done() // 等待停止完成
	}()

	// 4. 驗證結果
	// 我們預期只有 is_enabled=true 且 CronSpec 有效的排程會被加入
	// sch-4 的規格是無效的，所以 AddFunc 會失敗，因此只會有 2 個任務成功加入
	entries := scheduler.cron.Entries()
	require.Len(t, entries, 2, "預期只有 2 個啟用的、且規格有效的排程被加入")

	// 也可以檢查下一次執行的時間是否合理
	// 給予一些緩衝時間
	require.NotNil(t, entries[0].Next, "下一次執行時間不應為 nil")
	require.WithinDuration(t, time.Now(), entries[0].Next, 5*time.Second, "預期下一次執行時間在 5 秒內")
}
