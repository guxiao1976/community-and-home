<template>
  <div class="config-list">
    <div class="page-header">
      <h2>系统配置管理</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        新建配置
      </el-button>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="配置键">
            <el-input v-model="filters.keyword" placeholder="请输入配置键关键字" clearable style="width: 240px" @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="config_key" label="配置键" width="220" show-overflow-tooltip />
        <el-table-column prop="config_value" label="配置值" show-overflow-tooltip />
        <el-table-column label="值类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ row.value_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="updated_time" label="更新时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="560px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="配置键" prop="config_key">
          <el-input
            v-model="form.config_key"
            placeholder="如：max_upload_size"
            :disabled="isEditing"
          />
        </el-form-item>
        <el-form-item label="值类型" prop="value_type">
          <el-select v-model="form.value_type" placeholder="请选择" style="width: 100%">
            <el-option label="字符串 (string)" value="string" />
            <el-option label="数字 (number)" value="number" />
            <el-option label="布尔值 (boolean)" value="boolean" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置值" prop="config_value">
          <el-input
            v-if="form.value_type !== 'json'"
            v-model="form.config_value"
            :placeholder="getValuePlaceholder()"
          />
          <el-input
            v-else
            v-model="form.config_value"
            type="textarea"
            :rows="4"
            placeholder='{"key": "value"}'
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="2"
            placeholder="请输入配置描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmitForm">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import * as masterdataApi from '@/api/masterdata';
import { logger } from '@/utils/logger'

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const dialogTitle = ref('新建配置');
const tableData = ref<any[]>([]);
const formRef = ref<FormInstance>();

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const filters = reactive({
  keyword: ''
});

const form = reactive({
  id: '',
  config_key: '',
  config_value: '',
  value_type: 'string',
  description: ''
});

const isEditing = computed(() => !!form.id);

const validateValue = (_rule: any, value: any, callback: any) => {
  if (!value) {
    callback(new Error('请输入配置值'));
    return;
  }

  if (form.value_type === 'number' && isNaN(Number(value))) {
    callback(new Error('请输入有效的数字'));
    return;
  }

  if (form.value_type === 'boolean' && !['true', 'false'].includes(value.toLowerCase())) {
    callback(new Error('布尔值只能是 true 或 false'));
    return;
  }

  if (form.value_type === 'json') {
    try {
      JSON.parse(value);
    } catch (e) {
      callback(new Error('请输入有效的 JSON 格式'));
      return;
    }
  }

  callback();
};

const rules: FormRules = {
  config_key: [
    { required: true, message: '请输入配置键', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  value_type: [
    { required: true, message: '请选择值类型', trigger: 'change' }
  ],
  config_value: [
    { required: true, validator: validateValue, trigger: 'blur' }
  ]
};

onMounted(() => {
  logger.componentMounted('Config List');
  loadData();
});

const loadData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    };

    if (filters.keyword) params.keyword = filters.keyword;

    const response = await masterdataApi.getConfigurations(params);
    tableData.value = response?.list || [];
    pagination.total = response?.total || 0;
  } catch (error) {
    ElMessage.error('加载配置列表失败');
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.page = 1;
  loadData();
};

const handleReset = () => {
  filters.keyword = '';
  pagination.page = 1;
  loadData();
};

const handleCreate = () => {
  dialogTitle.value = '新建配置';
  resetForm();
  dialogVisible.value = true;
};

const handleEdit = (row: any) => {
  dialogTitle.value = '编辑配置';
  form.id = row.id;
  form.config_key = row.config_key;
  form.config_value = row.config_value;
  form.value_type = row.value_type;
  form.description = row.description;
  dialogVisible.value = true;
};

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除配置「${row.config_key}」吗？`, '提示', {
      type: 'warning'
    });

    await masterdataApi.deleteConfiguration(row.id);
    ElMessage.success('删除成功');
    loadData();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败');
    }
  }
};

const handleSubmitForm = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (form.id) {
        await masterdataApi.updateConfiguration(form.id, {
          config_value: form.config_value,
          value_type: form.value_type,
          description: form.description
        });
        ElMessage.success('更新成功');
      } else {
        await masterdataApi.createConfiguration({
          config_key: form.config_key,
          config_value: form.config_value,
          value_type: form.value_type,
          description: form.description
        });
        ElMessage.success('创建成功');
      }
      dialogVisible.value = false;
      loadData();
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败');
    } finally {
      submitting.value = false;
    }
  });
};

const getValuePlaceholder = () => {
  switch (form.value_type) {
    case 'number':
      return '请输入数字，如：10';
    case 'boolean':
      return '请输入 true 或 false';
    default:
      return '请输入配置值';
  }
};

const resetForm = () => {
  form.id = '';
  form.config_key = '';
  form.config_value = '';
  form.value_type = 'string';
  form.description = '';
  formRef.value?.resetFields();
};
</script>

<style scoped lang="scss">
.config-list {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 20px;
      font-weight: 500;
    }
  }

  .filter-bar {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
