package store

import (
	"context"
	"report-scheduler/backend/internal/models"
)

// MockStore is a configurable, in-memory implementation of the Store interface for testing.
type MockStore struct {
	// Public fields allow tests to set up specific data or errors to be returned.
	SchedulesToReturn  []models.Schedule
	HistoryLogToReturn *models.HistoryLog
	ErrToReturn        error
}

// NewMockStore creates a new MockStore.
func NewMockStore() *MockStore {
	return &MockStore{}
}

// --- DataSource Methods (Placeholders) ---
func (s *MockStore) GetDataSources(ctx context.Context) ([]models.DataSource, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	return []models.DataSource{}, nil
}
func (s *MockStore) GetDataSourceByID(ctx context.Context, id string) (*models.DataSource, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	return &models.DataSource{ID: id}, nil
}
func (s *MockStore) CreateDataSource(ctx context.Context, ds *models.DataSource) error {
	return s.ErrToReturn
}
func (s *MockStore) UpdateDataSource(ctx context.Context, id string, ds *models.DataSource) error {
	return s.ErrToReturn
}
func (s *MockStore) DeleteDataSource(ctx context.Context, id string) error {
	return s.ErrToReturn
}

// --- ReportDefinition Methods (Placeholders) ---
func (s *MockStore) GetReportDefinitions(ctx context.Context) ([]models.ReportDefinition, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	return []models.ReportDefinition{}, nil
}
func (s *MockStore) GetReportDefinitionByID(ctx context.Context, id string) (*models.ReportDefinition, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	return &models.ReportDefinition{ID: id}, nil
}
func (s *MockStore) CreateReportDefinition(ctx context.Context, rd *models.ReportDefinition) error {
	return s.ErrToReturn
}
func (s *MockStore) UpdateReportDefinition(ctx context.Context, id string, rd *models.ReportDefinition) error {
	return s.ErrToReturn
}
func (s *MockStore) DeleteReportDefinition(ctx context.Context, id string) error {
	return s.ErrToReturn
}

// --- Schedule Methods ---
func (s *MockStore) GetSchedules(ctx context.Context) ([]models.Schedule, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	if s.SchedulesToReturn != nil {
		return s.SchedulesToReturn, nil
	}
	return []models.Schedule{}, nil
}
func (s *MockStore) GetScheduleByID(ctx context.Context, id string) (*models.Schedule, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	for _, schedule := range s.SchedulesToReturn {
		if schedule.ID == id {
			return &schedule, nil
		}
	}
	return nil, nil // Not found
}
func (s *MockStore) CreateSchedule(ctx context.Context, sc *models.Schedule) error {
	return s.ErrToReturn
}
func (s *MockStore) UpdateSchedule(ctx context.Context, id string, sc *models.Schedule) error {
	return s.ErrToReturn
}
func (s *MockStore) DeleteSchedule(ctx context.Context, id string) error {
	return s.ErrToReturn
}

// Close is a no-op for the mock store.
func (s *MockStore) Close() error {
	return nil
}

// --- HistoryLog Methods (Placeholders) ---
func (s *MockStore) CreateHistoryLog(ctx context.Context, log *models.HistoryLog) error {
	return s.ErrToReturn
}
func (s *MockStore) GetHistoryLogs(ctx context.Context, scheduleID string) ([]models.HistoryLog, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	return []models.HistoryLog{}, nil
}

func (s *MockStore) GetHistoryLogByID(ctx context.Context, id string) (*models.HistoryLog, error) {
	if s.ErrToReturn != nil {
		return nil, s.ErrToReturn
	}
	if s.HistoryLogToReturn != nil && s.HistoryLogToReturn.ID == id {
		return s.HistoryLogToReturn, nil
	}
	return nil, nil // Not found
}
