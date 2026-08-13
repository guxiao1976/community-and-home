<template>
  <div class="template-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>提示模板管理</span>
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            新增模板
          </el-button>
        </div>
      </template>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="200" />
        <el-table-column prop="name" label="模板名称" min-width="150" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag>{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="model_name" label="关联模型" min-width="150" show-overflow-tooltip />
        <el-table-column prop="content" label="模板内容" min-width="250" show-overflow-tooltip />
        <el-table-column prop="created_time" label="创建时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row as PromptTemplate)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row as PromptTemplate)">
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

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑模板' : '新增模板'"
      width="700px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="模板名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: 代码审查模板" />
        </el-form-item>

        <el-form-item label="分类" prop="category">
          <el-select v-model="formData.category" placeholder="请选择分类" allow-create filterable>
            <el-option label="代码" value="code" />
            <el-option label="文档" value="document" />
            <el-option label="翻译" value="translation" />
            <el-option label="分析" value="analysis" />
            <el-option label="合规审核" value="moderation" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>

        <el-form-item label="模型组">
          <el-select
            v-model="selectedConfigKey"
            placeholder="选择模型组（按标识符聚合）"
            filterable
            @change="onConfigKeyChange"
          >
            <el-option
              v-for="g in modelGroups"
              :key="g.config_key"
              :label="`${g.config_key} (${g.count}个模型)`"
              :value="g.config_key"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="具体模型" prop="model_name">
          <el-select
            v-model="formData.model_name"
            placeholder="选择该组下的具体模型"
            filterable
            :disabled="!selectedConfigKey"
          >
            <el-option
              v-for="m in filteredModels"
              :key="m.name"
              :label="`${m.name} (${m.display_name || ''})`"
              :value="m.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="模板内容" prop="content">
          <el-input
            v-model="formData.content"
            type="textarea"
            :rows="10"
            placeholder="请输入模板内容，可以使用 {{variable}} 作为变量占位符"
          />
        </el-form-item>

        <div v-if="detectedVariables.length > 0" class="variables-display">
          <span class="var-label">已识别变量：</span>
          <el-tag v-for="v in detectedVariables" :key="v" size="small" type="info" effect="plain" style="margin-left: 4px;">
            {{ '{' + '{' + v + '}' + '}' }}
          </el-tag>
        </div>

        <template v-if="detectedVariables.length > 0 && formData.model_name">
          <el-divider content-position="left">即时测试</el-divider>

          <el-form-item v-for="v in detectedVariables" :key="v" :label="v">
            <el-input v-model="testVariables[v]" :placeholder="`输入 ${v} 的值`" />
          </el-form-item>

          <el-form-item>
            <el-button type="success" @click="handleTestTemplate" :loading="testRunning" :disabled="!canRunTest">
              <el-icon><VideoPlay /></el-icon>
              测试运行
            </el-button>
          </el-form-item>

          <div v-if="testOutput" class="test-output" style="margin-top: 12px;">
            <el-alert title="渲染结果" :closable="false" style="margin-bottom: 8px;">
              <pre style="white-space: pre-wrap; font-size: 13px;">{{ testOutput.rendered }}</pre>
            </el-alert>
            <el-alert v-if="testOutput.response" title="模型返回" type="success" :closable="false">
              <pre style="white-space: pre-wrap; font-size: 13px;">{{ testOutput.response }}</pre>
            </el-alert>
          </div>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable */
import { ref, reactive, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { Plus, Edit, Delete, VideoPlay } from '@element-plus/icons-vue';
import { getTemplates, createTemplate, updateTemplate, deleteTemplate, getModelConfigs, testTemplate } from '@/api/aimodel';
import type { PromptTemplate, ModelConfig } from '@/api/aimodel';

const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();

const tableData = ref<PromptTemplate[]>([]);
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
});

const formData = reactive({
  id: undefined as string | undefined,
  name: '',
  category: '',
  content: '',
  model_name: ''
});

const rules: FormRules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ],
  model_name: [
    { required: true, message: '请选择模型配置', trigger: 'change' }
  ],
  content: [
    { required: true, message: '请输入模板内容', trigger: 'blur' }
  ]
};

// Model selection & variable detection & instant test
const detectedVariables = computed(() => {
  const matches = formData.content.match(/\{\{(\w+)\}\}/g) || [];
  return [...new Set(matches.map(m => m.slice(2, -2)))];
});
const testVariables = ref<Record<string, string>>({});
const testRunning = ref(false);
const testOutput = ref<any>(null);
const canRunTest = computed(() =>
  detectedVariables.value.every(v => testVariables.value[v]?.trim())
);

// Two-level model selection: config_key → model_name
const selectedConfigKey = ref('');
const allModelConfigs = ref<any[]>([]);

// Group models by config_key
const modelGroups = computed(() => {
  const map = new Map<string, number>();
  for (const m of allModelConfigs.value) {
    const key = m.config_key || m.name;
    map.set(key, (map.get(key) || 0) + 1);
  }
  return Array.from(map.entries()).map(([config_key, count]) => ({ config_key, count }));
});

// Models filtered by selected config_key
const filteredModels = computed(() => {
  if (!selectedConfigKey.value) return [];
  return allModelConfigs.value.filter(m =>
    (m.config_key || m.name) === selectedConfigKey.value
  );
});

function onConfigKeyChange(_val: string) {
  formData.model_name = ''; // Reset model selection
}

async function loadModels() {
  try {
    const res = await getModelConfigs({ page: 1, page_size: 100 } as any);
    const list = (res as any)?.data || res;
    allModelConfigs.value = list.models || [];
  } catch (e) { /* silent */ }
}

async function handleTestTemplate() {
  if (!canRunTest.value) return;
  testRunning.value = true;
  testOutput.value = null;
  try {
    const res = await testTemplate({
      model_name: formData.model_name,
      content: formData.content,
      variables: testVariables.value,
    });
    testOutput.value = (res as any)?.data || res;
  } catch (e: any) {
    testOutput.value = { rendered: '', response: `Error: ${e?.message || e}` };
  } finally {
    testRunning.value = false;
  }
}

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getTemplates({
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    });
    const list = (res as any)?.data || res;
    tableData.value = list.templates || [];
    pagination.value.total = list.total || 0;
  } catch (error) {
    ElMessage.error('获取模板列表失败');
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleCreate = () => {
  isEdit.value = false;
  testOutput.value = null;
  testVariables.value = {};
  selectedConfigKey.value = '';
  Object.assign(formData, {
    id: undefined,
    name: '',
    category: '',
    content: '',
    model_name: ''
  });
  loadModels();
  dialogVisible.value = true;
};

const handleEdit = (row: PromptTemplate) => {
  isEdit.value = true;
  testOutput.value = null;
  testVariables.value = {};
  const configKey = row.config_key || '';
  selectedConfigKey.value = configKey;
  Object.assign(formData, {
    id: row.id,
    name: row.name,
    category: row.category,
    content: row.content,
    model_name: row.model_name || ''
  });
  loadModels();
  dialogVisible.value = true;
};

const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    submitting.value = true;
    try {
      if (isEdit.value && formData.id) {
        await updateTemplate({
          id: formData.id,
          name: formData.name,
          category: formData.category,
          content: formData.content,
          config_key: selectedConfigKey.value, // 模型组标识
          model_name: formData.model_name      // 具体模型
        });
        ElMessage.success('更新成功');
      } else {
        await createTemplate({
          name: formData.name,
          category: formData.category,
          content: formData.content,
          config_key: selectedConfigKey.value, // 模型组标识
          model_name: formData.model_name       // 具体模型
        });
        ElMessage.success('创建成功');
      }

      dialogVisible.value = false;
      fetchData();
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '创建失败');
      console.error(error);
    } finally {
      submitting.value = false;
    }
  });
};

const handleDelete = async (row: PromptTemplate) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${row.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    );

    await deleteTemplate(row.id);
    ElMessage.success('删除成功');
    fetchData();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
      console.error(error);
    }
  }
};

onMounted(() => {
  fetchData();
});
</script>

<style scoped lang="scss">
.template-list {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}

.variables-display {
  margin: 8px 0 16px 100px;

  .var-label {
    color: #909399;
    font-size: 12px;
    margin-right: 4px;
  }
}

.test-output pre {
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
}
</style>
