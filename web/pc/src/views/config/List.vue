<template>
  <div class="config-list">
    <div class="page-header">
      <h2>系统配置管理</h2>
      <div style="display: flex; gap: 10px;">
        <el-button type="warning" :disabled="selectedRows.length === 0" @click="handleBatchSubmit">
          批量提交 ({{ selectedRows.length }})
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建配置
        </el-button>
      </div>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="模块">
            <el-input v-model="filters.module" placeholder="请输入模块名" clearable style="width: 200px" />
          </el-form-item>
          <el-form-item label="配置键">
            <el-input v-model="filters.key" placeholder="请输入配置键" clearable style="width: 200px" />
          </el-form-item>
          <el-form-item label="提交状态">
            <el-select v-model="filters.submission_status" placeholder="全部" clearable style="width: 150px">
              <el-option label="待提交" :value="0" />
              <el-option label="已提交" :value="1" />
              <el-option label="已批准" :value="2" />
              <el-option label="已拒绝" :value="3" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" :selectable="canSelect" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="module" label="模块" width="150" />
        <el-table-column prop="key" label="配置键" width="200" />
        <el-table-column prop="value" label="配置值" show-overflow-tooltip />
        <el-table-column label="值类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.valueType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="公开" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.isPublic" type="success" size="small">是</el-tag>
            <el-tag v-else type="info" size="small">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交状态" width="100">
          <template #default="{ row }">
            <el-tag :type="submissionStatusMap[row.submission_status]?.type || 'info'" size="small">
              {{ submissionStatusMap[row.submission_status]?.label || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="canEdit(row)" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="canSubmit(row)" link type="warning" @click="handleSubmit(row)">提交</el-button>
            <el-button v-if="canDelete(row)" link type="danger" @click="handleDelete(row)">删除</el-button>
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
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="模块" prop="module">
          <el-input v-model="form.module" placeholder="如：认证管理、基础数据、物业管理" />
        </el-form-item>
        <el-form-item label="配置键" prop="key">
          <el-input
            v-model="form.key"
            placeholder="如：max_upload_size"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item label="值类型" prop="valueType">
          <el-select v-model="form.valueType" placeholder="请选择" style="width: 100%">
            <el-option label="字符串 (string)" value="string" />
            <el-option label="数字 (number)" value="number" />
            <el-option label="布尔值 (boolean)" value="boolean" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="配置值" prop="value">
          <el-input
            v-if="form.valueType !== 'json'"
            v-model="form.value"
            :placeholder="getValuePlaceholder()"
          />
          <el-input
            v-else
            v-model="form.value"
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
        <el-form-item label="公开配置">
          <el-switch v-model="form.isPublic" />
          <span style="margin-left: 10px; color: #909399; font-size: 12px">
            公开配置可能会暴露给前端
          </span>
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
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import * as masterdataApi from '@/api/masterdata';
import { logger } from '@/utils/logger'

type StatusType = 'info' | 'warning' | 'success' | 'danger'

const submissionStatusMap: Record<number, { label: string; type: StatusType }> = {
  0: { label: '待提交', type: 'info' },
  1: { label: '已提交', type: 'warning' },
  2: { label: '已批准', type: 'success' },
  3: { label: '已拒绝', type: 'danger' }
}

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const dialogTitle = ref('新建配置');
const tableData = ref<any[]>([]);
const selectedRows = ref<any[]>([]);
const formRef = ref<FormInstance>();

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const filters = reactive({
  module: '',
  key: '',
  submission_status: undefined as number | undefined
});

const form = reactive({
  id: 0,
  module: '',
  key: '',
  value: '',
  valueType: 'string',
  description: '',
  isPublic: false
});

const canEdit = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canSubmit = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canDelete = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canSelect = (row: any) => row.submission_status === 0 || row.submission_status === 3

const validateValue = (_rule: any, value: any, callback: any) => {
  if (!value) {
    callback(new Error('请输入配置值'));
    return;
  }

  if (form.valueType === 'number' && isNaN(Number(value))) {
    callback(new Error('请输入有效的数字'));
    return;
  }

  if (form.valueType === 'boolean' && !['true', 'false'].includes(value.toLowerCase())) {
    callback(new Error('布尔值只能是 true 或 false'));
    return;
  }

  if (form.valueType === 'json') {
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
  module: [
    { required: true, message: '请输入模块名', trigger: 'blur' }
  ],
  key: [
    { required: true, message: '请输入配置键', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线', trigger: 'blur' }
  ],
  valueType: [
    { required: true, message: '请选择值类型', trigger: 'change' }
  ],
  value: [
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

    if (filters.module) params.module = filters.module;
    if (filters.key) params.key = filters.key;
    if (filters.submission_status !== undefined) params.submission_status = filters.submission_status;

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
  filters.module = '';
  filters.key = '';
  filters.submission_status = undefined;
  pagination.page = 1;
  loadData();
};

const handleCreate = () => {
  dialogTitle.value = '新建配置';
  resetForm();
  dialogVisible.value = true;
};

const handleEdit = (row: any) => {
  if (!canEdit(row)) {
    ElMessage.warning('当前状态不允许编辑');
    return;
  }
  dialogTitle.value = '编辑配置';
  form.id = row.id;
  form.module = row.module;
  form.key = row.key;
  form.value = row.value;
  form.valueType = row.valueType;
  form.description = row.description;
  form.isPublic = row.isPublic;
  dialogVisible.value = true;
};

const handleDelete = async (row: any) => {
  if (!canDelete(row)) {
    ElMessage.warning('当前状态不允许删除');
    return;
  }
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '提示', {
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

const handleSubmit = async (row: any) => {
  if (!canSubmit(row)) {
    ElMessage.warning('当前状态不允许提交');
    return;
  }
  try {
    await ElMessageBox.confirm('确定要提交该配置进行审核吗？', '提示', { type: 'warning' });
    await masterdataApi.submitConfiguration(row.id);
    ElMessage.success('提交成功');
    loadData();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '提交失败');
    }
  }
};

const handleSelectionChange = (rows: any[]) => {
  selectedRows.value = rows;
};

const handleBatchSubmit = async () => {
  if (selectedRows.value.length === 0) return;
  try {
    await ElMessageBox.confirm(`确定要批量提交 ${selectedRows.value.length} 条配置吗？`, '提示', { type: 'warning' });
    const ids = selectedRows.value.map((r: any) => r.id);
    await masterdataApi.batchSubmitConfigurations(ids);
    ElMessage.success('批量提交成功');
    loadData();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量提交失败');
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
          value: form.value,
          description: form.description,
          is_public: form.isPublic
        });
        ElMessage.success('更新成功');
      } else {
        await masterdataApi.createConfiguration({
          module: form.module,
          key: form.key,
          value: form.value,
          value_type: form.valueType,
          description: form.description,
          is_public: form.isPublic
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
  switch (form.valueType) {
    case 'number':
      return '请输入数字，如：10';
    case 'boolean':
      return '请输入 true 或 false';
    default:
      return '请输入配置值';
  }
};

const resetForm = () => {
  form.id = 0;
  form.module = '';
  form.key = '';
  form.value = '';
  form.valueType = 'string';
  form.description = '';
  form.isPublic = false;
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
