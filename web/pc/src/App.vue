<script setup lang="ts">
import { onMounted } from 'vue';
import { RouterView } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { usePermissionStore } from '@/stores/permission';

const authStore = useAuthStore();
const permissionStore = usePermissionStore();

// Restore session on app mount
onMounted(async () => {
  authStore.restoreSession();
  // Load permissions if already authenticated and token not expired
  if (authStore.isAuthenticated && authStore.user?.id) {
    // Check if token is expired
    const now = Date.now();
    if (authStore.tokenExpiry && authStore.tokenExpiry > now) {
      try {
        await permissionStore.loadUserPermissionsAndMenus(authStore.user.id);
      } catch (error: any) {
        // If 401, clear session to stop the loop
        if (error.response?.status === 401) {
          authStore.clearSession();
        }
      }
    } else {
      // Token expired, clear session
      authStore.clearSession();
    }
  }
});
</script>

<template>
  <router-view v-slot="{ Component, route }">
    <transition name="fade" mode="out-in">
      <component :is="Component" :key="route.path" />
    </transition>
  </router-view>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
