<!-- @deprecated 请使用 /views/roles/List.vue + /views/roles/Permissions.vue，本文件为早期 mock 实现 -->
<template>
  <div class="role-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="handleAdd">新增角色</el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="queryForm" class="search-form">
        <el-form-item label="角色名称">
          <el-input v-model="queryForm.keyword" placeholder="请输入角色名称" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="请选择状态" clearable>
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 角色列表 -->
      <el-table :data="roleList" border style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="角色ID" width="180" />
        <el-table-column prop="role_code" label="角色编码" width="150" />
        <el-table-column prop="role_name" label="角色名称" width="150" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="is_system" label="系统角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_system ? 'danger' : 'info'">
              {{ row.is_system ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="primary" @click="handlePermission(row)">配置权限</el-button>
            <el-button link type="danger" @click="handleDelete(row)" v-if="!row.is_system">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleQuery"
        @current-change="handleQuery"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="角色编码" prop="role_code">
          <el-input v-model="formData.role_code" placeholder="请输入角色编码" />
        </el-form-item>
        <el-form-item label="角色名称" prop="role_name">
          <el-input v-model="formData.role_name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="formData.sort_order" :min="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限配置对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      title="配置权限"
      width="600px"
    >
      <el-tree
        ref="permissionTreeRef"
        :data="permissionTree"
        show-checkbox
        node-key="id"
        :props="{ children: 'children', label: 'name' }"
        :default-checked-keys="checkedPermissions"
      />
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handlePermissionSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

// 查询表单
const queryForm = reactive({
  keyword: '',
  status: undefined as number | undefined,
})

// 角色列表
const roleList = ref([])
const loading = ref(false)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('新增角色')
const formRef = ref<FormInstance>()
const submitting = ref(false)

// 表单数据
const formData = reactive({
  id: undefined as number | undefined,
  role_code: '',
  role_name: '',
  description: '',
  sort_order: 0,
  status: 1,
})

// 表单验证规则
const formRules: FormRules = {
  role_code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }],
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
}

// 权限配置对话框
const permissionDialogVisible = ref(false)
const permissionTreeRef = ref()
const permissionTree = ref([])
const checkedPermissions = ref<number[]>([])
const currentRoleId = ref<number>()

// 加载角色列表
const loadRoleList = async () => {
  loading.value = true
  try {
    // TODO: 调用 API
    // const res = await getRoleList({
    //   page: pagination.page,
    //   pageSize: pagination.pageSize,
    //   ...queryForm
    // })
    // roleList.value = res.data.roles
    // pagination.total = res.data.total

    // 模拟数据
    roleList.value = [
      {
        id: 1,
        role_code: 'owner',
        role_name: '业主',
        description: '小区业主',
        is_system: false,
        status: 1,
      },
      {
        id: 2,
        role_code: 'property_admin',
        role_name: '物业管理员',
        description: '物业管理人员',
        is_system: false,
        status: 1,
      },
    ]
    pagination.total = 2
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

// 查询
const handleQuery = () => {
  pagination.page = 1
  loadRoleList()
}

// 重置
const handleReset = () => {
  queryForm.keyword = ''
  queryForm.status = undefined
  handleQuery()
}

// 新增
const handleAdd = () => {
  dialogTitle.value = '新增角色'
  Object.assign(formData, {
    id: undefined,
    role_code: '',
    role_name: '',
    description: '',
    sort_order: 0,
    status: 1,
  })
  dialogVisible.value = true
}

// 编辑
const handleEdit = (row: any) => {
  dialogTitle.value = '编辑角色'
  Object.assign(formData, row)
  dialogVisible.value = true
}

// 删除
const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    // TODO: 调用删除 API
    ElMessage.success('删除成功')
    loadRoleList()
  } catch {
    // 用户取消
  }
}

// 配置权限
const handlePermission = async (row: any) => {
  currentRoleId.value = row.id
  // TODO: 加载权限树和已选权限
  permissionTree.value = [
    {
      id: 1,
      name: '用户管理',
      children: [
        { id: 11, name: '用户列表' },
        { id: 12, name: '用户新增' },
        { id: 13, name: '用户编辑' },
      ],
    },
    {
      id: 2,
      name: '角色管理',
      children: [
        { id: 21, name: '角色列表' },
        { id: 22, name: '角色新增' },
      ],
    },
  ]
  checkedPermissions.value = [11, 12, 21]
  permissionDialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate()
  submitting.value = true
  try {
    // TODO: 调用创建/更新 API
    ElMessage.success(formData.id ? '更新成功' : '创建成功')
    dialogVisible.value = false
    loadRoleList()
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

// 提交权限配置
const handlePermissionSubmit = async () => {
  const checkedKeys = permissionTreeRef.value.getCheckedKeys()
  submitting.value = true
  try {
    // TODO: 调用权限配置 API
    ElMessage.success('权限配置成功')
    permissionDialogVisible.value = false
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

// 对话框关闭
const handleDialogClose = () => {
  formRef.value?.resetFields()
}

// 初始化
onMounted(() => {
  loadRoleList()
})
</script>

<style scoped lang="scss">
.role-management {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-form {
    margin-bottom: 20px;
  }
}
</style>
