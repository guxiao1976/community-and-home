<template>
  <div class="role-permissions">
    <div class="page-header">
      <el-page-header @back="handleBack">
        <template #content>
          <h2>权限配置 - {{ roleName }}</h2>
        </template>
      </el-page-header>
    </div>

    <el-card v-loading="loading">
      <div class="permissions-content">
        <el-alert
          title="提示"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 20px"
        >
          勾选权限后，拥有该角色的用户将获得对应的菜单和按钮访问权限。权限变更将立即生效。
        </el-alert>

        <permission-tree
          v-if="permissions.length > 0"
          :permissions="permissions"
          :checked-ids="checkedPermissionIds"
          @update:checked-ids="handlePermissionChange"
        />

        <el-empty v-else description="暂无权限数据" />
      </div>

      <div class="actions">
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
import { ElMessage } from 'element-plus';
import type { Permission } from '@common/types/identity';
import * as identityApi from '@/api/identity';
import PermissionTree from '@/components/business/PermissionTree.vue';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const submitting = ref(false);
const roleName = ref('');
const permissions = ref<Permission[]>([]);
const checkedPermissionIds = ref<number[]>([]);
const roleId = computed(() => Number(route.params.id));

onMounted(() => {
  loadData();
});

const loadData = async () => {
  loading.value = true;
  try {
    // Load role info
    const roleResponse = await identityApi.getRoleById(roleId.value);
    roleName.value = roleResponse?.name || '';

    // Load all permissions
    const permissionsResponse = await identityApi.getPermissions();
    permissions.value = buildTree(permissionsResponse || []);

    // Load role's current permissions
    const rolePermissionsResponse = await identityApi.getRolePermissions(roleId.value);
    checkedPermissionIds.value = rolePermissionsResponse?.permissionIds || [];
  } catch (error: any) {
    ElMessage.error(error.message || '加载数据失败');
  } finally {
    loading.value = false;
  }
};

const buildTree = (items: Permission[], parentId: number = 0): Permission[] => {
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

const handlePermissionChange = (ids: number[]) => {
  checkedPermissionIds.value = ids;
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
