<template>
  <div class="model-form">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-button @click="handleBack" link>
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <span>{{ isEdit ? '编辑模型' : '创建模型' }}</span>
        </div>
      </template>

      <el-form ref="formRef" :model="formData" :rules="rules" label-width="130px" v-loading="loading" autocomplete="off">
        <el-form-item label="标识符" prop="name">
          <el-input v-model="formData.name" placeholder="模型唯一名称，如 deepseek-chat" />
          <template #extra>
            <div style="color: #909399; font-size: 12px; margin-top: 2px;">
              model_name — 模型唯一标识，创建后不建议改
            </div>
          </template>
        </el-form-item>

        <el-form-item label="分组标识">
          <el-input v-model="formData.config_key" placeholder="同组模型共享，如 deepseek。模板通过此字段关联模型组" />
          <template #extra>
            <div style="color: #909399; font-size: 12px; margin-top: 2px;">
              config_key — 留空则默认与标识符相同。提示模板引用此字段选择模型组
            </div>
          </template>
        </el-form-item>

        <el-form-item label="显示名" prop="display_name">
          <el-input v-model="formData.display_name" placeholder="给人看的名字，如 Qwen2.5 3B 审核" />
        </el-form-item>

        <el-form-item label="提供商" prop="provider">
          <el-select v-model="formData.provider" placeholder="请选择提供商">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Claude" value="claude" />
            <el-option label="Ollama" value="ollama" />
          </el-select>
        </el-form-item>

        <el-form-item label="API端点" prop="endpoint">
          <div style="display: flex; gap: 8px; align-items: center;">
            <el-input v-model="formData.endpoint" placeholder="基础URL，如 https://api.deepseek.com 或 http://localhost:11434" style="flex: 1" />
            <el-button size="small" type="primary" :loading="fetchingModels" :disabled="!formData.endpoint" @click="fetchModels">
              获取模型列表
            </el-button>
          </div>
        </el-form-item>

        <!-- 已获取的模型列表（复选框，仅创建模式） -->
        <el-form-item v-if="availableModels.length > 0" label="选择模型">
          <div style="border: 1px solid #dcdfe6; border-radius: 4px; padding: 12px; max-height: 220px; overflow-y: auto;">
            <el-checkbox-group v-model="selectedModels" :max="4">
              <div v-for="m in availableModels" :key="m.model_name" style="margin-bottom: 6px;">
                <el-checkbox :value="m.model_name" :label="m.model_name">
                  {{ m.display_name || m.model_name }}
                </el-checkbox>
              </div>
            </el-checkbox-group>
          </div>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            已选 {{ selectedModels.length }}/4。保存时每个模型创建一条配置，共享标识符和端点。
          </div>
        </el-form-item>

        <el-form-item v-if="availableModels.length === 0 && !fetchingModels && formData.endpoint && formData.provider && !isEdit" label="">
          <span style="color: #909399; font-size: 12px;">
            点击"获取模型列表"从供应商拉取可用模型
          </span>
        </el-form-item>

        <el-form-item label="部署方式" prop="model_type">
          <el-select v-model="formData.model_type" placeholder="请选择部署方式">
            <el-option label="云端模型" value="cloud" />
            <el-option label="本地模型" value="local" />
          </el-select>
        </el-form-item>

        <el-form-item label="最大Token" prop="max_tokens">
          <el-input-number v-model="formData.max_tokens" :min="1" :max="200000" :step="1000" />
        </el-form-item>

        <el-form-item label="温度" prop="temperature">
          <el-input-number v-model="formData.temperature" :min="0" :max="2" :precision="1" :step="0.1" />
        </el-form-item>

        <!-- API 密钥：创建必填，编辑选填（留空不修改） -->
        <el-form-item label="API 密钥" prop="api_key" :required="!isEdit && formData.model_type === 'cloud'">
          <el-input v-model="formData.api_key" type="password" show-password autocomplete="new-password"
            :placeholder="isEdit
              ? hasExistingKey ? '留空不修改；填写则替换现有密钥' : '填写以添加 API 密钥'
              : formData.model_type === 'cloud' ? '云端模型必须填写 API Key（如 sk-xxxx）' : '本地模型可跳过（如 Ollama）'" />
          <template #extra>
            <div v-if="!isEdit && formData.model_type === 'cloud'" style="color: #e6a23c; font-size: 12px; margin-top: 2px;">
              ⚠ 云端模型必须填写 API 密钥
            </div>
          </template>
        </el-form-item>

        <el-form-item label="连接测试">
          <el-button @click="handleTestConnection" :loading="testing" :disabled="!canTest">测试连接</el-button>
          <span v-if="testResult" :style="{ color: testResult.success ? '#67C23A' : '#F56C6C', marginLeft: '12px' }">
            {{ testResult.success ? `✅ 成功 (${testResult.latency_ms}ms)` : `❌ ${testResult.error || '连接失败'}` }}
          </span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
          <el-button @click="handleBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable */
import { ref, reactive, computed, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { ArrowLeft } from '@element-plus/icons-vue';
import { getModelConfigById, createModelConfig, updateModelConfig, getApiKeys, createApiKey, updateApiKey, testModelConnection, fetchProviderModels } from '@/api/aimodel';

const router = useRouter(); const route = useRoute(); const formRef = ref<FormInstance>();
const loading = ref(false); const submitting = ref(false);
const testing = ref(false);
const testResult = ref<{ success: boolean; latency_ms: number; error?: string } | null>(null);
const hasExistingKey = ref(false);
const existingKeyId = ref<string>('');

const canTest = computed(() => {
  if (!formData.endpoint) return false;
  if (formData.model_type === 'local') return true;
  // 编辑模式：有现有密钥或在表单中填了密钥
  if (isEdit.value) return !!formData.api_key || hasExistingKey.value;
  return !!formData.api_key;
});

const fetchingModels = ref(false);
const availableModels = ref<Array<{ model_name: string; display_name: string }>>([]);
const selectedModels = ref<string[]>([]);

const fetchModels = async () => {
  if (!formData.endpoint) return;
  if (formData.model_type === 'cloud' && !formData.api_key && !hasExistingKey.value) {
    ElMessage.warning('云端模型需要先填写 API 密钥');
    return;
  }
  fetchingModels.value = true; availableModels.value = []; selectedModels.value = [];
  try {
    const apiKey = formData.api_key || '';  // 编辑模式可能用已有密钥
    const res = await fetchProviderModels({ provider: formData.provider || 'ollama', endpoint: formData.endpoint, api_key: apiKey });
    const inner = (res as any)?.data || res;
    const models = Array.isArray(inner) ? inner : (inner?.data || inner?.models || []);
    availableModels.value = models;
    ElMessage.success(`获取到 ${models.length} 个模型`);
  } catch (e: any) { ElMessage.error('获取模型列表失败: ' + (e.message || '网络错误')); }
  finally { fetchingModels.value = false; }
};

const isEdit = ref(false); const modelId = ref<string>('');
const formData = reactive({ name: '', config_key: '', display_name: '', provider: '', type: 'chat', model_type: 'cloud', endpoint: '', max_tokens: 4096, temperature: 0.7, supported_features: 'streaming', cost_per_1k_input_tokens: 0, cost_per_1k_output_tokens: 0, description: '', api_key: '' });

const rules: FormRules = {
  name: [{ required: true, message: '请输入标识符', trigger: 'blur' }],
  display_name: [{ required: true, message: '请输入显示名', trigger: 'blur' }],
  provider: [{ required: true, message: '请选择提供商', trigger: 'change' }],
  type: [{ required: true, message: '请选择模型类型', trigger: 'change' }],
  max_tokens: [{ required: true, message: '请输入最大Token数', trigger: 'blur' }],
  supported_features: [{ required: true, message: '请输入支持的特性', trigger: 'blur' }],
  cost_per_1k_input_tokens: [{ required: true, message: '请输入输入成本', trigger: 'blur' }],
  cost_per_1k_output_tokens: [{ required: true, message: '请输入输出成本', trigger: 'blur' }],
  api_key: [{
    validator: (_rule: any, value: string, callback: any) => {
      // 编辑模式不强制要求填写（可留空保留现有密钥）
      if (isEdit.value) { callback(); return; }
      if (formData.model_type === 'cloud' && !value) {
        callback(new Error('云端模型必须填写 API 密钥'));
      } else {
        callback();
      }
    },
    trigger: 'blur'
  }]
};

const fetchModelData = async () => {
  if (!modelId.value) return; loading.value = true;
  try {
    const res = await getModelConfigById(modelId.value); const model = (res as any)?.data || res;
    Object.assign(formData, { name: model.name, config_key: model.config_key || '', display_name: model.display_name, provider: model.provider, type: model.type, model_type: (model.model_type === 'cloud' || model.model_type === 'local') ? model.model_type : 'cloud', endpoint: model.endpoint || '', max_tokens: model.max_tokens, temperature: model.temperature ?? 0.7, supported_features: model.supported_features, cost_per_1k_input_tokens: model.cost_per_1k_input_tokens, cost_per_1k_output_tokens: model.cost_per_1k_output_tokens, description: model.description || '' });

    // 获取现有密钥信息（仅用于判断是否存在和 canTest）
    try {
      const keyRes = await getApiKeys({ model_id: modelId.value } as any);
      const keyData = (keyRes as any)?.data || keyRes;
      const keys = keyData.keys || [];
      if (keys.length > 0) {
        hasExistingKey.value = true;
        existingKeyId.value = keys[0].id;
      }
    } catch { /* 密钥查询失败不影响模型信息展示 */ }
  } catch (error) { ElMessage.error('获取模型信息失败'); } finally { loading.value = false; }
};

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return;
    submitting.value = true;
    try {
      const baseData = {
        name: formData.name, config_key: formData.config_key || undefined, display_name: formData.display_name, provider: formData.provider,
        type: formData.type, model_type: formData.model_type, endpoint: formData.endpoint || undefined,
        max_tokens: formData.max_tokens, temperature: formData.temperature,
        supported_features: formData.supported_features,
        cost_per_1k_input_tokens: formData.cost_per_1k_input_tokens,
        cost_per_1k_output_tokens: formData.cost_per_1k_output_tokens,
        description: formData.description || undefined
      };

      if (isEdit.value && modelId.value) {
        await updateModelConfig({ id: modelId.value, ...baseData });

        // 编辑模式：如果填了 API 密钥，更新或创建
        if (formData.api_key?.trim()) {
          try {
            if (existingKeyId.value) {
              await updateApiKey({ id: existingKeyId.value, api_key: formData.api_key });
            } else {
              const keyRes = await createApiKey({
                model_id: modelId.value,
                key_name: `${formData.provider}-${formData.name}-默认`,
                api_key: formData.api_key,
                description: '编辑模型时添加'
              });
              existingKeyId.value = (keyRes as any)?.id || '';
              hasExistingKey.value = true;
            }
          } catch (keyErr: any) {
            console.error('API 密钥更新失败:', keyErr);
            ElMessage.warning('模型信息已更新，但密钥保存失败');
          }
        }
        ElMessage.success('更新成功');
      } else {
        // 支持批量创建：如果有勾选模型，每个模型创建一条
        const modelNames = selectedModels.value.length > 0 ? selectedModels.value : [formData.name];
        let created = 0;
        let keyErrors = 0;
        for (const modelName of modelNames) {
          try {
            const res = await createModelConfig({ ...baseData, name: modelName });
            created++;

            // 创建成功后，如果有 API 密钥，自动创建密钥记录
            if (formData.api_key?.trim() && res) {
              const modelId = res?.id;
              if (modelId) {
                try {
                  await createApiKey({
                    model_id: modelId,
                    key_name: `${formData.provider || modelName}-${modelName}-默认`,
                    api_key: formData.api_key,
                    description: `创建模型 ${modelName} 时自动添加`
                  });
                } catch (keyErr: any) {
                  keyErrors++;
                  console.error(`为模型 ${modelName} 创建 API 密钥失败:`, keyErr);
                }
              }
            }
          } catch (e: any) {
            if (e?.message?.includes('已存在') || e?.code === 30400) {
              console.warn(`模型 ${modelName} 已存在，跳过`);
            } else {
              throw e;
            }
          }
        }
        const msg = keyErrors > 0
          ? `创建成功 (${created}/${modelNames.length})，但 ${keyErrors} 个密钥保存失败，请在编辑页手动添加`
          : `创建成功 (${created}/${modelNames.length})`;
        ElMessage.success(msg);
      }
      router.push('/aimodel/models');
    } catch (error) { ElMessage.error(isEdit.value ? '更新失败' : '创建失败'); console.error(error); }
    finally { submitting.value = false; }
  });
};

async function handleTestConnection() {
  if (!canTest.value) return;
  testing.value = true; testResult.value = null;
  try {
    // 编辑模式下，如果表单没填密钥，测试连接仍可执行（后端会从 DB 取现有密钥）
    let apiKey = formData.api_key;
    const res = await testModelConnection({ config_key: formData.name, endpoint: formData.endpoint, api_key: apiKey, provider: formData.provider || undefined, model_name: formData.name });
    const inner = (res as any)?.data || res;
    testResult.value = { success: inner.success ?? false, latency_ms: inner.latency_ms ?? 0, error: inner.error || '' };
  } catch (e: any) { testResult.value = { success: false, latency_ms: 0, error: e?.message || String(e) }; }
  finally { testing.value = false; }
}

// 表单初始值（用于重置）
const INITIAL_FORM = { name: '', config_key: '', display_name: '', provider: '', type: 'chat', model_type: 'cloud', endpoint: '', max_tokens: 4096, temperature: 0.7, supported_features: 'streaming', cost_per_1k_input_tokens: 0, cost_per_1k_output_tokens: 0, description: '', api_key: '' };

const resetForm = () => {
  Object.assign(formData, INITIAL_FORM);
  hasExistingKey.value = false;
  existingKeyId.value = '';
  availableModels.value = [];
  selectedModels.value = [];
  testResult.value = null;
};

// 路由变化时重置状态（处理编辑→创建的组件复用 + 初始挂载）
watch(() => route.params.id, (newId) => {
  if (!newId) {
    // 进入创建模式：完全重置
    isEdit.value = false;
    modelId.value = '';
    resetForm();
  } else if (newId !== modelId.value) {
    // 切换到不同模型的编辑模式
    isEdit.value = true;
    modelId.value = newId as string;
    resetForm();
    fetchModelData();
  }
}, { immediate: true });

const handleBack = () => { router.back(); };
</script>

<style scoped lang="scss">
.model-form { padding: 20px; .card-header { display: flex; justify-content: space-between; align-items: center; } .el-form { max-width: 650px; } }
</style>
