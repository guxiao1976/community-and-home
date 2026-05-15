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

      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="120px"
        v-loading="loading"
      >
        <el-form-item label="模型名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: gpt-4, claude-opus-4" />
        </el-form-item>

        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="formData.display_name" placeholder="例如: GPT-4" />
        </el-form-item>

        <el-form-item label="提供商" prop="provider">
          <el-select v-model="formData.provider" placeholder="请选择提供商">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Claude" value="claude" />
            <el-option label="Ollama" value="ollama" />
          </el-select>
        </el-form-item>

        <el-form-item label="模型类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择模型类型">
            <el-option label="Chat" value="chat" />
            <el-option label="Completion" value="completion" />
            <el-option label="Embedding" value="embedding" />
          </el-select>
        </el-form-item>

        <el-form-item label="API端点" prop="endpoint">
          <el-input
            v-model="formData.endpoint"
            placeholder="例如: https://api.openai.com/v1/chat/completions"
          />
          <template #extra>
            <div style="color: #909399; font-size: 12px; margin-top: 4px;">
              留空则使用默认端点
            </div>
          </template>
        </el-form-item>

        <el-form-item label="最大Token数" prop="max_tokens">
          <el-input-number
            v-model="formData.max_tokens"
            :min="1"
            :max="200000"
            :step="1000"
            placeholder="例如: 4096"
          />
        </el-form-item>

        <el-form-item label="支持的特性" prop="supported_features">
          <el-input v-model="formData.supported_features" placeholder="例如: streaming,function_calling" />
        </el-form-item>

        <el-form-item label="输入成本($/1K)" prop="cost_per_1k_input_tokens">
          <el-input-number
            v-model="formData.cost_per_1k_input_tokens"
            :min="0"
            :precision="6"
            :step="0.001"
            placeholder="例如: 0.03"
          />
        </el-form-item>

        <el-form-item label="输出成本($/1K)" prop="cost_per_1k_output_tokens">
          <el-input-number
            v-model="formData.cost_per_1k_output_tokens"
            :min="0"
            :precision="6"
            :step="0.001"
            placeholder="例如: 0.06"
          />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="模型描述"
          />
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
import { ref, reactive, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { ArrowLeft } from '@element-plus/icons-vue';
import { getModelConfigById, createModelConfig, updateModelConfig } from '@/api/aimodel';

const router = useRouter();
const route = useRoute();
const formRef = ref<FormInstance>();
const loading = ref(false);
const submitting = ref(false);

const isEdit = ref(false);
const modelId = ref<number>();

const formData = reactive({
  name: '',
  display_name: '',
  provider: '',
  type: 'chat',
  endpoint: '',
  max_tokens: 4096,
  supported_features: 'streaming',
  cost_per_1k_input_tokens: 0,
  cost_per_1k_output_tokens: 0,
  description: ''
});

const rules: FormRules = {
  name: [
    { required: true, message: '请输入模型名称', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' }
  ],
  provider: [
    { required: true, message: '请选择提供商', trigger: 'change' }
  ],
  type: [
    { required: true, message: '请选择模型类型', trigger: 'change' }
  ],
  max_tokens: [
    { required: true, message: '请输入最大Token数', trigger: 'blur' }
  ],
  supported_features: [
    { required: true, message: '请输入支持的特性', trigger: 'blur' }
  ],
  cost_per_1k_input_tokens: [
    { required: true, message: '请输入输入成本', trigger: 'blur' }
  ],
  cost_per_1k_output_tokens: [
    { required: true, message: '请输入输出成本', trigger: 'blur' }
  ]
};

const fetchModelData = async () => {
  if (!modelId.value) return;

  loading.value = true;
  try {
    const res = await getModelConfigById(modelId.value);
    Object.assign(formData, {
      name: res.name,
      display_name: res.display_name,
      provider: res.provider,
      type: res.type,
      endpoint: res.endpoint || '',
      max_tokens: res.max_tokens,
      supported_features: res.supported_features,
      cost_per_1k_input_tokens: res.cost_per_1k_input_tokens,
      cost_per_1k_output_tokens: res.cost_per_1k_output_tokens,
      description: res.description || ''
    });
  } catch (error) {
    ElMessage.error('获取模型信息失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      const data = {
        name: formData.name,
        display_name: formData.display_name,
        provider: formData.provider,
        type: formData.type,
        endpoint: formData.endpoint || undefined,
        max_tokens: formData.max_tokens,
        supported_features: formData.supported_features,
        cost_per_1k_input_tokens: formData.cost_per_1k_input_tokens,
        cost_per_1k_output_tokens: formData.cost_per_1k_output_tokens,
        description: formData.description || undefined
      };

      if (isEdit.value && modelId.value) {
        await updateModelConfig({ id: modelId.value, ...data });
        ElMessage.success('更新成功');
      } else {
        await createModelConfig(data);
        ElMessage.success('创建成功');
      }

      router.push('/aimodel/models');
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败');
      console.error(error);
    } finally {
      submitting.value = false;
    }
  });
};

const handleBack = () => {
  router.back();
};

onMounted(() => {
  const id = route.params.id;
  if (id) {
    isEdit.value = true;
    modelId.value = Number(id);
    fetchModelData();
  }
});
</script>

<style scoped lang="scss">
.model-form {
  padding: 20px;

  .card-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .el-form {
    max-width: 600px;
  }
}
</style>
