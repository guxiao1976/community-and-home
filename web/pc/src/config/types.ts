import type { Component } from 'vue';
import type { RouteRecordRaw } from 'vue-router';

/**
 * 菜单项配置
 */
export interface MenuItemConfig {
  path: string;           // 路由路径
  title: string;          // 菜单标题
  icon?: Component;       // 图标组件
  permission?: string;    // 权限码（可选）
  children?: MenuItemConfig[];  // 子菜单
  hidden?: boolean;       // 是否隐藏（用于详情页等）
}

/**
 * 模块配置
 */
export interface ModuleConfig {
  name: string;                    // 模块名称（用于标识）
  menu: MenuItemConfig | MenuItemConfig[];  // 菜单配置（单个或多个）
  routes: RouteRecordRaw[];        // 路由配置
}
