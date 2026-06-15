<template>
  <div class="config-test-page">
    <el-card>
      <template #header>
        <div class="page-header">
          <span>内容审核配置测试</span>
          <el-button type="primary" :loading="saving" @click="handleSave">保存配置</el-button>
        </div>
      </template>

      <PipelineSelector ref="selectorRef" @select="onPipelineSelect" @new="onNewPipeline" />
      <LayerConfigPanel v-model="layerConfig" />
      <EscalationRuleEditor v-model="escalationRules" />
      <TestInputArea ref="testInputRef" @test="onTest" />
      <PipelineResultPanel :result="testResult" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import PipelineSelector from '@/components/moderation/PipelineSelector.vue';
import LayerConfigPanel from '@/components/moderation/LayerConfigPanel.vue';
import type { LayerConfigValues } from '@/components/moderation/LayerConfigPanel.vue';
import EscalationRuleEditor from '@/components/moderation/EscalationRuleEditor.vue';
import type { EscalationRuleValues } from '@/components/moderation/EscalationRuleEditor.vue';
import TestInputArea from '@/components/moderation/TestInputArea.vue';
import PipelineResultPanel from '@/components/moderation/PipelineResultPanel.vue';
import { testPipeline, createPipeline, updatePipeline, getPipeline } from '@/api/moderation';
import type { PipelineConfig, PipelineTestResponse } from '@common/types/moderation';

const saving = ref(false);
const testResult = ref<PipelineTestResponse | null>(null);
const selectorRef = ref<InstanceType<typeof PipelineSelector>>();
const testInputRef = ref<InstanceType<typeof TestInputArea>>();
const currentPipelineId = ref('');

const layerConfig = reactive<LayerConfigValues>({
  ac_enabled: 1, ac_severity_threshold: 1,
  small_model_template_id: '', small_model_config_key: '',
  large_model_template_id: '', large_model_config_key: '',
});

const escalationRules = reactive<EscalationRuleValues>({
  ac_to_small_condition: 'any_hit', ac_to_small_severity: 1, ac_to_small_categories: [],
  small_to_large_condition: 'confidence_lt', small_to_large_confidence_threshold: 0.90, small_to_large_categories: [],
  final_verdict_logic: 'last_model_wins',
});

const onPipelineSelect = (config: PipelineConfig) => {
  currentPipelineId.value = config.pipeline_id;
  Object.assign(layerConfig, {
    ac_enabled: config.ac_enabled, ac_severity_threshold: config.ac_severity_threshold,
    small_model_template_id: config.small_model_template_id, small_model_config_key: config.small_model_config_key,
    large_model_template_id: config.large_model_template_id, large_model_config_key: config.large_model_config_key,
  });
  Object.assign(escalationRules, {
    ac_to_small_condition: config.ac_to_small_condition, ac_to_small_severity: config.ac_to_small_severity,
    ac_to_small_categories: config.ac_to_small_categories || [],
    small_to_large_condition: config.small_to_large_condition,
    small_to_large_confidence_threshold: config.small_to_large_confidence_threshold,
    small_to_large_categories: config.small_to_large_categories || [],
    final_verdict_logic: config.final_verdict_logic,
  });
};

const onNewPipeline = () => {
  currentPipelineId.value = '';
  Object.assign(layerConfig, { ac_enabled: 1, ac_severity_threshold: 1, small_model_template_id: '', small_model_config_key: '', large_model_template_id: '', large_model_config_key: '' });
  Object.assign(escalationRules, { ac_to_small_condition: 'any_hit', ac_to_small_severity: 1, ac_to_small_categories: [], small_to_large_condition: 'confidence_lt', small_to_large_confidence_threshold: 0.90, small_to_large_categories: [], final_verdict_logic: 'last_model_wins' });
};

const onTest = async (content: string) => {
  testInputRef.value?.setLoading(true);
  testResult.value = null;
  try {
    const resp = await testPipeline({
      content,
      pipeline_id: currentPipelineId.value || undefined,
      small_model_template_id: layerConfig.small_model_template_id || undefined,
      large_model_template_id: layerConfig.large_model_template_id || undefined,
      small_to_large_confidence: escalationRules.small_to_large_confidence_threshold,
    });
    testResult.value = resp;
    ElMessage.success('测试完成');
  } catch (e: any) {
    ElMessage.error(e.message || '测试失败');
  } finally {
    testInputRef.value?.setLoading(false);
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    if (!currentPipelineId.value) {
      const newId = `pipeline_${Date.now()}`;
      await createPipeline({
        pipeline_id: newId,
        pipeline_name: `管线配置_${new Date().toLocaleDateString()}`,
        ac_enabled: layerConfig.ac_enabled,
        ac_severity_threshold: layerConfig.ac_severity_threshold,
        small_model_template_id: layerConfig.small_model_template_id,
        small_model_config_key: layerConfig.small_model_config_key,
        large_model_template_id: layerConfig.large_model_template_id,
        large_model_config_key: layerConfig.large_model_config_key,
        ac_to_small_condition: escalationRules.ac_to_small_condition,
        ac_to_small_severity: escalationRules.ac_to_small_severity,
        ac_to_small_categories: escalationRules.ac_to_small_categories,
        small_to_large_condition: escalationRules.small_to_large_condition,
        small_to_large_confidence_threshold: escalationRules.small_to_large_confidence_threshold,
        small_to_large_categories: escalationRules.small_to_large_categories,
        final_verdict_logic: escalationRules.final_verdict_logic,
      });
      currentPipelineId.value = newId;
      ElMessage.success('配置已保存');
      selectorRef.value?.loadPipelines();
    } else {
      const existing = await getPipeline(currentPipelineId.value);
      await updatePipeline({
        id: existing.id,
        pipeline_name: existing.pipeline_name,
        ac_enabled: layerConfig.ac_enabled,
        ac_severity_threshold: layerConfig.ac_severity_threshold,
        small_model_template_id: layerConfig.small_model_template_id,
        small_model_config_key: layerConfig.small_model_config_key,
        large_model_template_id: layerConfig.large_model_template_id,
        large_model_config_key: layerConfig.large_model_config_key,
        ac_to_small_condition: escalationRules.ac_to_small_condition,
        ac_to_small_severity: escalationRules.ac_to_small_severity,
        ac_to_small_categories: escalationRules.ac_to_small_categories,
        small_to_large_condition: escalationRules.small_to_large_condition,
        small_to_large_confidence_threshold: escalationRules.small_to_large_confidence_threshold,
        small_to_large_categories: escalationRules.small_to_large_categories,
        final_verdict_logic: escalationRules.final_verdict_logic,
      });
      ElMessage.success('配置已更新');
      selectorRef.value?.loadPipelines();
    }
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败');
  } finally {
    saving.value = false;
  }
};
</script>

<style scoped>
.config-test-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
</style>
