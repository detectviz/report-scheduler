# **Report Scheduler - 系統規格文件 (v2.0)**

## **1. 系統簡介**

本系統為一套專為 Kibana 與 Grafana 設計的儀表板擷取與排程報表寄送平台。旨在整合 Elasticsearch、Kibana、Grafana，提供一個集中式的可視化介面，讓使用者能輕鬆完成報表組態、內容挑選、排程寄送與紀錄追蹤。

### **1.1 核心功能模組**

  - **資料來源管理 (Data Source Management):** 集中管理與 Kibana/Grafana/Elasticsearch 的連線。
  - **報表定義 (Report Definition):** 彈性組合儀表板、圖表或表格，定義報表內容。
  - **排程管理 (Schedule Management):** 使用 Cron 語法設定週期性任務，自動寄送報表。
  - **報表產生與寄送 (Report Generation & Delivery):** 後端核心服務，負責擷取、格式化並寄送報表。
  - **歷史紀錄與稽核 (History & Auditing):** 追蹤所有排程的執行狀態，並提供重寄功能。

## **2. 技術棧與架構**

### **2.1 技術選型**

  - **後端 (Backend):** Go
  - **前端 (Frontend):** React, TypeScript, Ant Design
  - **身分驗證 (Authentication):** Keycloak (Factory Provider 模式)
  - **任務佇列 (Task Queue):** Redis 或 InMemory (Factory Provider 模式)
  - **資料庫 (Database):** PostgreSQL 或 MySQL (Factory Provider 模式)
  - **憑證管理 (Secrets Management):** HashiCorp Vault 或 Kubernetes Secret (Factory Provider 模式)
  - **物件儲存 (Object Storage):** MinIO, Google Cloud Storage, or Amazon S3

### **2.2 高階架構**

系統採用前後端分離架構，後端為無狀態 (Stateless) 服務，並搭配任務佇列與 Worker Pool 處理高資源消耗的報表產生任務，以確保系統的可擴展性與效能。

1.  **API Server (Go):** 負責處理前端請求、管理設定與排程。
2.  **Scheduler:** 內建於 API Server 或使用 K8s CronJob，負責在指定時間觸發任務。
3.  **Task Queue:** 接收由 Scheduler 發出的報表產生任務。
4.  **Worker (Go):** 獨立的執行緒或 Pod，從佇列中取得任務，執行耗時的報表擷取 (Puppeteer) 與格式化工作。
5.  **Object Storage:** 儲存產生的報表檔案 (PDF/CSV)，並提供有時效性的下載連結。

## **3. 登入與權限管理**

  - **系統入口:** `https://<domain>:<port>`
  - **身分驗證:** 透過 OIDC 協定與 **Keycloak** 整合，由 Keycloak 統一管理使用者帳號密碼。
  - **權限控制 (RBAC):**
      - **Admin:** 擁有所有權限，包含管理資料來源、使用者與系統設定。
      - **User:** 可建立/管理自己建立的報表定義與排程，但無法存取系統層級設定。

## **4. 資料來源管理**

管理與外部 BI 系統的連線設定。

### **4.1 資料庫欄位**

| 欄位名稱 | 型別 | 說明 | 範例 |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | 唯一識別碼 | `a3b8d4c2-6e7f-4b0a-9c1d-8e2f0a1b3c4d` |
| `name` | `String` | 資料來源顯示名稱 | `公司正式環境 Kibana` |
| `type` | `Enum` | 類型 | `kibana`, `grafana`, `elasticsearch` |
| `url` | `String` | 系統主頁面 URL | `https://kibana.mycompany.com` |
| `api_url` | `String` | 對應的 API Endpoint | `https://kibana.mycompany.com/api` |
| `auth_type` | `Enum` | 認證方式 | `basic_auth`, `api_token` |
| `credentials_ref` | `String` | **[安全]** 指向 Vault Secret 的路徑 | `kv/report-scheduler/kibana-prod` |
| `version` | `String` | **[選填]** 對應系統版本 | `8.5.1` |
| `status` | `Enum` | 連線狀態 | `verified`, `unverified`, `error` |
| `created_at` | `Timestamp` | 建立時間 | |
| `updated_at` | `Timestamp` | 更新時間 | |

### **4.2 操作功能**

  - **【新增】:** 開啟表單填寫上述資訊。`credentials_ref` 欄位由後端邏輯處理，前端僅需填寫帳號密碼/Token。
  - **【編輯】:** 修改已存在的連線設定。
  - **【刪除】:** 移除資料來源。若該來源仍被報表定義使用，應提示使用者無法刪除。
  - **【驗證連線】:** 觸發後端 API，使用儲存的憑證嘗試連線至目標系統。成功或失敗後更新 `status` 欄位。

## **5. 報表定義**

定義單份報表的具體內容。

### **5.1 報表建立流程**

1.  **基本資訊:** 填寫報表名稱、描述。
2.  **選擇資料來源:** 從已驗證的 `Data Source` 清單中選擇一個。
3.  **設定時間範圍:**
      - **快捷選單:** `過去 24 小時`, `過去 7 天`, `上個月` 等。
      - **相對時間:** 數字 + 單位 (例: `7` `天`)。後端儲存為 `now-7d` 格式。
4.  **挑選報表元素 (Element):**
      - 系統會透過 API 動態載入該資料來源下的可用元素清單 (附帶搜尋功能)。
      - 使用者可以挑選多個元素，並**拖曳排序**。
      - **支援類型:**
          - `Dashboard` (整個儀表板)
          - `Visualization` (單一圖表)
          - `Saved Search` (Kibana 已儲存的搜尋，用於匯出 CSV/Excel)

### **5.2 匯總表格 (Table)**

  - **MVP 階段:** 僅支援 Kibana 的 `Saved Search`。使用者在 Kibana 中定義好查詢與欄位，本系統負責執行並匯出。
  - **未來擴充:** 考慮內建一個簡易的查詢產生器，讓使用者直接在本系統中選擇 Index Pattern、聚合欄位與指標。

## **6. 排程管理**

定義報表的寄送時機與對象。

### **6.1 排程設定欄位**

| 欄位名稱 | 型別 | 說明 | 範例 |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | 唯一識別碼 | |
| `name` | `String` | 排程名稱 | `每日營運日報` |
| `cron_spec` | `String` | Cron 表達式 | `0 9 * * 1-5` (週一至週五早上 9 點) |
| `timezone` | `String` | 時區 | `Asia/Taipei` |
| `recipients` | `JSON` | 收件者 | `{ "to": [...], "cc": [...], "bcc": [...] }` |
| `email_subject` | `String` | 郵件主旨 (支援變數) | `[每日報表] {{report_name}} - {{date}}` |
| `email_body` | `String` | 郵件內文 (支援變數) | `您好，附件為今日的營運報表。` |
| `report_ids` | `Array<UUID>` | 綁定的報表定義 ID (一對多) | `["...", "..."]` |
| `is_enabled` | `Boolean` | 是否啟用此排程 | `true` |

### **6.2 錯誤處理**

  - **重試機制:** 若排程因暫時性問題 (如網路不穩、目標服務無回應) 執行失敗，系統將自動重試。
      - **重試次數:** 3 次
      - **重試間隔:** 10 分鐘

### **6.3 操作功能**

  - **【啟用/停用】:** 快速切換排程的生效狀態。
  - **【立即測試】:** 手動觸發一次排程，寄送測試郵件至當前登入使用者信箱，方便驗證設定是否正確。

## **7. 報表產生與寄送**

### **7.1 擷取方式**

  - **優先順序:**
    1.  **API 優先:** 盡可能使用 Kibana Reporting API 或 Grafana Image Renderer 等官方 API，效能較佳。
    2.  **Headless Browser 備用:** 若 API 無法滿足需求 (如需擷取整個互動式頁面)，則使用 Headless Browser (Puppeteer) 進行截圖。
  - **版面配置:**
      - 對於包含多個元素的 PDF 報表，MVP 階段採用**垂直堆疊**方式，每個元素佔據頁面完整寬度，並自動分頁。

### **7.2 格式支援**

  - **PDF:** 用於 Dashboard/Visualization 截圖。
  - **CSV/Excel:** 用於 Kibana Saved Search 或 Elasticsearch 查詢結果。

### **7.3 寄送方式**

  - **SMTP:** 系統設定中提供 SMTP Server 的組態介面。
  - **未來擴充:** 支援 SendGrid / Amazon SES 等外部郵件服務 API。

## **8. 歷史紀錄與稽核**

### **8.1 記錄內容**

  - `ScheduleName` (排程名稱)
  - `TriggerTime` (觸發時間)
  - `ExecutionDuration` (執行耗時)
  - `Status` (成功 / 失敗 / 重試中)
  - `ErrorMessage` (失敗原因)
  - `Recipients` (收件者資訊)
  - `ReportURL` (指向物件儲存的檔案連結，具備 TTL)

### **8.2 資料保留策略 (Data Retention)**

  - 執行紀錄預設保留 **90 天**，可由系統管理員設定。

### **8.3 重寄功能**

  - **【重寄】:** 允許使用者對「失敗」或「成功」的紀錄手動觸發一次重寄。
  - **執行邏輯:** 系統會使用**原始的 `TriggerTime`** 作為報表的時間基準點，重新產生並寄送一次。
  - **限制:** 若原始的報表定義或資料來源已被刪除，則無法重寄，並應在介面提示使用者。

## **9. API 端點定義 (初版)**

```
// 資料來源
POST   /api/v1/datasources
GET    /api/v1/datasources
GET    /api/v1/datasources/{id}
PUT    /api/v1/datasources/{id}
DELETE /api/v1/datasources/{id}
POST   /api/v1/datasources/{id}/validate

// 報表定義
POST   /api/v1/reports
GET    /api/v1/reports
GET    /api/v1/reports/{id}
PUT    /api/v1/reports/{id}
DELETE /api/v1/reports/{id}

// 排程管理
POST   /api/v1/schedules
GET    /api/v1/schedules
GET    /api/v1/schedules/{id}
PUT    /api/v1/schedules/{id}
DELETE /api/v1/schedules/{id}
POST   /api/v1/schedules/{id}/trigger // 立即測試

// 歷史紀錄
GET    /api/v1/history?schedule_id={id}
POST   /api/v1/history/{log_id}/resend // 重寄
```

## **10. MVP 範圍與未來展望**

### **10.1 MVP 核心功能**

  - **[單一資料來源]** 支援設定一個 Kibana 或 Grafana 資料來源。
  - **[核心報表類型]** 支援 `Dashboard`, `Visualization` 與 Kibana `Saved Search`。
  - **[PDF/CSV 產出]** 支援產生 PDF (截圖) 與 CSV (Saved Search) 格式。
  - **[基礎排程]** 支援 Cron 語法排程與手動觸發。
  - **[SMTP 寄送]** 支援透過標準 SMTP 服務寄送郵件。
  - **[歷史紀錄]** 提供完整的執行紀錄查詢。

### **10.2 未來展望 (Post-MVP)**

  - 支援多個資料來源。
  - 內建視覺化表格產生器。
  - 支援更豐富的報表版面配置。
  - 支援更多寄送渠道 (如 Slack, Webhook)。
  - 提供更詳細的權限管理模型 (如團隊/專案空間)。
