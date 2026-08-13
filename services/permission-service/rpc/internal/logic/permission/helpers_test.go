package permission

import (
	"testing"

	"github.com/guxiao1976/community-permission/model"
	"github.com/stretchr/testify/assert"
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
