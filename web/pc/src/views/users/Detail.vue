<template>
  <div class="user-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button @click="handleBack">返回</el-button>
          <span>用户详情</span>
          <el-button type="primary" v-permission="'user:update'" @click="handleEdit">编辑</el-button>
        </div>
      </template>

      <div v-if="user" class="detail-content">
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="用户ID">
            {{ user.id }}
          </el-descriptions-item>
          <el-descriptions-item label="手机号">
            {{ user.phone }}
          </el-descriptions-item>
          <el-descriptions-item label="昵称">
            {{ user.nickname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="用户类型">
            <el-tag :type="user.userType === 1 ? 'warning' : 'info'">
              {{ user.userType === 1 ? '员工' : '业主' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="user.status === 1 ? 'success' : 'danger'">
              {{ user.status === 1 ? '正常' : user.status === 2 ? '禁用' : '锁定' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="实名状态">
            <el-tag :type="user.verificationStatus === 1 ? 'success' : user.verificationStatus === 2 ? 'danger' : 'info'">
              {{ user.verificationStatus === 0 ? '未认证' : user.verificationStatus === 1 ? '已认证' : '已拒绝' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="行政范围" :span="2">
            {{ user.scope || '全国' }}
          </el-descriptions-item>
          <el-descriptions-item label="最后登录">
            {{ user.lastLoginAt || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ user.createdAt }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">角色与权限</el-divider>

        <div class="roles-section">
          <div class="section-header">
            <span>已分配角色</span>
            <el-button type="primary" size="small" v-permission="'user:role'" @click="handleAssignRole">
              分配角色
            </el-button>
          </div>
          <el-table :data="userRoles" v-loading="rolesLoading" stripe size="small">
            <el-table-column prop="name" label="角色名称" />
            <el-table-column prop="code" label="角色编码" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column label="系统角色" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.isSystem" type="warning" size="small">系统</el-tag>
                <el-tag v-else type="info" size="small">自定义</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.status === 1" type="success" size="small">启用</el-tag>
                <el-tag v-else type="danger" size="small">禁用</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button
                  link
                  type="danger"
                  size="small"
                  :disabled="row.isSystem"
                  @click="handleRemoveRole(row)"
                >
                  移除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="permissions-section" style="margin-top: 20px">
          <div class="section-header">
            <span>有效权限</span>
            <el-button type="primary" size="small" link @click="showPermissions = !showPermissions">
              {{ showPermissions ? '收起' : '展开' }}
            </el-button>
          </div>
          <el-collapse-transition>
            <div v-show="showPermissions">
              <permission-tree
                v-if="effectivePermissions.length > 0"
                :permissions="effectivePermissions"
                :checked-ids="effectivePermissions.map(p => p.id)"
              />
              <el-empty v-else description="暂无权限数据" />
            </div>
          </el-collapse-transition>
        </div>
      </div>
    </el-card>

    <!-- Assign Role Dialog -->
    <el-dialog v-model="assignDialogVisible" title="分配角色" width="500px">
      <el-checkbox-group v-model="selectedRoleIds">
        <el-checkbox
          v-for="role in availableRoles"
          :key="role.id"
          :value="role.id"
          :label="role.id"
          style="margin-bottom: 8px"
        >
          {{ role.name }}
          <el-tag v-if="role.isSystem" type="warning" size="small" style="margin-left: 8px">系统</el-tag>
          <span style="color: #999; margin-left: 8px">{{ role.code }}</span>
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="handleConfirmAssign">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserById, getUserRoles, assignUserRoles, removeUserRole, getUserPermissions } from '@/api/identity'
import PermissionTree from '@/components/business/PermissionTree.vue'
import type { User, Role, Permission } from '@common/types/identity'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const rolesLoading = ref(false)
const assigning = ref(false)
const user = ref<User | null>(null)
const userRoles = ref<Role[]>([])
const effectivePermissions = ref<Permission[]>([])
const showPermissions = ref(false)
const assignDialogVisible = ref(false)
const selectedRoleIds = ref<number[]>([])
const allRoles = ref<Role[]>([])

const userId = Number(route.params.id)

const availableRoles = computed(() => {
  const assignedIds = new Set(userRoles.value.map(r => r.id))
  return allRoles.value.filter(r => !assignedIds.has(r.id))
})

const loadUser = async () => {
  loading.value = true
  try {
    const response = await getUserById(userId)
    user.value = response
  } catch (error) {
    ElMessage.error('加载用户信息失败')
    handleBack()
  } finally {
    loading.value = false
  }
}

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    const response = await getUserRoles(userId)
    userRoles.value = response?.roles || []
  } catch {
    userRoles.value = []
  } finally {
    rolesLoading.value = false
  }
}

const loadPermissions = async () => {
  try {
    const response = await getUserPermissions(userId)
    effectivePermissions.value = response?.menus || []
  } catch {
    effectivePermissions.value = []
  }
}

const loadAllRoles = async () => {
  try {
    const { getRoles } = await import('@/api/identity')
    const response = await getRoles({ page: 1, page_size: 100 })
    allRoles.value = response?.list || []
  } catch {
    allRoles.value = []
  }
}

const handleAssignRole = () => {
  selectedRoleIds.value = []
  assignDialogVisible.value = true
}

const handleConfirmAssign = async () => {
  if (selectedRoleIds.value.length === 0) {
    ElMessage.warning('请选择至少一个角色')
    return
  }
  assigning.value = true
  try {
    await assignUserRoles(userId, selectedRoleIds.value)
    ElMessage.success('角色分配成功')
    assignDialogVisible.value = false
    loadRoles()
    loadPermissions()
  } catch (error: any) {
    ElMessage.error(error.message || '分配失败')
  } finally {
    assigning.value = false
  }
}

const handleRemoveRole = async (row: Role) => {
  try {
    await ElMessageBox.confirm(`确定要移除角色 "${row.name}" 吗？`, '提示', { type: 'warning' })
    await removeUserRole(userId, row.id)
    ElMessage.success('角色已移除')
    loadRoles()
    loadPermissions()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '移除失败')
    }
  }
}

const handleBack = () => {
  router.push({ name: 'UserList' })
}

const handleEdit = () => {
  router.push({ name: 'UserEdit', params: { id: userId } })
}

onMounted(() => {
  loadUser()
  loadRoles()
  loadPermissions()
  loadAllRoles()
})
</script>

<style scoped>
.user-detail {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.card-header span {
  flex: 1;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
}

.detail-content {
  margin-top: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: 500;
  font-size: 14px;
}
</style>
