# Report Scheduler

本系統為一套專為 Kibana 與 Grafana 設計的儀表板擷取與排程報表寄送平台。旨在整合 Elasticsearch、Kibana、Grafana，提供一個集中式的可視化介面，讓使用者能輕鬆完成報表組態、內容挑選、排程寄送與紀錄追蹤。

## 專案結構

本專案採用前後端分離架構：

-   `backend/`: Go 語言開發的後端服務。
-   `frontend/`: 使用 React, TypeScript, Ant Design 開發的前端應用程式。

## 開發入門

### 環境需求

-   Go (建議版本 1.18 或以上)

### 啟動後端伺服器

1.  進入後端專案目錄：
    ```bash
    cd backend
    ```

2.  執行 `go mod tidy` 下載依賴：
    ```bash
    go mod tidy
    ```

3.  啟動伺服器：
    ```bash
    go run ./cmd/server/main.go
    ```

伺服器將會啟動在 `http://localhost:8080`。
