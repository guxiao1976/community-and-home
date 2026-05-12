<template>
  <div class="moderation-test-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>内容审核测试</span>
          <el-tag type="info" size="small">开发测试工具</el-tag>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="文本测试" name="text">
          <TextTestTab />
        </el-tab-pane>

        <el-tab-pane label="图片测试" name="image">
          <ImageTestTab />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import TextTestTab from '@/components/moderation/TextTestTab.vue';
import ImageTestTab from '@/components/moderation/ImageTestTab.vue';

const activeTab = ref('text');
</script>

<style scoped>
.moderation-test-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
