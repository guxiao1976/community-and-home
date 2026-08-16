package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteContentPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteContentPostLogic {
	return &DeleteContentPostLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// DeleteContentPost 撤回代理（真正越权判定交 RPC 080002 作者校验，REST 权限码 427 由 PermMiddleware 强制）。
func (l *DeleteContentPostLogic) DeleteContentPost(req *types.DeleteContentPostReq) error {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return err
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.DeleteContentPost(callCtx, &communityv1.DeleteContentPostRequest{
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
