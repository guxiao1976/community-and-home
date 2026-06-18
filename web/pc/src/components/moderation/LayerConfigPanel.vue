<template>
  <el-row :gutter="16" class="layer-config-panel">
    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>AC 引擎</span>
            <el-switch v-model="localConfig.ac_enabled" :active-value="1" :inactive-value="0" />
          </div>
        </template>
        <el-form label-width="80px" size="small">
          <el-form-item label="严重度≥">
            <el-select v-model="localConfig.ac_severity_threshold" :disabled="!localConfig.ac_enabled" style="width: 100%;">
              <el-option :value="1" label="1 - 高危" />
              <el-option :value="2" label="2 - 中危" />
              <el-option :value="3" label="3 - 低危" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>小模型</span>
            <el-switch v-model="smallEnabled" @change="onSmallToggle" />
          </div>
        </template>
        <el-form label-width="60px" size="small">
          <el-form-item label="模板">
            <el-select v-model="localConfig.small_model_template_id" :disabled="!smallEnabled" placeholder="选择提示词模板（模型由模板关联）" style="width: 100%;">
              <el-option v-for="t in templates" :key="t.id" :label="`${t.name}（${t.model_name}）`" :value="String(t.id)" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>

    <el-col :span="8">
      <el-card shadow="never">
        <template #header>
          <div class="layer-header">
            <span>大模型</span>
            <el-switch v-model="largeEnabled" @change="onLargeToggle" />
          </div>
        </template>
        <el-form label-width="60px" size="small">
          <el-form-item label="模板">
            <el-select v-model="localConfig.large_model_template_id" :disabled="!largeEnabled" placeholder="选择提示词模板（模型由模板关联）" style="width: 100%;">
              <el-option v-for="t in templates" :key="t.id" :label="`${t.name}（${t.model_name}）`" :value="String(t.id)" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue';
import { getModerationTemplates } from '@/api/aimodel';

export interface LayerConfigValues {
  ac_enabled: number;
  ac_severity_threshold: number;
  small_model_template_id: string;
  small_model_config_key: string;
  large_model_template_id: string;
  large_model_config_key: string;
}

const props = defineProps<{ modelValue: LayerConfigValues }>();
const emit = defineEmits<{ (e: 'update:modelValue', v: LayerConfigValues): void }>();

const templates = ref<any[]>([]);

const localConfig = reactive<LayerConfigValues>({ ...props.modelValue });
const smallEnabled = ref(props.modelValue.small_model_template_id !== '');
const largeEnabled = ref(props.modelValue.large_model_template_id !== '');

watch(localConfig, (v) => emit('update:modelValue', { ...v }), { deep: true });
watch(() => props.modelValue, (v) => {
  Object.assign(localConfig, v);
  smallEnabled.value = v.small_model_template_id !== '';
  largeEnabled.value = v.large_model_template_id !== '';
});

const onSmallToggle = (val: boolean) => { if (!val) { localConfig.small_model_template_id = ''; localConfig.small_model_config_key = ''; } };
const onLargeToggle = (val: boolean) => { if (!val) { localConfig.large_model_template_id = ''; localConfig.large_model_config_key = ''; } };

onMounted(async () => {
  try {
    const tResp: any = await getModerationTemplates();
    // AI model service returns double-wrapped: {code:0, data:{code:0, data:{templates:[], total:N}}}
    // After axios interceptor strips one layer: {code:0, msg:"success", data:{templates:[], total:N}}
    templates.value = tResp?.data?.templates || tResp?.templates || [];
  } catch (e) { console.error('加载模板列表失败', e); }
});
</script>

<style scoped>
.layer-config-panel { margin-bottom: 16px; }
.layer-header { display: flex; justify-content: space-between; align-items: center; }
</style>
