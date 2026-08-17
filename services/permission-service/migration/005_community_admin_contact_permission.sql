-- ============================================================================
-- 005_community_admin_contact_permission.sql
-- 社区管理员角色强化：维护便民电话权限（用户拍板 2026-08-17）
--
-- 目标：
--   community_admin(3) 绑定 436 community:contact:upsert-api（POST:/api/community/contacts）
--   —— 既有库补齐该绑定（init_permissions.sql 段 7 从零建库已含，本迁移为存量库幂等补绑定）。
--
-- 幂等：INSERT IGNORE 可重复执行（uk_role_permission 唯一约束冲突静默跳过）。
-- 436 min_verf_level 0→2：436 为 POST 写（维护社区级联系方式，非 self-scoped），community_admin(3)
--   属服务角色（可免 membership 自助申请）——未认证(status=0) grant 不得行使破坏性写，
--   与 6.8 服务角色破坏性写加固同判据（审计回归测试强制）。owner/tenant 持 436 同步收窄为需已认证。
-- SEE: [[permission-seed-api-path-must-match-routes]] — 436 path 与 community-hub 实际 REST 路由一致
-- SEE: [[migration-must-execute]] — 迁移三步闭环（写 → 提交 → 执行），末尾 SELECT 验证
-- ============================================================================

INSERT IGNORE INTO rel_role_permission (role_id, permission_id) VALUES (3, 436);
UPDATE sys_permission SET min_verf_level = 2 WHERE code = 'community:contact:upsert-api';

-- 幂等验证：community_admin(3) 持 436 且 436 min_verf_level=2
SELECT 'community_admin → 436' AS check_type,
       (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((3,436))) AS bindings,
       (SELECT min_verf_level FROM sys_permission WHERE code = 'community:contact:upsert-api') AS level436,
       CASE WHEN (SELECT COUNT(*) FROM rel_role_permission WHERE (role_id, permission_id) IN ((3,436))) = 1
             AND (SELECT min_verf_level FROM sys_permission WHERE code = 'community:contact:upsert-api') = 2
            THEN '✅ PASS' ELSE '❌ FAIL' END AS status;
