<template>
  <div class="apikey-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>API密钥管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新增密钥
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="key_name" label="密钥名称" min-width="150" />
        <el-table-column prop="provider" label="提供商" width="120">
          <template #default="{ row }">
            <el-tag :type="getProviderType(row.provider)">
              {{ row.provider }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="api_key" label="API密钥" min-width="200">
          <template #default="{ row }">
            <span v-if="!row.showKey">{{ maskApiKey(row.api_key) }}</span>
            <span v-else style="font-family: monospace;">{{ row.api_key }}</span>
            <el-button
              link
              type="primary"
              @click="toggleKeyVisibility(row)"
              style="margin-left: 8px;"
            >
              <el-icon><View v-if="!row.showKey" /><Hide v-else /></el-icon>
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_time" label="创建时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑密钥' : '新增密钥'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="密钥名称" prop="key_name">
          <el-input v-model="formData.key_name" placeholder="例如: OpenAI Production Key" />
        </el-form-item>

        <el-form-item label="模型" prop="model_id">
          <el-select v-model="formData.model_id" placeholder="请选择模型" :disabled="isEdit">
            <el-option
              v-for="model in modelList"
              :key="model.id"
              :label="`${model.display_name} (${model.provider})`"
              :value="model.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="API密钥" prop="api_key" v-if="!isEdit">
          <el-input
            v-model="formData.api_key"
            type="textarea"
            :rows="3"
            placeholder="请输入API密钥"
          />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="2"
            placeholder="可选：添加密钥描述"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { Plus, Edit, Delete, View, Hide } from '@element-plus/icons-vue';
import { getApiKeys, createApiKey, updateApiKey, deleteApiKey, getModelConfigs } from '@/api/aimodel';
import type { ApiKey, ModelConfig } from '@/api/aimodel';

interface ApiKeyWithVisibility extends ApiKey {
  showKey?: boolean;
}

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();
const modelList = ref<ModelConfig[]>([]);

const tableData = ref<ApiKeyWithVisibility[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
});

const formData = reactive({
  id: undefined as number | undefined,
  key_name: '',
  model_id: undefined as number | undefined,
  api_key: '',
  description: ''
});

const rules: FormRules = {
  key_name: [
    { required: true, message: '请输入密钥名称', trigger: 'blur' }
  ],
  model_id: [
    { required: true, message: '请选择模型', trigger: 'change' }
  ],
  api_key: [
    { required: true, message: '请输入API密钥', trigger: 'blur' }
  ]
};

const getProviderType = (provider: string) => {
  const types: Record<string, string> = {
    openai: 'primary',
    claude: 'success',
    ollama: 'warning'
  };
  return types[provider] || 'info';
};

const maskApiKey = (key: string) => {
  if (!key || key.length < 8) return '***';
  return key.substring(0, 8) + '...' + key.substring(key.length - 4);
};

const toggleKeyVisibility = (row: ApiKeyWithVisibility) => {
  row.showKey = !row.showKey;
};

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getApiKeys({
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    });
    tableData.value = (res.keys || []).map(item => ({
      ...item,
      showKey: false
    }));
    pagination.value.total = res.total || 0;
  } catch (error) {
    ElMessage.error('获取密钥列表失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  isEdit.value = false;
  Object.assign(formData, {
    id: undefined,
    key_name: '',
    model_id: undefined,
    api_key: '',
    description: ''
  });
  dialogVisible.value = true;
};

const handleEdit = (row: ApiKey) => {
  isEdit.value = true;
  Object.assign(formData, {
    id: row.id,
    key_name: row.key_name,
    model_id: row.model_id,
    api_key: '',
    description: row.description || ''
  });
  dialogVisible.value = true;
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (isEdit.value && formData.id) {
        await updateApiKey({
          id: formData.id,
          key_name: formData.key_name
        });
        ElMessage.success('更新成功');
      } else {
        await createApiKey({
          model_id: formData.model_id!,
          key_name: formData.key_name,
          api_key: formData.api_key,
          description: formData.description
        });
        ElMessage.success('创建成功');
      }

      dialogVisible.value = false;
      fetchData();
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败');
      console.error(error);
    } finally {
      submitting.value = false;
    }
  });
};

const handleDelete = async (row: ApiKey) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除密钥 "${row.key_name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    await deleteApiKey(row.id);
    ElMessage.success('删除成功');
    fetchData();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
      console.error(error);
    }
  }
};

const fetchModelList = async () => {
  try {
    const res = await getModelConfigs({ page: 1, page_size: 100 });
    modelList.value = res.models || [];
  } catch (error) {
    console.error('获取模型列表失败:', error);
  }
};

onMounted(() => {
  fetchModelList();
  fetchData();
});
</script>

<style scoped lang="scss">
.apikey-list {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
