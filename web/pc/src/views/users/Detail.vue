<template>
  <div class="user-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button @click="handleBack">返回</el-button>
          <span>用户详情</span>
          <el-button type="primary" @click="handleEdit">编辑</el-button>
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
            <h4>已分配角色</h4>
            <el-button type="primary" size="small" @click="showAssignDialog">
              分配角色
            </el-button>
          </div>

          <el-table :data="userRoles" v-loading="rolesLoading" stripe border size="small">
            <el-table-column prop="name" label="角色名称" />
            <el-table-column prop="code" label="角色编码" />
            <el-table-column prop="description" label="描述" show-overflow-tooltip />
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.isSystem" type="warning" size="small">系统</el-tag>
                <el-tag v-else type="info" size="small">自定义</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-popconfirm
                  title="确定移除该角色吗？"
                  @confirm="handleRemoveRole(row.id)"
                >
                  <template #reference>
                    <el-button link type="danger" size="small">移除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!rolesLoading && userRoles.length === 0" description="暂未分配角色" />

          <div v-if="userRoles.length > 0" class="permissions-preview">
            <h4>权限概览</h4>
            <el-collapse v-model="expandedPermissions">
              <el-collapse-item
                v-for="role in userRoles"
                :key="role.id"
                :title="`${role.name} 权限`"
                :name="role.id"
              >
                <el-tree
                  v-if="rolePermissionMap[role.id]?.length"
                  :data="rolePermissionMap[role.id]"
                  :props="{ children: 'children', label: 'name' }"
                  node-key="id"
                  default-expand-all
                >
                  <template #default="{ data }">
                    <span class="perm-node">
                      <el-tag v-if="data.type === 2" size="small" type="info">按钮</el-tag>
                      <el-tag v-else size="small" type="success">菜单</el-tag>
                      <span>{{ data.name }}</span>
                      <span class="perm-code">{{ data.code }}</span>
                    </span>
                  </template>
                </el-tree>
                <el-empty v-else description="该角色暂无权限" :image-size="60" />
              </el-collapse-item>
            </el-collapse>
          </div>
        </div>
      </div>
    </el-card>

    <!-- Assign Role Dialog -->
    <el-dialog
      v-model="assignDialogVisible"
      title="分配角色"
      width="500px"
    >
      <el-select
        v-model="selectedRoleId"
        placeholder="请选择要分配的角色"
        filterable
        style="width: 100%"
      >
        <el-option
          v-for="role in availableRoles"
          :key="role.id"
          :label="`${role.name} (${role.code})`"
          :value="role.id"
          :disabled="role.status !== 1"
        />
      </el-select>
      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="handleAssignRole">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getUserById, getUserRoles, getUserPermissions, assignUserRole, removeUserRole, getAllRoles } from '@/api/identity'
import type { User, Role, Permission } from '@common/types/identity'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const rolesLoading = ref(false)
const assigning = ref(false)
const user = ref<User | null>(null)
const userRoles = ref<Role[]>([])
const allRoles = ref<Role[]>([])
const rolePermissionMap = ref<Record<number, Permission[]>>({})
const expandedPermissions = ref<number[]>([])

const assignDialogVisible = ref(false)
const selectedRoleId = ref<number | null>(null)

const userId = Number(route.params.id)

const availableRoles = ref<Role[]>([])

const loadUser = async () => {
  loading.value = true
  try {
    const response = await getUserById(userId)
    user.value = response
  } catch (error) {
    ElMessage.error('加载用户信息失败')
    console.error('Load user error:', error)
    handleBack()
  } finally {
    loading.value = false
  }
}

const loadUserRoles = async () => {
  rolesLoading.value = true
  try {
    const response = await getUserRoles(userId)
    userRoles.value = response || []
  } catch (error) {
    ElMessage.error('加载用户角色失败')
    console.error('Load user roles error:', error)
  } finally {
    rolesLoading.value = false
  }
}

const loadAllRoles = async () => {
  try {
    const response = await getAllRoles()
    allRoles.value = response || []
  } catch {
    try {
      const response = await getRoles({ page: 1, page_size: 100 })
      allRoles.value = response?.list || []
    } catch (error) {
      console.error('Load all roles error:', error)
    }
  }
}

const loadRolePermissions = async () => {
  try {
    const response = await getUserPermissions(userId)
    const menus = response?.menus || []
    const grouped: Record<number, Permission[]> = {}

    for (const perm of menus) {
      if (perm.parentId === 0 && perm.children) {
        for (const role of userRoles.value) {
          if (!grouped[role.id]) grouped[role.id] = []
        }
        grouped[0] = grouped[0] || []
        grouped[0].push(perm)
      }
    }

    rolePermissionMap.value = grouped
  } catch (error) {
    console.error('Load role permissions error:', error)
  }
}

const showAssignDialog = () => {
  const assignedIds = new Set(userRoles.value.map(r => r.id))
  availableRoles.value = allRoles.value.filter(r => !assignedIds.has(r.id))
  selectedRoleId.value = null

  if (availableRoles.value.length === 0) {
    ElMessage.info('没有可分配的角色')
    return
  }

  assignDialogVisible.value = true
}

const handleAssignRole = async () => {
  if (!selectedRoleId.value) {
    ElMessage.warning('请选择角色')
    return
  }

  assigning.value = true
  try {
    await assignUserRole(userId, selectedRoleId.value)
    ElMessage.success('角色分配成功')
    assignDialogVisible.value = false
    loadUserRoles()
  } catch (error: any) {
    ElMessage.error(error.message || '分配角色失败')
  } finally {
    assigning.value = false
  }
}

const handleRemoveRole = async (roleId: number) => {
  try {
    await removeUserRole(userId, roleId)
    ElMessage.success('角色已移除')
    loadUserRoles()
  } catch (error: any) {
    ElMessage.error(error.message || '移除角色失败')
  }
}

const handleBack = () => {
  router.push({ name: 'UserList' })
}

const handleEdit = () => {
  router.push({ name: 'UserEdit', params: { id: userId } })
}

onMounted(async () => {
  await Promise.all([loadUser(), loadUserRoles(), loadAllRoles()])
  if (userRoles.value.length > 0) {
    await loadRolePermissions()
  }
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

.roles-section {
  margin-top: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-header h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 500;
}

.permissions-preview {
  margin-top: 24px;
}

.permissions-preview h4 {
  margin: 0 0 12px 0;
  font-size: 15px;
  font-weight: 500;
}

.perm-node {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
}

.perm-code {
  color: #909399;
  font-size: 12px;
  font-family: monospace;
}
</style>
