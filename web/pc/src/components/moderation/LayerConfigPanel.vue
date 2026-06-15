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
            <el-select v-model="localConfig.small_model_template_id" :disabled="!smallEnabled" placeholder="选择提示词模板" style="width: 100%;">
              <el-option v-for="t in templates" :key="t.template_id" :label="`${t.template_name} v${t.version}`" :value="t.template_id" />
            </el-select>
          </el-form-item>
          <el-form-item label="模型">
            <el-select v-model="localConfig.small_model_config_key" :disabled="!smallEnabled" placeholder="默认使用模板关联模型" clearable style="width: 100%;">
              <el-option v-for="m in models" :key="m.config_key" :label="m.display_name" :value="m.config_key" />
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
            <el-select v-model="localConfig.large_model_template_id" :disabled="!largeEnabled" placeholder="选择提示词模板" style="width: 100%;">
              <el-option v-for="t in templates" :key="t.template_id" :label="`${t.template_name} v${t.version}`" :value="t.template_id" />
            </el-select>
          </el-form-item>
          <el-form-item label="模型">
            <el-select v-model="localConfig.large_model_config_key" :disabled="!largeEnabled" placeholder="默认使用模板关联模型" clearable style="width: 100%;">
              <el-option v-for="m in models" :key="m.config_key" :label="m.display_name" :value="m.config_key" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue';
import { getModerationTemplates, getAvailableModels } from '@/api/aimodel';

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
const models = ref<any[]>([]);

const localConfig = reactive<LayerConfigValues>({ ...props.modelValue });
const smallEnabled = ref(props.modelValue.small_model_template_id !== '');
const largeEnabled = ref(props.modelValue.large_model_template_id !== '');

watch(localConfig, (v) => emit('update:modelValue', { ...v }), { deep: true });
watch(() => props.modelValue, (v) => {
  Object.assign(localConfig, v);
  smallEnabled.value = v.small_model_template_id !== '';
  largeEnabled.value = v.large_model_template_id !== '';
});

const onSmallToggle = (val: boolean) => { if (!val) localConfig.small_model_template_id = ''; };
const onLargeToggle = (val: boolean) => { if (!val) localConfig.large_model_template_id = ''; };

onMounted(async () => {
  try {
    const [tResp, mResp] = await Promise.all([getModerationTemplates(), getAvailableModels()]);
    templates.value = (tResp as any).list || [];
    models.value = (mResp as any).models || (mResp as any).list || [];
  } catch (e) { console.error('加载模板/模型列表失败', e); }
});
</script>

<style scoped>
.layer-config-panel { margin-bottom: 16px; }
.layer-header { display: flex; justify-content: space-between; align-items: center; }
</style>
