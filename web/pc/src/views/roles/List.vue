<template>
  <div class="roles-list">
    <div class="page-header">
      <h2>角色管理</h2>
      <el-button type="primary" v-permission="'role:create'" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        新建角色
      </el-button>
    </div>

    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="角色名称" />
        <el-table-column prop="code" label="角色编码" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="系统角色" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.isSystem" type="warning" size="small">系统</el-tag>
            <el-tag v-else type="info" size="small">自定义</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.status === 1" type="success" size="small">启用</el-tag>
            <el-tag v-else type="danger" size="small">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'role:update'" link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button v-permission="'role:permission'" link type="primary" @click="handlePermissions(row)">权限配置</el-button>
            <el-button
              v-permission="'role:delete'"
              link
              type="danger"
              @click="handleDelete(row)"
              :disabled="row.isSystem"
            >
              删除
            </el-button>
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
          @size-change="loadRoles"
          @current-change="loadRoles"
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
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" prop="code">
          <el-input
            v-model="form.code"
            placeholder="请输入角色编码（英文字母、数字、下划线）"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
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
import type { Role } from '@common/types/identity';
import * as identityApi from '@/api/identity';
import { useRouter } from 'vue-router';

const router = useRouter();

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const dialogTitle = ref('新建角色');
const tableData = ref<Role[]>([]);
const formRef = ref<FormInstance>();

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const form = reactive({
  id: 0,
  name: '',
  code: '',
  description: ''
});

const rules: FormRules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: '只能包含字母、数字和下划线', trigger: 'blur' }
  ]
};

onMounted(() => {
  loadRoles();
});

const loadRoles = async () => {
  loading.value = true;
  try {
    const response = await identityApi.getRoles({
      page: pagination.page,
      page_size: pagination.pageSize
    });
    tableData.value = response?.list || [];
    pagination.total = response?.total || 0;
  } catch (error) {
    ElMessage.error('加载角色列表失败');
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  dialogTitle.value = '新建角色';
  resetForm();
  dialogVisible.value = true;
};

const handleEdit = (row: Role) => {
  dialogTitle.value = '编辑角色';
  form.id = row.id;
  form.name = row.name;
  form.code = row.code;
  form.description = row.description;
  dialogVisible.value = true;
};

const handlePermissions = (row: Role) => {
  router.push(`/roles/${row.id}/permissions`);
};

const handleDelete = async (row: Role) => {
  if (row.isSystem) {
    ElMessage.warning('系统角色不能删除');
    return;
  }

  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      type: 'warning'
    });

    await identityApi.deleteRole(row.id);
    ElMessage.success('删除成功');
    loadRoles();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败');
    }
  }
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (form.id) {
        await identityApi.updateRole(form.id, {
          name: form.name,
          description: form.description
        });
        ElMessage.success('更新成功');
      } else {
        await identityApi.createRole({
          name: form.name,
          code: form.code,
          description: form.description
        });
        ElMessage.success('创建成功');
      }
      dialogVisible.value = false;
      loadRoles();
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败');
    } finally {
      submitting.value = false;
    }
  });
};

const resetForm = () => {
  form.id = 0;
  form.name = '';
  form.code = '';
  form.description = '';
  formRef.value?.resetFields();
};
</script>

<style scoped lang="scss">
.roles-list {
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

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
