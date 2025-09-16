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
	// 開啟資料庫連線
	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		return nil, err
	}

	// 檢查連線是否成功
	if err := db.Ping(); err != nil {
		return nil, err
	}

	store := &SqliteStore{db: db}
	// 初始化資料表
	if err := store.initSchema(); err != nil {
		return nil, err
	}

	return store, nil
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
		version TEXT,
		status TEXT NOT NULL,
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

	query := `INSERT INTO datasources (id, name, type, url, api_url, auth_type, version, status, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, ds.ID, ds.Name, ds.Type, ds.URL, ds.APIURL, ds.AuthType, ds.Version, ds.Status, ds.CreatedAt, ds.UpdatedAt)
	return err
}

func (s *SqliteStore) GetDataSources(ctx context.Context) ([]models.DataSource, error) {
	query := `SELECT id, name, type, url, api_url, auth_type, version, status, created_at, updated_at FROM datasources`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataSources []models.DataSource
	for rows.Next() {
		var ds models.DataSource
		if err := rows.Scan(&ds.ID, &ds.Name, &ds.Type, &ds.URL, &ds.APIURL, &ds.AuthType, &ds.Version, &ds.Status, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, err
		}
		dataSources = append(dataSources, ds)
	}
	return dataSources, nil
}

func (s *SqliteStore) GetDataSourceByID(ctx context.Context, id string) (*models.DataSource, error) {
	query := `SELECT id, name, type, url, api_url, auth_type, version, status, created_at, updated_at FROM datasources WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)

	var ds models.DataSource
	err := row.Scan(&ds.ID, &ds.Name, &ds.Type, &ds.URL, &ds.APIURL, &ds.AuthType, &ds.Version, &ds.Status, &ds.CreatedAt, &ds.UpdatedAt)
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
	query := `UPDATE datasources SET name = ?, type = ?, url = ?, api_url = ?, auth_type = ?, version = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, ds.Name, ds.Type, ds.URL, ds.APIURL, ds.AuthType, ds.Version, ds.Status, ds.UpdatedAt, id)
	return err
}

func (s *SqliteStore) DeleteDataSource(ctx context.Context, id string) error {
	query := `DELETE FROM datasources WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
