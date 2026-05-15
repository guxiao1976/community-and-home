// Axios instance with interceptors for API requests

import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios';
import { ElMessage } from 'element-plus';
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@common/utils/auth';
import { ErrorCode } from '@common/constants/error-codes';
import type { ApiResponse } from '@common/types/common';
import { logger } from './logger';

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
});

logger.info('Axios initialized', {
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000
});

// Token refresh state
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach(promise => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(token);
    }
  });
  failedQueue = [];
};

// Request interceptor
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    logger.apiRequest(config.method?.toUpperCase() || 'GET', config.url || '', config.data || config.params);

    // Add access token to headers
    const token = getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
      logger.debug('Added Authorization header', { tokenLength: token.length });
    }
    return config;
  },
  (error: AxiosError) => {
    logger.apiError('REQUEST_INTERCEPTOR', 'Request failed', error);
    return Promise.reject(error);
  }
);

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse<any>>) => {
    const { code, message, data } = response.data;

    logger.apiResponse(
      response.config.method?.toUpperCase() || 'GET',
      response.config.url || '',
      response.status,
      { code, message, dataSize: JSON.stringify(data ?? response.data).length }
    );

    // If response doesn't have code field, treat it as success and return raw data
    // This handles APIs that don't follow the standard ApiResponse format
    if (code === undefined) {
      return response.data as any;
    }

    // Success - accept both 0 and 200 as success codes
    // Backend services may return different success codes
    if (code === ErrorCode.Success || code === 200) {
      return data as any;
    }

    // Business error
    logger.warn('Business error', { code, message });
    ElMessage.error(message || '请求失败');
    return Promise.reject(new Error(message || '请求失败'));
  },
  async (error: AxiosError<ApiResponse<any>>) => {
    logger.apiError(
      error.config?.method?.toUpperCase() || 'UNKNOWN',
      error.config?.url || 'unknown',
      error
    );

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Handle 401 Unauthorized - token expired
    if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
      if (isRefreshing) {
        // Queue this request until token refresh completes
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(token => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return request(originalRequest);
          })
          .catch(err => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Attempt to refresh token
        const refreshToken = getRefreshToken();
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        const response = await axios.post<ApiResponse<{
          accessToken: string;
          refreshToken: string;
          expiresIn: number;
        }>>('/api/identity/auth/token/refresh', {
          refreshToken
        });

        const { accessToken, refreshToken: newRefreshToken, expiresIn } = response.data.data;

        // Update tokens
        setTokens(accessToken, newRefreshToken, expiresIn);

        // Update default authorization header
        if (request.defaults.headers.common) {
          request.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
        }

        // Process queued requests
        processQueue(null, accessToken);

        // Retry original request
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        }
        return request(originalRequest);
      } catch (refreshError) {
        // Refresh failed, clear tokens and redirect to login
        processQueue(refreshError as AxiosError, null);
        clearTokens();

        // Clear auth store to prevent loops
        if (typeof window !== 'undefined') {
          // Use dynamic import to avoid circular dependency
          import('@/stores/auth').then(({ useAuthStore }) => {
            const authStore = useAuthStore();
            authStore.clearSession();
          });

          // Only redirect if not already on login page
          if (!window.location.pathname.startsWith('/login')) {
            window.location.href = '/login';
          }
        }

        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Handle other errors
    const message = error.response?.data?.message || error.message || '请求失败';

    if (error.response?.status === 403) {
      ElMessage.error('权限不足');
    } else if (error.response?.status === 404) {
      ElMessage.error('资源不存在');
    } else if (error.response?.status === 500) {
      ElMessage.error('服务器错误');
    } else {
      ElMessage.error(message);
    }

    return Promise.reject(error);
  }
);

// Type-safe request wrappers that account for the response interceptor unwrapping
interface RequestInstance {
  get<T = any>(url: string, config?: any): Promise<T>;
  post<T = any>(url: string, data?: any, config?: any): Promise<T>;
  put<T = any>(url: string, data?: any, config?: any): Promise<T>;
  delete<T = any>(url: string, config?: any): Promise<T>;
  patch<T = any>(url: string, data?: any, config?: any): Promise<T>;
}

export default request as unknown as RequestInstance;
