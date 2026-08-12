package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteNoticeLogic {
	return &DeleteNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DeleteNotice 删除通知公告。
//
// 注入 JWT 身份 metadata 供 rpc 层 AssertPublishScope 数据权限校验（T4.3，超范围不删），
// 并将 rpc 业务错误（如 080006）透出给客户端。
func (l *DeleteNoticeLogic) DeleteNotice(req *types.DeleteNoticeReq) error {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return err
	}

	resp, err := l.svcCtx.NoticeServiceRpc.DeleteNotice(callCtx, &communityv1.DeleteNoticeRequest{
		Id: req.Id,
	})
	if err != nil {
		return err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return responsex.ToError(resp.GetBase())
	}
	return nil
}
