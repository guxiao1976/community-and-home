package permission

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// fakeMasterDataClient 仅覆盖 ResolveScopeAncestors，其余接口方法嵌入不调用
type fakeMasterDataClient struct {
	masterdatav1.MasterdataServiceClient
	resolveFn func(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error)
}

func (f *fakeMasterDataClient) ResolveScopeAncestors(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error) {
	return f.resolveFn(ctx, in, opts...)
}

// TestAssertPublishScope — T1.7 统一判据：GLOBAL 放行 / EMPTY 拒绝(060007) / 逐 target 祖先链 ∩ ids
// SEE: [[grpc-timeout-layers]] — AssertPublishScope 内嵌 master-data ResolveScopeAncestors
func TestAssertPublishScope(t *testing.T) {
	// 祖先链表：节点 → 自包含祖先链（self-first, ≤6）
	ancestors := map[int64][]int64{
		100: {100, 90, 80}, // 小区A → 社区 → 街道
		200: {200, 95, 85}, // 小区B（A 的兄弟社区，非同链）
	}

	communityTarget := func(id int64) []*permissionv1.ScopeRef {
		return []*permissionv1.ScopeRef{{ScopeType: model.ScopeTypeCommunity, ScopeId: id}}
	}

	tests := []struct {
		name        string
		grants      []*model.UserRoleWithInfo
		targets     []*permissionv1.ScopeRef
		wantAllowed bool
		wantCode    int32
	}{
		{
			name:        "owner@A 发 A✅（社区100 命中）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets:     communityTarget(100),
			wantAllowed: true,
		},
		{
			name:        "owner@A 发 B❌（060007，非同链）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets:     communityTarget(200),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			name:        "global 审核员任意✅（GLOBAL 放行，不调祖先链）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 8, ScopeType: model.ScopeTypeGlobal, ScopeId: 0, URStatus: 2}},
			targets:     communityTarget(999),
			wantAllowed: true,
		},
		{
			name:   "多目标部分未覆盖→整体拒绝❌（060007）",
			grants: []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets: []*permissionv1.ScopeRef{
				{ScopeType: model.ScopeTypeCommunity, ScopeId: 100},
				{ScopeType: model.ScopeTypeCommunity, ScopeId: 200},
			},
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			name:        "empty 拒绝❌（060007，注册用户无数据范围）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 9, ScopeType: model.ScopeTypeEmpty, ScopeId: 0, URStatus: 2}},
			targets:     communityTarget(100),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			name:        "未知节点拒绝❌（060007，found=false 安全拒绝）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 1, ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets:     communityTarget(999),
			wantAllowed: false,
			wantCode:    60007,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := miniredis.RunT(t)
			defer mr.Close()
			redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

			mockUserRole := new(MockUserRoleModel)
			mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(2001)).Return(tt.grants, nil)

			fakeMD := &fakeMasterDataClient{
				resolveFn: func(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error) {
					chain, ok := ancestors[in.NodeId]
					if !ok {
						return &masterdatav1.ResolveScopeAncestorsResponse{Found: false}, nil
					}
					return &masterdatav1.ResolveScopeAncestorsResponse{AncestorIds: chain, Found: true}, nil
				},
			}

			svcCtx := &svc.ServiceContext{
				UserRoleModel:    mockUserRole,
				RedisClient:      redisClient,
				MasterDataClient: fakeMD,
			}

			logic := NewAssertPublishScopeLogic(context.Background(), svcCtx)
			resp, err := logic.AssertPublishScope(&permissionv1.AssertPublishScopeRequest{
				UserId:  2001,
				Targets: tt.targets,
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, resp.Allowed, "Allowed 应为 %v", tt.wantAllowed)
			if !tt.wantAllowed {
				assert.Equal(t, tt.wantCode, resp.Base.Code, "拒绝错误码应为 %d", tt.wantCode)
			}
			mockUserRole.AssertExpectations(t)
		})
	}
}
