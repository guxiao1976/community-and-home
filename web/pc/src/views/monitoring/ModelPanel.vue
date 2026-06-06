<template>
  <el-card class="panel" shadow="hover">
    <template #header>
      <span class="panel-title">AI 模型状态</span>
    </template>
    <el-table :data="models" stripe size="small">
      <el-table-column prop="display_name" label="模型名称" width="180" />
      <el-table-column prop="provider" label="提供商" width="120" />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <span>{{ row.status === 'healthy' ? '🟢' : '🔴' }}</span>
          {{ row.status === 'healthy' ? '正常' : '异常' }}
        </template>
      </el-table-column>
      <el-table-column label="异常信息" min-width="250">
        <template #default="{ row }">
          <span v-if="row.error" class="error-text">{{ row.error }}</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import type { AiModelHealth } from '@/api/monitoring'

defineProps<{ models: AiModelHealth[] }>()
</script>

<style scoped>
.panel { margin-bottom: 20px; }
.panel-title { font-weight: 600; font-size: 16px; }
.error-text { color: #f56c6c; font-size: 12px; }
</style>
