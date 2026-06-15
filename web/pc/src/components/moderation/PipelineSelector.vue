<template>
  <div class="pipeline-selector">
    <el-select
      v-model="selectedId"
      placeholder="选择管线配置"
      :loading="loading"
      @change="handleSelect"
      style="width: 240px;"
    >
      <el-option
        v-for="p in pipelines"
        :key="p.pipeline_id"
        :label="`${p.pipeline_name}${p.is_production ? ' (生产中)' : ''}`"
        :value="p.pipeline_id"
      />
    </el-select>

    <el-button @click="handleNew" :icon="Plus">新建</el-button>
    <el-button @click="handleCopy" :disabled="!selectedId" :icon="CopyDocument">复制</el-button>
    <el-button type="warning" @click="handleActivate" :disabled="!selectedId" :icon="CircleCheck">设为生产</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, CopyDocument, CircleCheck } from '@element-plus/icons-vue';
import { listPipelines, createPipeline, activatePipeline, getPipeline } from '@/api/moderation';
import type { PipelineConfig } from '@common/types/moderation';

const emit = defineEmits<{
  (e: 'select', config: PipelineConfig): void;
  (e: 'new'): void;
}>();

const selectedId = ref('');
const pipelines = ref<PipelineConfig[]>([]);
const loading = ref(false);

const loadPipelines = async () => {
  loading.value = true;
  try {
    const resp = await listPipelines({ page: 1, page_size: 50 });
    pipelines.value = resp.list;
    const prod = resp.list.find(p => p.is_production === 1);
    if (prod && !selectedId.value) {
      selectedId.value = prod.pipeline_id;
      emit('select', prod);
    }
  } catch (e: any) {
    ElMessage.error(e.message || '加载管线列表失败');
  } finally {
    loading.value = false;
  }
};

const handleSelect = async (pipelineId: string) => {
  try {
    const config = await getPipeline(pipelineId);
    emit('select', config);
  } catch (e: any) {
    ElMessage.error(e.message || '加载管线配置失败');
  }
};

const handleNew = () => { emit('new'); };

const handleCopy = async () => {
  if (!selectedId.value) return;
  try {
    const config = await getPipeline(selectedId.value);
    const newId = `${config.pipeline_id}_copy_${Date.now()}`;
    await createPipeline({
      pipeline_id: newId,
      pipeline_name: `${config.pipeline_name} (副本)`,
      description: config.description,
      ac_enabled: config.ac_enabled,
      ac_severity_threshold: config.ac_severity_threshold,
      small_model_template_id: config.small_model_template_id,
      small_model_config_key: config.small_model_config_key,
      large_model_template_id: config.large_model_template_id,
      large_model_config_key: config.large_model_config_key,
      ac_to_small_condition: config.ac_to_small_condition,
      ac_to_small_severity: config.ac_to_small_severity,
      ac_to_small_categories: config.ac_to_small_categories,
      small_to_large_condition: config.small_to_large_condition,
      small_to_large_confidence_threshold: config.small_to_large_confidence_threshold,
      small_to_large_categories: config.small_to_large_categories,
      final_verdict_logic: config.final_verdict_logic,
    });
    ElMessage.success('配置复制成功');
    await loadPipelines();
    selectedId.value = newId;
    handleSelect(newId);
  } catch (e: any) {
    ElMessage.error(e.message || '复制失败');
  }
};

const handleActivate = async () => {
  if (!selectedId.value) return;
  try {
    await ElMessageBox.confirm(
      '将此管线配置设为生产环境默认配置？',
      '确认操作',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    );
    await activatePipeline(selectedId.value);
    ElMessage.success('已设为生产配置');
    await loadPipelines();
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败');
  }
};

const reset = () => { selectedId.value = ''; };

onMounted(loadPipelines);
defineExpose({ loadPipelines, reset });
</script>

<style scoped>
.pipeline-selector { display: flex; align-items: center; gap: 8px; margin-bottom: 16px; }
</style>
