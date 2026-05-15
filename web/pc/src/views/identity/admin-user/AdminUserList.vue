<template>
  <div class="admin-user-container">
    <div class="page-header">
      <h2>管理员管理</h2>
      <el-button
        type="primary"
        v-permission="'identity:admin-user:create'"
        @click="handleCreate"
      >
        新建管理员
      </el-button>
    </div>

    <el-table
      :data="tableData"
      v-loading="loading"
      border
      stripe
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column prop="nickname" label="昵称" width="150" />
      <el-table-column label="角色" width="200">
        <template #default="{ row }">
          <el-tag
            v-for="role in row.roles"
            :key="role.id"
            size="small"
            style="margin-right: 5px"
          >
            {{ role.name }}
          </el-tag>
          <el-button
            link
            type="primary"
            size="small"
            v-permission="'identity:admin-user:assign-role'"
            @click="handleAssignRoles(row)"
          >
            分配角色
          </el-button>
        </template>
      </el-table-column>
      <el-table-column prop="scope" label="权限范围" width="150" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="2"
            v-permission="'identity:admin-user:update'"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column prop="lastLoginAt" label="最后登录" width="180" />
      <el-table-column prop="createdAt" label="创建时间" width="180" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button
            link
            type="primary"
            size="small"
            v-permission="'identity:admin-user:update'"
            @click="handleEdit(row)"
          >
            编辑
          </el-button>
          <el-button
            link
            type="danger"
            size="small"
            v-permission="'identity:admin-user:delete'"
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :page-sizes="[10, 20, 50, 100]"
      :total="pagination.total"
      layout="total, sizes, prev, pager, next, jumper"
      @size-change="fetchUsers"
      @current-change="fetchUsers"
      style="margin-top: 20px; justify-content: flex-end"
    />

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            :disabled="!!form.id"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!form.id">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="权限范围" prop="scope">
          <el-input
            v-model="form.scope"
            placeholder="请输入权限范围（可选）"
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

    <!-- Assign Roles Dialog -->
    <el-dialog
      v-model="roleDialogVisible"
      title="分配角色"
      width="500px"
      @close="resetRoleDialog"
    >
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="role in availableRoles"
          :key="role.id"
          :label="role.id"
          style="display: block; margin-bottom: 10px"
        >
          {{ role.name }} ({{ role.code }})
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleSubmitRoles"
          :loading="submitting"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import {
  getUsers,
  createUser,
  updateUser,
  disableUser,
  enableUser,
  getUserRoles,
  assignUserRoles,
  getRoles
} from '@/api/identity';
import { UserType, UserStatus, type User, type Role } from '@common/types/identity';

interface UserWithRoles extends User {
  roles?: Role[];
}

const loading = ref(false);
const submitting = ref(false);
const tableData = ref<UserWithRoles[]>([]);
const dialogVisible = ref(false);
const roleDialogVisible = ref(false);
const formRef = ref<FormInstance>();
const currentUserId = ref<number>();
const selectedRoleIds = ref<number[]>([]);
const availableRoles = ref<Role[]>([]);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const form = reactive({
  id: undefined as number | undefined,
  phone: '',
  password: '',
  nickname: '',
  scope: ''
});

const dialogTitle = computed(() => (form.id ? '编辑管理员' : '新建管理员'));

const formRules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' }
  ]
};

const fetchUsers = async () => {
  loading.value = true;
  try {
    const { data } = await getUsers({
      userType: UserType.Staff,
      page: pagination.page,
      page_size: pagination.pageSize
    });

    // Fetch roles for each user
    const usersWithRoles = await Promise.all(
      data.list.map(async (user) => {
        try {
          const { data: rolesData } = await getUserRoles(user.id);
          return { ...user, roles: rolesData.roles };
        } catch {
          return { ...user, roles: [] };
        }
      })
    );

    tableData.value = usersWithRoles;
    pagination.total = data.total;
  } catch (error) {
    ElMessage.error('获取管理员列表失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const fetchAvailableRoles = async () => {
  try {
    const { data } = await getRoles({ page: 1, page_size: 100 });
    availableRoles.value = data.list;
  } catch (error) {
    ElMessage.error('获取角色列表失败');
    console.error(error);
  }
};

const handleCreate = () => {
  dialogVisible.value = true;
};

const handleEdit = (row: User) => {
  form.id = row.id;
  form.phone = row.phone;
  form.nickname = row.nickname;
  form.scope = row.scope;
  dialogVisible.value = true;
};

const handleDelete = async (row: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除管理员 "${row.nickname}" 吗？`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    await disableUser(row.id);
    ElMessage.success('删除成功');
    fetchUsers();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
      console.error(error);
    }
  }
};

const handleStatusChange = async (row: User) => {
  try {
    if (row.status === UserStatus.Active) {
      await enableUser(row.id);
      ElMessage.success('已启用');
    } else {
      await disableUser(row.id);
      ElMessage.success('已禁用');
    }
  } catch (error) {
    ElMessage.error('状态更新失败');
    row.status = row.status === UserStatus.Active ? UserStatus.Disabled : UserStatus.Active;
    console.error(error);
  }
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (form.id) {
        await updateUser(form.id, {
          nickname: form.nickname,
          scope: form.scope
        });
        ElMessage.success('更新成功');
      } else {
        await createUser({
          phone: form.phone,
          password: form.password,
          nickname: form.nickname,
          user_type: UserType.Staff,
          scope: form.scope
        });
        ElMessage.success('创建成功');
      }

      dialogVisible.value = false;
      fetchUsers();
    } catch (error) {
      ElMessage.error(form.id ? '更新失败' : '创建失败');
      console.error(error);
    } finally {
      submitting.value = false;
    }
  });
};

const handleAssignRoles = async (row: UserWithRoles) => {
  currentUserId.value = row.id;
  selectedRoleIds.value = row.roles?.map(r => r.id) || [];
  roleDialogVisible.value = true;
};

const handleSubmitRoles = async () => {
  if (!currentUserId.value) return;

  submitting.value = true;
  try {
    await assignUserRoles(currentUserId.value, selectedRoleIds.value);
    ElMessage.success('角色分配成功');
    roleDialogVisible.value = false;
    fetchUsers();
  } catch (error) {
    ElMessage.error('角色分配失败');
    console.error(error);
  } finally {
    submitting.value = false;
  }
};

const resetForm = () => {
  form.id = undefined;
  form.phone = '';
  form.password = '';
  form.nickname = '';
  form.scope = '';
  formRef.value?.resetFields();
};

const resetRoleDialog = () => {
  currentUserId.value = undefined;
  selectedRoleIds.value = [];
};

onMounted(() => {
  fetchUsers();
  fetchAvailableRoles();
});
</script>

<style scoped>
.admin-user-container {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
</style>
