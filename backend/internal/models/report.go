package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ReportElementType 定義了報表元素的類型
type ReportElementType string

const (
	DashboardType     ReportElementType = "dashboard"
	VisualizationType ReportElementType = "visualization"
	SavedSearchType   ReportElementType = "saved_search"
)

// ReportElement 代表報表中的一個可排序項目
type ReportElement struct {
	ID    string            `json:"id"`
	Type  ReportElementType `json:"type"`
	Title string            `json:"title"` // 這個欄位可以在擷取時動態填入
}

// ReportElements 是一個 ReportElement 的切片，它實作了 sql.Scanner 和 driver.Valuer
// 這讓我們的資料庫驅動程式知道如何將 []ReportElement 存入 (轉為 JSON) 和讀出 (從 JSON 解析) 資料庫。
type ReportElements []ReportElement

// ReportDefinition 對應到資料庫中的 report_definitions 資料表
type ReportDefinition struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	DataSourceID string         `json:"datasource_id"`
	TimeRange    string         `json:"time_range"`
	Elements     ReportElements `json:"elements"` // 使用我們的自訂類型
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// Value 實作 driver.Valuer 介面，將 ReportElements 轉為可存入資料庫的 JSON 字串
func (e ReportElements) Value() (driver.Value, error) {
	if len(e) == 0 {
		return "[]", nil
	}
	return json.Marshal(e)
}

// Scan 實作 sql.Scanner 介面，從資料庫讀取資料並解析為 ReportElements
func (e *ReportElements) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		*e = make(ReportElements, 0)
		return nil
	default:
		return errors.New("incompatible type for ReportElements")
	}
	return json.Unmarshal(source, e)
}
