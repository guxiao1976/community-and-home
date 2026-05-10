<template>
  <div class="dashboard-container">
    <el-card class="welcome-card">
      <h2>欢迎使用社区家园管理平台</h2>
      <p v-if="authStore.user">
        您好，{{ authStore.user.nickname }}！
      </p>
      <p class="description">
        这是一个基于 Vue3 + TypeScript + Element Plus 的管理后台系统
      </p>

      <el-divider />

      <div class="user-info" v-if="authStore.user">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="手机号">
            {{ authStore.user.phone }}
          </el-descriptions-item>
          <el-descriptions-item label="用户类型">
            {{ authStore.user.userType === 1 ? '员工' : '业主' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="authStore.user.status === 1 ? 'success' : 'danger'">
              {{ authStore.user.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="认证状态">
            <el-tag :type="getVerificationStatusType(authStore.user.verificationStatus)">
              {{ getVerificationStatusLabel(authStore.user.verificationStatus) }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="actions">
        <el-button type="danger" @click="handleLogout">退出登录</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useAuthStore } from '@/stores/auth';

const router = useRouter();
const authStore = useAuthStore();

const getVerificationStatusType = (status: number) => {
  const map: Record<number, any> = {
    0: 'info',
    1: 'success',
    2: 'danger'
  };
  return map[status] || 'info';
};

const getVerificationStatusLabel = (status: number) => {
  const map: Record<number, string> = {
    0: '未认证',
    1: '已认证',
    2: '已拒绝'
  };
  return map[status] || '未知';
};

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });

    await authStore.logout();
    ElMessage.success('已退出登录');
    router.push('/login');
  } catch (error) {
    // User cancelled
  }
};
</script>

<style scoped lang="scss">
.dashboard-container {
  padding: 24px;
}

.welcome-card {
  max-width: 800px;
  margin: 0 auto;

  h2 {
    font-size: 24px;
    font-weight: 600;
    color: #303133;
    margin-bottom: 16px;
  }

  p {
    font-size: 14px;
    color: #606266;
    margin-bottom: 8px;
  }

  .description {
    color: #909399;
  }
}

.user-info {
  margin: 24px 0;
}

.actions {
  margin-top: 24px;
  text-align: center;
}
</style>
