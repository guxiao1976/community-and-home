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
        <el-alert
          title="角色与权限功能将在Phase 7实现"
          type="info"
          :closable="false"
          show-icon
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getUserById } from '@/api/identity'
import type { User } from '@common/types/identity'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const user = ref<User | null>(null)

const userId = Number(route.params.id)

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

const handleBack = () => {
  router.push({ name: 'UserList' })
}

const handleEdit = () => {
  router.push({ name: 'UserEdit', params: { id: userId } })
}

onMounted(() => {
  loadUser()
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
</style>
