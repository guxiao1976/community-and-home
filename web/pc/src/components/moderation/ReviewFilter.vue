<template>
  <div class="review-filter">
    <el-radio-group v-model="localSourceType" @change="onFilterChange">
      <el-radio-button value="">全部</el-radio-button>
      <el-radio-button value="notice">通知公告</el-radio-button>
      <el-radio-button value="lost_found">寻失互助</el-radio-button>
      <el-radio-button value="certification">房主认证</el-radio-button>
      <el-radio-button value="nickname">用户昵称</el-radio-button>
    </el-radio-group>

    <el-divider direction="vertical" />

    <el-radio-group v-model="localReviewStatus" @change="onFilterChange">
      <el-radio-button :value="0">未审核</el-radio-button>
      <el-radio-button :value="1">已通过</el-radio-button>
      <el-radio-button :value="2">已不通过</el-radio-button>
    </el-radio-group>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

const props = defineProps<{
  sourceType: string;
  reviewStatus: number;
}>();

const emit = defineEmits<{
  (e: 'update:sourceType', val: string): void;
  (e: 'update:reviewStatus', val: number): void;
  (e: 'change'): void;
}>();

const localSourceType = ref(props.sourceType);
const localReviewStatus = ref(props.reviewStatus);

watch(() => props.sourceType, (v) => { localSourceType.value = v; });
watch(() => props.reviewStatus, (v) => { localReviewStatus.value = v; });

function onFilterChange() {
  emit('update:sourceType', localSourceType.value);
  emit('update:reviewStatus', localReviewStatus.value);
  emit('change');
}
</script>
