package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateContentPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContentPostLogic {
	return &CreateContentPostLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// CreateContentPost 发布内容帖（代理到 RPC ContentPostService）。
//
// REST []string → RPC []int64 JS_STRING（接口 v4 INFO 1 对齐）；entry_status 同号数值映射
// （REST 0↔RPC 0=draft、REST 1↔RPC 1=submitted，数值即语义、无枚举偏移，评审 M2）。
// 身份经 CallCtx 注入出站 gRPC metadata（publisher_id/role/publisher 由 RPC 服务端派生，禁请求体信任）。
func (l *CreateContentPostLogic) CreateContentPost(req *types.CreateContentPostReq) (*types.CreateContentPostResp, error) {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	communityIDs, err := parseIDList(req.CommunityIds)
	if err != nil {
		return nil, err
	}
	attachmentIDs, err := parseIDList(req.AttachmentIds)
	if err != nil {
		return nil, err
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.CreateContentPost(callCtx, &communityv1.CreateContentPostRequest{
		SectionCode:   req.SectionCode,
		Title:         req.Title,
		Text:          req.Text,
		EntryStatus:   req.EntryStatus, // 同号映射
		CommunityIds:  communityIDs,
		AttachmentIds: attachmentIDs,
	})
	if err != nil {
		return nil, err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return nil, responsex.ToError(resp.GetBase())
	}
	return &types.CreateContentPostResp{Id: resp.Id}, nil
}
