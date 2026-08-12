package scope

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
)

// FilterAllowed 按调用方数据范围判断目标小区是否可读（读列表过滤，T4.6 / REQ-1.6）。
//
// 调 GetDataScopes(userID,'community')，语义：
//   - GLOBAL  → 不过滤（true）；
//   - LIMITED → requestedCommunityID ∈ scope_ids（true），否则 false（若请求带 community_id 且不在范围内 → 空列表）；
//   - EMPTY / UNSPECIFIED → false（空列表，不泄露小区内部内容）。
//
// 空范围必须在逻辑层返回空列表——SQL 不能拼空 IN() 子句，故本函数只返回布尔判定，
// 由调用方（列表逻辑）在 false 时直接返回空结果。
//
// 传输错误原样返回（调用方决定 fail-closed 行为）。
func FilterAllowed(ctx context.Context, client permissionv1.PermissionServiceClient, userID, requestedCommunityID int64) (bool, error) {
	// userID==0 恒拒绝：0 是系统审核身份（global scope），仅服务间回调可用，禁止用户读路径借用。
	if userID == 0 {
		return false, nil
	}

	resp, err := client.GetDataScopes(ctx, &permissionv1.GetDataScopesRequest{
		UserId:    userID,
		ScopeType: ScopeTypeCommunity,
	})
	if err != nil {
		return false, err
	}

	switch resp.GetState() {
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL:
		return true, nil
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED:
		for _, id := range resp.GetScopeIds() {
			if id == requestedCommunityID {
				return true, nil
			}
		}
		return false, nil
	default: // EMPTY / UNSPECIFIED
		return false, nil
	}
}
