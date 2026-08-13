<template>
  <div class="user-roles">
    <div class="page-header">
      <el-page-header @back="handleBack">
        <template #content>
          <h2>用户角色分配 - {{ userInfo?.nickname }}</h2>
        </template>
      </el-page-header>
    </div>

    <el-card v-loading="loading">
      <el-alert
        title="用户可以在不同的数据范围（如不同小区）拥有不同的角色"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <!-- 已分配角色列表 -->
      <div class="assigned-roles">
        <h3>已分配角色</h3>
        <el-table :data="userRoles" stripe>
          <el-table-column prop="role.name" label="角色名称" />
          <el-table-column prop="role.code" label="角色编码" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getRoleStatusType(row.status)">{{ getRoleStatusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="数据范围" width="200">
            <template #default="{ row }">
              <el-tag type="success">{{ getScopeLabel(row.scopeType) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button
                v-permission="'user:assign-role'"
                link
                type="danger"
                @click="handleRemoveRole(row as UserRole)"
              >
                移除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分配新角色 -->
      <div class="assign-new-role">
        <h3>分配新角色</h3>
        <el-form :model="assignForm" label-width="120px">
          <el-form-item label="选择角色" required>
            <el-select
              v-model="assignForm.roleId"
              placeholder="请选择角色"
              style="width: 100%"
            >
              <el-option
                v-for="role in availableRoles"
                :key="role.id"
                :label="role.name"
                :value="role.id"
              >
                <span>{{ role.name }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">
                  {{ role.code }}
                </span>
              </el-option>
            </el-select>
          </el-form-item>

          <!-- TODO: 数据范围选择暂未接入小区数据 API -->
          <el-form-item label="数据范围">
            <span style="color: #909399">暂不限制（后续支持选择小区/楼栋/单元）</span>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              :loading="assigning"
              :disabled="!canAssign"
              @click="handleAssignRole"
            >
              分配角色
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable */
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { User, Role, UserRole } from '@common/types/identity'
import * as identityApi from '@/api/identity'

const route = useRoute()
const router = useRouter()

const userId = computed(() => route.params.id as string)

const loading = ref(false)
const assigning = ref(false)
const userInfo = ref<User | null>(null)
const userRoles = ref<UserRole[]>([])
const availableRoles = ref<Role[]>([])
const scopeOptions = ref<any[]>([])

const assignForm = reactive({
  roleId: '',
  scopeType: 'community' as 'community' | 'building' | 'unit' | 'grid',
  scopeId: null as string | null
})

const cascaderProps = {
  value: 'id',
  label: 'name',
  children: 'children',
  checkStrictly: true
}

const canAssign = computed(() => {
  return !!assignForm.roleId
})

onMounted(() => {
  loadUserInfo()
  loadUserRoles()
  loadAvailableRoles()
  loadScopeData()
})

const loadUserInfo = async () => {
  try {
    const response = await identityApi.getUserById(userId.value)
    userInfo.value = response
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户信息失败')
  }
}

const loadUserRoles = async () => {
  loading.value = true
  try {
    const response = await identityApi.getUserRoles(userId.value)
    userRoles.value = response?.roles || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户角色失败')
  } finally {
    loading.value = false
  }
}

const loadAvailableRoles = async () => {
  try {
    const response = await identityApi.getRoles({ page: 1, page_size: 100 })
    availableRoles.value = response?.roles || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色列表失败')
  }
}

const loadScopeData = async () => {
  // TODO: 接入小区/楼栋/单元数据 API 后启用
  scopeOptions.value = []
}

const handleScopeTypeChange = () => {
  assignForm.scopeId = null
  // 根据不同的 scopeType 可以加载不同的数据源
  loadScopeData()
}

const getScopeLabel = (scopeType: string) => {
  const map: Record<string, string> = {
    community: '小区',
    building: '楼栋',
    unit: '单元',
    grid: '网格',
    global: '全局'
  }
  return map[scopeType] || scopeType
}

const getScopeEntityName = (scopeType: string, scopeId: string) => {
  // TODO: 从缓存或API获取实体名称
  return scopeId || '-'
}

// 个体角色生命周期状态展示
const getRoleStatusLabel = (status: number) => {
  const map: Record<number, string> = {
    0: '未认证',
    1: '待审核',
    2: '已认证',
    3: '已驳回',
    4: '已过期'
  }
  return map[status] ?? '未知'
}

const getRoleStatusType = (status: number): 'primary' | 'success' | 'warning' | 'info' | 'danger' => {
  const map: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    0: 'info',
    1: 'warning',
    2: 'success',
    3: 'danger',
    4: 'danger'
  }
  return map[status] ?? 'info'
}

const handleAssignRole = async () => {
  if (!canAssign.value) return

  assigning.value = true
  try {
    await identityApi.assignUserRole({
      userId: userId.value,
      roleId: assignForm.roleId,
      scopeType: 'global',
      scopeId: '0'
    })

    ElMessage.success('角色分配成功')

    // 重置表单
    assignForm.roleId = ''
    assignForm.scopeType = 'community'
    assignForm.scopeId = null

    // 重新加载用户角色列表
    loadUserRoles()
  } catch (error: any) {
    ElMessage.error(error.message || '角色分配失败')
  } finally {
    assigning.value = false
  }
}

const handleRemoveRole = async (userRole: UserRole) => {
  try {
    await ElMessageBox.confirm(
      `确定要移除用户在"${getScopeEntityName(userRole.scopeType, userRole.scopeId)}"的"${userRole.role.name}"角色吗？`,
      '提示',
      { type: 'warning' }
    )

    await identityApi.revokeUserRole({
      userId: userId.value,
      roleId: userRole.role.id,
      scopeType: userRole.scopeType,
      scopeId: userRole.scopeId
    })

    ElMessage.success('角色移除成功')
    loadUserRoles()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '角色移除失败')
    }
  }
}

const handleBack = () => {
  router.back()
}
</script>

<style scoped lang="scss">
.user-roles {
  .page-header {
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 20px;
      font-weight: 500;
    }
  }

  .assigned-roles {
    margin-bottom: 40px;

    h3 {
      margin: 0 0 16px 0;
      font-size: 16px;
      font-weight: 500;
    }
  }

  .assign-new-role {
    padding-top: 30px;
    border-top: 1px solid var(--el-border-color);

    h3 {
      margin: 0 0 20px 0;
      font-size: 16px;
      font-weight: 500;
    }
  }
}
</style>
