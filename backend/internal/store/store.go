package store

import (
	"context"
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
}
