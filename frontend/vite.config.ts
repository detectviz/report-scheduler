import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  // 重要：部署到 GitHub Pages 時，需要設定正確的 base 路徑
  // 這裡假設你的 repository 名稱為 "report-scheduler"
  // 如果你的 repository 名稱不同，請修改下面的字串
  base: '/report-scheduler/',
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8089',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
