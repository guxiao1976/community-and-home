// Authentication store

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { User, RegisterRequest, LoginResponse } from '@common/types/identity';
import { getAccessToken, getRefreshToken, setTokens, clearTokens, getTokenExpiry } from '@common/utils/auth';
import * as identityApi from '@/api/identity';
import { usePermissionStore } from '@/stores/permission';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null);
  const accessToken = ref<string | null>(null);
  const refreshToken = ref<string | null>(null);
  const tokenExpiry = ref<number>(0);

  // Computed
  const isAuthenticated = computed(() => !!accessToken.value);
  const isTokenExpiring = computed(() => {
    if (!tokenExpiry.value) return false;
    const now = Date.now();
    const fiveMinutes = 5 * 60 * 1000;
    return tokenExpiry.value - now < fiveMinutes;
  });

  // Actions
  const login = async (phone: string, password: string): Promise<void> => {
    const response = await identityApi.login({ phone, password });
    handleLoginResponse(response);
  };

  const loginWithSms = async (phone: string, smsCode: string): Promise<void> => {
    const response = await identityApi.loginWithSms({ phone, smsCode });
    handleLoginResponse(response);
  };

  const register = async (data: RegisterRequest): Promise<void> => {
    const response = await identityApi.register(data);
    handleLoginResponse(response);
  };

  const logout = async (): Promise<void> => {
    try {
      await identityApi.logout();
    } finally {
      clearSession();
    }
  };

  const refreshAccessToken = async (): Promise<void> => {
    const token = refreshToken.value || getRefreshToken();
    if (!token) {
      throw new Error('No refresh token available');
    }

    const response = await identityApi.refreshToken({ refreshToken: token });
    updateTokens(response.accessToken, response.refreshToken, response.expiresIn);
  };

  const handleLoginResponse = (response: LoginResponse): void => {
    user.value = response.user;
    updateTokens(response.accessToken, response.refreshToken, response.expiresIn);
    // Load user permissions after login
    if (response.user?.id) {
      const permissionStore = usePermissionStore();
      permissionStore.loadUserPermissionsAndMenus(response.user.id);
    }
  };

  const updateTokens = (access: string, refresh: string, expiresIn: number): void => {
    accessToken.value = access;
    refreshToken.value = refresh;
    tokenExpiry.value = Date.now() + expiresIn * 1000;
    setTokens(access, refresh, expiresIn);
  };

  const clearSession = (): void => {
    user.value = null;
    accessToken.value = null;
    refreshToken.value = null;
    tokenExpiry.value = 0;
    clearTokens();
    const permissionStore = usePermissionStore();
    permissionStore.clearPermissions();
  };

  const restoreSession = (): void => {
    const access = getAccessToken();
    const refresh = getRefreshToken();
    const expiry = getTokenExpiry();

    if (access && refresh && expiry) {
      accessToken.value = access;
      refreshToken.value = refresh;
      tokenExpiry.value = expiry;
    }
  };

  return {
    // State
    user,
    accessToken,
    refreshToken,
    tokenExpiry,
    // Computed
    isAuthenticated,
    isTokenExpiring,
    // Actions
    login,
    loginWithSms,
    register,
    logout,
    refreshAccessToken,
    restoreSession,
    clearSession
  };
}, {
  persist: true
});
