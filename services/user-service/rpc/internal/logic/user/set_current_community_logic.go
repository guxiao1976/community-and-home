package user

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SetCurrentCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetCurrentCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetCurrentCommunityLogic {
	return &SetCurrentCommunityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetCurrentCommunity 切换当前小区：校验目标小区 ∈ 数据范围（permission GetDataScopes），越界返回 10015。
// SEE: [[rpc-identity-spoofing-loopback-isolation]] — user_id 取自请求体，安全前提是 RPC 仅可信网络可达。
func (l *SetCurrentCommunityLogic) SetCurrentCommunity(in *userv1.SetCurrentCommunityRequest) (*userv1.SetCurrentCommunityResponse, error) {
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("SetCurrentCommunity: PermissionClient is nil")
		return nil, fmt.Errorf("permission client unavailable")
	}

	scopes, err := l.svcCtx.PermissionClient.GetDataScopes(l.ctx, &permissionv1.GetDataScopesRequest{
		UserId:    in.UserId,
		ScopeType: model.ScopeTypeCommunity,
	})
	if err != nil {
		l.Errorf("SetCurrentCommunity: GetDataScopes failed userId=%d err=%v", in.UserId, err)
		return nil, err
	}

	if !inScope(scopes.State, scopes.ScopeIds, in.CommunityId) {
		return &userv1.SetCurrentCommunityResponse{
			Base: responsex.NewBaseRespWithError(10015, "目标小区不在数据范围"),
		}, nil
	}

	if err := l.svcCtx.UserAppStateModel.Upsert(l.ctx, in.UserId, in.CommunityId); err != nil {
		l.Errorf("SetCurrentCommunity: upsert app state error userId=%d communityId=%d err=%v", in.UserId, in.CommunityId, err)
		return nil, err
	}

	return &userv1.SetCurrentCommunityResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}

// inScope 判定目标小区是否在数据范围三态内：GLOBAL 放行；EMPTY 拒绝；LIMITED 仅当命中 scope_ids。
func inScope(state permissionv1.DataScopeState, scopeIds []int64, communityID int64) bool {
	switch state {
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL:
		return true
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY:
		return false
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED:
		for _, id := range scopeIds {
			if id == communityID {
				return true
			}
		}
		return false
	default:
		// UNSPECIFIED 视为未声明 → fail-closed（拒绝）
		return false
	}
}
