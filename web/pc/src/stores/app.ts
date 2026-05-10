// App store for global UI state

import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { RouteLocationNormalized } from 'vue-router';

export interface Breadcrumb {
  title: string;
  path: string;
}

export const useAppStore = defineStore('app', () => {
  // State
  const loading = ref(false);
  const breadcrumbs = ref<Breadcrumb[]>([]);
  const sidebarCollapsed = ref(false);
  const theme = ref<'light' | 'dark'>('light');

  // Actions
  const setLoading = (value: boolean): void => {
    loading.value = value;
  };

  const setBreadcrumbs = (items: Breadcrumb[]): void => {
    breadcrumbs.value = items;
  };

  const updateBreadcrumb = (route: RouteLocationNormalized): void => {
    if (route.meta.breadcrumb) {
      breadcrumbs.value = route.meta.breadcrumb as Breadcrumb[];
    } else {
      const matched = route.matched.filter(r => r.meta?.title);
      breadcrumbs.value = matched.map(r => ({
        title: r.meta.title as string,
        path: r.path === route.path ? '' : r.path
      }));
    }
  };

  const addBreadcrumb = (item: Breadcrumb): void => {
    const exists = breadcrumbs.value.some(b => b.path === item.path);
    if (!exists) {
      breadcrumbs.value.push(item);
    }
  };

  const toggleSidebar = (): void => {
    sidebarCollapsed.value = !sidebarCollapsed.value;
  };

  const setSidebarCollapsed = (value: boolean): void => {
    sidebarCollapsed.value = value;
  };

  const setTheme = (value: 'light' | 'dark'): void => {
    theme.value = value;
    document.documentElement.setAttribute('data-theme', value);
  };

  return {
    // State
    loading,
    breadcrumbs,
    sidebarCollapsed,
    theme,
    // Actions
    setLoading,
    setBreadcrumbs,
    addBreadcrumb,
    updateBreadcrumb,
    toggleSidebar,
    setSidebarCollapsed,
    setTheme
  };
}, {
  persist: true
});
