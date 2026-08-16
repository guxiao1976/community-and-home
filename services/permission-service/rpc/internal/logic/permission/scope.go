package permission

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
)

// RoleCodeCommunityAdmin 社区管理员角色编码（resolvePublishScope 社区管理员角色感知展开判定）
// 与 sys_role.role_code / init_permissions.sql 常量一致（'community_admin'）。
// SEE: [[is-system-no-permission-shortcut]] — 角色判定仅驱动数据范围展开，非权限短路
const RoleCodeCommunityAdmin = "community_admin"

// resolveUserScope 合并用户在指定 scopeType 下的数据范围（REQ-A 三态合并）
//
//	供 GetDataScopes（读）与 AssertPublishScope（写）共用（SHOULD FIX S3）。
//
// 合并优先级：
//  1. ∃ grant.scope_type == 'global'          → state=GLOBAL  （global 支配）
//  2. limited_ids = ∪{scope_id : scope_type==scopeType AND scope_id!=0}
//     若 limited_ids≠∅                        → state=LIMITED  （并集）
//  3. 否则                                    → state=EMPTY    （空；”/0 占位零贡献）
//
// 三态互斥；EMPTY 永不等于 GLOBAL（REQ-1.2，杜绝「空当 global」灾难）。
// grants 由 FindActiveRolesByUserId 提供（status IN (0,1,2) 且未过期）。
// SEE: [[is-system-no-permission-shortcut]]
func resolveUserScope(ctx context.Context, urm model.RelUserRoleModel, userId int64, scopeType string) (permissionv1.DataScopeState, []int64) {
	grants, err := urm.FindActiveRolesByUserId(ctx, userId)
	if err != nil || len(grants) == 0 {
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, nil
	}

	// 1. global 支配
	for _, g := range grants {
		if !grantActive(g) {
			continue
		}
		if g.ScopeType == model.ScopeTypeGlobal {
			return permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL, nil
		}
	}

	// 2. limited 并集（排除 scope_id=0 占位；status=3,4 不计）
	ids := make([]int64, 0, len(grants))
	seen := make(map[int64]struct{}, len(grants))
	for _, g := range grants {
		if !grantActive(g) {
			continue
		}
		if g.ScopeType != scopeType || g.ScopeId == 0 {
			continue
		}
		if _, dup := seen[g.ScopeId]; dup {
			continue
		}
		seen[g.ScopeId] = struct{}{}
		ids = append(ids, g.ScopeId)
	}
	if len(ids) > 0 {
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, ids
	}

	// 3. empty
	return permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, nil
}

// grantActive 判定 grant 是否计入 scope 合并/能力聚合
// status ∈ {0,1,2} 视为活跃；3(已驳回)/4(已过期) 不计（防御性，SQL 已过滤）
func grantActive(g *model.UserRoleWithInfo) bool {
	return g != nil && g.URStatus >= 0 && g.URStatus <= 2
}

// resolvePublishScope 发布专用 scope 解析（Task 3.1，社区管理员角色感知展开，R1 grounded）
//
//	在 resolveUserScope 基线之上新增社区管理员角色感知展开：
//	  基线（同 resolveUserScope）：收集 scope_type='community' grant（ids 并集，含 GLOBAL 支配短路）
//	  社区管理员展开：用户持有 active community_admin 角色（grantActive，URStatus ∈ {0,1,2}）时，
//	    对每个 community_admin 的 community grant 的 scope_id（communityId）
//	    → masterdata GetResidentialArea(communityId).community_div_id → GetResidentialAreasByDivision(division, status=1)
//	    → approved 小区子树并入 ids（division 管辖权覆盖其下全部 approved 小区）
//	  非 community_admin（owner/tenant/committee/property_admin/grid_worker）语义完全不变（精确小区授权，不展开）
//
//	仅 AssertPublishScope 使用；GetDataScopes 读路径保持 resolveUserScope 不动。
//	已过期(4)/已驳回(3) 的 community_admin grant 不驱动展开（grantActive 语义，与 community-hub
//	ResolveAdminDivision URStatus 过滤对齐——防失效 grant 与另一 level-2 发布角色组合时越权放大授权）。
//	masterdata 展开失败 fail-closed（跳过该 grant 的子树贡献，不静默放大授权；目标级由 targetCovered 拒绝）。
//	SEE: [[is-system-no-permission-shortcut]]、[[grpc-timeout-layers]]（内嵌 GetResidentialArea/GetResidentialAreasByDivision）
func resolvePublishScope(ctx context.Context, urm model.RelUserRoleModel, mdClient masterdatav1.MasterdataServiceClient, userId int64) (permissionv1.DataScopeState, []int64) {
	grants, err := urm.FindActiveRolesByUserId(ctx, userId)
	if err != nil || len(grants) == 0 {
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, nil
	}

	// 1. global 支配（与 resolveUserScope 一致）
	for _, g := range grants {
		if !grantActive(g) {
			continue
		}
		if g.ScopeType == model.ScopeTypeGlobal {
			return permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL, nil
		}
	}

	// 2. community limited 基线并集（排除 scope_id=0 占位；status=3,4 不计）
	ids := make([]int64, 0, len(grants))
	seen := make(map[int64]struct{}, len(grants))
	for _, g := range grants {
		if !grantActive(g) {
			continue
		}
		if g.ScopeType != model.ScopeTypeCommunity || g.ScopeId == 0 {
			continue
		}
		if _, dup := seen[g.ScopeId]; dup {
			continue
		}
		seen[g.ScopeId] = struct{}{}
		ids = append(ids, g.ScopeId)
	}

	// 3. 社区管理员角色感知展开（Task 3.1，R1 grounded）
	if holdsCommunityAdminRole(grants) {
		ids = expandCommunityAdminDivision(ctx, mdClient, grants, ids, seen)
	}

	if len(ids) > 0 {
		return permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, ids
	}
	return permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY, nil
}

// holdsCommunityAdminRole 用户是否持有 active 社区管理员 grant（URStatus ∈ {0,1,2}）
// 已过期(4)/已驳回(3) 的 community_admin grant 不驱动 division 展开
func holdsCommunityAdminRole(grants []*model.UserRoleWithInfo) bool {
	for _, g := range grants {
		if g == nil || !grantActive(g) {
			continue
		}
		if g.RoleCode == RoleCodeCommunityAdmin {
			return true
		}
	}
	return false
}

// expandCommunityAdminDivision 社区管理员 division 子树展开（approved 小区并入 ids）
//
//	对每个 active community_admin 的 community grant（scope_id=communityId）：
//	  GetResidentialArea(communityId) → community_div_id（失败/无小区/Base 非零 → 跳过该 grant 的子树贡献，fail-closed）
//	  GetResidentialAreasByDivision(division, status=1) → approved 小区集并入 ids（去重）
//	展开空（division 无 approved 小区）→ 不贡献新 id，由 targetCovered 在目标级拒绝。
//	SEE: [[grpc-timeout-layers]] — GetResidentialArea/GetResidentialAreasByDivision 内嵌超时对齐
func expandCommunityAdminDivision(ctx context.Context, mdClient masterdatav1.MasterdataServiceClient, grants []*model.UserRoleWithInfo, ids []int64, seen map[int64]struct{}) []int64 {
	for _, g := range grants {
		if g == nil || !grantActive(g) {
			continue
		}
		if g.RoleCode != RoleCodeCommunityAdmin || g.ScopeType != model.ScopeTypeCommunity || g.ScopeId == 0 {
			continue
		}

		areaResp, err := mdClient.GetResidentialArea(ctx, &masterdatav1.GetResidentialAreaReq{Id: g.ScopeId})
		if err != nil || areaResp == nil || areaResp.GetBase().GetCode() != 0 || areaResp.GetResidentialArea() == nil || areaResp.GetResidentialArea().GetCommunityDivId() == 0 {
			continue // fail-closed：该 grant 的子树不展开，目标级 targetCovered 拒绝
		}

		divResp, err := mdClient.GetResidentialAreasByDivision(ctx, &masterdatav1.GetResidentialAreasByDivisionReq{
			CommunityDivId: areaResp.GetResidentialArea().GetCommunityDivId(),
			Status:         1, // approved only
		})
		if err != nil || divResp == nil || divResp.GetBase().GetCode() != 0 {
			continue
		}
		for _, a := range divResp.GetResidentialAreas() {
			if a == nil || a.GetId() == 0 {
				continue
			}
			if _, dup := seen[a.GetId()]; dup {
				continue
			}
			seen[a.GetId()] = struct{}{}
			ids = append(ids, a.GetId())
		}
	}
	return ids
}
