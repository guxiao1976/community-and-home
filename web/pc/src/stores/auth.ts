// Authentication store

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { User, LoginResponse } from '@common/types/identity';
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
    const response = await identityApi.login(phone, password);
    await handleLoginResponse(response, phone);
  };

  const loginWithSms = async (phone: string, smsCode: string): Promise<void> => {
    const response = await identityApi.loginWithSms(phone, smsCode);
    await handleLoginResponse(response, phone);
  };

  const register = async (data: {
    phone: string;
    password?: string;
    smsCode: string;
    nickname: string;
  }): Promise<void> => {
    const response = await identityApi.register(data);
    await handleLoginResponse(response, data.phone);
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
    updateTokens(response.accessToken, response.refreshToken, response.expiresAt);
  };

  const handleLoginResponse = async (response: LoginResponse, phone?: string): Promise<void> => {
    updateTokens(response.accessToken, response.refreshToken, response.expiresAt);

    // Construct minimal User from login response. Load full user + permissions in background.
    user.value = {
      id: response.userId,
      phone: phone || '',
      nickname: '',
      avatar: '',
      userType: 1,
      status: 1,
      verificationStatus: 0,
      scope: '',
      lastLoginAt: '',
      createdAt: '',
      updatedAt: '',
      deleteTime: 0,
    } as User;

    if (response.userId) {
      const permissionStore = usePermissionStore();
      permissionStore.loadUserPermissionsAndMenus(response.userId);
    }
  };

  const updateTokens = (access: string, refresh: string, expiresAt: number): void => {
    accessToken.value = access;
    refreshToken.value = refresh;
    tokenExpiry.value = expiresAt * 1000; // Convert Unix timestamp (seconds) to milliseconds
    setTokens(access, refresh, expiresAt);
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
