package permission

// seed_community_admin_contact_test.go — 社区管理员角色强化：维护便民电话权限绑定（436）种子结构测试
//
// 背景（用户拍板 2026-08-17）：community_admin(3) 可维护便民电话 —— 绑定 436 community:contact:upsert-api
// （POST:/api/community/contacts）。init_permissions.sql 新增段 7 + migration/005 幂等补绑定。
// 436 为 POST 写（社区级联系方式，非 self-scoped），community_admin(3) 属服务角色——未认证(status=0)
// grant 不得行使破坏性写，故 436 min_verf_level 0→2（与 6.8 服务角色破坏性写加固同判据，
// 审计回归测试 TestSeedPrivilegedRoles_DestructiveWritePerms_HardenedToLevel2 强制）。
//
// SEE: [[permission-seed-api-path-must-match-routes]] — 436 path 必须与实际 REST 路由一致
// SEE: [[is-system-no-permission-shortcut]] — 权限经 rel_role_permission 配置，认证要求经 min_verf_level 数据驱动
// SEE: [[migration-must-execute]] — migration 005 幂等 + 末尾 SELECT 验证（执行由 Owner 验证）

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSeedCommunityAdmin_ContactUpsertBound — init_permissions.sql 从零建库后 community_admin(3) 绑定 436
func TestSeedCommunityAdmin_ContactUpsertBound(t *testing.T) {
	root := repoRoot(t)
	sql := readSQLFile(t, filepath.Join(root, "services", "permission-service", "scripts", "init_permissions.sql"))

	rolePerms := parseRolePermBindings(sql)
	defs := parsePermissionDefs(sql)

	// 436 定义存在且 path 与实际 REST 路由一致（community-hub UpsertContacts）
	d, ok := defs[436]
	assert.True(t, ok, "436 community:contact:upsert-api 在种子中无定义（INSERT 缺失，防幻影绑定）")
	if ok {
		assert.Equal(t, "POST:/api/community/contacts", d.path,
			"436 path 必须与实际 REST 路由一致（[[permission-seed-api-path-must-match-routes]]）")
	}

	// community_admin(3) 绑定 436
	assert.Contains(t, rolePerms[3], int64(436),
		"community_admin(3) 必须绑定 436 community:contact:upsert-api（维护便民电话权限）")

	// 436 为 POST 写 + 服务角色绑定 → 有效 min_verf_level 必须 = 2（需已认证，与 6.8 同判据）
	// 放置于 4.8（436 INSERT + min_verf_level=0）之后恒生效；审计回归测试亦强制。
	eff := simulateEffectiveMinVerfLevel(sql)
	assert.Contains(t, eff, "community:contact:upsert-api", "436 在种子中无有效层级定义")
	assert.Equalf(t, 2, eff["community:contact:upsert-api"],
		"436 community:contact:upsert-api 从零建库后有效 min_verf_level 应为 2（服务角色破坏性写需已认证）；当前=%d", eff["community:contact:upsert-api"])
}

// TestMigrationCommunityAdmin_ContactUpsertBound — migration 005 为既有库幂等补绑定 + 436 加固
func TestMigrationCommunityAdmin_ContactUpsertBound(t *testing.T) {
	root := repoRoot(t)
	sql := readSQLFile(t, filepath.Join(root, "services", "permission-service",
		"migration", "005_community_admin_contact_permission.sql"))

	rolePerms := parseRolePermBindings(sql)

	assert.Contains(t, rolePerms[3], int64(436),
		"迁移 005 须为既有库补 community_admin(3) → 436 绑定（幂等）")

	eff := simulateEffectiveMinVerfLevel(sql)
	assert.Contains(t, eff, "community:contact:upsert-api", "迁移 005 未设置 436 有效层级")
	assert.Equalf(t, 2, eff["community:contact:upsert-api"],
		"迁移 005 须将 436 置 min_verf_level=2（既有库幂等应用，服务角色破坏性写需已认证）")
}
