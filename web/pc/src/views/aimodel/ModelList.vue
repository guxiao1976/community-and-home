<template>
  <div class="model-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>模型配置列表</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新增模型
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="模型名称" min-width="150" />
        <el-table-column prop="display_name" label="显示名称" min-width="120" />
        <el-table-column prop="provider" label="提供商" width="120">
          <template #default="{ row }">
            <el-tag :type="getProviderType(row.provider)">
              {{ row.provider }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="max_tokens" label="最大Token" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleHealthCheck(row)">
              <el-icon><CircleCheck /></el-icon>
              健康检查
            </el-button>
            <el-button link type="primary" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
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
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Edit, Delete, CircleCheck } from '@element-plus/icons-vue';
import { getModelConfigs, deleteModelConfig, triggerHealthCheck } from '@/api/aimodel';
import type { ModelConfig } from '@/api/aimodel';

const router = useRouter();
const loading = ref(false);
const tableData = ref<ModelConfig[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
});

const getProviderType = (provider: string) => {
  const types: Record<string, string> = {
    openai: 'primary',
    claude: 'success',
    ollama: 'warning'
  };
  return types[provider] || 'info';
};

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getModelConfigs({
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    });
    tableData.value = res.data.list || [];
    pagination.value.total = res.data.total || 0;
  } catch (error) {
    ElMessage.error('获取模型列表失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  router.push('/aimodel/models/create');
};

const handleEdit = (row: ModelConfig) => {
  router.push(`/aimodel/models/${row.id}/edit`);
};

const handleDelete = async (row: ModelConfig) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模型 "${row.model_name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    await deleteModelConfig(row.id);
    ElMessage.success('删除成功');
    fetchData();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
      console.error(error);
    }
  }
};

const handleHealthCheck = async (row: ModelConfig) => {
  loading.value = true;
  try {
    const res = await triggerHealthCheck(row.id);
    if (res.data.status === 1) {
      ElMessage.success(`健康检查成功，响应时间: ${res.data.response_time}ms`);
    } else {
      ElMessage.warning(`健康检查失败: ${res.data.error_message}`);
    }
  } catch (error) {
    ElMessage.error('健康检查失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchData();
});
</script>

<style scoped lang="scss">
.model-list {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
