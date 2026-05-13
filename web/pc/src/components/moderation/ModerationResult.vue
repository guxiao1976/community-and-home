<template>
  <div class="moderation-result">
    <el-alert
      :type="getResultType()"
      :title="getResultTitle()"
      :closable="false"
      show-icon
    />

    <div class="result-info" v-if="result.risk_level">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="风险等级">
          <el-tag :type="getRiskLevelType(result.risk_level)">
            {{ getRiskLevelText(result.risk_level) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="需要人工复审" v-if="result.need_review !== undefined">
          <el-tag :type="result.need_review ? 'warning' : 'info'">
            {{ result.need_review ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="原因" :span="2" v-if="result.reason">
          {{ result.reason }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- 检测详情 -->
    <div class="detection-details" v-if="result.details && result.details.length > 0">
      <h4>检测详情</h4>
      <el-table :data="result.details" border style="margin-top: 12px;">
        <el-table-column prop="layer" label="检测层" width="150">
          <template #default="{ row }">
            <el-tag size="small">{{ getLayerName(row.layer) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="matched_text" label="匹配文本" width="150" />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="severity" label="严重程度" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.severity" :type="getSeverityType(row.severity)" size="small">
              {{ row.severity }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="confidence" label="置信度" width="120">
          <template #default="{ row }">
            <span v-if="row.confidence">{{ (row.confidence * 100).toFixed(1) }}%</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  result: any; // 使用any以兼容后端实际返回的结构
}

const props = defineProps<Props>();

const getResultType = () => {
  return props.result.pass ? 'success' : 'error';
};

const getResultTitle = () => {
  return props.result.pass ? '审核通过' : '审核未通过';
};

const getRiskLevelType = (level: string) => {
  const map: Record<string, any> = {
    'safe': 'success',
    'low': 'info',
    'medium': 'warning',
    'high': 'danger'
  };
  return map[level] || 'info';
};

const getRiskLevelText = (level: string) => {
  const map: Record<string, string> = {
    'safe': '安全',
    'low': '低风险',
    'medium': '中风险',
    'high': '高风险'
  };
  return map[level] || level;
};

const getLayerName = (layer: string) => {
  const map: Record<string, string> = {
    'ac_engine': 'AC引擎',
    'small_model': '小模型',
    'large_model': '大模型'
  };
  return map[layer] || layer;
};

const getSeverityType = (severity: number) => {
  if (severity === 1) return 'danger';
  if (severity === 2) return 'warning';
  if (severity === 3) return 'info';
  return 'info';
};
</script>

<style scoped>
.moderation-result {
  margin-top: 20px;
}

.result-info {
  margin-top: 16px;
}

.detection-details {
  margin-top: 20px;
}

.detection-details h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #303133;
}
</style>
