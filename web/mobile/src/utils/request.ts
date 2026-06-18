// Axios instance with interceptors for Uni-app H5 API requests
import axios, {
  type AxiosInstance,
  type AxiosError,
  type InternalAxiosRequestConfig,
  type AxiosResponse,
} from 'axios';
import { parse, isLosslessNumber, type LosslessNumber } from 'lossless-json';
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@common/utils/auth';

// Custom JSON reviver: keep safe integers as numbers, convert large integers to strings.
// Snowflake IDs are 19-digit int64 values that exceed JS Number.MAX_SAFE_INTEGER (2^53).
function reviveLargeNumbers(_key: string, value: unknown): unknown {
  if (isLosslessNumber(value)) {
    const asNumber = Number((value as LosslessNumber).toString());
    if (Number.isSafeInteger(asNumber)) {
      return asNumber;
    }
    return (value as LosslessNumber).toString();
  }
  return value;
}

const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
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
    },
  ],
});

// Token refresh state — prevents concurrent refresh calls
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null, token: string | null = null) => {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error);
    } else {
      promise.resolve(token);
    }
  });
  failedQueue = [];
};

// --- Request Interceptor ---
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  },
);

// --- Response Interceptor ---
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { code, msg, message, data } = response.data as Record<string, unknown>;

    // Backend uses "msg" (go-zero), support both "msg" and "message"
    const errorMessage = (msg || message) as string | undefined;

    // If response doesn't have code field, treat it as success and return raw data
    if (code === undefined) {
      return response.data;
    }

    if (code === 0) {
      // Unwrap the "data" field — align with PC's interceptor
      return data !== undefined ? (data as any) : response.data;
    }

    // Business error
    uni.showToast({
      title: errorMessage || '请求失败',
      icon: 'none',
      duration: 2000,
    });
    return Promise.reject(new Error(errorMessage || `Business error code: ${code}`));
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Handle 401 — attempt token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      const refreshTokenValue = getRefreshToken();
      if (!refreshTokenValue) {
        clearTokens();
        uni.reLaunch({ url: '/pages/login/login' });
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return request(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      isRefreshing = true;
      originalRequest._retry = true;

      try {
        const response = await axios.post('/api/auth/token/refresh', {
          refreshToken: refreshTokenValue,
        });

        const data = response.data as {
          data?: { accessToken: string; refreshToken: string; expiresAt: number };
        };
        if (data.data) {
          const { accessToken, refreshToken: newRefreshToken, expiresAt } = data.data;
          setTokens(accessToken, newRefreshToken, expiresAt);
          processQueue(null, accessToken);
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${accessToken}`;
          }
          return request(originalRequest);
        }

        throw new Error('Token refresh returned no data');
      } catch (refreshError) {
        processQueue(refreshError as AxiosError, null);
        clearTokens();
        uni.reLaunch({ url: '/pages/login/login' });
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    // Network error
    if (!error.response) {
      uni.showToast({
        title: '网络连接失败，请检查网络',
        icon: 'none',
        duration: 2000,
      });
    }

    return Promise.reject(error);
  },
);

export default request;
export type { AxiosResponse, AxiosError };
