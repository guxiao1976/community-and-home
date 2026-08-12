package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
)

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
