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
        <el-table-column prop="id" label="ID" width="200" />
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
        <el-table-column label="允许登录端" width="160">
          <template #default="{ row }">
            <template v-if="row.platforms && row.platforms.length > 0">
              <el-tag
                v-for="p in row.platforms"
                :key="p"
                size="small"
                :type="p === 'pc' ? 'primary' : 'success'"
                class="platform-tag"
              >
                {{ platformLabel(p) }}
              </el-tag>
            </template>
            <el-tag v-else size="small" type="info">全部</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="380" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'role:update'" link type="primary" @click="handleEdit(row as Role)">编辑</el-button>
            <el-button v-permission="'role:permission'" link type="primary" @click="handlePermissions(row as Role)">权限配置</el-button>
            <el-button link type="primary" @click="handleRoleUsers(row as Role)">查看用户</el-button>
            <el-button
              v-permission="'role:delete'"
              link
              type="danger"
              @click="handleDelete(row as Role)"
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
        <el-form-item label="允许登录端" prop="platforms">
          <el-checkbox-group v-model="form.platforms">
            <el-checkbox
              v-for="opt in PLATFORM_OPTIONS"
              :key="opt.value"
              :value="opt.value"
              :label="opt.value"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
          <div class="platform-hint">未勾选 = 允许所有端登录</div>
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
/* eslint-disable */
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import type { Role } from '@common/types/identity';
import * as identityApi from '@/api/identity';
import { useRouter } from 'vue-router';

const router = useRouter();

// 允许登录端选项：值域与 permission-service sys_role.platforms 一致（pc/mobile）
// 空数组 = 未声明，运行时 fail-open（允许所有端），由 auth-service 权威判定
// SEE: [[frontend-business-rule-hardcode]] — 前端仅做选项与展示，端准入权威判定在后端
const PLATFORM_OPTIONS = [
  { value: 'pc', label: 'PC 管理端' },
  { value: 'mobile', label: '移动端' }
];

const platformLabel = (p: string) => {
  return p === 'pc' ? 'PC' : '移动端';
};

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
  id: '',
  name: '',
  code: '',
  description: '',
  platforms: [] as string[]
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
    tableData.value = response?.roles || [];
    pagination.total = response?.page?.total || 0;
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
  form.platforms = Array.isArray(row.platforms) ? [...row.platforms] : [];
  dialogVisible.value = true;
};

const handlePermissions = (row: Role) => {
  router.push(`/roles/${row.id}/permissions`);
};

const handleRoleUsers = (row: Role) => {
  router.push(`/roles/${row.id}/users`);
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
          description: form.description,
          platforms: form.platforms
        });
        ElMessage.success('更新成功');
      } else {
        await identityApi.createRole({
          name: form.name,
          code: form.code,
          description: form.description,
          platforms: form.platforms
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
  form.id = '';
  form.name = '';
  form.code = '';
  form.description = '';
  form.platforms = [];
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

  .platform-tag {
    margin-right: 4px;
  }

  .platform-hint {
    font-size: 12px;
    color: #86909c;
    line-height: 1.5;
  }
}
</style>
