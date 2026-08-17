-- ============================================================================
-- 004_privileged_role_min_verf_level.sql
-- 敏感写/管理权限 min_verf_level=2 加固（security-arch 评审 CRITICAL）
--
-- 目标：
--   服务角色（网格员 grid_worker / 社区管理员 community_admin / 物业管理员 property_admin）
--   可免 membership 自助申请（user-service 并行改造），但其未认证(status=0) grant
--   不得行使破坏性操作：删公告 / 改公告 / 建活动 / 查角色配置。
--
-- 加固权限（与 init_permissions.sql 段 6.8 同批）：
--   community:notice:delete-api    (427 DELETE /api/community/notices/:id)
--   community:notice:update-api    (428 PUT   /api/community/notices/:id)
--   community:activity:create-api  (432 POST  /api/community/activities)
--   role:read                      (210 查看角色按钮)
--   role:read:list-api             (211 GET /api/perm/roles)
--   role:read:detail-api           (212 GET /api/perm/roles/:id)
--
-- 约束机制：min_verf_level=2（需已认证）为数据驱动约束——CheckPermission 放行
--   ⟺ maxLevel(grantSatisfiedLevel) >= minLevel；未认证 grant 恒 level-0 → 无法放行。
--
-- 幂等：UPDATE 可重复执行（不依赖 information_schema guard）。
-- 既有库 sys_permission 行已存在（min_verf_level=0 或既有值），直接覆盖为 2。
-- SEE: [[auto-grant-unverified-grant-confers-scope-level0]]
-- SEE: [[is-system-no-permission-shortcut]]
-- SEE: [[permission-seed-api-path-must-match-routes]]
-- SEE: [[migration-must-execute]]
-- ============================================================================

UPDATE sys_permission SET min_verf_level = 2
WHERE code IN ('community:notice:delete-api', 'community:notice:update-api',
               'community:activity:create-api',
               'role:read', 'role:read:list-api', 'role:read:detail-api');

-- 幂等验证：6 个敏感码全部 min_verf_level=2
SELECT '敏感写/管理权限 min_verf_level=2' AS check_type,
       (SELECT COUNT(*) FROM sys_permission
        WHERE code IN ('community:notice:delete-api','community:notice:update-api',
                       'community:activity:create-api',
                       'role:read','role:read:list-api','role:read:detail-api')
          AND min_verf_level = 2) AS hardened,
       (SELECT COUNT(*) FROM sys_permission
        WHERE code IN ('community:notice:delete-api','community:notice:update-api',
                       'community:activity:create-api',
                       'role:read','role:read:list-api','role:read:detail-api')) AS total,
       CASE WHEN (SELECT COUNT(*) FROM sys_permission
                  WHERE code IN ('community:notice:delete-api','community:notice:update-api',
                                 'community:activity:create-api',
                                 'role:read','role:read:list-api','role:read:detail-api')
                    AND min_verf_level = 2) = 6
            THEN '✅ PASS' ELSE '❌ FAIL' END AS status;
