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
