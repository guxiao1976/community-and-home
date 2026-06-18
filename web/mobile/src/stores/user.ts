import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import {
  getAccessToken,
  getRefreshToken,
  setTokens,
  clearTokens,
  isAuthenticated,
} from '@common/utils/auth';
import type { User } from '@common/types/identity';

export const useUserStore = defineStore('user', () => {
  // --- State ---
  const user = ref<User | null>(null);
  const accessToken = ref<string | null>(getAccessToken());
  const refreshToken = ref<string | null>(getRefreshToken());

  // --- Getters ---
  const isLoggedIn = computed(() => isAuthenticated() && user.value !== null);
  const userId = computed(() => user.value?.id || '');
  const nickname = computed(() => user.value?.nickname || '');
  const avatar = computed(() => user.value?.avatar || '');
  const userType = computed(() => user.value?.userType);

  // --- Actions ---
  function setAuth(loginResponse: {
    accessToken: string;
    refreshToken: string;
    expiresAt: number;
  }) {
    accessToken.value = loginResponse.accessToken;
    refreshToken.value = loginResponse.refreshToken;
    setTokens(
      loginResponse.accessToken,
      loginResponse.refreshToken,
      loginResponse.expiresAt,
    );
  }

  function setUser(newUser: User) {
    user.value = newUser;
  }

  function logout() {
    user.value = null;
    accessToken.value = null;
    refreshToken.value = null;
    clearTokens();
  }

  return {
    // State
    user,
    accessToken,
    refreshToken,
    // Getters
    isLoggedIn,
    userId,
    nickname,
    avatar,
    userType,
    // Actions
    setAuth,
    setUser,
    logout,
  };
});
