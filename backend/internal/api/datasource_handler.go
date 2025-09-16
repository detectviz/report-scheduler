package api

import (
	"encoding/json"
	"net/http"
	"report-scheduler/backend/internal/models"
	"time"
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
	// 模擬的資料
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
	// 在實際應用中，我們會從 r.Body 解析請求內容
	// 這裡只是一個模擬的回應
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
// @Param   id   path    string  true  "資料來源 ID"
// @Success 200 {object} models.DataSource
// @Router /datasources/{id} [get]
func GetDataSourceByID(w http.ResponseWriter, r *http.Request) {
	// 在實際應用中，我們會從 URL 路徑中解析 id
	id := "a3b8d4c2-6e7f-4b0a-9c1d-8e2f0a1b3c4d" // 模擬的 ID
	mockDataSource := models.DataSource{
		ID:        id,
		Name:      "公司正式環境 Kibana",
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
// @Param   id   path    string  true  "資料來源 ID"
// @Param   datasource body models.DataSource true "資料來源資訊"
// @Success 200 {object} map[string]string
// @Router /datasources/{id} [put]
func UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "資料來源已成功更新"})
}

// DeleteDataSource 處理刪除資料來源的請求
// @Summary 刪除一個資料來源
// @Description 刪除指定 ID 的資料來源
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   id   path    string  true  "資料來源 ID"
// @Success 200 {object} map[string]string
// @Router /datasources/{id} [delete]
func DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "資料來源已成功刪除"})
}
