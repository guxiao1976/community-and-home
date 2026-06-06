<template>
  <div class="health-dashboard">
    <div class="dashboard-header">
      <h2>运行监控</h2>
      <el-tag :type="overallHealthy ? 'success' : 'danger'" size="large">
        {{ overallHealthy ? '🟢 全部正常' : '🔴 存在异常' }}
      </el-tag>
      <span class="refresh-info">每 10 秒自动刷新 · 上次: {{ lastRefresh }}</span>
    </div>

    <ServicePanel :services="data?.services || []" />
    <ContainerPanel :containers="data?.containers || []" />
    <ModelPanel :models="data?.ai_models || []" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { getHealthStatus, type HealthResponse } from '@/api/monitoring'
import ServicePanel from './ServicePanel.vue'
import ContainerPanel from './ContainerPanel.vue'
import ModelPanel from './ModelPanel.vue'

const data = ref<HealthResponse | null>(null)
const lastRefresh = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const overallHealthy = computed(() => data.value?.overall_status === 'healthy')

async function fetchHealth() {
  try {
    const res = await getHealthStatus()
    data.value = res.data
    lastRefresh.value = new Date().toLocaleTimeString()
  } catch (e) {
    console.error('Failed to fetch health status:', e)
  }
}

onMounted(() => {
  fetchHealth()
  timer = setInterval(fetchHealth, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.health-dashboard {
  padding: 20px;
}
.dashboard-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.refresh-info {
  color: #909399;
  font-size: 13px;
  margin-left: auto;
}
</style>
