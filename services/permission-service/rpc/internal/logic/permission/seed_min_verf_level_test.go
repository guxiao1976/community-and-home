package permission

// seed_min_verf_level_test.go — 敏感权限 min_verf_level=2 加固的种子结构测试
//
// 背景（security-arch 评审 CRITICAL）：服务角色（网格员/社区管理员/物业管理员）可免 membership
// 自助申请（user-service 并行改造），但其未认证(status=0) grant 不得行使破坏性操作
// （删公告/改公告/建活动/查角色配置）。init_permissions.sql 已在 4.3.1/6.1 加固
// user:read/moderation:*/notice:create-api/committee:election:vote，但以下破坏性权限仍未设
// min_verf_level=2：
//
//	community:notice:delete-api    (427 DELETE /api/community/notices/:id)
//	community:notice:update-api    (428 PUT   /api/community/notices/:id)
//	community:activity:create-api  (432 POST  /api/community/activities)
//	role:read                      (210 查看角色按钮)
//	role:read:list-api             (211 GET /api/perm/roles)
//	role:read:detail-api           (212 GET /api/perm/roles/:id)
//
// 本测试按「从零建库（init_permissions.sql 自上而下执行）」模拟每个权限码的有效 min_verf_level，
// 断言 6 个敏感码最终有效层级 = 2。此模拟能捕获「加固 UPDATE 放置早于 427/428 的 INSERT IGNORE」
// 导致的被默认 0 覆盖问题（INSERT IGNORE 未列 min_verf_level 列 → 新行取列默认 0）。
//
// SEE: [[auto-grant-unverified-grant-confers-scope-level0]] — 未认证 grant 立即生效的既有语义，
// 现经 min_verf_level=2 数据驱动收窄（CheckPermission: maxLevel >= minLevel）
// SEE: [[is-system-no-permission-shortcut]] — 权限经 rel_role_permission 配置，认证要求经 min_verf_level 数据驱动

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// hardenCodes 本任务加固的敏感写/管理权限码（security-arch CRITICAL）
var hardenCodes = []string{
	"community:notice:delete-api",   // 427 删公告
	"community:notice:update-api",   // 428 改公告
	"community:activity:create-api", // 432 建活动
	"role:read",                     // 210 查看角色（按钮）
	"role:read:list-api",            // 211 查角色配置列表
	"role:read:detail-api",          // 212 查角色配置详情
}

// repoRoot 返回仓库根目录（基于当前测试文件位置向上定位，`go test` cwd=包目录，不可依赖）
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	// .../services/permission-service/rpc/internal/logic/permission → 上溯 6 级到仓库根
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "services")); err != nil {
		t.Fatalf("无法定位仓库根目录（%s）：%v", root, err)
	}
	return root
}

func readSQLFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败：%v", path, err)
	}
	return string(b)
}

// --- 迷你 SQL 解析器（仅覆盖本测试关注的语句形态，不做通用 SQL 解析） ---

var (
	reUpdateLevel  = regexp.MustCompile(`(?i)UPDATE\s+sys_permission\s+SET\s+min_verf_level\s*=\s*(\d+)`)
	reWhereCodeIn  = regexp.MustCompile(`(?i)WHERE\s+code\s+IN\s*\(([^)]*)\)`)
	reQuotedString = regexp.MustCompile(`'([^']*)'`)
	reColumnList   = regexp.MustCompile(`(?i)INTO\s+sys_permission\s*\(([^)]*)\)\s*VALUES`)
	// 元组分隔：`), (`（逗号后可有空白）——NOW(), NOW()) 内部不匹配，仅元组边界 `) , (` 命中
	reTupleSep = regexp.MustCompile(`\),\s*\(`)
	// INSERT INTO <table> (...) VALUES 的列清单（表名通用，供 rel_role_permission/sys_permission 解析）
	reIntoColumns = regexp.MustCompile(`(?i)INTO\s+[a-z_][a-z0-9_]*\s*\(([^)]*)\)\s*VALUES`)
)

// permDef sys_permission 定义（审计用）
type permDef struct {
	id   int64
	code string
	typ  int64
	path string
}

// parseRolePermBindings 解析 rel_role_permission INSERT，返回 roleId → permissionId 集合
func parseRolePermBindings(sql string) map[int64]map[int64]struct{} {
	rolePerms := make(map[int64]map[int64]struct{})
	for _, stmt := range splitStatements(stripSQLComments(sql)) {
		one := strings.Join(strings.Fields(stmt), " ")
		if one == "" {
			continue
		}
		up := strings.ToUpper(one)
		if !strings.HasPrefix(up, "INSERT IGNORE INTO REL_ROLE_PERMISSION") &&
			!strings.HasPrefix(up, "INSERT INTO REL_ROLE_PERMISSION") {
			continue
		}
		roles := extractValueField(one, 0)
		perms := extractValueField(one, 1)
		for i, r := range roles {
			if i >= len(perms) {
				break
			}
			rid := int64(atoi(trimQuotes(strings.TrimSpace(r))))
			pid := int64(atoi(trimQuotes(strings.TrimSpace(perms[i]))))
			if rolePerms[rid] == nil {
				rolePerms[rid] = make(map[int64]struct{})
			}
			rolePerms[rid][pid] = struct{}{}
		}
	}
	return rolePerms
}

// parsePermissionDefs 解析 sys_permission INSERT，返回 id → 定义
func parsePermissionDefs(sql string) map[int64]permDef {
	defs := make(map[int64]permDef)
	for _, stmt := range splitStatements(stripSQLComments(sql)) {
		one := strings.Join(strings.Fields(stmt), " ")
		if one == "" {
			continue
		}
		up := strings.ToUpper(one)
		if !strings.HasPrefix(up, "INSERT IGNORE INTO SYS_PERMISSION") &&
			!strings.HasPrefix(up, "INSERT INTO SYS_PERMISSION") {
			continue
		}
		cols := extractColumns(one)
		if cols == nil {
			continue
		}
		idIdx, codeIdx, typeIdx, pathIdx := indexOf(cols, "id"), indexOf(cols, "code"), indexOf(cols, "type"), indexOf(cols, "path")
		if idIdx < 0 || codeIdx < 0 {
			continue
		}
		ids, codes, types, paths := extractValueField(one, idIdx), extractValueField(one, codeIdx), extractValueField(one, typeIdx), extractValueField(one, pathIdx)
		for i, idRaw := range ids {
			id := int64(atoi(trimQuotes(strings.TrimSpace(idRaw))))
			d := permDef{id: id}
			if i < len(codes) {
				d.code = trimQuotes(strings.TrimSpace(codes[i]))
			}
			if i < len(types) {
				d.typ = int64(atoi(trimQuotes(strings.TrimSpace(types[i]))))
			}
			if i < len(paths) {
				d.path = trimQuotes(strings.TrimSpace(paths[i]))
			}
			defs[id] = d
		}
	}
	return defs
}

// isWriteMethod 判断 API 权限 path 是否为写方法（POST/PUT/DELETE/PATCH）
func isWriteMethod(path string) bool {
	for _, m := range []string{"POST:", "PUT:", "DELETE:", "PATCH:"} {
		if strings.HasPrefix(path, m) {
			return true
		}
	}
	return false
}

// simulateEffectiveMinVerfLevel 按执行序模拟 init_permissions.sql 对每个权限码的最终 min_verf_level
//
//	从零建库语义：
//	  - INSERT IGNORE INTO sys_permission 未列 min_verf_level 列 → 新行取列默认 0
//	  - UPDATE sys_permission SET min_verf_level = N WHERE code IN (...) / code = 'x' 覆盖
//	  - 以「最后一次写入」为有效层级
func simulateEffectiveMinVerfLevel(sql string) map[string]int {
	eff := make(map[string]int)
	// 先剥离 `--` 注释（整行 + 行内；种子中字符串字面量不含 "--"，安全），
	// 否则无分号的注释块会与后续 SQL 合并成以注释开头的语句，破坏前缀检测。
	for _, stmt := range splitStatements(stripSQLComments(sql)) {
		// 折叠空白为单空格，简化单行正则
		one := strings.Join(strings.Fields(stmt), " ")
		if one == "" {
			continue
		}
		if m := reUpdateLevel.FindStringSubmatch(one); m != nil {
			applyUpdateMinLevel(one, atoi(m[1]), eff)
			continue
		}
		up := strings.ToUpper(one)
		if strings.HasPrefix(up, "INSERT IGNORE INTO SYS_PERMISSION") ||
			strings.HasPrefix(up, "INSERT INTO SYS_PERMISSION") {
			applyInsertDefault(one, eff)
		}
	}
	return eff
}

// stripSQLComments 剥离 `--` 注释（整行注释整行移除；行内注释仅保留 `--` 前内容）。
// 种子 SQL 中字符串字面量不含 `--`，直接按 `--` 切分安全。
func stripSQLComments(sql string) string {
	lines := strings.Split(sql, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		trimmed := strings.TrimLeft(ln, " \t")
		if strings.HasPrefix(trimmed, "--") {
			continue // 整行注释
		}
		if i := strings.Index(trimmed, "--"); i >= 0 {
			trimmed = trimmed[:i] // 行内注释（保留 `--` 前内容）
		}
		out = append(out, trimmed)
	}
	return strings.Join(out, "\n")
}

// splitStatements 按分号切分 SQL（保留无分号的尾段，去空）
func splitStatements(sql string) []string {
	var out []string
	for _, s := range strings.Split(sql, ";") {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// applyUpdateMinLevel 处理 UPDATE ... SET min_verf_level = N WHERE code IN (...) / code = 'x'
func applyUpdateMinLevel(stmt string, level int, eff map[string]int) {
	if in := reWhereCodeIn.FindStringSubmatch(stmt); in != nil {
		for _, q := range reQuotedString.FindAllString(in[1], -1) {
			eff[trimQuotes(q)] = level
		}
		return
	}
	// 单码形态 WHERE code = 'x'
	quoted := reQuotedString.FindAllString(stmt, -1)
	if len(quoted) == 1 {
		eff[trimQuotes(quoted[0])] = level
	}
}

// applyInsertDefault 处理 INSERT IGNORE INTO sys_permission(...)：未列 min_verf_level 列 → 新码默认 0
func applyInsertDefault(stmt string, eff map[string]int) {
	cols := reColumnList.FindStringSubmatch(stmt)
	if cols == nil {
		return
	}
	colList := splitCSV(cols[1])
	if indexOf(colList, "min_verf_level") >= 0 {
		return // 显式指定层级，非本测试关注的默认 0 形态
	}
	codeIdx := indexOf(colList, "code")
	if codeIdx < 0 {
		return
	}
	for _, code := range extractValueCodes(stmt, codeIdx) {
		eff[code] = 0
	}
}

// extractValueCodes 提取 VALUES 各元组第 codeIdx 列的权限码
func extractValueCodes(stmt string, codeIdx int) []string {
	var codes []string
	for _, raw := range extractValueField(stmt, codeIdx) {
		if c := trimQuotes(strings.TrimSpace(raw)); c != "" {
			codes = append(codes, c)
		}
	}
	return codes
}

// extractValueField 提取 VALUES 各元组第 fieldIdx 列的原始字段值（保留引号）
func extractValueField(stmt string, fieldIdx int) []string {
	vi := strings.Index(stmt, "VALUES")
	if vi < 0 {
		return nil
	}
	rest := strings.TrimSpace(stmt[vi+len("VALUES"):])
	tuples := reTupleSep.Split(rest, -1)
	if len(tuples) == 0 {
		return nil
	}
	tuples[0] = strings.TrimPrefix(strings.TrimSpace(tuples[0]), "(")
	last := tuples[len(tuples)-1]
	last = strings.TrimSuffix(strings.TrimSpace(last), ";")
	tuples[len(tuples)-1] = strings.TrimSuffix(strings.TrimSpace(last), ")")

	var vals []string
	for _, tpl := range tuples {
		fields := splitCSV(tpl)
		if len(fields) > fieldIdx {
			vals = append(vals, fields[fieldIdx])
		}
	}
	return vals
}

// extractColumns 提取 INSERT INTO <table> (...) VALUES 的列清单
func extractColumns(stmt string) []string {
	m := reIntoColumns.FindStringSubmatch(stmt)
	if m == nil {
		return nil
	}
	return splitCSV(m[1])
}

// splitCSV 按逗号切分（种子元组内字符串值无 ASCII 逗号，简单切分足够）
func splitCSV(s string) []string {
	var out []string
	for _, f := range strings.Split(s, ",") {
		out = append(out, strings.TrimSpace(f))
	}
	return out
}

func indexOf(list []string, want string) int {
	for i, v := range list {
		if v == want {
			return i
		}
	}
	return -1
}

func trimQuotes(s string) string {
	return strings.Trim(strings.TrimSpace(s), "'")
}

// --- 断言 ---

// TestSeedSensitivePermissions_HardenedToLevel2 — init_permissions.sql 从零建库后，
// 6 个敏感写/管理权限码的有效 min_verf_level 必须 = 2（需已认证）
func TestSeedSensitivePermissions_HardenedToLevel2(t *testing.T) {
	root := repoRoot(t)
	sql := readSQLFile(t, filepath.Join(root, "services", "permission-service", "scripts", "init_permissions.sql"))
	eff := simulateEffectiveMinVerfLevel(sql)

	for _, code := range hardenCodes {
		assert.Contains(t, eff, code, "权限码 %s 在种子中无定义（INSERT 缺失，防加固幻影）", code)
		assert.Equalf(t, 2, eff[code],
			"权限码 %s 从零建库后有效 min_verf_level 应为 2（需已认证）；当前=%d。"+
				"检查加固 UPDATE 是否放置于 427/428 的 INSERT IGNORE 之后（4.3.2 早于段 6.4 会被默认 0 覆盖）",
			code, eff[code])
	}
}

// TestMigrationPrivilegedRoleMinVerfLevel_HardenedToLevel2 — 迁移文件须为既有库重复应用同一加固
func TestMigrationPrivilegedRoleMinVerfLevel_HardenedToLevel2(t *testing.T) {
	root := repoRoot(t)
	sql := readSQLFile(t, filepath.Join(root, "services", "permission-service",
		"migration", "004_privileged_role_min_verf_level.sql"))
	eff := simulateEffectiveMinVerfLevel(sql)

	for _, code := range hardenCodes {
		assert.Equalf(t, 2, eff[code], "迁移 004 须将 %s 置 min_verf_level=2（既有库幂等应用）", code)
	}
}

// TestSeedPrivilegedRoles_DestructiveWritePerms_HardenedToLevel2 — 审计（security-arch CRITICAL）：
// 服务角色（property_admin 2 / community_admin 3 / grid_worker 4）绑定的所有破坏性写权限
// （type=3 API，path 为 POST/PUT/DELETE/PATCH）必须 min_verf_level=2；读权限保持 0 即可。
// 防止未来给服务角色新增破坏性写权限时漏设认证门槛（本变更修复 427/428/432 即属此类）。
func TestSeedPrivilegedRoles_DestructiveWritePerms_HardenedToLevel2(t *testing.T) {
	root := repoRoot(t)
	sql := readSQLFile(t, filepath.Join(root, "services", "permission-service",
		"scripts", "init_permissions.sql"))
	eff := simulateEffectiveMinVerfLevel(sql)
	rolePerms := parseRolePermBindings(sql)
	defs := parsePermissionDefs(sql)

	// 已知良性自域写权限：非社区级破坏性操作，可保持 min_verf_level=0
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]] — 该语义下 self-scoped 写不扩大数据范围
	benignWrites := map[string]string{
		"user:currentcommunity:write-api": "PUT /api/users/me/current-community（用户自设当前小区，self-scoped 非破坏性）",
	}

	for _, roleID := range []int64{2, 3, 4} {
		for pid := range rolePerms[roleID] {
			d, ok := defs[pid]
			if !ok {
				continue // 幻影绑定（如 435 lostfound 无 sys_permission 行），非本测试关注
			}
			if d.typ != 3 {
				continue // 仅审计 API 级叶子；菜单/按钮为分组节点（角色读按钮 210 由 hardenCodes 单独断言）
			}
			if !isWriteMethod(d.path) {
				continue // 读权限保持 0
			}
			if _, benign := benignWrites[d.code]; benign {
				continue
			}
			assert.Equalf(t, 2, eff[d.code],
				"角色 %d 绑定的破坏性写权限 %s(%s) 必须 min_verf_level=2（未认证 grant 不得行使）",
				roleID, d.code, d.path)
		}
	}
}
