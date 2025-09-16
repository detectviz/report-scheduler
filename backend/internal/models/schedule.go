package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Recipients 定義了郵件收件人
type Recipients struct {
	To  []string `json:"to"`
	Cc  []string `json:"cc,omitempty"`
	Bcc []string `json:"bcc,omitempty"`
}

// ReportIDList 是一個字串陣列，用於存放報表 ID
type ReportIDList []string

// Schedule 對應到資料庫中的 schedules 資料表
type Schedule struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	CronSpec     string       `json:"cron_spec"`
	Timezone     string       `json:"timezone"`
	Recipients   Recipients   `json:"recipients"`
	EmailSubject string       `json:"email_subject"`
	EmailBody    string       `json:"email_body"`
	ReportIDs    ReportIDList `json:"report_ids"`
	IsEnabled    bool         `json:"is_enabled"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// --- JSON (un)marshalling for Recipients ---

// Value 實作 driver.Valuer 介面
func (r Recipients) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 實作 sql.Scanner 介面
func (r *Recipients) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		// 如果資料庫中是 NULL，就使用 Recipients 的零值
		return nil
	default:
		return errors.New("incompatible type for Recipients")
	}
	return json.Unmarshal(source, r)
}

// --- JSON (un)marshalling for ReportIDList ---

// Value 實作 driver.Valuer 介面
func (l ReportIDList) Value() (driver.Value, error) {
	if len(l) == 0 {
		return "[]", nil
	}
	return json.Marshal(l)
}

// Scan 實作 sql.Scanner 介面
func (l *ReportIDList) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []byte:
		source = v
	case nil:
		*l = make(ReportIDList, 0)
		return nil
	default:
		return errors.New("incompatible type for ReportIDList")
	}
	return json.Unmarshal(source, l)
}
