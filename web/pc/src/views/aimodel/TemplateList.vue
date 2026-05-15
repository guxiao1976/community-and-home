<template>
  <div class="template-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>提示模板管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新增模板
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="模板名称" min-width="150" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="模板内容" min-width="300" show-overflow-tooltip />
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
      :title="isEdit ? '编辑模板' : '新增模板'"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: 代码审查模板" />
        </el-form-item>

        <el-form-item label="分类" prop="category">
          <el-select v-model="formData.category" placeholder="请选择分类" allow-create filterable>
            <el-option label="代码" value="code" />
            <el-option label="文档" value="document" />
            <el-option label="翻译" value="translation" />
            <el-option label="分析" value="analysis" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>

        <el-form-item label="模板内容" prop="content">
          <el-input
            v-model="formData.content"
            type="textarea"
            :rows="10"
            placeholder="请输入模板内容，可以使用 {{variable}} 作为变量占位符"
          />
          <template #extra>
            <div style="color: #909399; font-size: 12px; margin-top: 4px;">
              提示：使用 {{variable}} 格式定义变量，例如 {{code}}, {{language}}
            </div>
          </template>
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
import { Plus, Edit, Delete } from '@element-plus/icons-vue';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate } from '@/api/aimodel';
import type { PromptTemplate } from '@/api/aimodel';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();

const tableData = ref<PromptTemplate[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
});

const formData = reactive({
  id: undefined as number | undefined,
  name: '',
  category: '',
  content: ''
});

const rules: FormRules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ],
  content: [
    { required: true, message: '请输入模板内容', trigger: 'blur' }
  ]
};

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getTemplates({
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    });
    tableData.value = res.templates || [];
    pagination.value.total = res.total || 0;
  } catch (error) {
    ElMessage.error('获取模板列表失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  isEdit.value = false;
  Object.assign(formData, {
    id: undefined,
    name: '',
    category: '',
    content: ''
  });
  dialogVisible.value = true;
};

const handleEdit = (row: PromptTemplate) => {
  isEdit.value = true;
  Object.assign(formData, {
    id: row.id,
    name: row.name,
    category: row.category,
    content: row.content
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
        await updateTemplate({
          id: formData.id,
          name: formData.name,
          category: formData.category,
          content: formData.content
        });
        ElMessage.success('更新成功');
      } else {
        await createTemplate({
          name: formData.name,
          category: formData.category,
          content: formData.content
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

const handleDelete = async (row: PromptTemplate) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${row.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    await deleteTemplate(row.id);
    ElMessage.success('删除成功');
    fetchData();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
      console.error(error);
    }
  }
};

onMounted(() => {
  fetchData();
});
</script>

<style scoped lang="scss">
.template-list {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
