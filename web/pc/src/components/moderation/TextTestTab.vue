<template>
  <div class="text-test-tab">
    <el-form :model="form" label-width="100px">
      <el-form-item label="测试文本">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="8"
          placeholder="请输入要测试的文本内容（不超过500字）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="用户ID">
        <el-input
          v-model="form.userId"
          placeholder="可选，用于审计日志"
          clearable
        />
      </el-form-item>

      <el-form-item label="场景标识">
        <el-input
          v-model="form.scene"
          placeholder="可选，如 comment、post、chat"
          clearable
        />
      </el-form-item>

      <el-form-item>
        <el-button
          type="primary"
          :loading="loading"
          :disabled="!form.content.trim()"
          @click="handleSubmit"
        >
          开始检测
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <ModerationResult v-if="result" :result="result" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { checkText } from '@/api/moderation';
import ModerationResult from './ModerationResult.vue';
import type { TextModerationRequest, TextModerationResponse } from '@common/types/moderation';

const loading = ref(false);
const result = ref<TextModerationResponse | null>(null);

const form = reactive<TextModerationRequest>({
  content: '',
  userId: '',
  scene: ''
});

const handleSubmit = async () => {
  if (!form.content.trim()) {
    ElMessage.warning('请输入测试文本');
    return;
  }

  if (form.content.length > 500) {
    ElMessage.warning('文本长度不能超过500字');
    return;
  }

  loading.value = true;
  result.value = null;

  try {
    const response = await checkText({
      content: form.content,
      userId: form.userId || undefined,
      scene: form.scene || undefined
    });
    result.value = response;
    ElMessage.success('检测完成');
  } catch (error: any) {
    ElMessage.error(error.message || '检测失败，请稍后重试');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  form.content = '';
  form.userId = '';
  form.scene = '';
  result.value = null;
};
</script>

<style scoped>
.text-test-tab {
  padding: 20px;
}
</style>
