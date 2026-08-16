package notice

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/guxiao1976/community-hub/internal/contentcompat"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetContentPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContentPostLogic {
	return &GetContentPostLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// GetContentPost 详情代理（R2 wire 兼容：响应 JSON 键保持 notice；正文键 content）。
//
// community_id 兼容回退（R2）：req.CommunityId==0（移动端 getNoticeDetail(id) 不传）时，
// 经 contentcompat.ResolveReadableCommunityForCompat（scope 反查 + 逐小区 filterAllowed 任一允许即放行）
// 解析可读小区注入 RPC GetContentPost——多小区用户迁移后详情不 080005；
// 080001（全部不可读/不存在）/ 080005（帖无 scope，数据异常）透传。
func (l *GetContentPostLogic) GetContentPost(req *types.GetContentPostReq) (*types.GetContentPostResp, error) {
	callCtx, uid, err := l.svcCtx.CallCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	communityID := req.CommunityId
	if communityID == 0 {
		// R2 兼容回退（薄代理层）：RPC 层仍严格必填
		if l.svcCtx.ContentPostModel == nil || l.svcCtx.ContentPostScopeModel == nil {
			return nil, contentcompat.ErrorInvalidParam("缺少小区上下文")
		}
		resolved, err := contentcompat.ResolveReadableCommunityForCompat(l.ctx,
			l.svcCtx.ContentPostModel, l.svcCtx.ContentPostScopeModel, uid, req.Id,
			func(ctx context.Context, userID, cid int64) (bool, error) {
				return filterAllowed(ctx, l.svcCtx.PermClient, userID, cid)
			})
		if err != nil {
			return nil, err // 080001/080005 透传
		}
		communityID = resolved
	}

	resp, err := l.svcCtx.ContentPostServiceRpc.GetContentPost(callCtx, &communityv1.GetContentPostRequest{
		Id:          req.Id,
		CommunityId: communityID,
	})
	if err != nil {
		return nil, err
	}
	if err := responsex.ToError(resp.GetBase()); err != nil {
		return nil, err
	}
	return &types.GetContentPostResp{Notice: toContentPostInfo(resp.ContentPost)}, nil
}
