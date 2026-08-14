package perm

import (
	"context"
	"testing"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/guxiao1976/community-permission/rpc/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockPermRpc 最小 gRPC 客户端 mock：内嵌 permission.PermissionService 接口，仅覆写 ListRoles
// 其余接口方法由嵌入的 nil 接口兜底（不会在本次路径被调用）
type mockPermRpc struct {
	permission.PermissionService
	listRolesFn func(ctx context.Context, in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error)
}

func (m *mockPermRpc) ListRoles(ctx context.Context, in *permissionv1.ListRolesRequest, opts ...grpc.CallOption) (*permissionv1.ListRolesResponse, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn(ctx, in)
	}
	return &permissionv1.ListRolesResponse{Base: responsex.NewBaseResp()}, nil
}

// TestListRoles_SortPassThrough — req.SortBy/SortOrder → grpcReq.Sort.Field/Order 透传
// RED: API 层尚未构造 grpcReq.Sort → gotReq.Sort 为 nil → assert.NotNil FAIL
func TestListRoles_SortPassThrough(t *testing.T) {
	var gotReq *permissionv1.ListRolesRequest
	mockRpc := &mockPermRpc{
		listRolesFn: func(ctx context.Context, in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
			gotReq = in
			return &permissionv1.ListRolesResponse{
				Base:  responsex.NewBaseResp(),
				Roles: []*permissionv1.Role{},
				Page:  &commonv1.PageResponse{Page: 1, PageSize: 10, Total: 0, TotalPages: 0},
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	sortBy := "role_name"
	sortOrder := "desc"
	req := &types.ListRolesReq{Page: 1, PageSize: 10, SortBy: &sortBy, SortOrder: &sortOrder}

	resp, err := logic.ListRoles(req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	require.NotNil(t, gotReq, "gRPC 请求应被调用")
	require.NotNil(t, gotReq.Sort, "sort 参数应透传到 gRPC 请求")
	assert.Equal(t, "role_name", gotReq.Sort.Field)
	assert.Equal(t, "desc", gotReq.Sort.Order)
}

// TestListRoles_SortOnlyNoOrder — 仅 sortBy、sortOrder 为 nil → Sort.Order 透传空串（由 RPC 层默认 asc）
func TestListRoles_SortOnlyNoOrder(t *testing.T) {
	var gotReq *permissionv1.ListRolesRequest
	mockRpc := &mockPermRpc{
		listRolesFn: func(ctx context.Context, in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
			gotReq = in
			return &permissionv1.ListRolesResponse{
				Base:  responsex.NewBaseResp(),
				Roles: []*permissionv1.Role{},
				Page:  &commonv1.PageResponse{Page: 1, PageSize: 10},
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	sortBy := "role_name"
	req := &types.ListRolesReq{Page: 1, PageSize: 10, SortBy: &sortBy}

	_, err := logic.ListRoles(req)
	assert.NoError(t, err)
	require.NotNil(t, gotReq)
	require.NotNil(t, gotReq.Sort)
	assert.Equal(t, "role_name", gotReq.Sort.Field)
	assert.Equal(t, "", gotReq.Sort.Order, "sortOrder 为空串由 RPC 层默认 asc")
}

// TestListRoles_BaseErrorToGoError — mock 返回 Base.Code=99400 → ListRoles 返回 Go error（非 nil）
// RED: API 层当前未检查 grpcResp.Base → 返回 nil err → assert.Error FAIL
// SEE: [[rpc-callback-must-check-response-base]] — 业务校验错误不可被静默吞掉
func TestListRoles_BaseErrorToGoError(t *testing.T) {
	mockRpc := &mockPermRpc{
		listRolesFn: func(ctx context.Context, in *permissionv1.ListRolesRequest) (*permissionv1.ListRolesResponse, error) {
			return &permissionv1.ListRolesResponse{
				Base: responsex.NewBaseRespWithError(int32(errx.CodeInvalidParam), "非法排序字段: evil"),
			}, nil
		},
	}
	svcCtx := &svc.ServiceContext{PermissionRpc: mockRpc}
	logic := NewListRolesLogic(context.Background(), svcCtx)

	resp, err := logic.ListRoles(&types.ListRolesReq{Page: 1, PageSize: 10})
	assert.Error(t, err, "Base.Code=99400 应转换为 Go error")
	assert.Nil(t, resp)
}
