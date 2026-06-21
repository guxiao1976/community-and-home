<template>
  <div class="user-roles">
    <div class="page-header">
      <el-page-header @back="handleBack">
        <template #content>
          <h2>用户角色分配 - {{ userInfo?.name }}</h2>
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
          <el-table-column label="数据范围" width="200">
            <template #default="{ row }">
              <el-tag type="success">{{ getScopeLabel(row.scopeType) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="范围实体" width="200">
            <template #default="{ row }">
              {{ getScopeEntityName(row.scopeType, row.scopeId) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button
                v-permission="'user:assign-role'"
                link
                type="danger"
                @click="handleRemoveRole(row)"
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

          <el-form-item label="数据范围类型" required>
            <el-radio-group v-model="assignForm.scopeType" @change="handleScopeTypeChange">
              <el-radio value="community">小区</el-radio>
              <el-radio value="building">楼栋</el-radio>
              <el-radio value="unit">单元</el-radio>
              <el-radio value="grid">网格</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="选择范围实体" required>
            <el-cascader
              v-model="assignForm.scopeId"
              :options="scopeOptions"
              :props="cascaderProps"
              placeholder="请选择具体的小区/楼栋/单元/网格"
              clearable
              filterable
              style="width: 100%"
            />
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
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { User, Role, UserRole } from '@common/types/identity'
import * as identityApi from '@/api/identity'
import * as masterDataApi from '@/api/masterdata'

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
  return assignForm.roleId && assignForm.scopeType && assignForm.scopeId
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
    userRoles.value = response?.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户角色失败')
  } finally {
    loading.value = false
  }
}

const loadAvailableRoles = async () => {
  try {
    const response = await identityApi.getRoles({ page: 1, pageSize: 100 })
    availableRoles.value = response?.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载角色列表失败')
  }
}

const loadScopeData = async () => {
  try {
    // 加载小区数据（级联：小区→楼栋→单元）
    const communities = await masterDataApi.getCommunities()
    scopeOptions.value = communities?.list || []
  } catch (error: any) {
    ElMessage.error(error.message || '加载范围数据失败')
  }
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
    grid: '网格'
  }
  return map[scopeType] || scopeType
}

const getScopeEntityName = (scopeType: string, scopeId: string) => {
  // TODO: 从缓存或API获取实体名称
  return scopeId || '-'
}

const handleAssignRole = async () => {
  if (!canAssign.value) return

  assigning.value = true
  try {
    await identityApi.assignUserRole({
      userId: userId.value,
      roleId: assignForm.roleId,
      scopeType: assignForm.scopeType,
      scopeId: assignForm.scopeId!
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
