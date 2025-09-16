package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/models"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetDataSources 處理獲取所有資料來源的請求
// @Summary 獲取所有資料來源
// @Description 回傳一個包含所有資料來源的列表
// @Tags datasources
// @Accept  json
// @Produce  json
// @Success 200 {array} models.DataSource
// @Router /datasources [get]
func GetDataSources(w http.ResponseWriter, r *http.Request) {
	mockDataSources := []models.DataSource{
		{
			ID:        "a3b8d4c2-6e7f-4b0a-9c1d-8e2f0a1b3c4d",
			Name:      "公司正式環境 Kibana",
			Type:      models.Kibana,
			URL:       "https://kibana.mycompany.com",
			APIURL:    "https://kibana.mycompany.com/api",
			AuthType:  models.APIToken,
			Version:   "8.5.1",
			Status:    models.Verified,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockDataSources)
}

// CreateDataSource 處理新增資料來源的請求
// @Summary 新增一個資料來源
// @Description 根據傳入的資料建立一個新的資料來源
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   datasource body models.DataSource true "資料來源資訊"
// @Success 201 {object} map[string]string
// @Router /datasources [post]
func CreateDataSource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "資料來源已成功建立"})
}

// GetDataSourceByID 處理根據 ID 獲取單一資料來源的請求
// @Summary 根據 ID 獲取單一資料來源
// @Description 回傳指定 ID 的資料來源資訊
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   datasourceID   path    string  true  "資料來源 ID"
// @Success 200 {object} models.DataSource
// @Router /datasources/{datasourceID} [get]
func GetDataSourceByID(w http.ResponseWriter, r *http.Request) {
	// 從 chi 的 URL 參數中獲取 ID
	id := chi.URLParam(r, "datasourceID")

	mockDataSource := models.DataSource{
		ID:        id, // 使用從 URL 來的 ID
		Name:      "公司正式環境 Kibana (ID: " + id + ")", // 在名稱中也反映 ID
		Type:      models.Kibana,
		URL:       "https://kibana.mycompany.com",
		APIURL:    "https://kibana.mycompany.com/api",
		AuthType:  models.APIToken,
		Version:   "8.5.1",
		Status:    models.Verified,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockDataSource)
}

// UpdateDataSource 處理更新資料來源的請求
// @Summary 更新一個資料來源
// @Description 根據傳入的資料更新指定 ID 的資料來源
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   datasourceID   path    string  true  "資料來源 ID"
// @Param   datasource body models.DataSource true "資料來源資訊"
// @Success 200 {object} map[string]string
// @Router /datasources/{datasourceID} [put]
func UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "資料來源 " + id + " 已成功更新"})
}

// DeleteDataSource 處理刪除資料來源的請求
// @Summary 刪除一個資料來源
// @Description 刪除指定 ID 的資料來源
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   datasourceID   path    string  true  "資料來源 ID"
// @Success 200 {object} map[string]string
// @Router /datasources/{datasourceID} [delete]
func DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "datasourceID")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "資料來源 " + id + " 已成功刪除"})
}
