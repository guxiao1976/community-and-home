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
  const userRoles = ref<Role[]>([]);
  const currentUserId = ref<number | null>(null);
  const loaded = ref(false);

  // Computed
  const permissionTree = computed(() => {
    return buildTree(permissions.value);
  });

  const menus = computed(() => {
    return menuPermissions.value.filter(p => p.type === 1);
  });

  const isLoaded = computed(() => loaded.value);

  // Actions
  const loadPermissions = async (): Promise<void> => {
    const response = await identityApi.getPermissions();
    permissions.value = response || [];
  };

  const loadUserPermissionsAndMenus = async (userId: number): Promise<void> => {
    currentUserId.value = userId;
    try {
      const response = await identityApi.getUserPermissions(userId);
      userPermissions.value = response?.permissions || [];
      menuPermissions.value = response?.menus || [];
      loaded.value = true;
    } catch {
      userPermissions.value = [];
      menuPermissions.value = [];
      loaded.value = false;
    }
  };

  const loadUserRoles = async (userId: number): Promise<void> => {
    try {
      const response = await identityApi.getUserRoles(userId);
      userRoles.value = response?.roles || [];
    } catch {
      userRoles.value = [];
    }
  };

  const loadRoles = async (): Promise<void> => {
    const response = await identityApi.getRoles();
    roles.value = response?.list || [];
  };

  const hasPermission = (permissionCode: string): boolean => {
    return userPermissions.value.includes(permissionCode);
  };

  const hasAnyPermission = (codes: string[]): boolean => {
    return codes.some(code => userPermissions.value.includes(code));
  };

  const clearPermissions = (): void => {
    permissions.value = [];
    roles.value = [];
    userPermissions.value = [];
    menuPermissions.value = [];
    userRoles.value = [];
    currentUserId.value = null;
    loaded.value = false;
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
    userRoles,
    currentUserId,
    loaded,
    permissionTree,
    menus,
    isLoaded,
    loadPermissions,
    loadUserPermissionsAndMenus,
    loadUserRoles,
    loadRoles,
    hasPermission,
    hasAnyPermission,
    clearPermissions
  };
});
