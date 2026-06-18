<template>
  <div class="masterdata-query">
    <div class="page-header">
      <h2>数据查询</h2>
      <p class="page-desc">只读查询主数据信息</p>
    </div>

    <el-card>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="小区&村" name="residential">
          <ResidentialQuery />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ResidentialQuery from './ResidentialQuery.vue'

const activeTab = ref('residential')
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.masterdata-query {
  .page-header {
    margin-bottom: 20px;

    h2 {
      margin: 0 0 4px 0;
      font-size: 20px;
      font-weight: 500;
    }

    .page-desc {
      margin: 0;
      color: $text-secondary;
      font-size: 14px;
    }
  }
}
</style>
