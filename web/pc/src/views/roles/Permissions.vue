<template>
  <div class="role-permissions">
    <div class="page-header">
      <el-page-header @back="handleBack">
        <template #content>
          <h2 v-if="isEditMode">权限配置 - {{ roleName }}</h2>
          <h2 v-else>权限资源</h2>
        </template>
      </el-page-header>
    </div>

    <el-card v-loading="loading">
      <div class="permissions-content">
        <!-- 编辑模式：提示 + 全选 -->
        <template v-if="isEditMode">
          <el-alert
            title="提示"
            type="info"
            :closable="false"
            show-icon
            style="margin-bottom: 20px"
          >
            勾选权限后，拥有该角色的用户将获得对应的菜单和按钮访问权限。权限变更将立即生效。权限由配置决定，勾选全部即相当于超级管理员。
          </el-alert>

          <div class="select-all-bar">
            <el-switch
              :model-value="isSelectAll"
              active-text="全选所有权限"
              inactive-text="自定义权限"
              @change="handleSelectAll"
            />
          </div>
        </template>

        <!-- 只读模式：提示 + 自动发现按钮 -->
        <div v-else class="readonly-toolbar">
          <el-alert
            title="系统全部权限资源"
            type="info"
            :closable="false"
            show-icon
          >
            以下为系统当前定义的全部权限资源，按"菜单 → 操作 → API"三级结构展示。当后端新增 API 后，点击右侧"自动发现"按钮自动注册缺失的权限节点。
          </el-alert>
          <el-button
            type="primary"
            :loading="discovering"
            :icon="RefreshRight"
            @click="handleAutoDiscover"
          >
            自动发现
          </el-button>
        </div>

        <permission-tree
          v-if="permissions.length > 0"
          :permissions="permissions"
          :checked-ids="checkedPermissionIds"
          :readonly="!isEditMode"
          @update:checked-ids="handlePermissionChange"
        />

        <el-empty v-else description="暂无权限数据" />
      </div>

      <div v-if="isEditMode" class="actions">
        <el-button @click="handleBack">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          保存
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { RefreshRight } from '@element-plus/icons-vue';
import type { Permission } from '@common/types/identity';
import * as identityApi from '@/api/identity';
import PermissionTree from '@/components/business/PermissionTree.vue';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const submitting = ref(false);
const discovering = ref(false);
const roleName = ref('');
const permissions = ref<Permission[]>([]);
const checkedPermissionIds = ref<string[]>([]);
const roleId = computed(() => route.params.id as string);

// 有 roleId → 编辑模式（角色权限配置）；无 roleId → 只读模式（权限资源浏览）
const isEditMode = computed(() => !!roleId.value);

onMounted(() => {
  loadData();
});

const loadData = async () => {
  loading.value = true;
  try {
    // Load all permissions (both modes need this)
    const permissionsResponse = await identityApi.getPermissions();
    permissions.value = buildTree(permissionsResponse?.permissions || []);

    if (isEditMode.value) {
      // 编辑模式：加载角色信息 + 已有权限
      const roleResponse = await identityApi.getRoleById(roleId.value);
      roleName.value = roleResponse?.role?.name || '';

      const rolePermissionsResponse = await identityApi.getRolePermissions(roleId.value);
      checkedPermissionIds.value = rolePermissionsResponse?.permissionIds || [];
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载数据失败');
  } finally {
    loading.value = false;
  }
};

const buildTree = (items: Permission[], parentId: string = '0'): Permission[] => {
  const tree: Permission[] = [];

  for (const item of items) {
    if (item.parentId === parentId) {
      const children = buildTree(items, item.id);
      if (children.length > 0) {
        item.children = children;
      }
      tree.push(item);
    }
  }

  return tree.sort((a, b) => a.sortOrder - b.sortOrder);
};

const handlePermissionChange = (ids: string[]) => {
  checkedPermissionIds.value = ids;
};

// Collect all leaf permission IDs (for "select all" feature)
const getAllPermissionIds = (items: Permission[]): string[] => {
  const ids: string[] = [];
  const walk = (nodes: Permission[]) => {
    for (const node of nodes) {
      ids.push(node.id);
      if (node.children && node.children.length > 0) {
        walk(node.children);
      }
    }
  };
  walk(items);
  return ids;
};

const isSelectAll = computed({
  get: () => {
    const all = getAllPermissionIds(permissions.value);
    return all.length > 0 && all.every(id => checkedPermissionIds.value.includes(id));
  },
  set: () => {} // no-op, handled by @change
});

const handleSelectAll = (val: boolean) => {
  if (val) {
    checkedPermissionIds.value = getAllPermissionIds(permissions.value);
  } else {
    checkedPermissionIds.value = [];
  }
};

const handleSubmit = async () => {
  submitting.value = true;
  try {
    await identityApi.assignRolePermissions(roleId.value, checkedPermissionIds.value);
    ElMessage.success('权限配置保存成功');
    router.back();
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败');
  } finally {
    submitting.value = false;
  }
};

const handleBack = () => {
  router.back();
};

const handleAutoDiscover = async () => {
  try {
    await ElMessageBox.confirm(
      '自动发现将扫描系统已注册的 API 路由，对比当前权限表，自动注册缺失的权限节点。是否继续？',
      '自动发现权限',
      { type: 'info', confirmButtonText: '开始扫描', cancelButtonText: '取消' }
    );
  } catch {
    return; // 用户取消
  }

  discovering.value = true;
  try {
    const resp = await identityApi.autoDiscoverPermissions();
    if (resp.total > 0) {
      ElMessage.success(resp.message);
    } else {
      ElMessage.info(resp.message);
    }
    // 重新加载权限树
    await loadData();
  } catch (error: any) {
    ElMessage.error(error.message || '自动发现失败');
  } finally {
    discovering.value = false;
  }
};
</script>

<style scoped lang="scss">
.role-permissions {
  .page-header {
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 20px;
      font-weight: 500;
    }
  }

  .select-all-bar {
    margin-bottom: 16px;
    padding: 10px 16px;
    background: var(--el-color-info-light-9);
    border-radius: 4px;
  }

  .readonly-toolbar {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 20px;

    .el-alert {
      flex: 1;
    }

    .el-button {
      flex-shrink: 0;
      margin-top: 2px;
    }
  }

  .permissions-content {
    min-height: 400px;
  }

  .actions {
    margin-top: 30px;
    padding-top: 20px;
    border-top: 1px solid var(--el-border-color);
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>
