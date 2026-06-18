<template>
  <div class="error-page">
    <div class="error-content">
      <div class="error-code">403</div>
      <div class="error-title">访问被拒绝</div>
      <div class="error-description">
        抱歉，您没有权限访问此页面
      </div>
      <el-button type="primary" @click="goBack">返回上一页</el-button>
      <el-button @click="goHome">返回首页</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();

const goBack = () => {
  router.back();
};

const goHome = () => {
  router.push('/');
};
</script>

<style scoped lang="scss">
.error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: #f5f7fa;
}

.error-content {
  text-align: center;
}

.error-code {
  font-size: 120px;
  font-weight: bold;
  color: #0091FF;
  line-height: 1;
  margin-bottom: 20px;
}

.error-title {
  font-size: 24px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 12px;
}

.error-description {
  font-size: 14px;
  color: #909399;
  margin-bottom: 32px;
}

.el-button {
  margin: 0 8px;
}
</style>
