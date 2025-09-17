package generator

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	creds, err := g.Secrets.GetCredentials(ds.CredentialsRef)
	if err != nil {
		return nil, fmt.Errorf("無法獲取 Kibana 憑證 for ref %s: %w", ds.CredentialsRef, err)
	}

	// 2. 組合 URL (簡化邏輯)
	// 假設報表只有一個元素
	if len(report.Elements) == 0 {
		return nil, fmt.Errorf("報表 '%s' 中沒有任何元素", report.Name)
	}
	elementID := report.Elements[0].ID
	// 真正的實作需要處理 space, RISON 編碼、時間範圍等
	generationURL := fmt.Sprintf("%s/api/reporting/generate/dashboard/%s", ds.URL, elementID)
	log.Printf("[Generator] Kibana: 準備請求 URL: %s", generationURL)

	// 3. 建立並執行 HTTP 請求
	req, err := http.NewRequest("POST", generationURL, nil)
	if err != nil {
		return nil, fmt.Errorf("無法建立請求: %w", err)
	}

	// 根據認證類型設定標頭
	switch ds.AuthType {
	case models.APIToken:
		req.Header.Set("Authorization", "ApiKey "+creds.Token)
	case models.BasicAuth:
		req.SetBasicAuth(creds.Username, creds.Password)
	}
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[Generator] Kibana: Request Headers: %v", req.Header) // DEBUG LOGGING

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("請求 Kibana API 失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Kibana API 回應非 200 狀態: %d, body: %s", resp.StatusCode, string(body))
	}

	// 4. 將回應內容儲存到暫存檔案
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取回應內容失敗: %w", err)
	}

	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("report-%s-*.pdf", task.ID))
	if err != nil {
		return nil, fmt.Errorf("建立暫存檔案失敗: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(body); err != nil {
		return nil, fmt.Errorf("寫入暫存檔案失敗: %w", err)
	}

	log.Printf("[Generator] Kibana: 成功產生報告，檔案儲存於 %s", tmpFile.Name())

	return &GenerateResult{
		FilePath: tmpFile.Name(),
		MimeType: "application/pdf",
	}, nil
}
