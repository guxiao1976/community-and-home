// Permission store

import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Permission, Role } from '@common/types/identity';
import * as identityApi from '@/api/identity';

export const usePermissionStore = defineStore('permission', () => {
  // State
  const permissions = ref<Permission[]>([]);
  const roles = ref<Role[]>([]);
  const userPermissions = ref<string[]>([]);
  const menuPermissions = ref<Permission[]>([]);

  // Computed
  const permissionTree = computed(() => {
    return buildTree(permissions.value);
  });

  const menus = computed(() => {
    return menuPermissions.value.filter(p => p.type === 1); // Menu type
  });

  // Actions
  const loadPermissions = async (): Promise<void> => {
    const response = await identityApi.getPermissions();
    permissions.value = response || [];
  };

  const loadUserPermissions = async (userId: number): Promise<void> => {
    const response = await identityApi.getUserPermissions(userId);
    userPermissions.value = response?.permissions || [];
    menuPermissions.value = response?.menus || [];
  };

  const hasPermission = (permissionCode: string): boolean => {
    return userPermissions.value.includes(permissionCode);
  };

  const loadRoles = async (): Promise<void> => {
    const response = await identityApi.getRoles();
    roles.value = response?.list || [];
  };

  const clearPermissions = (): void => {
    permissions.value = [];
    roles.value = [];
    userPermissions.value = [];
    menuPermissions.value = [];
  };

  // Helper function to build tree structure
  function buildTree(items: Permission[], parentId: number = 0): Permission[] {
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
  }

  return {
    permissions,
    roles,
    userPermissions,
    menuPermissions,
    permissionTree,
    menus,
    loadPermissions,
    loadUserPermissions,
    hasPermission,
    loadRoles,
    clearPermissions
  };
});
