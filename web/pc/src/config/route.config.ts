import type { RouteRecordRaw } from 'vue-router';
import { dashboardModule } from './modules/dashboard.config';
import { userPermissionModule } from './modules/user-permission.config';
import { communityModule } from './modules/community.config';
import { reviewCenterModule } from './modules/review-center.config';
import { aimodelModule } from './modules/aimodel.config';
import { systemOpsModule } from './modules/system-ops.config';

/**
 * 聚合所有模块的路由配置
 */
export function getModuleRoutes(): RouteRecordRaw[] {
  const modules = [
    dashboardModule,
    userPermissionModule,
    communityModule,
    reviewCenterModule,
    aimodelModule,
    systemOpsModule
  ];

  return modules.flatMap(module => module.routes);
}
