package generator

import (
	"fmt"
	"log"
	"report-scheduler/backend/internal/models"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"time"
)

// KibanaGenerator 負責從 Kibana 產生報表
type KibanaGenerator struct {
	Secrets secrets.SecretsManager
}

// NewKibanaGenerator 建立一個新的 KibanaGenerator
func NewKibanaGenerator(sm secrets.SecretsManager) *KibanaGenerator {
	return &KibanaGenerator{
		Secrets: sm,
	}
}

// Generate 實作報表產生邏輯
func (g *KibanaGenerator) Generate(task *queue.Task, ds *models.DataSource, report *models.ReportDefinition) (*GenerateResult, error) {
	log.Printf("[Generator] Kibana: 正在為報表 '%s' 產生報告...", report.Name)

	// 1. 獲取憑證
	// TODO: ds.CredentialsRef 應該被傳入，而不是寫死的
	creds, err := g.Secrets.GetCredentials("kv/report-scheduler/kibana-prod")
	if err != nil {
		return nil, fmt.Errorf("無法獲取 Kibana 憑證: %w", err)
	}
	log.Printf("[Generator] Kibana: 成功獲取憑證 (Token: ...%s)", creds.Token[len(creds.Token)-5:])

	// 2. 組合 URL (此處為簡化邏輯)
	// 真正的實作需要處理 space, RISON 編碼、時間範圍等
	reportURL := fmt.Sprintf("%s/api/reporting/generate/...", ds.URL) // Simplified
	log.Printf("[Generator] Kibana: 產生報告 URL: %s", reportURL)

	// 3. 模擬呼叫 Kibana Reporting API 並儲存檔案
	log.Println("[Generator] Kibana: 正在呼叫 Kibana API... (模擬)")
	time.Sleep(2 * time.Second) // 模擬網路延遲和產生時間
	mockFilePath := fmt.Sprintf("/tmp/report-%s.pdf", task.ID)
	log.Printf("[Generator] Kibana: 成功產生報告，檔案儲存於: %s", mockFilePath)

	return &GenerateResult{
		FilePath: mockFilePath,
		MimeType: "application/pdf",
	}, nil
}
