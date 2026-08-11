<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <span>用户管理</span>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="queryForm" class="search-form">
        <el-form-item label="关键词">
          <el-input
            v-model="queryForm.keyword"
            placeholder="手机号/昵称"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryForm.status" placeholder="请选择" clearable>
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedUsers.length > 0">
        <el-button type="warning" @click="handleBatchUpdateStatus(2)">
          批量禁用
        </el-button>
        <el-button type="success" @click="handleBatchUpdateStatus(1)">
          批量启用
        </el-button>
        <span style="margin-left: 10px">已选 {{ selectedUsers.length }} 个用户</span>
      </div>

      <!-- 用户列表 -->
      <el-table
        :data="userList"
        border
        style="width: 100%; margin-top: 20px"
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="用户ID" width="180" />
        <el-table-column prop="phone" label="手机号" width="150" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="avatar_url" label="头像" width="80">
          <template #default="{ row }">
            <el-avatar :src="row.avatar_url" v-if="row.avatar_url" />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="credit_score" label="信用分" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_time" label="注册时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看</el-button>
            <el-button link type="primary" @click="handleAssignRole(row)">
              分配角色
            </el-button>
            <el-button
              link
              :type="row.status === 1 ? 'warning' : 'success'"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
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

    <!-- 分配角色对话框 -->
    <el-dialog v-model="roleDialogVisible" title="分配角色" width="500px">
      <el-form :model="roleForm" label-width="100px">
        <el-form-item label="用户">
          <span>{{ currentUser?.nickname || currentUser?.phone }}</span>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="roleForm.role_ids" multiple placeholder="请选择角色">
            <el-option
              v-for="role in availableRoles"
              :key="role.id"
              :label="role.role_name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="数据范围">
          <el-select v-model="roleForm.scope_type" placeholder="请选择">
            <el-option label="小区" value="community" />
            <el-option label="楼栋" value="building" />
            <el-option label="单元" value="unit" />
          </el-select>
        </el-form-item>
        <el-form-item label="范围ID">
          <el-input v-model="roleForm.scope_id" placeholder="请输入范围ID" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRoleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// 查询表单
const queryForm = reactive({
  keyword: '',
  status: undefined as number | undefined,
})

// 用户列表
const userList = ref([])
const loading = ref(false)
const selectedUsers = ref<any[]>([])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

// 角色对话框
const roleDialogVisible = ref(false)
const currentUser = ref<any>()
const submitting = ref(false)
const roleForm = reactive({
  role_ids: [] as number[],
  scope_type: 'community',
  scope_id: '',
})

// 可用角色列表
const availableRoles = ref([
  { id: 1, role_name: '业主' },
  { id: 2, role_name: '物业管理员' },
  { id: 3, role_name: '网格员' },
])

// 加载用户列表
const loadUserList = async () => {
  loading.value = true
  try {
    // TODO: 调用 API
    // const res = await getUserList({
    //   page: pagination.page,
    //   pageSize: pagination.pageSize,
    //   ...queryForm
    // })
    // userList.value = res.data.users
    // pagination.total = res.data.total

    // 模拟数据
    userList.value = [
      {
        id: '1001',
        phone: '138****8001',
        nickname: '张三',
        avatar_url: '',
        credit_score: 100,
        status: 1,
        created_time: '2024-01-01 10:00:00',
      },
      {
        id: '1002',
        phone: '138****8002',
        nickname: '李四',
        avatar_url: '',
        credit_score: 95,
        status: 1,
        created_time: '2024-01-02 10:00:00',
      },
    ]
    pagination.total = 2
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 查询
const handleQuery = () => {
  pagination.page = 1
  loadUserList()
}

// 重置
const handleReset = () => {
  queryForm.keyword = ''
  queryForm.status = undefined
  handleQuery()
}

// 选择变化
const handleSelectionChange = (selection: any[]) => {
  selectedUsers.value = selection
}

// 查看
const handleView = (row: any) => {
  ElMessage.info('查看用户详情')
}

// 切换状态
const handleToggleStatus = async (row: any) => {
  const newStatus = row.status === 1 ? 2 : 1
  const action = newStatus === 1 ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(`确定要${action}该用户吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    // TODO: 调用 API
    ElMessage.success(`${action}成功`)
    row.status = newStatus
  } catch {
    // 用户取消
  }
}

// 批量更新状态
const handleBatchUpdateStatus = async (status: number) => {
  const action = status === 1 ? '启用' : '禁用'

  try {
    await ElMessageBox.confirm(
      `确定要${action} ${selectedUsers.value.length} 个用户吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    // TODO: 调用批量更新 API
    // await batchUpdateUsers({
    //   user_ids: selectedUsers.value.map(u => u.id),
    //   status
    // })
    ElMessage.success(`批量${action}成功`)
    loadUserList()
  } catch {
    // 用户取消
  }
}

// 分配角色
const handleAssignRole = (row: any) => {
  currentUser.value = row
  roleForm.role_ids = []
  roleForm.scope_type = 'community'
  roleForm.scope_id = ''
  roleDialogVisible.value = true
}

// 提交角色分配
const handleRoleSubmit = async () => {
  submitting.value = true
  try {
    // TODO: 调用角色分配 API
    ElMessage.success('角色分配成功')
    roleDialogVisible.value = false
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

// 初始化
onMounted(() => {
  loadUserList()
})
</script>

<style scoped lang="scss">
.user-management {
  .search-form {
    margin-bottom: 0;
  }

  .batch-actions {
    margin-top: 20px;
    padding: 10px;
    background: #f5f7fa;
    border-radius: 4px;
  }
}
</style>
