package models

import "time"

// LogStatus 定義了歷史紀錄的狀態
type LogStatus string

const (
	LogStatusSuccess  LogStatus = "success"
	LogStatusFailed   LogStatus = "failed"
	LogStatusRetrying LogStatus = "retrying"
)

// HistoryLog 對應到資料庫中的 history_logs 資料表
type HistoryLog struct {
	ID                string     `json:"id"`
	ScheduleID        string     `json:"schedule_id"`
	ScheduleName      string     `json:"schedule_name"`
	TriggerTime       time.Time  `json:"trigger_time"`
	ExecutionDuration int64      `json:"execution_duration_ms"` // 執行耗時 (毫秒)
	Status            LogStatus  `json:"status"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	Recipients        Recipients `json:"recipients"` // 重用 Schedule 的 Recipients 結構
	ReportURL         string     `json:"report_url,omitempty"`
}
