package models

import (
	"encoding/json"
	"time"
)

// DataSourceType 代表資料來源的類型
type DataSourceType string

const (
	Kibana  DataSourceType = "kibana"
	Grafana DataSourceType = "grafana"
	// Elasticsearch is no longer a separate type, it's part of Kibana
)

// AuthType 代表認證方式
type AuthType string

const (
	BasicAuth AuthType = "basic_auth"
	APIToken  AuthType = "api_token"
	AuthNone  AuthType = "none"
)

// ConnectionStatus 代表連線狀態
type ConnectionStatus string

const (
	Verified   ConnectionStatus = "verified"
	Unverified ConnectionStatus = "unverified"
	Error      ConnectionStatus = "error"
)

// DataSource 對應到資料庫中的 datasources 資料表
type DataSource struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Type           DataSourceType   `json:"type"`
	URL            string           `json:"url"`
	APIURL         string           `json:"api_url,omitempty"`
	AuthType       AuthType         `json:"auth_type"`
	CredentialsRef string           `json:"credentials_ref,omitempty"` // Allow reading from JSON, but will be hidden on write
	Version        string           `json:"version,omitempty"`
	Status         ConnectionStatus `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// MarshalJSON is a custom marshaller for DataSource to hide sensitive fields.
func (ds DataSource) MarshalJSON() ([]byte, error) {
	// Use a type alias to avoid an infinite loop
	type Alias DataSource

	// Create a new struct that omits the sensitive field
	return json.Marshal(&struct {
		Alias
		CredentialsRef string `json:"credentials_ref,omitempty"` // This will be omitted because it's the zero value
	}{
		Alias: (Alias)(ds),
	})
}

// AvailableElement represents an element that can be chosen from a data source.
// It's defined here because it's an attribute of a data source.
type AvailableElement struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // Using string to avoid circular dependency on report.go's ReportElementType
	Title string `json:"title"`
}
