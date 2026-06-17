package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserModerationStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserModerationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserModerationStatusLogic {
	return &UpdateUserModerationStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserModerationStatusLogic) UpdateUserModerationStatus(in *userv1.UpdateModerationStatusRequest) (*userv1.UpdateModerationStatusResponse, error) {
	switch in.Target {
	case "nickname":
		if err := l.svcCtx.UserBaseModel.UpdateNicknameModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
			l.Errorf("UpdateNicknameModerationStatus failed: %v", err)
			return nil, err
		}
	case "certification":
		if err := l.svcCtx.UserCertificationModel.UpdateModerationStatus(l.ctx, in.Id, int64(in.ModerationStatus)); err != nil {
			l.Errorf("UpdateCertificationModerationStatus failed: %v", err)
			return nil, err
		}
	default:
		return &userv1.UpdateModerationStatusResponse{
			Base: responsex.NewBaseRespWithError(10005, "invalid target: must be 'nickname' or 'certification'"),
		}, nil
	}
	return &userv1.UpdateModerationStatusResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}
