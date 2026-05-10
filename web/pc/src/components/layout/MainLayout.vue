<template>
  <div class="main-layout">
    <app-header />
    <div class="layout-container">
      <app-sidebar />
      <div class="main-content">
        <app-breadcrumb />
        <div class="content-wrapper">
          <router-view />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import AppHeader from './AppHeader.vue';
import AppSidebar from './AppSidebar.vue';
import AppBreadcrumb from './AppBreadcrumb.vue';
import { logger } from '@/utils/logger';

onMounted(() => {
  logger.componentMounted('MainLayout');
});

onUnmounted(() => {
  logger.componentUnmounted('MainLayout');
});
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.main-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.layout-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: $background-base;
}

.content-wrapper {
  flex: 1;
  padding: $spacing-lg;
  overflow-y: auto;
}
</style>
