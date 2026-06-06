<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">微服务运行状态</span>
    </template>
    <el-table :data="services" stripe size="small">
      <el-table-column prop="display_name" label="服务名称" width="140" />
      <el-table-column label="API 端口" width="120" align="center">
        <template #default="{ row }">
          <span>{{ row.api_status === 'healthy' ? '🟢' : row.api_status === 'unhealthy' ? '🔴' : '⚪' }}</span>
          {{ row.api_port || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="RPC 端口" width="120" align="center">
        <template #default="{ row }">
          <span>{{ row.rpc_status === 'healthy' ? '🟢' : row.rpc_status === 'unhealthy' ? '🔴' : '⚪' }}</span>
          {{ row.rpc_port || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="异常信息" min-width="200">
        <template #default="{ row }">
          <span v-if="row.api_error" class="error-text">{{ row.api_error }}</span>
          <span v-if="row.rpc_error" class="error-text">{{ row.rpc_error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { ServiceHealth } from '@/api/monitoring'

defineProps<{ services: ServiceHealth[] }>()
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; display: block; }
</style>
