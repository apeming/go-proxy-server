import axios from 'axios';
import { message } from 'antd';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const errorMessage = error.response?.data || error.message || '请求失败';
    message.error(errorMessage);
    return Promise.reject(error);
  }
);

export default api;
