import type { MenuItemConfig } from './types';
import { dashboardModule } from './modules/dashboard.config';
import { userPermissionModule } from './modules/user-permission.config';
import { communityModule } from './modules/community.config';
import { reviewCenterModule } from './modules/review-center.config';
import { aimodelModule } from './modules/aimodel.config';
import { systemOpsModule } from './modules/system-ops.config';

/**
 * 聚合所有模块的菜单配置
 *
 * 菜单结构（按 RBAC 理念重构）：
 *   仪表板 → 用户与权限 → 社区管理 → 审核中心 → AI模型管理 → 系统运维
 */
export function getMenuItems(): MenuItemConfig[] {
  const modules = [
    dashboardModule,
    userPermissionModule,
    communityModule,
    reviewCenterModule,
    aimodelModule,
    systemOpsModule
  ];

  return modules.map(module => {
    // 如果 menu 是数组，展开；否则直接返回
    return Array.isArray(module.menu) ? module.menu : [module.menu];
  }).flat();
}
