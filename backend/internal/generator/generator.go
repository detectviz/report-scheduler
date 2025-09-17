package generator

import (
	"fmt"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"report-scheduler/backend/internal/store"
)

// GenerateResult 包含報表產生後的結果資訊
type GenerateResult struct {
	FilePath string
	MimeType string
	// 可以加入檔案大小、錯誤訊息等
}

// Generator 是報表產生器的介面，定義了所有產生器都必須實作的方法
type Generator interface {
	Generate(task *queue.Task, ds *models.DataSource, report *models.ReportDefinition) (*GenerateResult, error)
}

// Factory 用於根據資料來源類型建立對應的 Generator
type Factory struct {
	Store   store.Store
	Secrets secrets.SecretsManager
	// 未來可能需要 http.Client 等其他依賴
}

// NewFactory 建立一個新的 Generator 工廠
func NewFactory(s store.Store, sm secrets.SecretsManager) *Factory {
	return &Factory{
		Store:   s,
		Secrets: sm,
	}
}

// GetGenerator 根據資料來源類型回傳一個 Generator 實例
func (f *Factory) GetGenerator(dsType models.DataSourceType) (Generator, error) {
	switch dsType {
	case models.Kibana:
		return NewKibanaGenerator(f.Secrets), nil
	// case models.Grafana:
	// 	return NewGrafanaGenerator(f.Secrets), nil
	default:
		return nil, fmt.Errorf("不支援的資料來源類型: %s", dsType)
	}
}
