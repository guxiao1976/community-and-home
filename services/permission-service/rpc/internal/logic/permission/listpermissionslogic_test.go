package permission

import (
	"context"
	"database/sql"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestListPermissions_BuildTree 覆盖 ListPermissions 的树形构建（此前无直接单测）。
// 验证：roots 收集 + parent_id 挂载 children 的正确性。
func TestListPermissions_BuildTree(t *testing.T) {
	perms := []*model.SysPermission{
		{Id: 1, ParentId: sql.NullInt64{}, Name: "根1", Code: "root1", Type: 1, Status: 1},
		{Id: 2, ParentId: sql.NullInt64{Int64: 1, Valid: true}, Name: "子1", Code: "child1", Type: 2, Status: 1},
		{Id: 3, ParentId: sql.NullInt64{Int64: 1, Valid: true}, Name: "子2", Code: "child2", Type: 3, Status: 1},
		{Id: 4, ParentId: sql.NullInt64{}, Name: "根2", Code: "root2", Type: 1, Status: 1},
	}

	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindWithFilter", mock.Anything, (*int64)(nil), (*int64)(nil)).Return(perms, nil)

	sc := &svc.ServiceContext{PermissionModel: mockPerm}
	l := NewListPermissionsLogic(context.Background(), sc)
	resp, err := l.ListPermissions(&permissionv1.ListPermissionsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetPermissions(), 2, "两个根节点")

	roots := resp.GetPermissions()
	assert.Equal(t, int64(1), roots[0].GetId())
	assert.Equal(t, int64(4), roots[1].GetId())
	require.Len(t, roots[0].GetChildren(), 2, "根1 有两个子节点")
	assert.Equal(t, int64(2), roots[0].GetChildren()[0].GetId())
	assert.Equal(t, int64(3), roots[0].GetChildren()[1].GetId())
	mockPerm.AssertExpectations(t)
}

// TestListPermissions_FilterPassthrough 验证 type/status 筛选参数正确透传给 FindWithFilter。
func TestListPermissions_FilterPassthrough(t *testing.T) {
	mockPerm := new(MockPermissionModel)
	mockPerm.On("FindWithFilter", mock.Anything, mock.Anything, mock.Anything).Return([]*model.SysPermission{}, nil)

	typ := int32(3)
	status := int32(1)
	sc := &svc.ServiceContext{PermissionModel: mockPerm}
	l := NewListPermissionsLogic(context.Background(), sc)
	resp, err := l.ListPermissions(&permissionv1.ListPermissionsRequest{Type: &typ, Status: &status})
	require.NoError(t, err)
	assert.Empty(t, resp.GetPermissions(), "空结果 → 空树")
	mockPerm.AssertExpectations(t)
}
