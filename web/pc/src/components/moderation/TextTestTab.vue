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

      <el-form-item label="审核模式">
        <el-radio-group v-model="form.check_mode">
          <el-radio value="ac_only">AC 引擎</el-radio>
          <el-radio value="model_only">AI 模型</el-radio>
          <el-radio value="combined">组合模式</el-radio>
        </el-radio-group>
        <div style="color: #909399; font-size: 12px; margin-top: 4px;">
          AC 引擎：仅敏感词匹配 | AI 模型：仅语义审核 | 组合模式：先AC后模型
        </div>
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

    <!-- 组合模式：显示AC和AI两个结果 -->
    <div v-if="form.check_mode === 'combined' && acResult && aiResult">
      <el-card class="json-result-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>AC 引擎结果</span>
            <el-button size="small" @click="copyJson(acResult)">复制</el-button>
          </div>
        </template>
        <pre class="json-content">{{ JSON.stringify(acResult, null, 2) }}</pre>
      </el-card>

      <el-card class="json-result-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>AI 模型结果</span>
            <el-button size="small" @click="copyJson(aiResult)">复制</el-button>
          </div>
        </template>
        <pre class="json-content">{{ JSON.stringify(aiResult, null, 2) }}</pre>
      </el-card>
    </div>

    <!-- 单一模式：显示一个结果 -->
    <el-card v-if="result && form.check_mode !== 'combined'" class="json-result-card" shadow="never">
      <template #header>
        <div class="card-header">
          <span>原始JSON响应</span>
          <el-button size="small" @click="copyJson(result)">复制</el-button>
        </div>
      </template>
      <pre class="json-content">{{ JSON.stringify(result, null, 2) }}</pre>
    </el-card>
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
const acResult = ref<TextModerationResponse | null>(null);
const aiResult = ref<TextModerationResponse | null>(null);

const form = reactive({
  content: '',
  check_mode: 'combined'
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
  acResult.value = null;
  aiResult.value = null;

  try {
    if (form.check_mode === 'combined') {
      // 组合模式：分别调用AC和AI
      const [acResponse, aiResponse] = await Promise.all([
        checkText({ content: form.content, check_mode: 'ac_only' }),
        checkText({ content: form.content, check_mode: 'model_only' })
      ]);

      acResult.value = acResponse;
      aiResult.value = aiResponse;

      // 最终结果：如果AC不通过就用AC结果，否则用AI结果
      result.value = acResponse.pass ? aiResponse : acResponse;
    } else {
      // 单一模式
      const response = await checkText({
        content: form.content,
        check_mode: form.check_mode
      });
      result.value = response;
    }

    ElMessage.success('检测完成');
  } catch (error: any) {
    ElMessage.error(error.message || '检测失败，请稍后重试');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  form.content = '';
  form.check_mode = 'combined';
  result.value = null;
  acResult.value = null;
  aiResult.value = null;
};

const copyJson = (data?: any) => {
  const jsonData = data || result.value;
  if (jsonData) {
    const jsonStr = JSON.stringify(jsonData, null, 2);
    navigator.clipboard.writeText(jsonStr).then(() => {
      ElMessage.success('JSON已复制到剪贴板');
    }).catch(() => {
      ElMessage.error('复制失败');
    });
  }
};
</script>

<style scoped>
.text-test-tab {
  padding: 20px;
}

.json-result-card {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.json-content {
  background-color: #f5f7fa;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-x: auto;
  margin: 0;
  color: #303133;
}
</style>
