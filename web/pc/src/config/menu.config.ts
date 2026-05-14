import type { MenuItemConfig } from './types';
import { dashboardModule } from './modules/dashboard.config';
import { masterdataModule } from './modules/masterdata.config';
import { identityModule } from './modules/identity.config';
import { moderationModule } from './modules/moderation.config';
import { aimodelModule } from './modules/aimodel.config';

/**
 * 聚合所有模块的菜单配置
 */
export function getMenuItems(): MenuItemConfig[] {
  const modules = [
    dashboardModule,
    masterdataModule,
    identityModule,
    moderationModule,
    aimodelModule
  ];

  return modules.map(module => {
    // 如果 menu 是数组，展开；否则直接返回
    return Array.isArray(module.menu) ? module.menu : [module.menu];
  }).flat();
}
