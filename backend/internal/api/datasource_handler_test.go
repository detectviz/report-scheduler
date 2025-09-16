package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"report-scheduler/backend/internal/models"
	"testing"
)

// TestGetDataSources 驗證 GetDataSources 處理器是否能正常運作
func TestGetDataSources(t *testing.T) {
	// 建立一個路由器，並註冊我們的處理器函式
	// 這裡我們直接測試 GetDataSources 函式
	handler := http.HandlerFunc(GetDataSources)

	// 使用 httptest 建立一個測試伺服器
	server := httptest.NewServer(handler)
	defer server.Close()

	// 對測試伺服器發送 GET 請求
	// 注意：因為我們直接測試處理器，所以路徑是什麼都沒關係
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("無法發送 GET 請求: %v", err)
	}
	defer resp.Body.Close()

	// 1. 檢查 HTTP 狀態碼
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("預期狀態碼為 %v, 但得到 %v", http.StatusOK, status)
	}

	// 2. 讀取並解析回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("無法讀取回應內容: %v", err)
	}

	var dataSources []models.DataSource
	if err := json.Unmarshal(body, &dataSources); err != nil {
		t.Fatalf("無法將 JSON 解碼為 DataSource: %v", err)
	}

	// 3. 驗證回應內容是否符合預期
	if len(dataSources) != 1 {
		t.Errorf("預期有 1 個資料來源, 但得到 %d 個", len(dataSources))
		return // 如果長度不對，後續的檢查就沒有意義了
	}

	expectedName := "公司正式環境 Kibana"
	if dataSources[0].Name != expectedName {
		t.Errorf("預期的資料來源名稱為 '%s', 但得到 '%s'", expectedName, dataSources[0].Name)
	}

	expectedType := models.Kibana
	if dataSources[0].Type != expectedType {
		t.Errorf("預期的資料來源類型為 '%s', 但得到 '%s'", expectedType, dataSources[0].Type)
	}
}
