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
  const currentUserId = ref<string | null>(null);
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
    permissions.value = response?.permissions || [];
  };

  const loadUserPermissionsAndMenus = async (userId: string): Promise<void> => {
    currentUserId.value = userId;
    try {
      const response = await identityApi.getUserPermissions(userId);
      // API 返回 { permissionCodes: string[] }，修正字段名
      userPermissions.value = response?.permissionCodes || [];
      loaded.value = true;
    } catch {
      userPermissions.value = [];
      loaded.value = false;
    }
  };

  const loadUserRoles = async (userId: string): Promise<void> => {
    try {
      const response = await identityApi.getUserRoles(userId);
      userRoles.value = (response?.roles || []).map((ur) => ur.role as Role);
    } catch {
      userRoles.value = [];
    }
  };

  const loadRoles = async (): Promise<void> => {
    const response = await identityApi.getRoles();
    roles.value = response?.roles || [];
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
  function buildTree(items: Permission[], parentId: string = '0'): Permission[] {
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
