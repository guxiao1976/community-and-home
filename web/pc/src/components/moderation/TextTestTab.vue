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
          <el-radio value="small_model">小模型</el-radio>
          <el-radio value="large_model">大模型</el-radio>
          <el-radio value="combined">组合模式</el-radio>
        </el-radio-group>
        <div style="color: #909399; font-size: 12px; margin-top: 4px;">
          AC 引擎：敏感词匹配 | 小模型：语义审核 | 大模型：深度审核 | 组合模式：AC→小模型→大模型 级联
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

    <!-- 执行总览表 -->
    <div v-if="showOverview" class="overview-section">
      <h4>审核层执行总览</h4>
      <el-table :data="overviewRows" border style="margin-top: 12px;">
        <el-table-column prop="layer" label="审核层" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.label }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="called" label="调用情况" width="100">
          <template #default="{ row }">
            <el-tag :type="row.called ? 'success' : 'info'" size="small">
              {{ row.called ? '已调用' : '未调用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="passed" label="结果" width="100">
          <template #default="{ row }">
            <template v-if="row.called">
              <el-tag :type="row.passed ? 'success' : 'danger'" size="small">
                {{ row.passed ? '通过' : '未通过' }}
              </el-tag>
            </template>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="confidence" label="置信度" width="100">
          <template #default="{ row }">
            <span v-if="row.called && row.confidence !== undefined" style="font-weight: 500;">
              {{ row.confidence }}%
            </span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情">
          <template #default="{ row }">
            <span v-if="row.detail" style="color: #909399; font-size: 13px;">{{ row.detail }}</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 各层原始 JSON（组合模式） -->
    <template v-if="form.check_mode === 'combined'">
      <el-card v-if="acResult" class="json-result-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>AC 引擎结果</span>
            <el-button size="small" @click="copyJson(acResult)">复制</el-button>
          </div>
        </template>
        <pre class="json-content">{{ JSON.stringify(acResult, null, 2) }}</pre>
      </el-card>

      <el-card v-if="smallModelResult" class="json-result-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>小模型结果</span>
            <el-button size="small" @click="copyJson(smallModelResult)">复制</el-button>
          </div>
        </template>
        <pre class="json-content">{{ JSON.stringify(smallModelResult, null, 2) }}</pre>
      </el-card>

      <el-card v-if="largeModelResult" class="json-result-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>大模型结果</span>
            <el-button size="small" @click="copyJson(largeModelResult)">复制</el-button>
          </div>
        </template>
        <pre class="json-content">{{ JSON.stringify(largeModelResult, null, 2) }}</pre>
      </el-card>
    </template>

    <!-- 单一模式：显示原始 JSON -->
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
import { ref, reactive, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { checkText } from '@/api/moderation';
import type { TextModerationResponse } from '@common/types/moderation';

const loading = ref(false);
const result = ref<TextModerationResponse | null>(null);

// 各层独立结果（组合模式使用）
const acResult = ref<TextModerationResponse | null>(null);
const smallModelResult = ref<TextModerationResponse | null>(null);
const largeModelResult = ref<TextModerationResponse | null>(null);

// 追踪各层调用状态和结果
const acCalled = ref(false);
const smallModelCalled = ref(false);
const largeModelCalled = ref(false);
const acPassed = ref(false);
const smallModelPassed = ref(false);
const largeModelPassed = ref(false);
const acConfidence = ref<number | undefined>(undefined);
const smallModelConfidence = ref<number | undefined>(undefined);
const largeModelConfidence = ref<number | undefined>(undefined);

const form = reactive({
  content: '',
  check_mode: 'combined' as 'ac_only' | 'small_model' | 'large_model' | 'combined'
});

const showOverview = computed(() => result.value !== null);

interface OverviewRow {
  layer: string;
  label: string;
  tagType: string;
  called: boolean;
  passed: boolean;
  confidence?: number;
  detail: string;
}

/** 从响应中提取置信度百分比（0-100 整数） */
const extractConfidence = (resp: TextModerationResponse, mode: 'ac' | 'small' | 'large'): number | undefined => {
  if (mode === 'ac') {
    return resp.traditional?.score != null ? Math.round(resp.traditional.score * 100) : undefined;
  }
  if (mode === 'small') {
    return resp.smallModel?.confidence != null ? Math.round(resp.smallModel.confidence * 100) : undefined;
  }
  // large
  return resp.largeModel?.confidence != null ? Math.round(resp.largeModel.confidence * 100) : undefined;
};

const overviewRows = computed<OverviewRow[]>(() => {
  const mode = form.check_mode;

  if (mode === 'combined') {
    return [
      {
        layer: 'ac',
        label: 'AC引擎',
        tagType: 'ac',
        called: acCalled.value,
        passed: acPassed.value,
        confidence: acConfidence.value,
        detail: acCalled.value ? (acPassed.value ? '' : '命中敏感词') : ''
      },
      {
        layer: 'small',
        label: '小模型',
        tagType: 'small',
        called: smallModelCalled.value,
        passed: smallModelPassed.value,
        confidence: smallModelConfidence.value,
        detail: smallModelCalled.value ? (smallModelPassed.value ? '' : '语义审核未通过') : (acPassed.value ? 'AC已通过，跳过' : '')
      },
      {
        layer: 'large',
        label: '大模型',
        tagType: 'large',
        called: largeModelCalled.value,
        passed: largeModelPassed.value,
        confidence: largeModelConfidence.value,
        detail: largeModelCalled.value
          ? (largeModelPassed.value ? '' : '深度审核未通过，交人工审核')
          : (acPassed.value ? 'AC已通过，跳过' : (smallModelPassed.value ? '小模型已通过，跳过' : ''))
      }
    ];
  }

  return [
    {
      layer: 'ac',
      label: 'AC引擎',
      tagType: 'ac',
      called: mode === 'ac_only',
      passed: mode === 'ac_only' ? result.value?.finalResult ?? false : false,
      confidence: mode === 'ac_only' ? extractConfidence(result.value!, 'ac') : undefined,
      detail: mode === 'ac_only' ? (result.value?.finalResult ? '' : '命中敏感词') : ''
    },
    {
      layer: 'small',
      label: '小模型',
      tagType: 'small',
      called: mode === 'small_model',
      passed: mode === 'small_model' ? result.value?.finalResult ?? false : false,
      confidence: mode === 'small_model' ? extractConfidence(result.value!, 'small') : undefined,
      detail: mode === 'small_model' ? (result.value?.finalResult ? '' : '语义审核未通过') : ''
    },
    {
      layer: 'large',
      label: '大模型',
      tagType: 'large',
      called: mode === 'large_model',
      passed: mode === 'large_model' ? result.value?.finalResult ?? false : false,
      confidence: mode === 'large_model' ? extractConfidence(result.value!, 'large') : undefined,
      detail: mode === 'large_model' ? (result.value?.finalResult ? '' : '深度审核未通过，交人工审核') : ''
    }
  ];
});

const callCheck = async (mode: 'ac_only' | 'small_model' | 'large_model'): Promise<TextModerationResponse> => {
  return checkText({
    content: form.content,
    check_mode: mode
  });
};

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
  smallModelResult.value = null;
  largeModelResult.value = null;
  acCalled.value = false;
  smallModelCalled.value = false;
  largeModelCalled.value = false;
  acPassed.value = false;
  smallModelPassed.value = false;
  largeModelPassed.value = false;
  acConfidence.value = undefined;
  smallModelConfidence.value = undefined;
  largeModelConfidence.value = undefined;

  try {
    if (form.check_mode === 'combined') {
      // 组合模式：级联调用 AC → 小模型 → 大模型
      const acResp = await callCheck('ac_only');
      acResult.value = acResp;
      acCalled.value = true;
      acPassed.value = acResp.finalResult;
      acConfidence.value = extractConfidence(acResp, 'ac');

      if (acResp.finalResult) {
        result.value = acResp;
        ElMessage.success('AC 引擎审核通过，内容放行');
        return;
      }

      const smallResp = await callCheck('small_model');
      smallModelResult.value = smallResp;
      smallModelCalled.value = true;
      smallModelPassed.value = smallResp.finalResult;
      smallModelConfidence.value = extractConfidence(smallResp, 'small');

      if (smallResp.finalResult) {
        result.value = {
          ...smallResp,
          traditional: acResp.traditional,
          largeModel: undefined
        };
        ElMessage.success('小模型判断合规，内容放行');
        return;
      }

      const largeResp = await callCheck('large_model');
      largeModelResult.value = largeResp;
      largeModelCalled.value = true;
      largeModelPassed.value = largeResp.finalResult;
      largeModelConfidence.value = extractConfidence(largeResp, 'large');

      if (largeResp.finalResult) {
        result.value = {
          ...largeResp,
          traditional: acResp.traditional,
          smallModel: smallResp.smallModel,
        };
        ElMessage.success('大模型判断合规，内容放行');
        return;
      }

      result.value = {
        requestId: largeResp.requestId,
        finalResult: false,
        traditional: acResp.traditional,
        smallModel: smallResp.smallModel,
        largeModel: largeResp.largeModel,
        processingTime: acResp.processingTime + smallResp.processingTime + largeResp.processingTime
      };
      ElMessage.warning('所有审核层未通过，内容需交人工审核');
    } else {
      const response = await callCheck(form.check_mode);
      result.value = response;

      if (form.check_mode === 'ac_only') {
        acCalled.value = true;
        acPassed.value = response.finalResult;
        acConfidence.value = extractConfidence(response, 'ac');
      } else if (form.check_mode === 'small_model') {
        smallModelCalled.value = true;
        smallModelPassed.value = response.finalResult;
        smallModelConfidence.value = extractConfidence(response, 'small');
      } else if (form.check_mode === 'large_model') {
        largeModelCalled.value = true;
        largeModelPassed.value = response.finalResult;
        largeModelConfidence.value = extractConfidence(response, 'large');
      }

      ElMessage.success('检测完成');
    }
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
  smallModelResult.value = null;
  largeModelResult.value = null;
  acCalled.value = false;
  smallModelCalled.value = false;
  largeModelCalled.value = false;
  acPassed.value = false;
  smallModelPassed.value = false;
  largeModelPassed.value = false;
  acConfidence.value = undefined;
  smallModelConfidence.value = undefined;
  largeModelConfidence.value = undefined;
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

.overview-section {
  margin-top: 24px;
}

.overview-section h4 {
  margin: 0 0 4px 0;
  font-size: 16px;
  color: #303133;
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
