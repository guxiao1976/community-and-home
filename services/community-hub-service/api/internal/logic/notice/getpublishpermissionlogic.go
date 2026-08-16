package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPublishPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPublishPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPublishPermissionLogic {
	return &GetPublishPermissionLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetPublishPermission 发布权限代理（响应 can_publish + publishable_roles）。
func (l *GetPublishPermissionLogic) GetPublishPermission(req *types.GetPublishPermissionReq) (*types.GetPublishPermissionResp, error) {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.GetPublishPermission(callCtx, &communityv1.GetPublishPermissionRequest{})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return nil, responsex.ToError(resp.GetBase())
	}

	roles := make([]int32, 0, len(resp.PublishableRoles))
	for _, r := range resp.PublishableRoles {
		roles = append(roles, int32(r))
	}
	return &types.GetPublishPermissionResp{
		CanPublish:       resp.CanPublish,
		PublishableRoles: roles,
	}, nil
}
