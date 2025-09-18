# React + TypeScript + Vite (繁體中文)

此樣板提供了一個最小化的設定，讓 React 可以在 Vite 中運作，並包含 HMR (熱模組替換) 和一些 ESLint 規則。

目前有兩個官方外掛可用：

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) 使用 [Babel](https://babeljs.io/) 來實現快速刷新 (Fast Refresh)。
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) 使用 [SWC](https://swc.rs/) 來實現快速刷新。

## 擴充 ESLint 設定

如果您正在開發一個生產環境的應用程式，我們建議您更新設定以啟用類型感知的 lint 規則：

```javascript
export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // 其他設定...

      // 移除 tseslint.configs.recommended 並替換為此設定
      ...tseslint.configs.recommendedTypeChecked,
      // 或者，使用此設定以獲得更嚴格的規則
      ...tseslint.configs.strictTypeChecked,
      // 選擇性地，加入此設定以應用風格規則
      ...tseslint.configs.stylisticTypeChecked,

      // 其他設定...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // 其他選項...
    },
  },
])
```

您也可以安裝 [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) 和 [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) 來獲得 React 專用的 lint 規則：

```javascript
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // 其他設定...
      // 啟用 React 的 lint 規則
      reactX.configs['recommended-typescript'],
      // 啟用 React DOM 的 lint 規則
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // 其他選項...
    },
  },
])
```
