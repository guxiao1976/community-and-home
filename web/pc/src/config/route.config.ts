import type { RouteRecordRaw } from 'vue-router';
import { dashboardModule } from './modules/dashboard.config';
import { masterdataModule } from './modules/masterdata.config';
import { identityModule } from './modules/identity.config';
import { moderationModule } from './modules/moderation.config';
import { aimodelModule } from './modules/aimodel.config';

/**
 * 聚合所有模块的路由配置
 */
export function getModuleRoutes(): RouteRecordRaw[] {
  const modules = [
    dashboardModule,
    masterdataModule,
    identityModule,
    moderationModule,
    aimodelModule
  ];

  return modules.flatMap(module => module.routes);
}
