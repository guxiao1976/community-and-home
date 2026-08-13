<template>
  <div class="pipeline-list-page">
    <div class="page-header">
      <h2>内容审核管线配置</h2>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>新建管线
      </el-button>
    </div>

    <el-card>
      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="pipeline_name" label="管线名称" width="180" />
        <el-table-column label="生产" width="70" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_production === 1" type="success" size="small">生产中</el-tag>
            <span v-else style="color:#999">-</span>
          </template>
        </el-table-column>
        <el-table-column label="AC引擎" width="80" align="center">
          <template #default="{ row }">
            <span :style="{ color: row.ac_enabled ? '#67c23a' : '#999' }">{{ row.ac_enabled ? '✓' : '✗' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="审核模型" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ templateName(row.large_model_template_id) || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="升级规则" width="100">
          <template #default="{ row }">
            <span style="font-size:12px;color:#666">AC命中 → 大模型</span>
          </template>
        </el-table-column>
        <el-table-column label="终判逻辑" width="110">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ verdictLabel(row.final_verdict_logic) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click.stop="handleTestOpen(row as PipelineConfig)">测试</el-button>
            <el-button link type="primary" @click.stop="handleEdit(row as PipelineConfig)">编辑</el-button>
            <el-button link type="primary" @click.stop="handleCopy(row as PipelineConfig)">复制</el-button>
            <el-button link type="warning" @click.stop="handleSetProduction(row as PipelineConfig)">设生产</el-button>
            <el-button link type="danger" @click.stop="handleDelete(row as PipelineConfig)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadPipelines"
          @current-change="loadPipelines"
        />
      </div>
    </el-card>

    <!-- 测试弹窗 -->
    <el-dialog v-model="testDialogVisible" :title="`管线测试 — ${testPipelineName}`" width="900px" :close-on-click-modal="false">
      <el-input
        v-model="testContent"
        type="textarea"
        :rows="3"
        placeholder="输入要检测的文本内容，如：毛主席是伟大的领导"
        maxlength="500"
        show-word-limit
      />
      <div style="text-align:right;margin:12px 0">
        <el-button type="primary" :loading="testLoading" @click="handleTestRun">执行测试</el-button>
      </div>
      <el-table v-if="testLayers.length" :data="testLayers" stripe size="small">
        <el-table-column prop="layer" label="层级" width="80" />
        <el-table-column label="调用" width="70" align="center">
          <template #default="{ row }"><el-tag :type="row.called ? 'success' : 'info'" size="small">{{ row.called ? '是' : '否' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="结果" width="80" align="center">
          <template #default="{ row }"><el-tag :type="row.passed ? 'success' : 'danger'" size="small">{{ row.called ? (row.passed ? '通过' : '拦截') : '-' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="置信度" width="90" align="center">
          <template #default="{ row }">{{ row.called ? (row.confidence != null ? (row.confidence * 100).toFixed(0) + '%' : '-') : '-' }}</template>
        </el-table-column>
        <el-table-column prop="reason" label="原因/返回值" min-width="200" show-overflow-tooltip />
        <el-table-column label="耗时" width="80" align="center">
          <template #default="{ row }">{{ row.called ? row.latencyMs + 'ms' : '-' }}</template>
        </el-table-column>
      </el-table>
      <div v-if="testFinalVerdict" style="margin-top:12px;text-align:right">
        最终判定：
        <el-tag :type="testFinalVerdict === 'pass' ? 'success' : 'danger'" size="large">{{ testFinalVerdict === 'pass' ? '通过' : '拦截' }}</el-tag>
      </div>
    </el-dialog>

    <!-- 新建/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="800px"
      :close-on-click-modal="false"
      @close="resetForm"
    >
      <el-form label-width="80px" class="dialog-form">
        <el-form-item label="管线名称" required>
          <el-input v-model="form.pipeline_name" placeholder="如：生产环境审核管线" maxlength="200" />
        </el-form-item>
      </el-form>

      <el-row :gutter="16" style="margin-bottom:16px">
        <el-col :span="12">
          <el-card shadow="never">
            <template #header><div style="display:flex;justify-content:space-between;align-items:center"><span>AC 引擎</span><el-switch v-model="formAcEnabled" :active-value="1" :inactive-value="0" /></div></template>
            <el-form label-width="80px" size="small">
              <el-form-item label="严重度≥"><el-select v-model="formAcSeverity" :disabled="!formAcEnabled" style="width:100%"><el-option :value="1" label="1 - 高危"/><el-option :value="2" label="2 - 中危"/><el-option :value="3" label="3 - 低危"/></el-select></el-form-item>
            </el-form>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="never">
            <template #header><span>大模型（终审）</span></template>
            <el-form label-width="60px" size="small">
              <el-form-item label="模板"><el-select v-model="formLargeModelTemplateId" placeholder="选择审核提示词模板" style="width:100%"><el-option v-for="t in templateOptions" :key="t.id" :label="`${t.name}（${t.model_name}）`" :value="String(t.id)"/></el-select></el-form-item>
            </el-form>
          </el-card>
        </el-col>
      </el-row>
      <p style="color:#999;font-size:12px;margin:0 0 12px">AC引擎命中敏感词 → 直接调用大模型终审（两层架构）</p>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable */
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import {
  listPipelines, createPipeline, updatePipeline, deletePipeline,
  getPipeline, activatePipeline, testPipeline,
} from '@/api/moderation';
import { getModerationTemplates } from '@/api/aimodel';
import type { PipelineConfig } from '@common/types/moderation';

// ── 表格 ──
const loading = ref(false);
const tableData = ref<PipelineConfig[]>([]);

// ── 测试弹窗 ──
const testDialogVisible = ref(false);
const testPipelineName = ref('');
const testPipelineId = ref('');
const testConfig = ref<PipelineConfig | null>(null);
const testContent = ref('');
const testLoading = ref(false);
const testLayers = ref<{ layer: string; called: boolean; passed: boolean; confidence: number | null; reason: string; latencyMs: number }[]>([]);
const testFinalVerdict = ref('');
const pagination = reactive({ page: 1, pageSize: 10, total: 0 });

// ── 模板名映射 ──
const templateMap = ref<Record<string, string>>({});
const templateName = (id: string | undefined) => {
  if (!id) return '';
  return templateMap.value[id] || id;
};

// ── 弹窗 ──
const dialogVisible = ref(false);
const dialogTitle = ref('');
const submitting = ref(false);
const isEditing = ref(false);
const form = reactive<{ id: string; pipeline_id: string; pipeline_name: string }>({ id: '', pipeline_id: '', pipeline_name: '' });
const formAcEnabled = ref(1);
const formAcSeverity = ref(1);
const formLargeModelTemplateId = ref('');
const templateOptions = ref<any[]>([]);

// ── 辅助 ──
const verdictLabel = (v: string) =>
  ({ last_model_wins: '末模胜出', large_overrides: '大模优先', small_overrides: '小模优先' } as Record<string, string>)[v] || v;

const defaultForm = () => {
  form.id = ''; form.pipeline_id = ''; form.pipeline_name = '';
  formAcEnabled.value = 1; formAcSeverity.value = 1; formLargeModelTemplateId.value = '';
};

const resetForm = () => { defaultForm(); };

const populateForm = (config: PipelineConfig) => {
  form.id = config.id || '';
  form.pipeline_id = config.pipeline_id;
  form.pipeline_name = config.pipeline_name;
  formAcEnabled.value = config.ac_enabled ?? 1;
  formAcSeverity.value = config.ac_severity_threshold ?? 1;
  formLargeModelTemplateId.value = config.large_model_template_id || '';
};

// ── 加载数据 ──
const loadPipelines = async () => {
  loading.value = true;
  try {
    const resp = await listPipelines({ page: pagination.page, page_size: pagination.pageSize });
    tableData.value = resp.list || [];
    pagination.total = resp.total || 0;
  } catch (e: any) {
    ElMessage.error(e.message || '加载列表失败');
  } finally {
    loading.value = false;
  }
};

const loadTemplates = async () => {
  try {
    const resp: any = await getModerationTemplates();
    const list = resp?.data?.templates || resp?.templates || [];
    templateOptions.value = list;
    for (const t of list) templateMap.value[String(t.id)] = t.name;
  } catch { /* 静默 */ }
};

// ── 表格行操作 ──

const handleCreate = () => {
  isEditing.value = false;
  dialogTitle.value = '新建管线';
  defaultForm();
  dialogVisible.value = true;
};

const handleEdit = async (row: PipelineConfig) => {
  isEditing.value = true;
  dialogTitle.value = '编辑管线';
  try {
    const config = await getPipeline(row.pipeline_id);
    populateForm(config);
  } catch (e: any) {
    ElMessage.error('加载管线详情失败');
    return;
  }
  dialogVisible.value = true;
};

const handleCopy = async (row: PipelineConfig) => {
  try {
    const config = await getPipeline(row.pipeline_id);
    const newId = `${config.pipeline_id}_copy_${Date.now()}`;
    await createPipeline({
      pipeline_id: newId, pipeline_name: `${config.pipeline_name} (副本)`,
      description: (config as any).description,
      ac_enabled: config.ac_enabled, ac_severity_threshold: config.ac_severity_threshold,
      small_model_template_id: '', small_model_config_key: '', large_model_template_id: config.large_model_template_id, large_model_config_key: '',
      ac_to_small_condition: 'any_hit', ac_to_small_severity: 1, ac_to_small_categories: [],
      small_to_large_condition: 'always', small_to_large_confidence_threshold: 0, small_to_large_categories: [],
      final_verdict_logic: 'last_model_wins',
    });
    ElMessage.success('复制成功');
    await loadPipelines();
  } catch (e: any) { ElMessage.error(e.message || '复制失败'); }
};

const handleSetProduction = async (row: PipelineConfig) => {
  try {
    await ElMessageBox.confirm(`将「${row.pipeline_name}」设为生产环境默认管线？`, '确认操作', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' });
    await activatePipeline(row.pipeline_id);
    ElMessage.success('已设为生产配置');
    await loadPipelines();
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '操作失败'); }
};

const handleDelete = async (row: PipelineConfig) => {
  try {
    await ElMessageBox.confirm(`确定要删除管线「${row.pipeline_name}」吗？`, '确认删除', { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' });
    await deletePipeline(row.pipeline_id);
    ElMessage.success('已删除');
    await loadPipelines();
  } catch (e: any) { if (e !== 'cancel') ElMessage.error(e.message || '删除失败'); }
};

// ── 弹窗提交 ──
const handleSubmit = async () => {
  if (!form.pipeline_name.trim()) { ElMessage.warning('请输入管线名称'); return; }
  submitting.value = true;
  try {
    if (isEditing.value && form.id) {
      await updatePipeline({
        id: form.id, pipeline_name: form.pipeline_name,
        ac_enabled: formAcEnabled.value, ac_severity_threshold: formAcSeverity.value,
        small_model_template_id: '', small_model_config_key: '', large_model_template_id: formLargeModelTemplateId.value, large_model_config_key: '',
        ac_to_small_condition: 'any_hit', ac_to_small_severity: 1, ac_to_small_categories: [],
        small_to_large_condition: 'always', small_to_large_confidence_threshold: 0, small_to_large_categories: [],
        final_verdict_logic: 'last_model_wins',
      });
      ElMessage.success('管线已更新');
    } else {
      const newId = `pipeline_${Date.now()}`;
      await createPipeline({
        pipeline_id: newId, pipeline_name: form.pipeline_name.trim(),
        ac_enabled: formAcEnabled.value, ac_severity_threshold: formAcSeverity.value,
        small_model_template_id: '', small_model_config_key: '', large_model_template_id: formLargeModelTemplateId.value, large_model_config_key: '',
        ac_to_small_condition: 'any_hit', ac_to_small_severity: 1, ac_to_small_categories: [],
        small_to_large_condition: 'always', small_to_large_confidence_threshold: 0, small_to_large_categories: [],
        final_verdict_logic: 'last_model_wins',
      });
      ElMessage.success('管线已创建');
    }
    dialogVisible.value = false;
    await loadPipelines();
  } catch (e: any) { ElMessage.error(e.message || '保存失败'); }
  finally { submitting.value = false; }
};

// ── 测试弹窗 ──
const handleTestOpen = async (row: PipelineConfig) => {
  try {
    testConfig.value = await getPipeline(row.pipeline_id);
  } catch { testConfig.value = row; }
  testPipelineName.value = row.pipeline_name;
  testPipelineId.value = row.pipeline_id;
  testContent.value = '';
  testLayers.value = [];
  testFinalVerdict.value = '';
  testDialogVisible.value = true;
};

const handleTestRun = async () => {
  if (!testContent.value.trim()) { ElMessage.warning('请输入测试内容'); return; }
  testLoading.value = true;
  testLayers.value = [];
  testFinalVerdict.value = '';
  try {
    const cfg = testConfig.value;
    const resp: any = await testPipeline({
      content: testContent.value.trim(),
      pipeline_id: testPipelineId.value,
      ac_enabled: Number(cfg?.ac_enabled ?? 1),
      ac_severity_threshold: Number(cfg?.ac_severity_threshold ?? 1),
      large_model_template_id: cfg?.large_model_template_id || undefined,
      small_to_large_condition: 'always',
    });
    const layers: any[] = [];
    // AC 层
    const ac = resp.ac_result;
    layers.push({ layer: 'AC引擎', called: ac?.called ?? false, passed: ac?.passed ?? false, confidence: ac?.confidence ?? null, reason: ac?.reason || ac?.skipped_reason || '', latencyMs: ac?.latency_ms ?? 0 });
    // 小模型层
    const sm = resp.small_model_result;
    layers.push({ layer: '小模型', called: sm?.called ?? false, passed: sm?.passed ?? false, confidence: sm?.confidence ?? null, reason: sm?.reason || sm?.skipped_reason || '', latencyMs: sm?.latency_ms ?? 0 });
    // 大模型层
    const lm = resp.large_model_result;
    layers.push({ layer: '大模型', called: lm?.called ?? false, passed: lm?.passed ?? false, confidence: lm?.confidence ?? null, reason: lm?.reason || lm?.skipped_reason || '', latencyMs: lm?.latency_ms ?? 0 });
    testLayers.value = layers;
    testFinalVerdict.value = resp.final_verdict || '';
    ElMessage.success('测试完成');
  } catch (e: any) { ElMessage.error(e.message || '测试失败'); }
  finally { testLoading.value = false; }
};

onMounted(() => { loadPipelines(); loadTemplates(); });
</script>

<style scoped>
.pipeline-list-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { margin: 0; }
.pagination { margin-top: 16px; display: flex; justify-content: flex-end; }
.test-area { margin-top: 16px; }
.dialog-form { margin-bottom: 16px; }
</style>
