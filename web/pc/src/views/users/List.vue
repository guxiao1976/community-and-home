<template>
  <div class="users-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" v-permission="'user:create'" @click="handleCreate">创建用户</el-button>
        </div>
      </template>

      <!-- Filters -->
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="手机号">
          <el-input v-model="filterForm.phone" placeholder="请输入手机号" clearable />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="filterForm.nickname" placeholder="请输入昵称" clearable />
        </el-form-item>
        <el-form-item label="用户类型">
          <el-select v-model="filterForm.user_type" placeholder="请选择" clearable style="width: 150px">
            <el-option label="普通用户" :value="0" />
            <el-option label="管理员" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="filterForm.status" placeholder="请选择" clearable style="width: 150px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilter">查询</el-button>
          <el-button @click="handleResetFilter">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- Table -->
      <el-table :data="userStore.users" v-loading="userStore.loading" border stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="phone" label="手机号" width="140">
          <template #default="{ row }">
            {{ maskPhone(row.phone) }}
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="user_type" label="用户类型" width="120">
          <template #default="{ row }">
            <el-tag :type="row.user_type === 1 ? 'warning' : 'info'">
              {{ row.user_type === 1 ? '管理员' : '普通用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scope" label="行政范围" width="200" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'user:detail'" link type="primary" size="small" @click="handleView(row.id)">
              查看
            </el-button>
            <el-button v-permission="'user:update'" link type="primary" size="small" @click="handleEdit(row.id)">
              编辑
            </el-button>
            <el-button
              v-permission="'user:status'"
              link
              :type="row.status === 1 ? 'warning' : 'success'"
              size="small"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <el-pagination
        v-model:current-page="userStore.filters.page"
        v-model:page-size="userStore.filters.page_size"
        :total="userStore.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadUsers"
        @current-change="loadUsers"
        class="pagination"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { getUsers, disableUser, enableUser } from '@/api/identity'
import type { User } from '@common/types/identity'

const router = useRouter()
const userStore = useUserStore()

const filterForm = ref({
  phone: '',
  nickname: '',
  user_type: undefined as number | undefined,
  status: undefined as number | undefined
})

// Mask phone number (show first 3 and last 4 digits)
const maskPhone = (phone: string) => {
  if (!phone || phone.length !== 11) return phone
  return phone.substring(0, 3) + '****' + phone.substring(7)
}

const loadUsers = async () => {
  try {
    userStore.setLoading(true)
    const params = {
      ...userStore.filters,
      ...(filterForm.value.phone && { phone: filterForm.value.phone }),
      ...(filterForm.value.nickname && { nickname: filterForm.value.nickname }),
      ...(filterForm.value.user_type !== undefined && { user_type: filterForm.value.user_type }),
      ...(filterForm.value.status !== undefined && { status: filterForm.value.status })
    }
    const response = await getUsers(params)
    userStore.setUsers(response.list, response.total)
  } catch (error) {
    ElMessage.error('加载用户列表失败')
    console.error('Load users error:', error)
  } finally {
    userStore.setLoading(false)
  }
}

const handleFilter = () => {
  userStore.setFilters({ page: 1 })
  loadUsers()
}

const handleResetFilter = () => {
  filterForm.value = {
    phone: '',
    nickname: '',
    user_type: undefined,
    status: undefined
  }
  userStore.resetFilters()
  loadUsers()
}

const handleCreate = () => {
  router.push({ name: 'UserCreate' })
}

const handleView = (id: number) => {
  router.push({ name: 'UserDetail', params: { id } })
}

const handleEdit = (id: number) => {
  router.push({ name: 'UserEdit', params: { id } })
}

const handleToggleStatus = async (user: User) => {
  const action = user.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确认${action}用户 "${user.nickname || user.phone}" 吗？`,
      `${action}确认`,
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    if (user.status === 1) {
      await disableUser(user.id)
    } else {
      await enableUser(user.id)
    }

    ElMessage.success(`用户已${action}`)
    userStore.updateUserInList(user.id, { status: user.status === 1 ? 2 : 1 })
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}用户失败`)
      console.error('Toggle user status error:', error)
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
