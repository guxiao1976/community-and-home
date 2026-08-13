package permission

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
