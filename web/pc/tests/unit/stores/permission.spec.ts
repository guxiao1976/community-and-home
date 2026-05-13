import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { usePermissionStore } from '@/stores/permission';
import * as identityApi from '@/api/identity';

vi.mock('@/api/identity');

describe('Permission Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it('should initialize with default state', () => {
    const store = usePermissionStore();

    expect(store.userPermissions).toEqual([]);
    expect(store.menuPermissions).toEqual([]);
    expect(store.userRoles).toEqual([]);
    expect(store.isLoaded).toBe(false);
    expect(store.currentUserId).toBeNull();
  });

  describe('loadUserPermissionsAndMenus', () => {
    it('should load permissions and menus for a user', async () => {
      vi.mocked(identityApi.getUserPermissions).mockResolvedValue({
        permissions: ['user:list', 'user:create'],
        menus: [
          { id: 1, name: 'User', code: 'user', type: 1, parentId: 0, path: '/users', sortOrder: 1, status: 1 },
          { id: 2, name: 'User List', code: 'user:list', type: 2, parentId: 1, path: '', sortOrder: 1, status: 1 }
        ]
      } as any);

      const store = usePermissionStore();
      await store.loadUserPermissionsAndMenus(5);

      expect(store.userPermissions).toEqual(['user:list', 'user:create']);
      expect(store.menuPermissions).toHaveLength(2);
      expect(store.isLoaded).toBe(true);
      expect(store.currentUserId).toBe(5);
    });

    it('should handle API errors gracefully', async () => {
      vi.mocked(identityApi.getUserPermissions).mockRejectedValue(new Error('Network error'));

      const store = usePermissionStore();
      await store.loadUserPermissionsAndMenus(5);

      expect(store.userPermissions).toEqual([]);
      expect(store.menuPermissions).toEqual([]);
      expect(store.isLoaded).toBe(false);
    });

    it('should handle null response', async () => {
      vi.mocked(identityApi.getUserPermissions).mockResolvedValue(null as any);

      const store = usePermissionStore();
      await store.loadUserPermissionsAndMenus(5);

      expect(store.userPermissions).toEqual([]);
      expect(store.menuPermissions).toEqual([]);
    });
  });

  describe('loadUserRoles', () => {
    it('should load user roles', async () => {
      vi.mocked(identityApi.getUserRoles).mockResolvedValue({
        roles: [
          { id: 1, name: 'Admin', code: 'admin', isSystem: true, status: 1 }
        ]
      } as any);

      const store = usePermissionStore();
      await store.loadUserRoles(5);

      expect(store.userRoles).toHaveLength(1);
      expect(store.userRoles[0].name).toBe('Admin');
    });

    it('should handle API errors gracefully', async () => {
      vi.mocked(identityApi.getUserRoles).mockRejectedValue(new Error('Network error'));

      const store = usePermissionStore();
      await store.loadUserRoles(5);

      expect(store.userRoles).toEqual([]);
    });
  });

  describe('hasPermission', () => {
    it('should return true when user has the permission', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list', 'user:create'];

      expect(store.hasPermission('user:list')).toBe(true);
      expect(store.hasPermission('user:create')).toBe(true);
    });

    it('should return false when user does not have the permission', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list'];

      expect(store.hasPermission('user:delete')).toBe(false);
    });

    it('should return false when permissions are empty', () => {
      const store = usePermissionStore();

      expect(store.hasPermission('user:list')).toBe(false);
    });
  });

  describe('hasAnyPermission', () => {
    it('should return true when user has any of the permissions', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list', 'user:create'];

      expect(store.hasAnyPermission(['user:list', 'user:delete'])).toBe(true);
    });

    it('should return true when user has all of the permissions', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list', 'user:create'];

      expect(store.hasAnyPermission(['user:list', 'user:create'])).toBe(true);
    });

    it('should return false when user has none of the permissions', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list'];

      expect(store.hasAnyPermission(['user:delete', 'user:update'])).toBe(false);
    });

    it('should return false for empty codes array', () => {
      const store = usePermissionStore();
      store.userPermissions = ['user:list'];

      expect(store.hasAnyPermission([])).toBe(false);
    });
  });

  describe('clearPermissions', () => {
    it('should reset all state', async () => {
      vi.mocked(identityApi.getUserPermissions).mockResolvedValue({
        permissions: ['user:list'],
        menus: []
      } as any);

      const store = usePermissionStore();
      await store.loadUserPermissionsAndMenus(5);
      await store.loadUserRoles(5);

      expect(store.isLoaded).toBe(true);

      store.clearPermissions();

      expect(store.userPermissions).toEqual([]);
      expect(store.menuPermissions).toEqual([]);
      expect(store.userRoles).toEqual([]);
      expect(store.isLoaded).toBe(false);
      expect(store.currentUserId).toBeNull();
    });
  });

  describe('buildTree', () => {
    it('should build a tree from flat permissions', async () => {
      vi.mocked(identityApi.getPermissions).mockResolvedValue([
        { id: 1, name: 'User', code: 'user', type: 1, parentId: 0, sortOrder: 1, status: 1 },
        { id: 2, name: 'User List', code: 'user:list', type: 2, parentId: 1, sortOrder: 1, status: 1 },
        { id: 3, name: 'Role', code: 'role', type: 1, parentId: 0, sortOrder: 2, status: 1 }
      ] as any);

      const store = usePermissionStore();
      await store.loadPermissions();

      expect(store.permissionTree).toHaveLength(2);
      expect(store.permissionTree[0].children).toHaveLength(1);
      expect(store.permissionTree[0].code).toBe('user');
    });
  });
});
