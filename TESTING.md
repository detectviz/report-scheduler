# 前端功能驗證步驟

本文件說明如何對前端的變更進行端對端 (E2E) 的視覺化驗證。

由於目前的執行環境存在問題，導致無法成功啟動 Go 後端伺服器，因此下列步驟是在一個正常的 Go 環境中應如何操作的指南。

## 1. 啟動後端伺服器

首先，需要啟動後端應用程式。在專案的根目錄 (即 `go.mod` 檔案所在的目錄) 下，執行以下指令：

```bash
# 此指令會啟動後端伺服器，並監聽在 http://localhost:8080
go run ./cmd/server/main.go &
```

## 2. 準備 Playwright 驗證環境

我們的驗證腳本使用 Playwright。請確保已安裝相關依賴。

```bash
# 安裝 playwright 的 python 套件
pip install playwright

# 安裝 playwright 所需的瀏覽器核心
playwright install --with-deps
```

## 3. 執行驗證腳本

我們提供了一個 Playwright 腳本來自動化測試「報表定義」頁面的功能。

**腳本內容 (`verify_report_definition.py`):**

```python
import re
from playwright.sync_api import sync_playwright, Page, expect

def run_verification(page: Page):
    # 導航至應用程式
    page.goto("http://localhost:8080")

    # 點擊「報表定義」選單項目
    page.get_by_role("menuitem", name="報表定義").click()

    # 預期頁面標題為「報表定義」
    expect(page.get_by_role("heading", name="報表定義")).to_be_visible()

    # 填寫表單
    page.get_by_label("報表名稱").fill("My Automated Test Report")

    # 選擇一個資料來源
    page.get_by_label("選擇資料來源").click()
    # 點擊第一個可用的選項
    first_option = page.locator(".ant-select-item-option-content").first
    expect(first_option).to_be_visible(timeout=10000) # 等待 API 回應
    first_option.click()

    # 選擇時間範圍
    page.get_by_label("設定時間範圍").click()
    page.get_by_text("10", exact=True).first.click()
    page.get_by_text("15", exact=True).first.click()
    page.get_by_role("button", name="OK").click()

    # 點擊儲存按鈕
    page.get_by_role("button", name="儲存報表定義").click()

    # 預期看到成功訊息
    success_message = page.locator(".ant-message-notice-content", has_text="報表定義已成功儲存！")
    expect(success_message).to_be_visible(timeout=5000)

    # 擷取螢幕截圖
    page.screenshot(path="verification.png")
    print("成功產生驗證截圖: verification.png")

# 執行腳本的標準程式碼
if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            run_verification(page)
            print("驗證腳本成功完成。")
        except Exception as e:
            print(f"發生錯誤: {e}")
            page.screenshot(path="error.png")
        finally:
            browser.close()
```

**執行指令:**

將上面的腳本內容儲存為 `verify_report_definition.py`，然後執行：

```bash
python verify_report_definition.py
```

## 4. 驗證結果

如果腳本執行成功，將會產生一張名為 `verification.png` 的螢幕截圖。這張截圖會顯示成功儲存報表定義後的畫面，包含了右上角的成功提示訊息。

如果執行失敗，則會產生 `error.png`，可用於除錯。
