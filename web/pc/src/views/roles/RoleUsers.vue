<template>
  <div class="role-users">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色用户 - {{ roleName }}</span>
          <el-button @click="handleBack">返回</el-button>
        </div>
      </template>

      <el-table :data="users" v-loading="loading" border stripe>
        <el-table-column prop="userId" label="用户ID" width="200" />
        <el-table-column label="手机号" width="150">
          <template #default="{ row }">{{ maskPhone(row.phone) }}</template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" />
      </el-table>

      <el-empty v-if="!loading && users.length === 0" description="该角色下暂无用户" />

      <el-pagination
        v-if="total > 0"
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
        class="pagination"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as identityApi from '@/api/identity'

const route = useRoute()
const router = useRouter()

const roleId = route.params.id as string
const roleName = ref('')
const users = ref<Array<{ userId: string; phone: string; nickname: string }>>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const maskPhone = (phone: string) => {
  if (!phone || phone.length !== 11) return phone
  return phone.substring(0, 3) + '****' + phone.substring(7)
}

const loadData = async () => {
  loading.value = true
  try {
    const resp = await identityApi.getRoleUsers(roleId, page.value, pageSize.value)
    users.value = resp?.users || []
    total.value = resp?.total || 0
  } catch (e: any) {
    ElMessage.error(e.message || '加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const role = await identityApi.getRoleById(roleId)
    roleName.value = role?.role?.name || roleId
  } catch { roleName.value = roleId }
  loadData()
})

const handleBack = () => router.push({ name: 'RoleList' })
</script>

<style scoped>
.role-users { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 20px; justify-content: flex-end; }
</style>
