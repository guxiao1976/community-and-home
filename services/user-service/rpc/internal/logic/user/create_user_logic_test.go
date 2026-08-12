package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.1: CreateUser 自动分配 registered_user（scope_type='', scope_id=0, status=2）
// role_id=9 经 roleMapper（ListRoles）解析；失败仅告警不阻塞注册；重复注册不重复分配
// =============================================================================

// mockRegisteredUserRoles 构造包含 registered_user(role_id=9) 的角色表响应
func mockRegisteredUserRoles() *permissionv1.ListRolesResponse {
	return &permissionv1.ListRolesResponse{
		Roles: []*permissionv1.Role{
			{Id: 9, Code: "registered_user"},
			{Id: 1, Code: "owner"},
			{Id: 5, Code: "tenant"},
		},
	}
}

func TestCreateUser_AssignsRegisteredUser(t *testing.T) {
	// C-RU-01: 注册成功 → 自动分配 registered_user grant（role_id=9, scope_type='', scope_id=0, status=2）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockRegisteredUserRoles(), nil).AnyTimes()

	var assigned *permissionv1.AssignRoleRequest
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.AssignRoleRequest, _ ...interface{}) (*permissionv1.AssignRoleResponse, error) {
			assigned = req
			return &permissionv1.AssignRoleResponse{}, nil
		}).Times(1)

	logic := NewCreateUserLogic(context.Background(), svc)
	resp, err := logic.CreateUser(&userv1.CreateUserRequest{Phone: "13800000001", Nickname: "张三"})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	// registered_user grant 参数断言
	require.NotNil(t, assigned, "AssignRole 应被调用（自动分配 registered_user）")
	assert.Equal(t, int64(9), assigned.RoleId)
	assert.Equal(t, "", assigned.ScopeType, "registered_user 为空 scope_type")
	assert.Equal(t, int64(0), assigned.ScopeId)
	require.NotNil(t, assigned.Status, "status 必须显式指定")
	assert.Equal(t, int32(2), *assigned.Status, "registered_user 永久有效(status=2)")

	// 用户已落库
	u, err := ub.FindOne(context.Background(), resp.UserId)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.Status)
}

func TestCreateUser_DuplicatePhone_NoReAssign(t *testing.T) {
	// C-RU-02: 手机号已注册 → 10002 且不再分配（幂等，无重复行）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockRegisteredUserRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(&permissionv1.AssignRoleResponse{}, nil).Times(1)

	logic := NewCreateUserLogic(context.Background(), svc)
	resp1, err := logic.CreateUser(&userv1.CreateUserRequest{Phone: "13900000001", Nickname: "李四"})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp1.Base.Code)

	// 同手机号二次注册 → 已注册失败，不得再次 AssignRole（无重复 grant 行）
	resp2, err := logic.CreateUser(&userv1.CreateUserRequest{Phone: "13900000001", Nickname: "李四2"})
	require.NoError(t, err)
	assert.Equal(t, int32(10002), resp2.Base.Code)

	// 仅一条用户记录
	assert.Len(t, ub.data, 1, "重复注册不应产生重复用户行")
}

func TestCreateUser_AssignRoleFailure_DoesNotBlockRegistration(t *testing.T) {
	// C-RU-03: AssignRole 失败 → 仅告警不阻塞注册（用户仍落库成功）
	svc, permMock := certTestSvc(t)
	ub := userBaseModel(svc)

	permMock.EXPECT().ListRoles(gomock.Any(), gomock.Any()).Return(mockRegisteredUserRoles(), nil).AnyTimes()
	permMock.EXPECT().AssignRole(gomock.Any(), gomock.Any()).Return(nil, errors.New("permission service unavailable")).Times(1)

	logic := NewCreateUserLogic(context.Background(), svc)
	resp, err := logic.CreateUser(&userv1.CreateUserRequest{Phone: "13700000001", Nickname: "王五"})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code, "注册不应被 registered_user 分配失败阻塞")
	assert.Len(t, ub.data, 1, "用户应已落库")
}
