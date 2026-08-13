package user

import (
	"context"
	"fmt"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// ==================== Get App State ====================

type GetAppStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAppStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppStateLogic {
	return &GetAppStateLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetAppStateLogic) GetAppState() (*types.GetAppStateResp, error) {
	userId := getUserIdFromJwt(l.ctx)
	if userId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}

	resp, err := l.svcCtx.UserRpc.GetAppState(l.ctx, &userv1.GetAppStateRequest{UserId: userId})
	if err != nil {
		return nil, err
	}
	if resp.Base != nil && resp.Base.GetCode() != 0 {
		return nil, responsex.ToError(resp.Base)
	}
	return &types.GetAppStateResp{
		CurrentCommunityId: resp.CurrentCommunityId,
		UpdatedAt:          resp.UpdatedAt,
	}, nil
}

// ==================== Set Current Community ====================

type SetCurrentCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetCurrentCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetCurrentCommunityLogic {
	return &SetCurrentCommunityLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SetCurrentCommunityLogic) SetCurrentCommunity(req *types.SetCurrentCommunityReq) (*types.SetCurrentCommunityResp, error) {
	userId := getUserIdFromJwt(l.ctx)
	if userId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}

	resp, err := l.svcCtx.UserRpc.SetCurrentCommunity(l.ctx, &userv1.SetCurrentCommunityRequest{
		UserId:      userId,
		CommunityId: req.CommunityId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base != nil && resp.Base.GetCode() != 0 {
		// 透出业务错误码（如 10015 目标小区不在数据范围）
		return nil, responsex.ToError(resp.Base)
	}
	return &types.SetCurrentCommunityResp{}, nil
}
