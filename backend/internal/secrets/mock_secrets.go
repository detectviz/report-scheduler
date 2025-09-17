package secrets

import "fmt"

// MockSecretsManager 是一個用於測試的 SecretsManager 介面實作
type MockSecretsManager struct {
	// 可選的，用於在測試中設定要回傳的憑證
	CredsToReturn *Credentials
	ErrToReturn   error
}

// NewMockSecretsManager 建立一個新的 MockSecretsManager
func NewMockSecretsManager() *MockSecretsManager {
	return &MockSecretsManager{}
}

// GetCredentials 實作 SecretsManager 介面
func (m *MockSecretsManager) GetCredentials(ref string) (*Credentials, error) {
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}
	if m.CredsToReturn != nil {
		return m.CredsToReturn, nil
	}
	// 如果沒有特別設定，就回傳一個預設的模擬憑證
	// 這個 ref 來自規格文件中的範例
	if ref == "kv/report-scheduler/kibana-prod" {
		return &Credentials{
			Token: "mock-api-token-12345",
		}, nil
	}
	return nil, fmt.Errorf("找不到對應的模擬憑證: %s", ref)
}
