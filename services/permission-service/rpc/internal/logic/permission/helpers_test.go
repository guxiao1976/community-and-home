package permission

import (
	"context"
	"database/sql"
	"testing"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestSplitPlatforms table-driven：platforms 字符串 → 切片
// SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，空值运行时 fail-open（透出空切片）
func TestSplitPlatforms(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "空串 → 空切片（非 nil）", in: "", want: []string{}},
		{name: "单端 pc", in: "pc", want: []string{"pc"}},
		{name: "单端 mobile", in: "mobile", want: []string{"mobile"}},
		{name: "双端 pc,mobile", in: "pc,mobile", want: []string{"pc", "mobile"}},
		{name: "含空格清理", in: "pc, mobile", want: []string{"pc", "mobile"}},
		{name: "尾部逗号过滤空元素", in: "pc,", want: []string{"pc"}},
		{name: "连续逗号过滤空元素", in: "pc,,mobile", want: []string{"pc", "mobile"}},
		{name: "多空格清理", in: "  pc  ,  mobile  ", want: []string{"pc", "mobile"}},
		{name: "全空格串 → 空切片", in: "  ", want: []string{}},
		{name: "纯逗号 → 空切片", in: ",,", want: []string{}},
		{name: "首尾都有空格", in: " pc , mobile ", want: []string{"pc", "mobile"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, splitPlatforms(tt.in))
		})
	}
}

// TestJoinPlatforms split/join 互为逆操作
func TestJoinPlatforms(t *testing.T) {
	assert.Equal(t, "pc,mobile", joinPlatforms([]string{"pc", "mobile"}))
	assert.Equal(t, "", joinPlatforms([]string{}))
	assert.Equal(t, "", joinPlatforms(nil))

	// 往返：split → join → split 结果不变
	orig := "pc,mobile"
	assert.Equal(t, orig, joinPlatforms(splitPlatforms(orig)))
}

// TestToRolePbPlatforms 验证 toRolePb 透出 platforms（空串 → 空切片，非 nil）
func TestToRolePbPlatforms(t *testing.T) {
	r := &model.SysRole{RoleCode: "community_admin", RoleName: "社区管理员", Platforms: "pc,mobile"}
	assert.Equal(t, []string{"pc", "mobile"}, toRolePb(r).Platforms)

	r2 := &model.SysRole{RoleCode: "owner", Platforms: ""}
	assert.Equal(t, []string{}, toRolePb(r2).Platforms)
}

// TestToRolePbFields 覆盖 toRolePb 各字段映射（L104 存活变异：IsSystem 的 == 比较族）
func TestToRolePbFields(t *testing.T) {
	r := &model.SysRole{
		Id: 1, RoleCode: "owner", RoleName: "业主",
		Description: sql.NullString{String: "desc", Valid: true},
		IsSystem:    1, Status: 1, SortOrder: 5, Platforms: "pc",
	}
	pb := toRolePb(r)
	assert.Equal(t, int64(1), pb.Id)
	assert.Equal(t, "owner", pb.Code)
	assert.Equal(t, "业主", pb.Name)
	assert.Equal(t, "desc", pb.Description)
	assert.True(t, pb.IsSystem)   // IsSystem==1 → true
	assert.Equal(t, int32(1), pb.Status)
	assert.Equal(t, int32(5), pb.SortOrder)

	// IsSystem==0 → false（杀 == 比较族的反向）
	r2 := &model.SysRole{IsSystem: 0}
	assert.False(t, toRolePb(r2).IsSystem)
}

// TestGrantSatisfiedLevel 覆盖层级判定的全分支（L22-32 存活变异：grantActive 翻转、==/比较符、return 0）
func TestGrantSatisfiedLevel(t *testing.T) {
	now := time.Now()
	verified := sql.NullTime{Time: now, Valid: true}
	expired := sql.NullTime{Time: now.Add(-time.Hour), Valid: true}
	never := sql.NullTime{}

	tests := []struct {
		name string
		g    *model.UserRoleWithInfo
		want int
	}{
		{name: "nil grant → -1", g: nil, want: -1},
		{name: "URStatus=3 非活跃 → -1", g: &model.UserRoleWithInfo{URStatus: 3}, want: -1},
		{name: "URStatus=4 非活跃 → -1", g: &model.UserRoleWithInfo{URStatus: 4}, want: -1},
		{name: "过期（ExpiresAt 早于现在）→ -1", g: &model.UserRoleWithInfo{URStatus: 2, VerifiedAt: verified, ExpiresAt: expired}, want: -1},
		{name: "status=2 + verified → level-2", g: &model.UserRoleWithInfo{URStatus: 2, VerifiedAt: verified}, want: 2},
		{name: "status=2 + 未 verified → level-0", g: &model.UserRoleWithInfo{URStatus: 2, VerifiedAt: never}, want: 0},
		{name: "status=0 → level-0", g: &model.UserRoleWithInfo{URStatus: 0}, want: 0},
		{name: "status=1 → level-0", g: &model.UserRoleWithInfo{URStatus: 1}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, grantSatisfiedLevel(tt.g))
		})
	}
}

// TestRoleToPbWithPermissions 覆盖 permIds 空/err/成功三路（L120/125 存活变异：==、||、!= 比较族）
func TestRoleToPbWithPermissions(t *testing.T) {
	r := &model.SysRole{Id: 1, RoleCode: "owner", RoleName: "业主", Platforms: "mobile"}
	perm := &model.SysPermission{Id: 435, Code: "community:lostfound:create-api", Type: 3, MinVerfLevel: 0}
	ctx := context.Background()

	t.Run("permIds 空 → 返回不含权限", func(t *testing.T) {
		pb := roleToPbWithPermissions(ctx, r, []int64{}, nil)
		assert.NotNil(t, pb)
		assert.Nil(t, pb.Permissions)
		assert.Equal(t, []string{"mobile"}, pb.Platforms)
	})

	t.Run("FindByIds 返回 err → 返回不含权限", func(t *testing.T) {
		mockPerm := new(MockPermissionModel)
		mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return(nil, assert.AnError)
		pb := roleToPbWithPermissions(ctx, r, []int64{435}, mockPerm)
		assert.NotNil(t, pb)
		assert.Nil(t, pb.Permissions)
	})

	t.Run("FindByIds 返回空 → 返回不含权限", func(t *testing.T) {
		mockPerm := new(MockPermissionModel)
		mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{}, nil)
		pb := roleToPbWithPermissions(ctx, r, []int64{435}, mockPerm)
		assert.NotNil(t, pb)
		assert.Nil(t, pb.Permissions)
	})

	t.Run("err 非空但 len 非零 → 返回不含权限（|| 短路）", func(t *testing.T) {
		mockPerm := new(MockPermissionModel)
		// 返回 err + 非空 perms，验证 || 的 err 分支优先（杀 || → && 变异）
		mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{perm}, assert.AnError)
		pb := roleToPbWithPermissions(ctx, r, []int64{435}, mockPerm)
		assert.NotNil(t, pb)
		assert.Nil(t, pb.Permissions)
	})

	t.Run("成功 → 返回含权限", func(t *testing.T) {
		mockPerm := new(MockPermissionModel)
		mockPerm.On("FindByIds", mock.Anything, []int64{435}).Return([]*model.SysPermission{perm}, nil)
		pb := roleToPbWithPermissions(ctx, r, []int64{435}, mockPerm)
		assert.NotNil(t, pb)
		assert.Len(t, pb.Permissions, 1)
		assert.Equal(t, "community:lostfound:create-api", pb.Permissions[0].Code)
	})
}

// TestScopeStateString 覆盖三态枚举→字符串的每个 case + default（L43-52 存活变异：global/empty 返回值被改）
func TestScopeStateString(t *testing.T) {
	assert.Equal(t, "global", scopeStateString(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL))
	assert.Equal(t, "limited", scopeStateString(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED))
	assert.Equal(t, "empty", scopeStateString(permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY))
	assert.Equal(t, "empty", scopeStateString(permissionv1.DataScopeState_DATA_SCOPE_STATE_UNSPECIFIED))
}

// TestScopeStateFromString 覆盖字符串→三态枚举的每个 case + default（未知/空串 → EMPTY 安全默认）
func TestScopeStateFromString(t *testing.T) {
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL, scopeStateFromString("global"))
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, scopeStateFromString("limited"))
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, scopeStateFromString("empty"))
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, scopeStateFromString("unknown"))
	assert.Equal(t, permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, scopeStateFromString(""))
}

// TestSqlNullString 覆盖空串→NULL、非空→Valid 两端（L68 存活变异：空串分支真假）
func TestSqlNullString(t *testing.T) {
	assert.Equal(t, sql.NullString{}, sqlNullString(""))
	assert.Equal(t, sql.NullString{String: "pc", Valid: true}, sqlNullString("pc"))
}
