<template>
  <div class="sensitive-words-list">
    <div class="page-header">
      <h2>敏感词管理</h2>
      <div style="display: flex; gap: 10px;">
        <el-button type="warning" :disabled="selectedRows.length === 0" @click="handleBatchSubmit">
          批量提交 ({{ selectedRows.length }})
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          添加敏感词
        </el-button>
      </div>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="分类">
            <el-input v-model="filters.category" placeholder="请输入分类" clearable style="width: 150px" />
          </el-form-item>
          <el-form-item label="严重等级">
            <el-select v-model="filters.severity" placeholder="全部" clearable style="width: 120px">
              <el-option label="低 (1)" :value="1" />
              <el-option label="中 (2)" :value="2" />
              <el-option label="高 (3)" :value="3" />
            </el-select>
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
        <el-table-column prop="word" label="敏感词" width="200" />
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column label="严重等级" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.severity === 1" type="info" size="small">低</el-tag>
            <el-tag v-else-if="row.severity === 2" type="warning" size="small">中</el-tag>
            <el-tag v-else-if="row.severity === 3" type="danger" size="small">高</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="处理动作" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.action === 'warn'" type="warning" size="small">警告</el-tag>
            <el-tag v-else-if="row.action === 'block'" type="danger" size="small">拦截</el-tag>
            <el-tag v-else-if="row.action === 'review'" type="info" size="small">审核</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="提交状态" width="100">
          <template #default="{ row }">
            <el-tag :type="submissionStatusMap[row.submission_status]?.type || 'info'" size="small">
              {{ submissionStatusMap[row.submission_status]?.label || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
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
        <el-form-item label="敏感词" prop="word">
          <el-input
            v-model="form.word"
            placeholder="请输入敏感词"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-input v-model="form.category" placeholder="如：政治、色情、暴力、广告" />
        </el-form-item>
        <el-form-item label="严重等级" prop="severity">
          <el-radio-group v-model="form.severity">
            <el-radio :label="1">低</el-radio>
            <el-radio :label="2">中</el-radio>
            <el-radio :label="3">高</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="处理动作" prop="action">
          <el-radio-group v-model="form.action">
            <el-radio label="warn">警告</el-radio>
            <el-radio label="block">拦截</el-radio>
            <el-radio label="review">审核</el-radio>
          </el-radio-group>
          <div style="margin-top: 8px; color: #909399; font-size: 12px">
            警告：提示用户但允许发布；拦截：禁止发布；审核：进入人工审核
          </div>
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
const dialogTitle = ref('添加敏感词');
const tableData = ref<any[]>([]);
const selectedRows = ref<any[]>([]);
const formRef = ref<FormInstance>();

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const filters = reactive({
  category: '',
  severity: undefined as number | undefined,
  submission_status: undefined as number | undefined
});

const form = reactive({
  id: 0,
  word: '',
  category: '',
  severity: 1,
  action: 'warn'
});

const canEdit = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canSubmit = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canDelete = (row: any) => row.submission_status === 0 || row.submission_status === 3
const canSelect = (row: any) => row.submission_status === 0 || row.submission_status === 3

const rules: FormRules = {
  word: [
    { required: true, message: '请输入敏感词', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请输入分类', trigger: 'blur' }
  ],
  severity: [
    { required: true, message: '请选择严重等级', trigger: 'change' }
  ],
  action: [
    { required: true, message: '请选择处理动作', trigger: 'change' }
  ]
};

onMounted(() => {
  logger.componentMounted('Sensitive Words List');
  loadData();
});

const loadData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    };

    if (filters.category) params.category = filters.category;
    if (filters.severity !== undefined) params.severity = filters.severity;
    if (filters.submission_status !== undefined) params.submission_status = filters.submission_status;

    const response = await masterdataApi.getSensitiveWords(params);
    tableData.value = response?.list || [];
    pagination.total = response?.total || 0;
  } catch (error) {
    ElMessage.error('加载敏感词列表失败');
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.page = 1;
  loadData();
};

const handleReset = () => {
  filters.category = '';
  filters.severity = undefined;
  filters.submission_status = undefined;
  pagination.page = 1;
  loadData();
};

const handleCreate = () => {
  dialogTitle.value = '添加敏感词';
  resetForm();
  dialogVisible.value = true;
};

const handleEdit = (row: any) => {
  if (!canEdit(row)) {
    ElMessage.warning('当前状态不允许编辑');
    return;
  }
  dialogTitle.value = '编辑敏感词';
  form.id = row.id;
  form.word = row.word;
  form.category = row.category;
  form.severity = row.severity;
  form.action = row.action;
  dialogVisible.value = true;
};

const handleDelete = async (row: any) => {
  if (!canDelete(row)) {
    ElMessage.warning('当前状态不允许删除');
    return;
  }
  try {
    await ElMessageBox.confirm('确定要删除该敏感词吗？', '提示', {
      type: 'warning'
    });

    await masterdataApi.deleteSensitiveWord(row.id);
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
    await ElMessageBox.confirm('确定要提交该敏感词进行审核吗？', '提示', { type: 'warning' });
    await masterdataApi.submitSensitiveWord(row.id);
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
    await ElMessageBox.confirm(`确定要批量提交 ${selectedRows.value.length} 条敏感词吗？`, '提示', { type: 'warning' });
    const ids = selectedRows.value.map((r: any) => r.id);
    await masterdataApi.batchSubmitSensitiveWords(ids);
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
        await masterdataApi.updateSensitiveWord(form.id, {
          category: form.category,
          severity: form.severity,
          action: form.action
        });
        ElMessage.success('更新成功');
      } else {
        await masterdataApi.createSensitiveWord({
          word: form.word,
          category: form.category,
          severity: form.severity,
          action: form.action
        });
        ElMessage.success('添加成功');
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

const resetForm = () => {
  form.id = 0;
  form.word = '';
  form.category = '';
  form.severity = 1;
  form.action = 'warn';
  formRef.value?.resetFields();
};
</script>

<style scoped lang="scss">
.sensitive-words-list {
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
