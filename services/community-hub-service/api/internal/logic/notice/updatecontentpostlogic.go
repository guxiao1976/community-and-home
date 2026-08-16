package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateContentPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentPostLogic {
	return &UpdateContentPostLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// UpdateContentPost 更新代理（V5 presence 语义转发，评审 interface v4 MUST 1/S3）。
//
// 转发规则：
//   - Title/Text/SectionCode/IsPinned 指针非 nil 才填 RPC optional 字段（nil 不填，presence 不丢失）；
//   - HasScopeChange/HasAttachmentChange 标志 + 对应数组直传（true 时空数组也转发——「清空」语义必须到达 RPC 层）；
//   - Status 同号映射转发。
//
// 禁止「指针解引用 + omitempty 后裸透传」将 *false/*空串 坍缩成未携带。
func (l *UpdateContentPostLogic) UpdateContentPost(req *types.UpdateContentPostReq) error {
	callCtx, _, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return err
	}

	rpcReq := &communityv1.UpdateContentPostRequest{Id: req.Id}
	if req.Title != nil {
		rpcReq.Title = req.Title
	}
	if req.Text != nil {
		rpcReq.Text = req.Text
	}
	if req.SectionCode != nil {
		rpcReq.SectionCode = req.SectionCode
	}
	rpcReq.HasScopeChange = req.HasScopeChange
	if req.HasScopeChange {
		ids, err := parseIDList(req.CommunityIds)
		if err != nil {
			return err
		}
		rpcReq.CommunityIds = ids // true 时空数组也转发（清空语义）
	}
	rpcReq.HasAttachmentChange = req.HasAttachmentChange
	if req.HasAttachmentChange {
		ids, err := parseIDList(req.AttachmentIds)
		if err != nil {
			return err
		}
		rpcReq.AttachmentIds = ids // true 时空数组也转发（清空语义）
	}
	if req.IsPinned != nil {
		rpcReq.IsPinned = req.IsPinned // *true 置顶 / *false 取消置顶，presence 不坍缩
	}
	rpcReq.Status = req.Status

	resp, err := l.svcCtx.ContentPostServiceRpc.UpdateContentPost(callCtx, rpcReq)
	if err != nil {
		return err
	}
	if !responsex.IsSuccess(resp.GetBase()) {
		return responsex.ToError(resp.GetBase())
	}
	return nil
}
