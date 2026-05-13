// v-permission directive for button-level access control

import type { App, Directive, DirectiveBinding } from 'vue';
import { usePermissionStore } from '@/stores/permission';

function checkPermission(binding: DirectiveBinding): boolean {
  const permissionStore = usePermissionStore();
  const value = binding.value;

  if (!value) return true;

  if (Array.isArray(value)) {
    return permissionStore.hasAnyPermission(value);
  }

  return permissionStore.hasPermission(String(value));
}

const permissionDirective: Directive = {
  mounted(el: HTMLElement, binding: DirectiveBinding) {
    if (!checkPermission(binding)) {
      el.style.display = 'none';
    }
  },
  updated(el: HTMLElement, binding: DirectiveBinding) {
    if (!checkPermission(binding)) {
      el.style.display = 'none';
    } else {
      el.style.display = '';
    }
  }
};

export function setupPermissionDirective(app: App): void {
  app.directive('permission', permissionDirective);
}

export default permissionDirective;
