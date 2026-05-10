<template>
  <div class="app-sidebar" :class="{ collapsed: sidebarCollapsed }">
    <!-- 收缩按钮 -->
    <div class="sidebar-toggle" @click="toggleSidebar">
      <el-icon :size="16">
        <Fold v-if="!sidebarCollapsed" />
        <Expand v-else />
      </el-icon>
      <span v-if="!sidebarCollapsed" class="toggle-text">收起菜单</span>
    </div>

    <el-menu
      :default-active="activeMenu"
      :collapse="sidebarCollapsed"
      :unique-opened="true"
      background-color="#ffffff"
      text-color="#4e5969"
      active-text-color="#1890ff"
      @select="handleMenuSelect"
    >
      <template v-for="item in menuItems" :key="item.path">
        <!-- Single menu item -->
        <el-menu-item v-if="!item.children" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>

        <!-- Menu with children -->
        <el-sub-menu v-else :index="item.path">
          <template #title>
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </template>
          <el-menu-item
            v-for="child in item.children"
            :key="child.path"
            :index="child.path"
          >
            <el-icon><component :is="child.icon" /></el-icon>
            <template #title>{{ child.title }}</template>
          </el-menu-item>
        </el-sub-menu>
      </template>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  HomeFilled,
  User,
  Location,
  OfficeBuilding,
  Setting,
  Document,
  ChatDotSquare,
  Stamp,
  RefreshRight,
  DataAnalysis,
  Download,
  Fold,
  Expand
} from '@element-plus/icons-vue';
import { useAppStore } from '@/stores/app';

interface MenuItem {
  path: string;
  title: string;
  icon?: any;
  children?: MenuItem[];
  permission?: string;
}

const route = useRoute();
const router = useRouter();
const appStore = useAppStore();

const menuItems: MenuItem[] = [
  {
    path: '/dashboard',
    title: '仪表板',
    icon: HomeFilled
  },
  {
    path: '/masterdata',
    title: '主数据管理',
    icon: Setting,
    children: [
      {
        path: '/masterdata/divisions',
        title: '行政区划',
        icon: Location
      },
      {
        path: '/masterdata/grassroots',
        title: '基层组织',
        icon: OfficeBuilding
      },
      {
        path: '/masterdata/residential-areas',
        title: '住宅小区',
        icon: OfficeBuilding
      },
      {
        path: '/masterdata/sensitive-words',
        title: '敏感词管理',
        icon: ChatDotSquare
      },
      {
        path: '/masterdata/configs',
        title: '系统配置',
        icon: Setting
      },
      {
        path: '/masterdata/approval-center',
        title: '审核中心',
        icon: Stamp
      },
      {
        path: '/masterdata/deleted-recovery',
        title: '删除数据恢复',
        icon: RefreshRight
      },
      {
        path: '/masterdata/statistics/division-counts',
        title: '小区数据统计',
        icon: DataAnalysis
      },
      {
        path: '/masterdata/amap-sync',
        title: '高德地图同步',
        icon: Download
      }
    ]
  },
  {
    path: '/users',
    title: '用户管理',
    icon: User,
    children: [
      {
        path: '/users/list',
        title: '用户列表',
        icon: User
      },
      {
        path: '/users/verifications',
        title: '实名审核',
        icon: Document
      }
    ]
  }
];

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed);
const activeMenu = computed(() => route.path);

const toggleSidebar = () => {
  appStore.toggleSidebar();
};

const handleMenuSelect = (index: string) => {
  router.push(index);
};
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.app-sidebar {
  width: $sidebar-width;
  height: 100%;
  background: $sidebar-bg;
  border-right: 1px solid $border-light;
  transition: width $transition-base;
  overflow-x: hidden;
  overflow-y: auto;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;

  &.collapsed {
    width: $sidebar-collapsed-width;
  }

  // 顶部收缩按钮
  .sidebar-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 40px;
    padding: 0 16px;
    cursor: pointer;
    color: $text-secondary;
    font-size: 13px;
    border-bottom: 1px solid $border-lighter;
    flex-shrink: 0;
    transition: color 0.15s;

    &:hover {
      color: $primary-color;
    }

    .toggle-text {
      white-space: nowrap;
    }
  }

  &.collapsed .sidebar-toggle {
    justify-content: center;
    padding: 0;
  }

  :deep(.el-menu) {
    border-right: none;
    padding-top: 4px;
  }

  // 一级菜单项（含子菜单的标题）
  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 40px;
    line-height: 40px;
    margin: 0;
    border-radius: 0;
    transition: all 0.15s ease;
    font-size: 14px;

    &:hover {
      background-color: $sidebar-hover-bg !important;
    }

    .el-icon {
      font-size: 16px;
      width: 16px;
      margin-right: 8px;
    }
  }

  // 选中项：左侧蓝色竖条 + 浅蓝背景（高德控制台风格）
  :deep(.el-menu-item.is-active) {
    background-color: $sidebar-active-bg !important;
    color: $sidebar-active-text !important;
    font-weight: 500;
    position: relative;

    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 4px;
      bottom: 4px;
      width: 3px;
      background-color: $sidebar-active-text;
      border-radius: 0 2px 2px 0;
    }
  }

  // 展开的子菜单标题
  :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
    color: $text-primary !important;
    font-weight: 500;
  }

  // 二级子菜单项
  :deep(.el-sub-menu .el-menu-item) {
    padding-left: 48px !important;
    min-width: auto;
    height: 36px;
    line-height: 36px;
    margin: 0;
    font-size: 13px;
    color: $sidebar-text-color;

    &:hover {
      background-color: $sidebar-hover-bg !important;
    }

    .el-icon {
      font-size: 14px;
      width: 14px;
    }
  }

  // 二级选中项
  :deep(.el-sub-menu .el-menu-item.is-active) {
    background-color: $sidebar-active-bg !important;
    color: $sidebar-active-text !important;

    &::before {
      left: 0;
      top: 3px;
      bottom: 3px;
      width: 3px;
    }
  }

  // 子菜单展开区域背景透明
  :deep(.el-sub-menu .el-menu) {
    background-color: transparent !important;
  }

  // 展开箭头
  :deep(.el-sub-menu__icon-arrow) {
    font-size: 12px;
    color: $text-secondary;
  }

  // Collapse 模式
  &.collapsed {
    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      margin: 0;
    }
  }
}
</style>
