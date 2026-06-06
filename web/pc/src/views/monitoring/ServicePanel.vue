<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">微服务运行状态</span>
      <span class="panel-hint">API/RPC = 端口检测 · 健康 = 服务自检端点</span>
    </template>
    <el-table :data="services" stripe size="small">
      <el-table-column prop="display_name" label="服务名称" width="140" />
      <el-table-column label="API" width="90" align="center">
        <template #default="{ row }">
          <span>{{ statusDot(row.api_status) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="RPC" width="90" align="center">
        <template #default="{ row }">
          <span>{{ statusDot(row.rpc_status) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="健康" width="90" align="center">
        <template #default="{ row }">
          <span>{{ statusDot(row.health_status) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="异常信息" min-width="220">
        <template #default="{ row }">
          <span v-if="row.api_error" class="error-text">API: {{ row.api_error }}</span>
          <span v-if="row.rpc_error" class="error-text">RPC: {{ row.rpc_error }}</span>
          <span v-if="row.health_error" class="error-text">健康: {{ row.health_error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { ServiceHealth } from '@/api/monitoring'

defineProps<{ services: ServiceHealth[] }>()

function statusDot(s: string) {
  if (s === 'healthy') return '🟢'
  if (s === 'unhealthy') return '🔴'
  return '⚪'
}
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.panel-hint { color: #909399; font-size: 12px; margin-left: 12px; }
.error-text { color: #f56c6c; font-size: 12px; display: block; }
</style>
