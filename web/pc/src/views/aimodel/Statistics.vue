<template>
  <div class="statistics">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>使用统计</span>
          <el-space>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              @change="fetchData"
            />
            <el-select
              v-model="selectedModel"
              placeholder="选择模型"
              clearable
              @change="fetchData"
              style="width: 200px;"
            >
              <el-option
                v-for="model in modelList"
                :key="model.id"
                :label="model.model_name"
                :value="model.id"
              />
            </el-select>
          </el-space>
        </div>
      </template>

      <!-- Summary Cards -->
      <el-row :gutter="20" style="margin-bottom: 20px;">
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="总调用次数" :value="summary.totalCalls">
              <template #suffix>次</template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="成功率" :value="summary.successRate" :precision="2">
              <template #suffix>%</template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="总消耗Token" :value="summary.totalTokens">
              <template #suffix>个</template>
            </el-statistic>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card shadow="hover">
            <el-statistic title="总成本" :value="summary.totalCost" :precision="2">
              <template #prefix>$</template>
            </el-statistic>
          </el-card>
        </el-col>
      </el-row>

      <!-- Statistics Table -->
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="model_name" label="模型名称" min-width="150" />
        <el-table-column prop="total_calls" label="总调用" width="100" align="right" />
        <el-table-column prop="success_calls" label="成功" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #67c23a;">{{ row.success_calls }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="failed_calls" label="失败" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #f56c6c;">{{ row.failed_calls }}</span>
          </template>
        </el-table-column>
        <el-table-column label="成功率" width="100" align="right">
          <template #default="{ row }">
            {{ ((row.success_calls / row.total_calls) * 100).toFixed(2) }}%
          </template>
        </el-table-column>
        <el-table-column prop="total_tokens" label="Token消耗" width="120" align="right">
          <template #default="{ row }">
            {{ formatNumber(row.total_tokens) }}
          </template>
        </el-table-column>
        <el-table-column prop="total_cost" label="成本($)" width="120" align="right">
          <template #default="{ row }">
            {{ row.total_cost.toFixed(4) }}
          </template>
        </el-table-column>
        <el-table-column prop="avg_latency" label="平均延迟(ms)" width="130" align="right">
          <template #default="{ row }">
            {{ row.avg_latency.toFixed(0) }}
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { getUsageStatistics, getModelConfigs } from '@/api/aimodel';
import type { UsageStatistics, ModelConfig } from '@/api/aimodel';

interface StatisticsWithModel extends UsageStatistics {
  model_name?: string;
}

const loading = ref(false);
const dateRange = ref<[Date, Date]>();
const selectedModel = ref<number>();
const modelList = ref<ModelConfig[]>([]);

const tableData = ref<StatisticsWithModel[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
});

const summary = computed(() => {
  const totalCalls = tableData.value.reduce((sum, item) => sum + item.total_calls, 0);
  const successCalls = tableData.value.reduce((sum, item) => sum + item.success_calls, 0);
  const totalTokens = tableData.value.reduce((sum, item) => sum + item.total_tokens, 0);
  const totalCost = tableData.value.reduce((sum, item) => sum + item.total_cost, 0);
  const successRate = totalCalls > 0 ? (successCalls / totalCalls) * 100 : 0;

  return {
    totalCalls,
    successRate,
    totalTokens,
    totalCost
  };
});

const formatNumber = (num: number) => {
  return num.toLocaleString();
};

const fetchModelList = async () => {
  try {
    const res = await getModelConfigs({ page: 1, page_size: 100 });
    modelList.value = res.data.list || [];
  } catch (error) {
    console.error('获取模型列表失败:', error);
  }
};

const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    };

    if (selectedModel.value) {
      params.model_config_id = selectedModel.value;
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0].toISOString().split('T')[0];
      params.end_date = dateRange.value[1].toISOString().split('T')[0];
    }

    const res = await getUsageStatistics(params);

    // Merge model names
    const dataWithModelNames = (res.data.list || []).map(item => {
      const model = modelList.value.find(m => m.id === item.model_config_id);
      return {
        ...item,
        model_name: model?.model_name || `模型 #${item.model_config_id}`
      };
    });

    tableData.value = dataWithModelNames;
    pagination.value.total = res.data.total || 0;
  } catch (error) {
    ElMessage.error('获取统计数据失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  await fetchModelList();
  fetchData();
});
</script>

<style scoped lang="scss">
.statistics {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  :deep(.el-statistic__head) {
    font-size: 14px;
    color: #909399;
  }

  :deep(.el-statistic__content) {
    font-size: 24px;
    font-weight: 600;
  }
}
</style>
