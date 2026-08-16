package permission

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssertPublishScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssertPublishScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssertPublishScopeLogic {
	return &AssertPublishScopeLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AssertPublishScope 发布数据权限校验（T1.7，统一判据封装 + Task 3.1 社区管理员角色感知展开）
//
//	scope 解析（Task 3.1 改用 resolvePublishScope——社区管理员角色感知展开，resolveUserScope 读路径不动）：
//	  1. state==GLOBAL → allowed（审核员/sys_admin/系统身份）
//	  2. state==EMPTY  → denied（060007，registered_user 等无数据范围）
//	  3. LIMITED：逐 target——
//	     - target.scope_type != 'community' → denied（060005 不支持的类型）
//	     - master-data ResolveScopeAncestors(target.scope_id) found=false → denied（安全拒绝未知节点）
//	     - 祖先链 ∩ ids ≠ ∅ → 该 target covered，否则整体 denied（060007）
//	ids 含 community_admin 的 division 子树展开（其 community grant 展开为唯一 division 下全部 approved 小区）；
//	非 community_admin（owner/tenant/committee/property_admin/grid_worker）精确小区授权不展开。
//	gRPC 超时三层对齐：AssertPublishScope(≤500ms) 内嵌 ResolveScopeAncestors/GetResidentialArea/GetResidentialAreasByDivision(≤500ms)
//
// SEE: [[grpc-timeout-layers]]、[[is-system-no-permission-shortcut]]
func (l *AssertPublishScopeLogic) AssertPublishScope(in *permissionv1.AssertPublishScopeRequest) (*permissionv1.AssertPublishScopeResponse, error) {
	// 1. 解析用户 community 发布 scope → {state, ids}（Task 3.1 resolvePublishScope 社区管理员角色感知展开）
	state, ids := resolvePublishScope(l.ctx, l.svcCtx.UserRoleModel, l.svcCtx.MasterDataClient, in.UserId)

	// 2. GLOBAL 支配 → 放行
	if state == permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL {
		return &permissionv1.AssertPublishScopeResponse{
			Base:    responsex.NewBaseResp(),
			Allowed: true,
		}, nil
	}

	// 3. EMPTY / 无 target → 拒绝（060007 无数据范围权限）
	if state == permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY || len(in.Targets) == 0 {
		return deniedPublishScope(), nil
	}

	// 4. LIMITED：逐 target 校验
	for _, target := range in.Targets {
		if target.ScopeType != model.ScopeTypeCommunity {
			return &permissionv1.AssertPublishScopeResponse{
				Base: responsex.NewBaseRespWithError(60005, "不支持的 scope_type"),
			}, nil
		}
		if !l.targetCovered(target.ScopeId, ids) {
			return deniedPublishScope(), nil
		}
	}

	return &permissionv1.AssertPublishScopeResponse{
		Base:    responsex.NewBaseResp(),
		Allowed: true,
	}, nil
}

// targetCovered 校验目标节点是否被授权 id 集覆盖：祖先链 ∩ ids ≠ ∅
// 未知/失效节点（found=false）→ 安全拒绝
func (l *AssertPublishScopeLogic) targetCovered(nodeID int64, ids []int64) bool {
	resp, err := l.svcCtx.MasterDataClient.ResolveScopeAncestors(l.ctx, &masterdatav1.ResolveScopeAncestorsRequest{
		NodeId: nodeID,
	})
	if err != nil || resp == nil || !resp.Found {
		l.Infof("AssertPublishScope: node=%d resolve failed/not-found, deny", nodeID)
		return false
	}
	for _, anc := range resp.AncestorIds {
		for _, id := range ids {
			if anc == id {
				return true
			}
		}
	}
	return false
}

// deniedPublishScope 数据权限拒绝响应（060007，唯一语义：目标小区超出发布者数据范围）
// 060006 已登记为「角色编码已存在」（createrolelogic.go:30），本语义必须用独立错误码
// SEE: [[tdd-red-evidence-requires-fail-excerpt]]
func deniedPublishScope() *permissionv1.AssertPublishScopeResponse {
	return &permissionv1.AssertPublishScopeResponse{
		Base: responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"),
	}
}
