// API configuration

export const API_CONFIG = {
  identity: {
    baseURL: 'http://localhost:8888/api/identity',
    timeout: 30000
  },
  masterdata: {
    baseURL: 'http://localhost:8889/api/masterdata',
    timeout: 30000
  }
};

// Token configuration
export const TOKEN_CONFIG = {
  accessTokenKey: 'accessToken',
  refreshTokenKey: 'refreshToken',
  tokenExpiryKey: 'tokenExpiry',
  accessTokenExpiry: 86400, // 24 hours in seconds
  refreshTokenExpiry: 604800 // 7 days in seconds
};

// Pagination defaults
export const PAGINATION_CONFIG = {
  defaultPage: 1,
  defaultPageSize: 20,
  pageSizes: [10, 20, 50, 100]
};
