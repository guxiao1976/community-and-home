package permission

import (
	"context"
	"testing"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestListRoles_DefaultPage 覆盖 Page nil → 默认 page=1, pageSize=10
func TestListRoles_DefaultPage(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.Anything, int64(1), int64(10)).
		Return([]*model.SysRole{}, int64(0), nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int32(1), resp.Page.Page)
	assert.Equal(t, int32(10), resp.Page.PageSize)
	assert.Equal(t, int64(0), resp.Page.Total)
	assert.Equal(t, int32(0), resp.Page.TotalPages)
	mockRole.AssertExpectations(t)
}

// TestListRoles_CustomPage 覆盖 Page/PageSize 显式设置 → 透传给 FindList
func TestListRoles_CustomPage(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.Anything, int64(2), int64(50)).
		Return([]*model.SysRole{
			{Id: 1, RoleCode: "owner", RoleName: "业主", Status: 1},
		}, int64(51), nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{
		Page: &commonv1.PageRequest{Page: 2, PageSize: 50},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), resp.Page.Page)
	assert.Equal(t, int32(50), resp.Page.PageSize)
	assert.Equal(t, int64(51), resp.Page.Total)
	assert.Equal(t, int32(2), resp.Page.TotalPages, "total=51, pageSize=50 → 2 页")
	assert.Len(t, resp.Roles, 1)
	mockRole.AssertExpectations(t)
}

// TestListRoles_PageZeroFallsBack 覆盖 Page 设置了但 Page=0/PageSize=0 → 回落默认 1/10
func TestListRoles_PageZeroFallsBack(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.Anything, int64(1), int64(10)).
		Return([]*model.SysRole{}, int64(0), nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{
		Page: &commonv1.PageRequest{Page: 0, PageSize: 0},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(1), resp.Page.Page)
	assert.Equal(t, int32(10), resp.Page.PageSize)
	mockRole.AssertExpectations(t)
}

// TestListRoles_StatusFilter 覆盖 Status 非 nil → 透传状态指针
func TestListRoles_StatusFilter(t *testing.T) {
	st := int32(1)
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.MatchedBy(func(s *int64) bool {
		return s != nil && *s == 1
	}), int64(1), int64(10)).
		Return([]*model.SysRole{}, int64(0), nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{Status: &st})
	assert.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	mockRole.AssertExpectations(t)
}

// TestListRoles_TotalPagesCeil 覆盖分页总页数上取整边界（total 整除 pageSize 时 -1/+1 区分）
func TestListRoles_TotalPagesCeil(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.Anything, int64(1), int64(10)).
		Return([]*model.SysRole{}, int64(100), nil)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int32(10), resp.Page.TotalPages, "total=100, pageSize=10 → 恰好 10 页")
	mockRole.AssertExpectations(t)
}

// TestListRoles_Error 覆盖 FindList 报错 → 透传 error
func TestListRoles_Error(t *testing.T) {
	mockRole := new(MockRoleModel)
	mockRole.On("FindList", mock.Anything, mock.Anything, int64(1), int64(10)).
		Return(nil, int64(0), assert.AnError)

	svcCtx := &svc.ServiceContext{RoleModel: mockRole}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRole.AssertExpectations(t)
}

// TestListRoles_PlatformsTransparency table-driven：ListRoles 透出 Role.Platforms
// SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，空值运行时 fail-open（透出空切片）
func TestListRoles_PlatformsTransparency(t *testing.T) {
	tests := []struct {
		name      string
		platforms string
		want      []string
	}{
		{name: "空串 → 空切片（fail-open）", platforms: "", want: []string{}},
		{name: "单端 pc", platforms: "pc", want: []string{"pc"}},
		{name: "双端 pc,mobile", platforms: "pc,mobile", want: []string{"pc", "mobile"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRole := new(MockRoleModel)
			mockRole.On("FindList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return([]*model.SysRole{
					{
						Id:        1,
						RoleCode:  "owner",
						RoleName:  "业主",
						Status:    1,
						SortOrder: 1,
						Platforms: tt.platforms,
					},
				}, int64(1), nil)

			svcCtx := &svc.ServiceContext{RoleModel: mockRole}
			logic := NewListRolesLogic(context.Background(), svcCtx)

			resp, err := logic.ListRoles(&permissionv1.ListRolesRequest{})
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Roles, 1)
			assert.Equal(t, tt.want, resp.Roles[0].Platforms)

			mockRole.AssertExpectations(t)
		})
	}
}
