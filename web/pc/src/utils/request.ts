// Axios instance with interceptors for API requests

import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig, type AxiosResponse } from 'axios';
import { ElMessage } from 'element-plus';
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@common/utils/auth';
import { ErrorCode } from '@common/constants/error-codes';
import type { ApiResponse } from '@common/types/common';
import { logger } from './logger';
import { parse, isLosslessNumber, LosslessNumber } from 'lossless-json';

// 端限制拒绝（auth-service 50007）：账号角色 platforms 与当前登录端不匹配
// SEE: [[error-code-collision-and-namespace-alignment]] — 50007 为 auth 端限制拒绝专用码，前端仅消费不另设语义
export const ERROR_CODE_PLATFORM_RESTRICTED = 50007;
// 端限制拒绝引导文案（与后端 auth-service 50007 返回 msg 一致）
export const PLATFORM_RESTRICTED_MESSAGE = '该账号为移动端用户，请使用移动端 APP';

// Custom JSON reviver: keep safe integers as numbers, convert large integers to strings.
// Snowflake IDs are 19-digit int64 values that exceed JS Number.MAX_SAFE_INTEGER (2^53).
// By converting only unsafe integers to strings, normal number fields (status, page, total)
// continue to work as expected.
function reviveLargeNumbers(_key: string, value: unknown): unknown {
  if (isLosslessNumber(value)) {
    const asNumber = Number((value as LosslessNumber).toString());
    if (Number.isSafeInteger(asNumber)) {
      return asNumber;
    }
    // Large integer (e.g., snowflake ID) — return as string to preserve precision
    return (value as LosslessNumber).toString();
  }
  return value;
}

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  transformResponse: [
    (data: string) => {
      if (typeof data === 'string') {
        try {
          return parse(data, reviveLargeNumbers);
        } catch {
          return data;
        }
      }
      return data;
    }
  ]
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
    const { code, message, msg, data } = response.data as any;

    // Backend uses "msg" (go-zero), frontend uses "message" — support both
    const errorMessage = message || msg;

    logger.apiResponse(
      response.config.method?.toUpperCase() || 'GET',
      response.config.url || '',
      response.status,
      { code, message: errorMessage, dataSize: JSON.stringify(data ?? response.data).length }
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
    logger.warn('Business error', { code, message: errorMessage });

    // 携带错误码供调用方（登录/注册页）按 code 分支处理
    const bizError = new Error(errorMessage || '请求失败') as Error & { code?: number };
    bizError.code = code;

    // 50007 端限制拒绝：由登录/注册页展示专属引导文案，此处不发通用 toast（避免双 toast 且文案更精准）
    if (code !== ERROR_CODE_PLATFORM_RESTRICTED) {
      ElMessage.error(errorMessage || '请求失败');
    }
    return Promise.reject(bizError);
  },
  async (error: AxiosError<ApiResponse<any>>) => {
    logger.apiError(
      error.config?.method?.toUpperCase() || 'UNKNOWN',
      error.config?.url || 'unknown',
      error
    );

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // logout 请求的 401 直接静默失败，不触发 Token 刷新
    if (originalRequest?.url?.includes('/api/auth/logout')) {
      return Promise.reject(error);
    }

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
          expiresAt: number;
        }>>('/api/auth/token/refresh', {
          refreshToken
        });

        const respData = response.data as any;
        // 刷新 Token 也可能触发端限制（50007）：角色 platforms 变更后旧会话刷新被拒
        if (respData && typeof respData.code === 'number' && respData.code !== ErrorCode.Success && respData.code !== 200) {
          const refreshErr = new Error(respData.msg || respData.message || '登录状态已失效') as Error & { code?: number };
          refreshErr.code = respData.code;
          throw refreshErr;
        }

        const { accessToken, refreshToken: newRefreshToken, expiresAt } = respData.data;

        // Update tokens
        setTokens(accessToken, newRefreshToken, expiresAt);

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

        // 50007 端限制拒绝：刷新被拒时提示用户换端登录
        const refreshErr = refreshError as Error & { code?: number };
        if (refreshErr?.code === ERROR_CODE_PLATFORM_RESTRICTED) {
          ElMessage.error(PLATFORM_RESTRICTED_MESSAGE);
        }

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
    const errData = error.response?.data as any;
    const message = errData?.msg || errData?.message || error.message || '请求失败';

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
