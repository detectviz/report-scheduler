package models

import "time"

// DataSourceType 代表資料來源的類型
type DataSourceType string

const (
	Kibana        DataSourceType = "kibana"
	Grafana       DataSourceType = "grafana"
	Elasticsearch DataSourceType = "elasticsearch"
)

// AuthType 代表認證方式
type AuthType string

const (
	BasicAuth AuthType = "basic_auth"
	APIToken  AuthType = "api_token"
)

// ConnectionStatus 代表連線狀態
type ConnectionStatus string

const (
	Verified   ConnectionStatus = "verified"
	Unverified ConnectionStatus = "unverified"
	Error      ConnectionStatus = "error"
)

// DataSource 對應到資料庫中的 datasources 資料表
// 這是系統中用來連線到外部 BI 系統的設定
type DataSource struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Type           DataSourceType   `json:"type"`
	URL            string           `json:"url"`
	APIURL         string           `json:"api_url,omitempty"` // 根據規格，此欄位應為 api_url
	AuthType       AuthType         `json:"auth_type"`
	CredentialsRef string           `json:"-"` // 這個欄位是安全的參考，不應該在 JSON 中傳輸
	Version        string           `json:"version,omitempty"` // omitempty 表示如果為空值，JSON 中就省略此欄位
	Status         ConnectionStatus `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}
