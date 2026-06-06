<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">Docker 容器状态</span>
    </template>
    <el-table :data="containers" stripe size="small">
      <el-table-column prop="display_name" label="容器名称" width="160" />
      <el-table-column prop="image" label="镜像" width="180" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span>{{ row.status === 'healthy' ? '🟢' : '🔴' }}</span>
          {{ row.status === 'healthy' ? '运行中' : '异常' }}
        </template>
      </el-table-column>
      <el-table-column prop="running_for" label="运行时长" width="140" />
      <el-table-column label="异常信息" min-width="200">
        <template #default="{ row }">
          <span v-if="row.error" class="error-text">{{ row.error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { ContainerHealth } from '@/api/monitoring'

defineProps<{ containers: ContainerHealth[] }>()
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; }
</style>
