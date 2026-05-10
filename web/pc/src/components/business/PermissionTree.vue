<template>
  <div class="permission-tree">
    <el-tree
      ref="treeRef"
      :data="treeData"
      :props="treeProps"
      :default-checked-keys="checkedKeys"
      show-checkbox
      node-key="id"
      :check-strictly="false"
      @check="handleCheck"
    >
      <template #default="{ node: _node, data }">
        <span class="tree-node">
          <el-icon v-if="data.icon" class="node-icon">
            <component :is="data.icon" />
          </el-icon>
          <span class="node-label">{{ data.name }}</span>
          <el-tag v-if="data.type === 2" size="small" type="info">按钮</el-tag>
        </span>
      </template>
    </el-tree>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import type { Permission } from '@common/types/identity';
import type { ElTree } from 'element-plus';

interface Props {
  permissions: Permission[];
  checkedIds?: number[];
}

interface Emits {
  (e: 'update:checkedIds', ids: number[]): void;
}

const props = withDefaults(defineProps<Props>(), {
  checkedIds: () => []
});

const emit = defineEmits<Emits>();

const treeRef = ref<InstanceType<typeof ElTree>>();
const checkedKeys = ref<number[]>([]);

const treeProps = {
  children: 'children',
  label: 'name'
};

const treeData = ref<Permission[]>([]);

// Watch for permissions changes
watch(() => props.permissions, (newVal) => {
  treeData.value = newVal;
}, { immediate: true });

// Watch for checked IDs changes
watch(() => props.checkedIds, (newVal) => {
  checkedKeys.value = newVal;
  if (treeRef.value) {
    treeRef.value.setCheckedKeys(newVal);
  }
}, { immediate: true });

const handleCheck = () => {
  if (treeRef.value) {
    const checkedNodes = treeRef.value.getCheckedKeys() as number[];
    const halfCheckedNodes = treeRef.value.getHalfCheckedKeys() as number[];
    emit('update:checkedIds', [...checkedNodes, ...halfCheckedNodes]);
  }
};

// Expose methods
defineExpose({
  getCheckedKeys: () => treeRef.value?.getCheckedKeys() || [],
  getHalfCheckedKeys: () => treeRef.value?.getHalfCheckedKeys() || [],
  setCheckedKeys: (keys: number[]) => treeRef.value?.setCheckedKeys(keys)
});
</script>

<style scoped lang="scss">
.permission-tree {
  :deep(.el-tree) {
    background: transparent;
  }

  .tree-node {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;

    .node-icon {
      font-size: 16px;
      color: var(--el-color-primary);
    }

    .node-label {
      flex: 1;
    }
  }
}
</style>
