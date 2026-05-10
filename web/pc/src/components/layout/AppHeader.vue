<template>
  <div class="app-header">
    <div class="header-content">
      <div class="header-left">
        <div class="app-brand">
          <span class="app-title">社区管理系统</span>
        </div>
      </div>

      <div class="header-right">
        <el-dropdown @command="handleCommand">
          <div class="user-info">
            <el-avatar :size="28" :src="userAvatar">
              <el-icon><User /></el-icon>
            </el-avatar>
            <span class="username">{{ currentUser?.nickname || '未登录' }}</span>
            <el-icon class="arrow-icon"><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><User /></el-icon>
                个人信息
              </el-dropdown-item>
              <el-dropdown-item command="settings">
                <el-icon><Setting /></el-icon>
                系统设置
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { User, ArrowDown, Setting, SwitchButton } from '@element-plus/icons-vue';
import { useAuthStore } from '@/stores/auth';
const router = useRouter();
const authStore = useAuthStore();

const currentUser = computed(() => authStore.user);
const userAvatar = computed(() => currentUser.value?.avatar || '');

const handleCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile');
      break;
    case 'settings':
      router.push('/settings');
      break;
    case 'logout':
      try {
        await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        });
        await authStore.logout();
        ElMessage.success('退出登录成功');
        router.push('/login');
      } catch (error) {
        // User cancelled
      }
      break;
  }
};
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.app-header {
  position: relative;
  flex-shrink: 0;
  z-index: 100;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: $header-height;
  padding: 0 $spacing-lg;
  background: #1e293e;
}

.header-left {
  .app-brand {
    .app-title {
      font-size: 15px;
      font-weight: 500;
      color: #ffffff;
      letter-spacing: 0.3px;
    }
  }
}

.header-right {
  .user-info {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    padding: 4px 12px;
    cursor: pointer;
    border-radius: $border-radius-base;
    transition: $transition-fast;

    &:hover {
      background: rgba(255, 255, 255, 0.1);
    }

    .username {
      font-size: $font-size-base;
      color: rgba(255, 255, 255, 0.85);
    }

    .arrow-icon {
      font-size: 12px;
      color: rgba(255, 255, 255, 0.6);
    }
  }
}
</style>
