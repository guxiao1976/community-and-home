// Authentication helpers for token storage

import { TOKEN_CONFIG } from '../constants/config';

/**
 * Get access token from sessionStorage
 */
export function getAccessToken(): string | null {
  return sessionStorage.getItem(TOKEN_CONFIG.accessTokenKey);
}

/**
 * Get refresh token from localStorage
 */
export function getRefreshToken(): string | null {
  return localStorage.getItem(TOKEN_CONFIG.refreshTokenKey);
}

/**
 * Get token expiry timestamp from localStorage
 */
export function getTokenExpiry(): number {
  const expiry = localStorage.getItem(TOKEN_CONFIG.tokenExpiryKey);
  return expiry ? parseInt(expiry, 10) : 0;
}

/**
 * Set tokens in storage
 */
export function setTokens(accessToken: string, refreshToken: string, expiresIn: number): void {
  const expiryTime = Date.now() + expiresIn * 1000;

  sessionStorage.setItem(TOKEN_CONFIG.accessTokenKey, accessToken);
  localStorage.setItem(TOKEN_CONFIG.refreshTokenKey, refreshToken);
  localStorage.setItem(TOKEN_CONFIG.tokenExpiryKey, expiryTime.toString());
}

/**
 * Clear all tokens from storage
 */
export function clearTokens(): void {
  sessionStorage.removeItem(TOKEN_CONFIG.accessTokenKey);
  localStorage.removeItem(TOKEN_CONFIG.refreshTokenKey);
  localStorage.removeItem(TOKEN_CONFIG.tokenExpiryKey);
}

/**
 * Check if token is expiring soon (within 5 minutes)
 */
export function isTokenExpiring(): boolean {
  const expiry = getTokenExpiry();
  if (!expiry) return false;

  const now = Date.now();
  const fiveMinutes = 5 * 60 * 1000;
  return expiry - now < fiveMinutes;
}

/**
 * Check if user is authenticated (has valid access token)
 */
export function isAuthenticated(): boolean {
  return !!getAccessToken();
}
