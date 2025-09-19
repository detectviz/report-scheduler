/**
 * 開發模式下的 API 模擬開關
 *
 * 將此值設為 `true` 來啟用 API 模擬，所有 API 請求將會回傳預先定義好的假資料。
 * 將此值設為 `false` 來停用 API 模擬，所有 API 請求將會正常發送到後端伺服器。
 *
 * 注意：此設定僅在開發模式 (`import.meta.env.DEV` 為 true) 下生效。
 */
export const MOCK_ENABLED = true;
