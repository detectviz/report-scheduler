package store

import (
	"context"
	"database/sql"
	"report-scheduler/backend/internal/config"
	"report-scheduler/backend/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Import the sqlite3 driver
)

// SqliteStore 提供了使用 SQLite 資料庫的 Store 介面實作
type SqliteStore struct {
	db *sql.DB
}

// newSqliteStore 建立一個新的 SqliteStore 實例，並初始化資料庫
func newSqliteStore(cfg config.Config) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &SqliteStore{db: db}
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	// 每次啟動時都植入種子資料，以確保 Demo 環境一致
	if err := store.seedData(); err != nil {
		return nil, err
	}

	return store, nil
}

// seedData 植入用於開發和測試的初始資料
func (s *SqliteStore) seedData() error {
	// 檢查資料是否已存在，避免重複植入
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM datasources WHERE id = 'ds-4'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		// 資料已存在，直接返回
		return nil
	}

	// 植入資料來源 'ds-4'
	ds4 := &models.DataSource{
		ID:        "ds-4",
		Name:      "公開 Kibana 體驗環境",
		Type:      models.Kibana,
		URL:       "https://demo.elastic.co",
		AuthType:  models.AuthNone,
		Status:    models.Verified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	queryDs := `INSERT INTO datasources (id, name, type, url, api_url, auth_type, credentials_ref, version, status, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(queryDs, ds4.ID, ds4.Name, ds4.Type, ds4.URL, ds4.APIURL, ds4.AuthType, ds4.CredentialsRef, ds4.Version, ds4.Status, ds4.CreatedAt, ds4.UpdatedAt)
	if err != nil {
		return err
	}

	// 植入報表定義 'report-1'
	elements := models.ReportElements{
		{
			ID:    "elastic_agent-0600ffa0-6b5e-11ed-98de-67bdecd21824",
			Type:  models.DashboardType,
			Title: "Elastic Agent dashboard",
			Order: 1,
		},
	}

	report1 := &models.ReportDefinition{
		ID:           "report-1",
		Name:         "Elastic Agent 狀態儀表板",
		DataSourceID: "ds-4",
		TimeRange:    "now-7d",
		Elements:     elements,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	queryReport := `INSERT INTO report_definitions (id, name, description, datasource_id, time_range, elements, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(queryReport, report1.ID, report1.Name, report1.Description, report1.DataSourceID, report1.TimeRange, report1.Elements, report1.CreatedAt, report1.UpdatedAt)
	return err
}


// initSchema 建立資料庫資料表 (如果不存在)
func (s *SqliteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS datasources (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		url TEXT NOT NULL,
		api_url TEXT,
		auth_type TEXT NOT NULL,
		credentials_ref TEXT,
		version TEXT,
		status TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	schema = `
	CREATE TABLE IF NOT EXISTS report_definitions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		datasource_id TEXT NOT NULL,
		time_range TEXT NOT NULL,
		elements TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		FOREIGN KEY(datasource_id) REFERENCES datasources(id)
	);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	schema = `
	CREATE TABLE IF NOT EXISTS history_logs (
		id TEXT PRIMARY KEY,
		schedule_id TEXT NOT NULL,
		schedule_name TEXT NOT NULL,
		trigger_time TIMESTAMP NOT NULL,
		execution_duration_ms INTEGER NOT NULL,
		status TEXT NOT NULL,
		error_message TEXT,
		recipients TEXT,
		report_url TEXT
	);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	schema = `
	CREATE TABLE IF NOT EXISTS schedules (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		cron_spec TEXT NOT NULL,
		timezone TEXT NOT NULL,
		recipients TEXT,
		email_subject TEXT,
		email_body TEXT,
		report_ids TEXT,
		is_enabled BOOLEAN NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// --- Store Interface Implementation ---

func (s *SqliteStore) CreateDataSource(ctx context.Context, ds *models.DataSource) error {
	// 使用 UUID 作為唯一識別碼
	ds.ID = uuid.New().String()
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()

	query := `INSERT INTO datasources (id, name, type, url, api_url, auth_type, credentials_ref, version, status, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, ds.ID, ds.Name, ds.Type, ds.URL, ds.APIURL, ds.AuthType, ds.CredentialsRef, ds.Version, ds.Status, ds.CreatedAt, ds.UpdatedAt)
	return err
}

func (s *SqliteStore) GetDataSources(ctx context.Context) ([]models.DataSource, error) {
	query := `SELECT id, name, type, url, api_url, auth_type, credentials_ref, version, status, created_at, updated_at FROM datasources`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataSources []models.DataSource
	for rows.Next() {
		var ds models.DataSource
		if err := rows.Scan(&ds.ID, &ds.Name, &ds.Type, &ds.URL, &ds.APIURL, &ds.AuthType, &ds.CredentialsRef, &ds.Version, &ds.Status, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, err
		}
		dataSources = append(dataSources, ds)
	}
	return dataSources, nil
}

func (s *SqliteStore) GetDataSourceByID(ctx context.Context, id string) (*models.DataSource, error) {
	query := `SELECT id, name, type, url, api_url, auth_type, credentials_ref, version, status, created_at, updated_at FROM datasources WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var ds models.DataSource
	err := row.Scan(&ds.ID, &ds.Name, &ds.Type, &ds.URL, &ds.APIURL, &ds.AuthType, &ds.CredentialsRef, &ds.Version, &ds.Status, &ds.CreatedAt, &ds.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 找不到時回傳 nil, nil，讓 handler 處理 404
		}
		return nil, err
	}
	return &ds, nil
}

func (s *SqliteStore) UpdateDataSource(ctx context.Context, id string, ds *models.DataSource) error {
    ds.UpdatedAt = time.Now()
	query := `UPDATE datasources SET name = ?, type = ?, url = ?, api_url = ?, auth_type = ?, credentials_ref = ?, version = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, ds.Name, ds.Type, ds.URL, ds.APIURL, ds.AuthType, ds.CredentialsRef, ds.Version, ds.Status, ds.UpdatedAt, id)
	return err
}

func (s *SqliteStore) DeleteDataSource(ctx context.Context, id string) error {
	query := `DELETE FROM datasources WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// --- HistoryLog Methods ---

func (s *SqliteStore) CreateHistoryLog(ctx context.Context, log *models.HistoryLog) error {
	log.ID = uuid.New().String()
	query := `INSERT INTO history_logs (id, schedule_id, schedule_name, trigger_time, execution_duration_ms, status, error_message, recipients, report_url)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, log.ID, log.ScheduleID, log.ScheduleName, log.TriggerTime, log.ExecutionDuration, log.Status, log.ErrorMessage, log.Recipients, log.ReportURL)
	return err
}

func (s *SqliteStore) GetHistoryLogByID(ctx context.Context, id string) (*models.HistoryLog, error) {
	query := `SELECT id, schedule_id, schedule_name, trigger_time, execution_duration_ms, status, error_message, recipients, report_url FROM history_logs WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var log models.HistoryLog
	err := row.Scan(&log.ID, &log.ScheduleID, &log.ScheduleName, &log.TriggerTime, &log.ExecutionDuration, &log.Status, &log.ErrorMessage, &log.Recipients, &log.ReportURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 找不到時回傳 nil, nil，讓 handler 處理 404
		}
		return nil, err
	}
	return &log, nil
}

func (s *SqliteStore) GetHistoryLogs(ctx context.Context, scheduleID string) ([]models.HistoryLog, error) {
	query := `SELECT id, schedule_id, schedule_name, trigger_time, execution_duration_ms, status, error_message, recipients, report_url FROM history_logs WHERE schedule_id = ? ORDER BY trigger_time DESC`
	rows, err := s.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.HistoryLog
	for rows.Next() {
		var log models.HistoryLog
		if err := rows.Scan(&log.ID, &log.ScheduleID, &log.ScheduleName, &log.TriggerTime, &log.ExecutionDuration, &log.Status, &log.ErrorMessage, &log.Recipients, &log.ReportURL); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// Close 關閉資料庫連線
func (s *SqliteStore) Close() error {
	return s.db.Close()
}

// --- Schedule Methods ---

func (s *SqliteStore) CreateSchedule(ctx context.Context, sc *models.Schedule) error {
	sc.ID = uuid.New().String()
	sc.CreatedAt = time.Now()
	sc.UpdatedAt = time.Now()

	query := `INSERT INTO schedules (id, name, cron_spec, timezone, recipients, email_subject, email_body, report_ids, is_enabled, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, sc.ID, sc.Name, sc.CronSpec, sc.Timezone, sc.Recipients, sc.EmailSubject, sc.EmailBody, sc.ReportIDs, sc.IsEnabled, sc.CreatedAt, sc.UpdatedAt)
	return err
}

func (s *SqliteStore) GetSchedules(ctx context.Context) ([]models.Schedule, error) {
	query := `SELECT id, name, cron_spec, timezone, recipients, email_subject, email_body, report_ids, is_enabled, created_at, updated_at FROM schedules`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.Schedule
	for rows.Next() {
		var sc models.Schedule
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.CronSpec, &sc.Timezone, &sc.Recipients, &sc.EmailSubject, &sc.EmailBody, &sc.ReportIDs, &sc.IsEnabled, &sc.CreatedAt, &sc.UpdatedAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, sc)
	}
	return schedules, nil
}

func (s *SqliteStore) GetScheduleByID(ctx context.Context, id string) (*models.Schedule, error) {
	query := `SELECT id, name, cron_spec, timezone, recipients, email_subject, email_body, report_ids, is_enabled, created_at, updated_at FROM schedules WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var sc models.Schedule
	err := row.Scan(&sc.ID, &sc.Name, &sc.CronSpec, &sc.Timezone, &sc.Recipients, &sc.EmailSubject, &sc.EmailBody, &sc.ReportIDs, &sc.IsEnabled, &sc.CreatedAt, &sc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sc, nil
}

func (s *SqliteStore) UpdateSchedule(ctx context.Context, id string, sc *models.Schedule) error {
	sc.UpdatedAt = time.Now()
	query := `UPDATE schedules SET name = ?, cron_spec = ?, timezone = ?, recipients = ?, email_subject = ?, email_body = ?, report_ids = ?, is_enabled = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, sc.Name, sc.CronSpec, sc.Timezone, sc.Recipients, sc.EmailSubject, sc.EmailBody, sc.ReportIDs, sc.IsEnabled, sc.UpdatedAt, id)
	return err
}

func (s *SqliteStore) DeleteSchedule(ctx context.Context, id string) error {
	query := `DELETE FROM schedules WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// --- ReportDefinition Methods ---

func (s *SqliteStore) CreateReportDefinition(ctx context.Context, rd *models.ReportDefinition) error {
	rd.ID = uuid.New().String()
	rd.CreatedAt = time.Now()
	rd.UpdatedAt = time.Now()

	query := `INSERT INTO report_definitions (id, name, description, datasource_id, time_range, elements, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, rd.ID, rd.Name, rd.Description, rd.DataSourceID, rd.TimeRange, rd.Elements, rd.CreatedAt, rd.UpdatedAt)
	return err
}

func (s *SqliteStore) GetReportDefinitions(ctx context.Context) ([]models.ReportDefinition, error) {
	query := `SELECT id, name, description, datasource_id, time_range, elements, created_at, updated_at FROM report_definitions`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.ReportDefinition
	for rows.Next() {
		var rd models.ReportDefinition
		if err := rows.Scan(&rd.ID, &rd.Name, &rd.Description, &rd.DataSourceID, &rd.TimeRange, &rd.Elements, &rd.CreatedAt, &rd.UpdatedAt); err != nil {
			return nil, err
		}
		reports = append(reports, rd)
	}
	return reports, nil
}

func (s *SqliteStore) GetReportDefinitionByID(ctx context.Context, id string) (*models.ReportDefinition, error) {
	query := `SELECT id, name, description, datasource_id, time_range, elements, created_at, updated_at FROM report_definitions WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var rd models.ReportDefinition
	err := row.Scan(&rd.ID, &rd.Name, &rd.Description, &rd.DataSourceID, &rd.TimeRange, &rd.Elements, &rd.CreatedAt, &rd.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rd, nil
}

func (s *SqliteStore) UpdateReportDefinition(ctx context.Context, id string, rd *models.ReportDefinition) error {
	rd.UpdatedAt = time.Now()
	query := `UPDATE report_definitions SET name = ?, description = ?, datasource_id = ?, time_range = ?, elements = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, rd.Name, rd.Description, rd.DataSourceID, rd.TimeRange, rd.Elements, rd.UpdatedAt, id)
	return err
}

func (s *SqliteStore) DeleteReportDefinition(ctx context.Context, id string) error {
	query := `DELETE FROM report_definitions WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
