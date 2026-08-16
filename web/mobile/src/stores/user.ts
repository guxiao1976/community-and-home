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
  // 登录态修复：token 是权威（isAuthenticated），user 只是 profile 缓存。
  // 修复前 = isAuthenticated() && user !== null —— 登录后 profile 拉取失败或 tab 在登录前已挂载
  // → user=null → 页面误判未登录。改为 token 为准，user 由登录流程/App.vue onLaunch 懒加载恢复。
  const isLoggedIn = computed(() => isAuthenticated());
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
    // 登录期写入的兜底缓存（user_phone）一并清除，防共享设备跨账号串号泄漏。
    // SEE: [[logout-clear-login-cache]]
    try {
      uni.removeStorageSync('user_phone');
    } catch {
      // ignore storage errors
    }
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
