package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNoticeLogic {
	return &CreateNoticeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateNotice 发布通知公告。
//
// 身份规范化（T4.1）：publisher_id 一律取 JWT 认证身份，覆盖写入 gRPC 请求，
// 忽略客户端 body 传值（防伪造）。JWT 身份同时注入出站 gRPC metadata，
// 供 rpc 层落库前执行 AssertPublishScope 数据权限校验（T4.3）。
//
// SEE: [[verify-api-before-calling]] — 身份取自 JWT 而非客户端 body
func (l *CreateNoticeLogic) CreateNotice(req *types.CreateNoticeReq) (*types.CreateNoticeResp, error) {
	callCtx, uid, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.NoticeServiceRpc.CreateNotice(callCtx, &communityv1.CreateNoticeRequest{
		CommunityId: req.CommunityId,
		Title:       req.Title,
		Content:     req.Content,
		Role:        communityv1.NoticeRole(req.Role),
		Publisher:   req.Publisher,
		PublisherId: uid, // 覆盖客户端 body 值，忽略 req.PublisherId
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return nil, responsex.ToError(resp.GetBase())
	}
	return &types.CreateNoticeResp{Id: resp.Id}, nil
}
