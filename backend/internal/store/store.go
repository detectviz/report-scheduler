package store

import (
	"context"
	"fmt"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/models"
)

// Store 是我們資料存取層的介面。
// 這種設計允許我們輕鬆地替換底層的資料庫實作（例如，從模擬的 store 換成真實的 PostgreSQL store），
// 而不需要修改 API 處理器的程式碼，這正符合了 Factory Provider 的模式。
type Store interface {
	// GetDataSources 返回所有資料來源
	GetDataSources(ctx context.Context) ([]models.DataSource, error)
	// GetDataSourceByID 根據 ID 返回單一資料來源
	GetDataSourceByID(ctx context.Context, id string) (*models.DataSource, error)
	// CreateDataSource 建立一個新的資料來源。在成功時，傳入的 DataSource 物件應被更新 (例如，填上 ID 和時間戳)。
	CreateDataSource(ctx context.Context, ds *models.DataSource) error
	// UpdateDataSource 更新一個已有的資料來源
	UpdateDataSource(ctx context.Context, id string, ds *models.DataSource) error
	// DeleteDataSource 根據 ID 刪除一個資料來源
	DeleteDataSource(ctx context.Context, id string) error

	// --- ReportDefinition Methods ---
	GetReportDefinitions(ctx context.Context) ([]models.ReportDefinition, error)
	GetReportDefinitionByID(ctx context.Context, id string) (*models.ReportDefinition, error)
	CreateReportDefinition(ctx context.Context, rd *models.ReportDefinition) error
	UpdateReportDefinition(ctx context.Context, id string, rd *models.ReportDefinition) error
	DeleteReportDefinition(ctx context.Context, id string) error

	// --- Schedule Methods ---
	GetSchedules(ctx context.Context) ([]models.Schedule, error)
	GetScheduleByID(ctx context.Context, id string) (*models.Schedule, error)
	CreateSchedule(ctx context.Context, s *models.Schedule) error
	UpdateSchedule(ctx context.Context, id string, s *models.Schedule) error
	DeleteSchedule(ctx context.Context, id string) error

	// --- HistoryLog Methods ---
	CreateHistoryLog(ctx context.Context, log *models.HistoryLog) error
	GetHistoryLogs(ctx context.Context, scheduleID string) ([]models.HistoryLog, error)
	GetHistoryLogByID(ctx context.Context, id string) (*models.HistoryLog, error)

	// Close 關閉與資料庫的連線
	Close() error
}

// NewStore 是資料儲存層的工廠函式。
// 它會根據設定檔中的 database.type 來決定要回傳哪一種 Store 實作。
func NewStore(cfg config.Config) (Store, error) {
	switch cfg.Database.Type {
	case "sqlite":
		return newSqliteStore(cfg)
	// 未來可以在這裡新增 "postgres" 的 case
	default:
		return nil, fmt.Errorf("不支援的資料庫類型: %s", cfg.Database.Type)
	}
}
