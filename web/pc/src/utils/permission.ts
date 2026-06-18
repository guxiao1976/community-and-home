// Permission check helper functions

/**
 * Check if user has specific permission
 */
export function hasPermission(permission: string | string[], userPermissions: string[]): boolean {
  if (!userPermissions || userPermissions.length === 0) {
    return false;
  }

  if (Array.isArray(permission)) {
    // Check if user has at least one of the permissions (OR logic)
    return permission.some(p => userPermissions.includes(p));
  }

  // Check single permission
  return userPermissions.includes(permission);
}

/**
 * Check if user has all specified permissions
 */
export function hasAllPermissions(permissions: string[], userPermissions: string[]): boolean {
  if (!userPermissions || userPermissions.length === 0) {
    return false;
  }

  // Check if user has all permissions (AND logic)
  return permissions.every(p => userPermissions.includes(p));
}

/**
 * Check if user has specific role
 */
export function hasRole(roleCode: string, userRoles: string[]): boolean {
  if (!userRoles || userRoles.length === 0) {
    return false;
  }

  return userRoles.includes(roleCode);
}

/**
 * Check if user has any of the specified roles
 */
export function hasAnyRole(roleCodes: string[], userRoles: string[]): boolean {
  if (!userRoles || userRoles.length === 0) {
    return false;
  }

  return roleCodes.some(code => userRoles.includes(code));
}
