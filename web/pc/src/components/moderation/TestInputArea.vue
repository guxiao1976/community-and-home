<template>
  <el-card shadow="never" class="test-input-area">
    <template #header><span>测试区域</span></template>
    <el-input v-model="content" type="textarea" :rows="6" placeholder="请输入要测试的文本内容（不超过500字）" maxlength="500" show-word-limit />
    <div class="test-actions">
      <el-button type="primary" :loading="loading" :disabled="!content.trim()" @click="handleTest">执行测试</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const emit = defineEmits<{ (e: 'test', content: string): void }>();
const content = ref('');
const loading = ref(false);

const handleTest = () => { emit('test', content.value); };
const handleReset = () => { content.value = ''; };
const setLoading = (val: boolean) => { loading.value = val; };

defineExpose({ setLoading });
</script>

<style scoped>
.test-input-area { margin-bottom: 16px; }
.test-actions { margin-top: 12px; display: flex; gap: 8px; }
</style>
