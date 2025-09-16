package store

import (
	"context"
	"report-scheduler/backend/internal/models"
	"time"
)

// MockStore 是一個用於測試的 Store 介面實作。
// 它回傳預先定義好的模擬資料，讓我們可以在沒有真實資料庫的情況下測試 API 層。
type MockStore struct{}

// NewMockStore 建立一個新的 MockStore 實例
func NewMockStore() *MockStore {
	return &MockStore{}
}

// GetDataSources 實作 Store 介面的 GetDataSources 方法
func (s *MockStore) GetDataSources(ctx context.Context) ([]models.DataSource, error) {
	mockData := []models.DataSource{
		{
			ID:        "mock-ds-1",
			Name:      "模擬的 Kibana 來源",
			Type:      models.Kibana,
			URL:       "http://mock-kibana.com",
			Status:    models.Verified,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	return mockData, nil
}

// GetDataSourceByID 實作 Store 介面的 GetDataSourceByID 方法
func (s *MockStore) GetDataSourceByID(ctx context.Context, id string) (*models.DataSource, error) {
	// 在模擬中，我們總是回傳一個成功的結果，並使用傳入的 ID
	return &models.DataSource{
		ID:        id,
		Name:      "模擬的特定來源 (ID: " + id + ")",
		Type:      models.Grafana,
		URL:       "http://mock-grafana.com",
		Status:    models.Verified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// CreateDataSource 實作 Store 介面的 CreateDataSource 方法
func (s *MockStore) CreateDataSource(ctx context.Context, ds *models.DataSource) error {
	// 模擬成功建立。在真實應用中，這裡會將 ds 插入資料庫。
	// 我們可以模擬資料庫行為，例如填上 ID 和時間戳。
	ds.ID = "new-mock-id"
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	return nil
}

// UpdateDataSource 實作 Store 介面的 UpdateDataSource 方法
func (s *MockStore) UpdateDataSource(ctx context.Context, id string, ds *models.DataSource) error {
	// 模擬成功更新
	return nil
}

// DeleteDataSource 實作 Store 介面的 DeleteDataSource 方法
func (s *MockStore) DeleteDataSource(ctx context.Context, id string) error {
	// 模擬成功刪除
	return nil
}
