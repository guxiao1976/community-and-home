# Implementation Plan: Content Moderation Test UI

**Created**: 2026-05-12  
**OpenSpec Change**: content-moderation-test-ui  
**Estimated Time**: 4-6 hours  
**Complexity**: Medium

## Overview

Add a content moderation testing interface to the admin panel with two capabilities:
1. **Text Testing**: Input text (≤500 chars) and view multi-layer moderation results (traditional tech, small model, large model)
2. **Image Testing**: Upload images (≤5MB) and view moderation results from small model (large model reserved)

This enables developers and operators to monitor, test, and optimize the content moderation microservice.

## File Structure

```
web/pc/
├── src/
│   ├── router/index.ts                          [MODIFY] Add moderation route
│   ├── api/moderation.ts                        [CREATE]  API service layer
│   ├── views/moderation/
│   │   └── ModerationTest.vue                   [CREATE]  Main page component
│   ├── components/moderation/
│   │   ├── TextTestTab.vue                      [CREATE]  Text testing tab
│   │   ├── ImageTestTab.vue                     [CREATE]  Image testing tab
│   │   └── ModerationResult.vue                 [CREATE]  Result display component
│   └── components/layout/AppSidebar.vue         [MODIFY] Add menu item
├── vite.config.ts                               [MODIFY] Add proxy config
└── web/common/types/moderation.d.ts             [CREATE]  TypeScript types

Total: 3 modifications, 6 new files
```

## Prerequisites

- Content moderation microservice running on port 8890
- Endpoints available:
  - POST /api/moderation/text/check
  - POST /api/moderation/image/check

## Implementation Steps

### Step 1: Add Proxy Configuration

**File**: `web/pc/vite.config.ts`

**Action**: Add moderation service proxy to the server.proxy section.

```typescript
// In vite.config.ts, modify the server.proxy object:
server: {
  port: 3000,
  proxy: {
    '/api/identity': {
      target: 'http://localhost:8888',
      changeOrigin: true
    },
    '/api/masterdata': {
      target: 'http://localhost:8889',
      changeOrigin: true
    },
    '/api/moderation': {
      target: 'http://localhost:8890',
      changeOrigin: true
    }
  }
}
```

**Test**: Restart dev server and verify no errors.

```bash
cd web/pc
npm run dev
# Should see: "Local: http://localhost:3000/"
```

**Commit**: `chore: add moderation service proxy to vite config`

---

### Step 2: Create TypeScript Type Definitions

**File**: `web/common/types/moderation.d.ts`

**Action**: Create comprehensive type definitions for moderation API.

```typescript
// Traditional technology check result
export interface TraditionalCheckResult {
  passed: boolean;
  reason?: string;
  keywords?: string[];
  score?: number;
}

// Small model check result
export interface SmallModelCheckResult {
  passed: boolean;
  confidence: number;
  categories?: string[];
  reason?: string;
}

// Large model check result
export interface LargeModelCheckResult {
  passed: boolean;
  confidence: number;
  analysis?: string;
  categories?: string[];
  reason?: string;
}

// Text moderation request
export interface TextModerationRequest {
  content: string;
  userId?: string;
  scene?: string;
}

// Text moderation response
export interface TextModerationResponse {
  requestId: string;
  finalResult: boolean;
  traditional: TraditionalCheckResult;
  smallModel: SmallModelCheckResult;
  largeModel?: LargeModelCheckResult;
  processingTime: number;
}

// Image moderation request
export interface ImageModerationRequest {
  imageBase64: string;
  userId?: string;
  scene?: string;
}

// Image moderation response
export interface ImageModerationResponse {
  requestId: string;
  finalResult: boolean;
  smallModel: SmallModelCheckResult;
  largeModel?: LargeModelCheckResult;
  processingTime: number;
}
```

**Test**: Run TypeScript compiler to verify no syntax errors.

```bash
cd web/pc
npx tsc --noEmit
# Should complete with no errors
```

**Commit**: `feat: add moderation type definitions`

---

### Step 3: Create API Service Layer

**File**: `web/pc/src/api/moderation.ts`

**Action**: Create API service with proper timeout configuration.

```typescript
import request from '@/utils/request';
import type {
  TextModerationRequest,
  TextModerationResponse,
  ImageModerationRequest,
  ImageModerationResponse
} from '@common/types/moderation';

/**
 * Check text content for moderation
 * @param data Text moderation request
 * @returns Moderation result with multi-layer checks
 */
export function checkText(data: TextModerationRequest) {
  return request.post<TextModerationResponse>(
    '/api/moderation/text/check',
    data,
    { timeout: 30000 }
  );
}

/**
 * Check image content for moderation
 * @param data Image moderation request (base64 encoded)
 * @returns Moderation result with model checks
 */
export function checkImage(data: ImageModerationRequest) {
  return request.post<ImageModerationResponse>(
    '/api/moderation/image/check',
    data,
    { timeout: 60000 }
  );
}
```

**Test**: Import the module to verify no errors.

```bash
cd web/pc
npx tsc --noEmit
# Should complete with no errors
```

**Commit**: `feat: add moderation API service layer`

---

### Step 4: Create Result Display Component

**File**: `web/pc/src/components/moderation/ModerationResult.vue`

**Action**: Create reusable component for displaying multi-layer results.

```vue
<template>
  <div class="moderation-result">
    <el-alert
      :type="result.finalResult ? 'success' : 'error'"
      :title="result.finalResult ? '审核通过' : '审核未通过'"
      :closable="false"
      show-icon
    />

    <div class="result-meta">
      <span>请求ID: {{ result.requestId }}</span>
      <span>处理时间: {{ result.processingTime }}ms</span>
    </div>

    <el-collapse v-model="activeNames" class="result-layers">
      <!-- Traditional Technology Layer -->
      <el-collapse-item
        v-if="result.traditional"
        name="traditional"
        title="传统技术检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.traditional.passed ? 'success' : 'danger'">
                {{ result.traditional.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.score" label="评分">
              {{ result.traditional.score }}
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.keywords" label="关键词">
              <el-tag
                v-for="keyword in result.traditional.keywords"
                :key="keyword"
                type="warning"
                size="small"
                style="margin-right: 8px"
              >
                {{ keyword }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.reason" label="原因">
              {{ result.traditional.reason }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>

      <!-- Small Model Layer -->
      <el-collapse-item
        v-if="result.smallModel"
        name="smallModel"
        title="小模型检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.smallModel.passed ? 'success' : 'danger'">
                {{ result.smallModel.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="置信度">
              <el-progress
                :percentage="Math.round(result.smallModel.confidence * 100)"
                :color="getConfidenceColor(result.smallModel.confidence)"
              />
            </el-descriptions-item>
            <el-descriptions-item v-if="result.smallModel.categories" label="分类">
              <el-tag
                v-for="category in result.smallModel.categories"
                :key="category"
                type="info"
                size="small"
                style="margin-right: 8px"
              >
                {{ category }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.smallModel.reason" label="原因">
              {{ result.smallModel.reason }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>

      <!-- Large Model Layer -->
      <el-collapse-item
        v-if="result.largeModel"
        name="largeModel"
        title="大模型检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.largeModel.passed ? 'success' : 'danger'">
                {{ result.largeModel.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="置信度">
              <el-progress
                :percentage="Math.round(result.largeModel.confidence * 100)"
                :color="getConfidenceColor(result.largeModel.confidence)"
              />
            </el-descriptions-item>
            <el-descriptions-item v-if="result.largeModel.categories" label="分类">
              <el-tag
                v-for="category in result.largeModel.categories"
                :key="category"
                type="info"
                size="small"
                style="margin-right: 8px"
              >
                {{ category }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.largeModel.analysis" label="分析">
              {{ result.largeModel.analysis }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Raw JSON View -->
    <el-collapse v-model="showRaw" class="raw-data">
      <el-collapse-item name="raw" title="原始数据 (JSON)">
        <pre>{{ JSON.stringify(result, null, 2) }}</pre>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { TextModerationResponse, ImageModerationResponse } from '@common/types/moderation';

interface Props {
  result: TextModerationResponse | ImageModerationResponse;
}

defineProps<Props>();

const activeNames = ref(['traditional', 'smallModel', 'largeModel']);
const showRaw = ref<string[]>([]);

const getConfidenceColor = (confidence: number) => {
  if (confidence >= 0.8) return '#67c23a';
  if (confidence >= 0.5) return '#e6a23c';
  return '#f56c6c';
};
</script>

<style scoped>
.moderation-result {
  margin-top: 20px;
}

.result-meta {
  display: flex;
  justify-content: space-between;
  margin: 16px 0;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
  color: #606266;
}

.result-layers {
  margin-bottom: 16px;
}

.layer-content {
  padding: 12px;
}

.raw-data pre {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
```

**Test**: Component syntax check.

```bash
cd web/pc
npx tsc --noEmit
```

**Commit**: `feat: add moderation result display component`

---

### Step 5: Create Text Test Tab Component

**File**: `web/pc/src/components/moderation/TextTestTab.vue`

**Action**: Create text testing interface with validation.

```vue
<template>
  <div class="text-test-tab">
    <el-form :model="form" label-width="100px">
      <el-form-item label="测试文本">
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="8"
          placeholder="请输入要测试的文本内容（不超过500字）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="用户ID">
        <el-input
          v-model="form.userId"
          placeholder="可选，用于审计日志"
          clearable
        />
      </el-form-item>

      <el-form-item label="场景标识">
        <el-input
          v-model="form.scene"
          placeholder="可选，如 comment、post、chat"
          clearable
        />
      </el-form-item>

      <el-form-item>
        <el-button
          type="primary"
          :loading="loading"
          :disabled="!form.content.trim()"
          @click="handleSubmit"
        >
          开始检测
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <ModerationResult v-if="result" :result="result" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { checkText } from '@/api/moderation';
import ModerationResult from './ModerationResult.vue';
import type { TextModerationRequest, TextModerationResponse } from '@common/types/moderation';

const loading = ref(false);
const result = ref<TextModerationResponse | null>(null);

const form = reactive<TextModerationRequest>({
  content: '',
  userId: '',
  scene: ''
});

const handleSubmit = async () => {
  if (!form.content.trim()) {
    ElMessage.warning('请输入测试文本');
    return;
  }

  if (form.content.length > 500) {
    ElMessage.warning('文本长度不能超过500字');
    return;
  }

  loading.value = true;
  result.value = null;

  try {
    const response = await checkText({
      content: form.content,
      userId: form.userId || undefined,
      scene: form.scene || undefined
    });
    result.value = response;
    ElMessage.success('检测完成');
  } catch (error: any) {
    ElMessage.error(error.message || '检测失败，请稍后重试');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  form.content = '';
  form.userId = '';
  form.scene = '';
  result.value = null;
};
</script>

<style scoped>
.text-test-tab {
  padding: 20px;
}
</style>
```

**Test**: Component syntax check.

```bash
cd web/pc
npx tsc --noEmit
```

**Commit**: `feat: add text test tab component`

---

### Step 6: Create Image Test Tab Component

**File**: `web/pc/src/components/moderation/ImageTestTab.vue`

**Action**: Create image upload and testing interface.

```vue
<template>
  <div class="image-test-tab">
    <el-form :model="form" label-width="100px">
      <el-form-item label="测试图片">
        <el-upload
          class="image-uploader"
          :auto-upload="false"
          :show-file-list="false"
          :on-change="handleImageChange"
          accept="image/jpeg,image/png,image/gif"
        >
          <img v-if="imageUrl" :src="imageUrl" class="preview-image" />
          <el-icon v-else class="image-uploader-icon"><Plus /></el-icon>
        </el-upload>
        <div class="upload-tip">
          支持 JPG、PNG、GIF 格式，文件大小不超过 5MB
        </div>
      </el-form-item>

      <el-form-item label="用户ID">
        <el-input
          v-model="form.userId"
          placeholder="可选，用于审计日志"
          clearable
        />
      </el-form-item>

      <el-form-item label="场景标识">
        <el-input
          v-model="form.scene"
          placeholder="可选，如 avatar、post、comment"
          clearable
        />
      </el-form-item>

      <el-form-item>
        <el-button
          type="primary"
          :loading="loading"
          :disabled="!form.imageBase64"
          @click="handleSubmit"
        >
          开始检测
        </el-button>
        <el-button @click="handleReset">重置</el-button>
      </el-form-item>
    </el-form>

    <ModerationResult v-if="result" :result="result" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElMessage } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import { checkImage } from '@/api/moderation';
import ModerationResult from './ModerationResult.vue';
import type { ImageModerationRequest, ImageModerationResponse } from '@common/types/moderation';
import type { UploadFile } from 'element-plus';

const loading = ref(false);
const result = ref<ImageModerationResponse | null>(null);
const imageUrl = ref('');

const form = reactive<ImageModerationRequest>({
  imageBase64: '',
  userId: '',
  scene: ''
});

const handleImageChange = (file: UploadFile) => {
  const rawFile = file.raw;
  if (!rawFile) return;

  // Validate file type
  const validTypes = ['image/jpeg', 'image/png', 'image/gif'];
  if (!validTypes.includes(rawFile.type)) {
    ElMessage.error('只支持 JPG、PNG、GIF 格式的图片');
    return;
  }

  // Validate file size (5MB)
  const maxSize = 5 * 1024 * 1024;
  if (rawFile.size > maxSize) {
    ElMessage.error('图片大小不能超过 5MB');
    return;
  }

  // Convert to base64
  const reader = new FileReader();
  reader.onload = (e) => {
    const base64 = e.target?.result as string;
    imageUrl.value = base64;
    // Remove data URL prefix for API
    form.imageBase64 = base64.split(',')[1];
  };
  reader.readAsDataURL(rawFile);
};

const handleSubmit = async () => {
  if (!form.imageBase64) {
    ElMessage.warning('请先上传图片');
    return;
  }

  loading.value = true;
  result.value = null;

  try {
    const response = await checkImage({
      imageBase64: form.imageBase64,
      userId: form.userId || undefined,
      scene: form.scene || undefined
    });
    result.value = response;
    ElMessage.success('检测完成');
  } catch (error: any) {
    ElMessage.error(error.message || '检测失败，请稍后重试');
  } finally {
    loading.value = false;
  }
};

const handleReset = () => {
  form.imageBase64 = '';
  form.userId = '';
  form.scene = '';
  imageUrl.value = '';
  result.value = null;
};
</script>

<style scoped>
.image-test-tab {
  padding: 20px;
}

.image-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  overflow: hidden;
  transition: border-color 0.3s;
}

.image-uploader:hover {
  border-color: #409eff;
}

.image-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 178px;
  height: 178px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-image {
  width: 178px;
  height: 178px;
  object-fit: contain;
  display: block;
}

.upload-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
```

**Test**: Component syntax check.

```bash
cd web/pc
npx tsc --noEmit
```

**Commit**: `feat: add image test tab component`

---

### Step 7: Create Main Page Component

**File**: `web/pc/src/views/moderation/ModerationTest.vue`

**Action**: Create main page with tab navigation.

```vue
<template>
  <div class="moderation-test-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>内容审核测试</span>
          <el-tag type="info" size="small">开发测试工具</el-tag>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="文本测试" name="text">
          <TextTestTab />
        </el-tab-pane>

        <el-tab-pane label="图片测试" name="image">
          <ImageTestTab />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import TextTestTab from '@/components/moderation/TextTestTab.vue';
import ImageTestTab from '@/components/moderation/ImageTestTab.vue';

const activeTab = ref('text');
</script>

<style scoped>
.moderation-test-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
```

**Test**: Component syntax check.

```bash
cd web/pc
npx tsc --noEmit
```

**Commit**: `feat: add moderation test main page component`

---

### Step 8: Add Route Configuration

**File**: `web/pc/src/router/index.ts`

**Action**: Add moderation test route to the router.

Find the children array inside the MainLayout route and add:

```typescript
{
  path: '/moderation/test',
  name: 'ModerationTest',
  component: () => import('@/views/moderation/ModerationTest.vue'),
  meta: {
    title: '内容审核测试',
    requiresAuth: true
  }
}
```

**Complete modification** (insert after existing routes, before the closing bracket):

```typescript
// Around line 280, add to the children array:
{
  path: '/moderation/test',
  name: 'ModerationTest',
  component: () => import('@/views/moderation/ModerationTest.vue'),
  meta: {
    title: '内容审核测试',
    requiresAuth: true
  }
}
```

**Test**: Check router configuration loads without errors.

```bash
cd web/pc
npm run dev
# Should start without errors
```

**Commit**: `feat: add moderation test route`

---

### Step 9: Add Menu Item

**File**: `web/pc/src/components/layout/AppSidebar.vue`

**Action**: Add "内容审核" menu with "内容审核测试" submenu.

Find the `menuItems` array and add a new top-level menu item:

```typescript
{
  title: '内容审核',
  icon: 'Shield',
  children: [
    {
      title: '内容审核测试',
      path: '/moderation/test',
      icon: 'Monitor'
    }
  ]
}
```

**Complete modification** (insert after existing menu items, before the closing bracket):

```typescript
// Around line 50-60, add to menuItems array:
{
  title: '内容审核',
  icon: 'Shield',
  children: [
    {
      title: '内容审核测试',
      path: '/moderation/test',
      icon: 'Monitor'
    }
  ]
}
```

**Test**: Start dev server and verify menu appears.

```bash
cd web/pc
npm run dev
# Navigate to http://localhost:3000
# Should see "内容审核" menu with "内容审核测试" submenu
```

**Commit**: `feat: add moderation menu to sidebar`

---

### Step 10: Integration Testing

**Actions**:

1. **Start all services**:
```bash
# Terminal 1: Start moderation service (port 8890)
cd services/moderation
go run main.go

# Terminal 2: Start frontend dev server
cd web/pc
npm run dev
```

2. **Test text moderation**:
   - Navigate to http://localhost:3000/moderation/test
   - Click "文本测试" tab
   - Enter test text: "这是一段测试文本"
   - Click "开始检测"
   - Verify results display with all three layers (if available)
   - Check raw JSON view

3. **Test image moderation**:
   - Click "图片测试" tab
   - Upload a test image (< 5MB, JPG/PNG/GIF)
   - Click "开始检测"
   - Verify results display
   - Check raw JSON view

4. **Test error handling**:
   - Stop moderation service
   - Try to submit a test
   - Verify error message displays
   - Restart service and verify recovery

5. **Test validation**:
   - Text tab: Try submitting empty text (should be disabled)
   - Text tab: Enter > 500 characters (should show warning)
   - Image tab: Try uploading > 5MB file (should show error)
   - Image tab: Try uploading non-image file (should show error)

**Expected Results**:
- All validations work correctly
- API calls succeed with proper timeout
- Results display in collapsible sections
- Raw JSON view shows complete response
- Error messages are user-friendly

**Commit**: `test: verify moderation test UI integration`

---

## Testing Strategy

### Unit Testing (Optional)

If you want to add unit tests:

```bash
cd web/pc
npm install -D @vue/test-utils vitest jsdom
```

Create test files:
- `src/components/moderation/__tests__/TextTestTab.spec.ts`
- `src/components/moderation/__tests__/ImageTestTab.spec.ts`
- `src/components/moderation/__tests__/ModerationResult.spec.ts`

### Manual Testing Checklist

- [ ] Menu item appears in sidebar
- [ ] Route navigation works
- [ ] Text tab loads without errors
- [ ] Image tab loads without errors
- [ ] Text validation (empty, > 500 chars)
- [ ] Image validation (size, format)
- [ ] API call succeeds (text)
- [ ] API call succeeds (image)
- [ ] Results display correctly
- [ ] Collapsible sections work
- [ ] Raw JSON view works
- [ ] Error handling (network failure)
- [ ] Error handling (service down)
- [ ] Loading states display correctly
- [ ] Reset button works

## Rollback Plan

If issues occur, rollback in reverse order:

```bash
# Revert all commits
git log --oneline  # Find the commit before Step 1
git reset --hard <commit-hash>

# Or revert specific commits
git revert <commit-hash>
```

## Execution Options

### Option A: Subagent-Driven (Recommended)

Each step is executed by a dedicated subagent. Benefits:
- Parallel execution where possible
- Isolated context per step
- Clear progress tracking

**Command**: Let me know if you want me to execute this plan using subagents.

### Option B: Inline Execution

Execute all steps in the current session. Benefits:
- Single context
- Faster for simple changes
- Direct feedback

**Command**: Let me know if you want me to execute this plan inline.

### Option C: Manual Execution

Follow the steps manually. Benefits:
- Full control
- Learn the codebase
- Customize as needed

**Command**: Use this document as a reference guide.

## Notes

- **Proxy Configuration**: Ensure moderation service is running on port 8890
- **Type Safety**: All API calls are fully typed with TypeScript
- **Error Handling**: Axios interceptor handles response unwrapping
- **Component Auto-Import**: Element Plus components are auto-imported
- **Icon Usage**: Element Plus icons are auto-imported (Shield, Monitor, Plus)
- **Large Model**: Interface is ready but large model endpoint may not be deployed yet

## Next Steps After Implementation

1. Add audit logging for test requests
2. Add request history/cache
3. Add batch testing capability
4. Add performance metrics dashboard
5. Add export results feature (CSV/JSON)

---

**Plan Status**: Ready for execution  
**Last Updated**: 2026-05-12
