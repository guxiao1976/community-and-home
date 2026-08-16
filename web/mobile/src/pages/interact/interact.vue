<template>
  <view class="page">
    <view class="header">
      <text class="title">邻里互动</text>
    </view>
    <view class="content">
      <text class="placeholder">邻里互动功能开发中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
  padding: 2rem 1.5rem;
}

.header {
  margin-bottom: 2rem;

  .title {
    display: block;
    font-size: 1.5rem;
    font-weight: 700;
    color: $uni-text-color;
  }
}

.content {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3.75rem 0;

  .placeholder {
    font-size: 0.875rem;
    color: $uni-text-color-placeholder;
  }
}
</style>
