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

// divisionFakeMD master-data 客户端 fake（支持 ResolveScopeAncestors / GetResidentialArea / GetResidentialAreasByDivision）
// Task 3.1 division 展开测试专用；既有 fakeMasterDataClient 仅覆盖 ResolveScopeAncestors。
type divisionFakeMD struct {
	masterdatav1.MasterdataServiceClient
	resolveFn  func(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error)
	areaFn     func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error)
	areasDivFn func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error)
}

func (f *divisionFakeMD) ResolveScopeAncestors(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error) {
	return f.resolveFn(ctx, in, opts...)
}

func (f *divisionFakeMD) GetResidentialArea(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
	return f.areaFn(ctx, in, opts...)
}

func (f *divisionFakeMD) GetResidentialAreasByDivision(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
	return f.areasDivFn(ctx, in, opts...)
}

// divisionTopology 测试用 division 拓扑：
//
//	Division 10（D1）→ 小区 100（C_admin）/ 101 / 102
//	Division 20（D2）→ 小区 200 / 201
//	999 = 不存在节点（ResolveScopeAncestors found=false）
//
//	祖先链（self-first）：100/101/102 → 90 → 80；200/201 → 95 → 85
var divisionTopology = struct {
	ancestors  map[int64][]int64
	areaByID   map[int64]*masterdatav1.ResidentialArea
	areasByDiv map[int64][]*masterdatav1.ResidentialArea
}{
	ancestors: map[int64][]int64{
		100: {100, 90, 80},
		101: {101, 90, 80},
		102: {102, 90, 80},
		200: {200, 95, 85},
		201: {201, 95, 85},
	},
	areaByID: map[int64]*masterdatav1.ResidentialArea{
		100: {Id: 100, CommunityDivId: 10},
		200: {Id: 200, CommunityDivId: 20},
	},
	areasByDiv: map[int64][]*masterdatav1.ResidentialArea{
		10: {{Id: 100, CommunityDivId: 10}, {Id: 101, CommunityDivId: 10}, {Id: 102, CommunityDivId: 10}},
		20: {{Id: 200, CommunityDivId: 20}, {Id: 201, CommunityDivId: 20}},
	},
}

func communityTargetRef(id int64) []*permissionv1.ScopeRef {
	return []*permissionv1.ScopeRef{{ScopeType: model.ScopeTypeCommunity, ScopeId: id}}
}

// TestAssertPublishScope_CommunityAdminDivisionExpansion — Task 3.1 Design Gate 验证门禁
//
//	社区管理员角色感知展开：community_admin 的 community grant（scope_id=communityId）
//	  → GetResidentialArea(communityId).community_div_id → GetResidentialAreasByDivision(division, status=1)
//	  → approved 小区子树并入授权 ids；非 community_admin 角色语义完全不变（精确小区授权）。
//
// SEE: [[is-system-no-permission-shortcut]]（无字段短路）、[[grpc-timeout-layers]]（内嵌 master-data RPC）
func TestAssertPublishScope_CommunityAdminDivisionExpansion(t *testing.T) {
	adminGrant := func(scopeID, urStatus int64) *model.UserRoleWithInfo {
		return &model.UserRoleWithInfo{RoleId: 3, RoleCode: "community_admin", ScopeType: model.ScopeTypeCommunity, ScopeId: scopeID, URStatus: urStatus}
	}

	tests := []struct {
		name        string
		grants      []*model.UserRoleWithInfo
		targets     []*permissionv1.ScopeRef
		wantAllowed bool
		wantCode    int32
		areaErr     bool // 注入 GetResidentialArea 传输错误（fail-closed 验证）
	}{
		{
			// 门禁场景 1：community_admin 持 community grant（scope_id=C_admin=100 ∈ D1）发布 D1 内另一小区 101 → allowed
			name:        "community_admin@100 发同division 101 ✅（子树展开生效）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(101),
			wantAllowed: true,
		},
		{
			name:        "community_admin@100 发同division 102 ✅（子树展开全部 approved 小区）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(102),
			wantAllowed: true,
		},
		{
			// 门禁场景 2：发布 D1 外小区 C2=200 → 060007 denied
			name:        "community_admin@100 发 division 外 200 ❌（060007）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(200),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 门禁场景 3：目标小区不存在（found=false）→ denied（安全拒绝未知节点）
			name:        "community_admin@100 发不存在 999 ❌（found=false 安全拒绝）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(999),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 门禁场景 4：非 community_admin 不回归——grid_worker 精确小区授权不展开
			name:        "grid_worker@100 发同division 101 ❌（非admin不展开）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 4, RoleCode: "grid_worker", ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 2}},
			targets:     communityTargetRef(101),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			name:        "grid_worker@100 发本小区 100 ✅（精确授权保持）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 4, RoleCode: "grid_worker", ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 2}},
			targets:     communityTargetRef(100),
			wantAllowed: true,
		},
		{
			name:        "owner@100 发101 ❌ / owner@100 发100 ✅（不展开）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 1, RoleCode: "owner", ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets:     communityTargetRef(101),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 门禁场景 5：共享调用方回归（AssertPublishScope 还被 lostfound 创建、contacts upsert 调用）
			// community_admin 持 community grant（C_admin=100）时 lostfound/contacts 写到同 division 小区 101 → allowed
			name:        "共享调用方：community_admin 发同division 101 ✅（lostfound/contacts 同判据）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(101),
			wantAllowed: true,
		},
		{
			name:        "共享调用方：owner@100 发 division 外 101 ❌（精确授权边界保持）",
			grants:      []*model.UserRoleWithInfo{{RoleId: 1, RoleCode: "owner", ScopeType: model.ScopeTypeCommunity, ScopeId: 100, URStatus: 0}},
			targets:     communityTargetRef(101),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 门禁场景 6：社区管理员多 community grant 映射不同 division → 并集展开
			name: "community_admin 多grant多division 并集展开 201 ✅",
			grants: []*model.UserRoleWithInfo{
				adminGrant(100, 2), // D1
				adminGrant(200, 2), // D2
			},
			targets:     communityTargetRef(201),
			wantAllowed: true,
		},
		{
			name: "community_admin 多grant 并集外 999 ❌",
			grants: []*model.UserRoleWithInfo{
				adminGrant(100, 2),
				adminGrant(200, 2),
			},
			targets:     communityTargetRef(999),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 防御：GetResidentialArea 传输错误 → fail-closed，只保留基线小区 100，子树不展开
			name:        "community_admin GetResidentialArea 失败 → 101 ❌（fail-closed 不放大授权）",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 2)},
			targets:     communityTargetRef(101),
			wantAllowed: false,
			wantCode:    60007,
			areaErr:     true,
		},
		{
			// 防御：已过期(4) community_admin grant 不驱动展开（grantActive 语义对齐，与 community-hub ResolveAdminDivision URStatus 过滤一致）
			name:        "过期 community_admin grant(URStatus=4) 不展开 → 101 ❌",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 4)},
			targets:     communityTargetRef(101),
			wantAllowed: false,
			wantCode:    60007,
		},
		{
			// 防御：驳回(3) community_admin grant 不计入展开
			name:        "驳回 community_admin grant(URStatus=3) 不展开 → 101 ❌",
			grants:      []*model.UserRoleWithInfo{adminGrant(100, 3)},
			targets:     communityTargetRef(101),
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

			fakeMD := &divisionFakeMD{
				resolveFn: func(ctx context.Context, in *masterdatav1.ResolveScopeAncestorsRequest, opts ...grpc.CallOption) (*masterdatav1.ResolveScopeAncestorsResponse, error) {
					chain, ok := divisionTopology.ancestors[in.NodeId]
					if !ok {
						return &masterdatav1.ResolveScopeAncestorsResponse{Found: false}, nil
					}
					return &masterdatav1.ResolveScopeAncestorsResponse{AncestorIds: chain, Found: true}, nil
				},
				areaFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
					if tt.areaErr {
						return nil, assert.AnError
					}
					a, ok := divisionTopology.areaByID[in.Id]
					if !ok {
						return &masterdatav1.GetResidentialAreaResp{}, nil
					}
					return &masterdatav1.GetResidentialAreaResp{ResidentialArea: a}, nil
				},
				areasDivFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
					areas, ok := divisionTopology.areasByDiv[in.CommunityDivId]
					if !ok {
						return &masterdatav1.GetResidentialAreasByDivisionResp{}, nil
					}
					return &masterdatav1.GetResidentialAreasByDivisionResp{ResidentialAreas: areas}, nil
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
