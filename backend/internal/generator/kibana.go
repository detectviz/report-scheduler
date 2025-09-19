package generator

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"report-scheduler/backend/internal/models"
	"net/url"
	"report-scheduler/backend/internal/queue"
	"report-scheduler/backend/internal/secrets"
	"strconv"
	"time"

	"github.com/sakura-internet/go-rison/v4"
)

// parseTimeRange 解析相對時間字串 (例如 "now-7d") 並回傳 from 和 to 的 ISO 8601 時間
func parseTimeRange(timeRange string) (from, to string, err error) {
	now := time.Now()
	to = now.Format(time.RFC3339) // 'to' is always now for relative times

	// Regex to capture the number and unit (d, h, m)
	re := regexp.MustCompile(`^now-(\d+)([dhm])$`)
	matches := re.FindStringSubmatch(timeRange)

	if len(matches) != 3 {
		// Handle fixed time ranges like "now/d" or other formats if needed
		// For now, we only support "now-Xd", "now-Xh", "now-Xm"
		return "", "", fmt.Errorf("不支援的時間範圍格式: %s", timeRange)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	var duration time.Duration
	switch unit {
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	case "h":
		duration = time.Duration(value) * time.Hour
	case "m":
		duration = time.Duration(value) * time.Minute
	default:
		return "", "", fmt.Errorf("不支援的時間單位: %s", unit)
	}

	fromTime := now.Add(-duration)
	from = fromTime.Format(time.RFC3339)

	return from, to, nil
}


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

	// 1. 組合 URL (簡化邏輯)
	// 假設報表只有一個元素
	if len(report.Elements) == 0 {
		return nil, fmt.Errorf("報表 '%s' 中沒有任何元素", report.Name)
	}
	elementID := report.Elements[0].ID
	// 真正的實作需要處理 space, RISON 編碼、時間範圍等
	var spacePrefix string
	if report.Space != "" && report.Space != "default" {
		spacePrefix = fmt.Sprintf("/s/%s", report.Space)
	}
	generationURL := fmt.Sprintf("%s%s/api/reporting/generate/dashboard/%s", ds.URL, spacePrefix, elementID)

	// 處理時間範圍
	if report.TimeRange != "" {
		from, to, err := parseTimeRange(report.TimeRange)
		if err != nil {
			log.Printf("[Generator] Kibana: 無法解析時間範圍 '%s': %v", report.TimeRange, err)
			// 忽略時間範圍錯誤，繼續執行
		} else {
			// 建構 RISON 物件
			gParam := map[string]interface{}{
				"time": map[string]string{
					"from": from,
					"to":   to,
				},
			}
			risonBytes, err := rison.Marshal(gParam, rison.Rison)
			if err != nil {
				log.Printf("[Generator] Kibana: RISON 編碼失敗: %v", err)
			} else {
				// 將 RISON 字串附加到 URL
				generationURL = fmt.Sprintf("%s?_g=%s", generationURL, url.QueryEscape(string(risonBytes)))
			}
		}
	}

	log.Printf("[Generator] Kibana: 準備請求 URL: %s", generationURL)

	// 2. 建立並執行 HTTP 請求
	req, err := http.NewRequest("POST", generationURL, nil)
	if err != nil {
		return nil, fmt.Errorf("無法建立請求: %w", err)
	}

	// 3. 只有在需要認證時才獲取憑證並設定標頭
	if ds.AuthType != models.AuthNone {
		creds, err := g.Secrets.GetCredentials(ds.CredentialsRef)
		if err != nil {
			return nil, fmt.Errorf("無法獲取 Kibana 憑證 for ref %s: %w", ds.CredentialsRef, err)
		}

		switch ds.AuthType {
		case models.APIToken:
			req.Header.Set("Authorization", "ApiKey "+creds.Token)
		case models.BasicAuth:
			req.SetBasicAuth(creds.Username, creds.Password)
		}
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
