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
  padding: 32px 24px;
}

.header {
  margin-bottom: 32px;

  .title {
    display: block;
    font-size: 48rpx;
    font-weight: 700;
    color: $uni-text-color;
  }
}

.content {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 120rpx 0;

  .placeholder {
    font-size: 28rpx;
    color: $uni-text-color-placeholder;
  }
}
</style>
