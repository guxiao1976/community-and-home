<template>
  <div class="division-counts-container">
    <el-card>
      <template #header>
        <span>社区/村数量统计</span>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="昨日数据" name="yesterday">
          <StatsPanel :data-loader="loadYesterdayData" />
        </el-tab-pane>
        <el-tab-pane label="实时数据" name="realtime">
          <StatsPanel :data-loader="loadRealtimeData" lazy />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getDivisionCounts, getDivisionCountsRealtime } from '@/api/masterdata'
import type { DivisionCountItem } from '@common/types/masterdata'
import StatsPanel from './StatsPanel.vue'

const activeTab = ref('yesterday')

function loadYesterdayData(parentId?: number) {
  return getDivisionCounts(parentId ? { parent_id: parentId } : undefined)
    .then(res => res?.list || [])
}

function loadRealtimeData(parentId?: number) {
  return getDivisionCountsRealtime(parentId ? { parent_id: parentId } : undefined)
    .then(res => res?.list || [])
}
</script>

<style scoped lang="scss">
.division-counts-container {
  padding: 20px;
}
</style>
