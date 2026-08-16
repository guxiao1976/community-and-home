package notice

import (
	"context"
	"database/sql"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteContentPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteContentPostLogic {
	return &DeleteContentPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteContentPost 撤回（仅发布者本人，REVISION）。
//
// 可用状态：draft/submitted/approved 均可删（draft 可删、submitted 不可编辑但可删）；
// 作者校验：publisher_id == JWT user_id，否则 080002（仅发布者本人，REUSE:notice-D19）；
// 单事务：Withdraw（软删 + status=withdrawn 单语句原子）——content_post_scope 行与附件行全部保留
// （帖的撤回由主表软删+withdrawn 表达，REQ-CPB-10）；不推 Kafka（撤回非审核提交）。
func (l *DeleteContentPostLogic) DeleteContentPost(in *communityv1.DeleteContentPostRequest) (*communityv1.DeleteContentPostResponse, error) {
	post, err := l.svcCtx.ContentPostModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return l.err(CodePostNotFound, "内容帖不存在"), nil
		}
		l.Errorf("DeleteContentPost: find failed: %v", err)
		return nil, err
	}

	// 作者校验（仅发布者本人，REUSE:notice-D19）
	userID := scope.UserIDFromCtx(l.ctx)
	if userID == 0 || post.PublisherId == nil || *post.PublisherId != userID {
		return l.err(CodePublishDenied, "非帖作者"), nil
	}

	// 单事务：Withdraw（软删 + status=withdrawn 单语句原子，无半态）；scope/附件行保留（REQ-CPB-10）
	if err := l.svcCtx.ContentPostModel.Withdraw(l.ctx, post.Id); err != nil {
		l.Errorf("DeleteContentPost: withdraw failed: %v", err)
		return nil, err
	}

	l.Infof("DeleteContentPost success: id=%d", post.Id)
	return &communityv1.DeleteContentPostResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}

func (l *DeleteContentPostLogic) err(code int32, msg string) *communityv1.DeleteContentPostResponse {
	return &communityv1.DeleteContentPostResponse{Base: responsex.NewBaseRespWithError(code, msg)}
}
