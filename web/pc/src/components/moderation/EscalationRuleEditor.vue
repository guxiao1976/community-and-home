<template>
  <el-card shadow="never" class="escalation-editor">
    <template #header><span>升级规则</span></template>

    <div class="rule-row">
      <span class="rule-label">AC → 小模型：</span>
      <el-select v-model="localRules.ac_to_small_condition" style="width: 140px;" @change="onAcConditionChange">
        <el-option value="any_hit" label="任何命中" />
        <el-option value="severity_gte" label="严重度 ≥" />
        <el-option value="category_in" label="分类包含" />
        <el-option value="never" label="从不" />
      </el-select>
      <template v-if="localRules.ac_to_small_condition === 'severity_gte'">
        <el-input-number v-model="localRules.ac_to_small_severity" :min="1" :max="3" style="width: 80px; margin-left: 8px;" />
      </template>
      <template v-if="localRules.ac_to_small_condition === 'category_in'">
        <el-select v-model="localRules.ac_to_small_categories" multiple filterable allow-create placeholder="输入分类名" style="width: 240px; margin-left: 8px;" />
      </template>
    </div>

    <div class="rule-row">
      <span class="rule-label">小模型 → 大模型：</span>
      <el-select v-model="localRules.small_to_large_condition" style="width: 140px;" @change="onSlConditionChange">
        <el-option value="confidence_lt" label="置信度 &lt;" />
        <el-option value="category_in" label="分类包含" />
        <el-option value="always" label="总是" />
        <el-option value="never" label="从不" />
      </el-select>
      <template v-if="localRules.small_to_large_condition === 'confidence_lt'">
        <el-input-number v-model="localRules.small_to_large_confidence_threshold" :min="0" :max="1" :step="0.05" :precision="2" style="width: 100px; margin-left: 8px;" />
      </template>
      <template v-if="localRules.small_to_large_condition === 'category_in'">
        <el-select v-model="localRules.small_to_large_categories" multiple filterable allow-create placeholder="输入分类名" style="width: 240px; margin-left: 8px;" />
      </template>
    </div>

    <div class="rule-row">
      <span class="rule-label">终判逻辑：</span>
      <el-select v-model="localRules.final_verdict_logic" style="width: 180px;">
        <el-option value="last_model_wins" label="最后模型判定" />
        <el-option value="large_overrides" label="大模型覆盖" />
        <el-option value="small_overrides" label="小模型覆盖" />
      </el-select>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue';

export interface EscalationRuleValues {
  ac_to_small_condition: string;
  ac_to_small_severity: number;
  ac_to_small_categories: string[];
  small_to_large_condition: string;
  small_to_large_confidence_threshold: number;
  small_to_large_categories: string[];
  final_verdict_logic: string;
}

const props = defineProps<{ modelValue: EscalationRuleValues }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: EscalationRuleValues): void }>();

const localRules = reactive<EscalationRuleValues>({
  ac_to_small_condition: props.modelValue.ac_to_small_condition || 'any_hit',
  ac_to_small_severity: props.modelValue.ac_to_small_severity || 1,
  ac_to_small_categories: props.modelValue.ac_to_small_categories || [],
  small_to_large_condition: props.modelValue.small_to_large_condition || 'confidence_lt',
  small_to_large_confidence_threshold: props.modelValue.small_to_large_confidence_threshold || 0.90,
  small_to_large_categories: props.modelValue.small_to_large_categories || [],
  final_verdict_logic: props.modelValue.final_verdict_logic || 'last_model_wins',
});

watch(localRules, (v) => emit('update:modelValue', { ...v }), { deep: true });
watch(() => props.modelValue, (v) => {
  Object.assign(localRules, {
    ac_to_small_condition: v.ac_to_small_condition || 'any_hit',
    ac_to_small_severity: v.ac_to_small_severity || 1,
    ac_to_small_categories: v.ac_to_small_categories || [],
    small_to_large_condition: v.small_to_large_condition || 'confidence_lt',
    small_to_large_confidence_threshold: v.small_to_large_confidence_threshold || 0.90,
    small_to_large_categories: v.small_to_large_categories || [],
    final_verdict_logic: v.final_verdict_logic || 'last_model_wins',
  });
});

const onAcConditionChange = () => {
  if (localRules.ac_to_small_condition !== 'severity_gte') localRules.ac_to_small_severity = 1;
  if (localRules.ac_to_small_condition !== 'category_in') localRules.ac_to_small_categories = [];
};
const onSlConditionChange = () => {
  if (localRules.small_to_large_condition !== 'confidence_lt') localRules.small_to_large_confidence_threshold = 0.90;
  if (localRules.small_to_large_condition !== 'category_in') localRules.small_to_large_categories = [];
};
</script>

<style scoped>
.escalation-editor { margin-bottom: 16px; }
.rule-row { display: flex; align-items: center; margin-bottom: 14px; }
.rule-label { width: 140px; color: #606266; font-size: 14px; flex-shrink: 0; }
</style>
