package perm

import (
	"testing"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/stretchr/testify/assert"
)

// TestToPermissionInfo_MinVerfLevelPassthrough — T1.1 MinVerfLevel 从 proto 透传到 types.PermissionInfo
// 断言：level-0 发布权限透传 0、level-2 选举权限透传 2（+ Timestamps 透传）
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — QA 补测，RED 摘录见 CHANGELOG 2026-08-12 补测节
func TestToPermissionInfo_MinVerfLevelPassthrough(t *testing.T) {
	// level-0 发布权限
	publish := &permissionv1.Permission{
		Id: 435, ParentId: 0, Code: "community:lostfound:create-api", Name: "发布-寻物启事",
		Type: 3, Path: "POST:/api/community/lostfound", Icon: "", SortOrder: 1, Status: 1,
		MinVerfLevel: 0,
		Timestamps:   &commonv1.Timestamps{CreatedAt: 1723420800, UpdatedAt: 1723420800},
	}
	info := toPermissionInfo(publish)
	assert.Equal(t, int32(0), info.MinVerfLevel, "发布类权限 min_verf_level=0 透传")
	assert.Equal(t, int64(435), info.Id)
	assert.Equal(t, int64(1723420800), info.CreatedAt, "Timestamps.CreatedAt 透传")

	// level-2 选举权限
	election := &permissionv1.Permission{
		Id: 600, ParentId: 0, Code: "committee:election:vote", Name: "业主委员会选举投票",
		Type: 2, MinVerfLevel: 2,
	}
	info2 := toPermissionInfo(election)
	// GREEN：level-2 选举权限透传 minVerfLevel=2
	assert.Equal(t, int32(2), info2.MinVerfLevel, "level-2 选举权限应透传 minVerfLevel=2")
}

// TestToRoleInfo_FieldMapping — toRoleInfo 字段映射（含嵌套 Permissions 的 MinVerfLevel）
func TestToRoleInfo_FieldMapping(t *testing.T) {
	r := &permissionv1.Role{
		Id: 1, Code: "owner", Name: "业主", Description: "业主角色",
		IsSystem: false, Status: 1, SortOrder: 1,
		Permissions: []*permissionv1.Permission{
			{Id: 435, Code: "community:lostfound:create-api", MinVerfLevel: 0},
		},
		Timestamps: &commonv1.Timestamps{CreatedAt: 1723420800, UpdatedAt: 1723420810},
	}

	info := toRoleInfo(r)
	assert.Equal(t, int64(1), info.Id)
	assert.Equal(t, "owner", info.Code)
	assert.Equal(t, "业主", info.Name)
	assert.Equal(t, "业主角色", info.Description)
	assert.False(t, info.IsSystem)
	assert.Equal(t, int32(1), info.Status)
	assert.Equal(t, int32(1), info.SortOrder)
	assert.Equal(t, int64(1723420800), info.CreatedAt)
	assert.Equal(t, int64(1723420810), info.UpdatedAt)
	assert.Len(t, info.Permissions, 1)
	assert.Equal(t, int32(0), info.Permissions[0].MinVerfLevel, "嵌套权限 MinVerfLevel 透传")
}

// TestToPermissionInfoList_RecursiveChildren — 树形结构子节点递归转换
func TestToPermissionInfoList_RecursiveChildren(t *testing.T) {
	parent := &permissionv1.Permission{
		Id: 10, ParentId: 0, Code: "community:lostfound", Name: "寻物启事", Type: 1,
		MinVerfLevel: 0,
		Children: []*permissionv1.Permission{
			{Id: 435, ParentId: 10, Code: "community:lostfound:create-api", Type: 2, MinVerfLevel: 0},
		},
	}

	infos := toPermissionInfoList([]*permissionv1.Permission{parent})
	assert.Len(t, infos, 1)
	assert.Len(t, infos[0].Children, 1)
	assert.Equal(t, "community:lostfound:create-api", infos[0].Children[0].Code)
	assert.Equal(t, int64(10), infos[0].Children[0].ParentId)
	assert.Equal(t, int32(0), infos[0].Children[0].MinVerfLevel)
}

// TestToPermissionInfoList_EmptyReturnsNil — 空列表返回 nil（not empty slice）
func TestToPermissionInfoList_EmptyReturnsNil(t *testing.T) {
	assert.Nil(t, toPermissionInfoList(nil))
	assert.Nil(t, toPermissionInfoList([]*permissionv1.Permission{}))
}
