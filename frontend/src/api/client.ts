import axios from 'axios';
import { message } from 'antd';

// 建立一個 Axios 實例
const apiClient = axios.create({
    baseURL: '/api/v1', // 後端 API 的基礎路徑
    timeout: 5000, // 請求超時時間
});

// 請求攔截器
apiClient.interceptors.request.use(
    (config) => {
        // 在這裡可以添加認證 Token 等
        return config;
    },
    (error) => {
        message.error('請求發送失敗');
        return Promise.reject(error);
    }
);

// 回應攔截器
apiClient.interceptors.response.use(
    (response) => {
        // 直接回傳 response.data，簡化後續操作
        return response.data;
    },
    (error) => {
        // 統一處理錯誤
        const errorMessage = error.response?.data?.error || error.message || '發生未知錯誤';
        message.error(errorMessage);
        return Promise.reject(error);
    }
);

export default apiClient;
