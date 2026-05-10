<template>
  <div class="breadcrumb-bar">
    <el-breadcrumb separator="/" class="app-breadcrumb">
      <el-breadcrumb-item
        v-for="(item, index) in breadcrumbs"
        :key="index"
        :to="item.path ? { path: item.path } : undefined"
      >
        {{ item.title }}
      </el-breadcrumb-item>
    </el-breadcrumb>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/stores/app';

const appStore = useAppStore();
const breadcrumbs = computed(() => appStore.breadcrumbs);
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.breadcrumb-bar {
  padding: $spacing-sm $spacing-lg;
  background: #fff;
  border-bottom: 1px solid $border-light;
}

.app-breadcrumb {
  font-size: $font-size-sm;

  :deep(.el-breadcrumb__inner) {
    color: $text-secondary;

    &.is-link {
      color: $text-regular;
      font-weight: normal;

      &:hover {
        color: $primary-color;
      }
    }
  }

  :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
    color: $text-primary;
  }
}
</style>
