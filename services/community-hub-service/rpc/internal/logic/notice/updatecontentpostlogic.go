package notice

import (
	"context"
	"database/sql"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateContentPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContentPostLogic {
	return &UpdateContentPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateContentPost draft 编辑 + attachment_count 重算 + is_pinned + submit（V5 presence 语义权威实现）。
//
// presence 语义（评审 interface v4 MUST 1）：Title/Text/SectionCode/IsPinned 为 proto3 `optional`
// 生成 `*string`/`*bool`（指针 != nil = 携带）；附件/scope 以 HasAttachmentChange/HasScopeChange bool 标志
// 判别（false=不改，true=全量替换集，空集=清空/080005）。分支判定以 presence/标志位为准，禁止 value 非空启发式。
//
// 授权分流（评审 data-model v3 M1 + V5 修订）：
//   - (a) 内容/附件/scope 编辑路径（含 status==1 submit）→ 先作者校验（非发布者 → 080002），
//     再按 draft/非 draft 走 080005（仅 draft 可内容编辑）；
//   - (b) 仅 is_pinned 路径 → 跳过作者校验，改验操作者授权（draft 发布者即可；
//     submitted/approved → PublishRolesFrom 非空 + AssertCommunitiesScope 覆盖帖小区，scope 不覆盖 → 080006）。
//
// 不再 LPUSH Redis（D3/评审 M3）：本逻辑整体移除原 updatenoticelogic 的 CreateAuditLog + LpushCtx 块，
// submit 路径只推 Kafka（Task 1.18），不既推 Kafka 又 LPUSH Redis。
func (l *UpdateContentPostLogic) UpdateContentPost(in *communityv1.UpdateContentPostRequest) (*communityv1.UpdateContentPostResponse, error) {
	post, err := l.svcCtx.ContentPostModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return l.err(CodePostNotFound, "内容帖不存在"), nil
		}
		l.Errorf("UpdateContentPost: find failed: %v", err)
		return nil, err
	}

	userID := scope.UserIDFromCtx(l.ctx)
	if userID == 0 {
		return l.err(CodePublishDenied, "无发布权限"), nil
	}

	// 授权分流（V5 presence 判定：IsPinned 为 *bool，指针 != nil = 携带）
	isPinnedOnly := in.IsPinned != nil && in.Title == nil && in.Text == nil && in.SectionCode == nil &&
		!in.HasScopeChange && !in.HasAttachmentChange && in.GetStatus() == 0

	if isPinnedOnly {
		return l.applyIsPinned(post, in, userID)
	}
	return l.applyContentEdit(post, in, userID)
}

// applyIsPinned (b) 仅 is_pinned 路径：跳过作者校验，改验操作者授权。
// 置顶（*true）与取消置顶（*false）均走本路径（presence 可判，取消置顶确定性可达，评审 MUST 1(a)）。
func (l *UpdateContentPostLogic) applyIsPinned(post *model.ContentPost, in *communityv1.UpdateContentPostRequest, userID int64) (*communityv1.UpdateContentPostResponse, error) {
	switch post.Status {
	case model.StatusDraft:
		// draft 帖 → 发布者即可
		if post.PublisherId == nil || *post.PublisherId != userID {
			return l.err(CodePublishDenied, "非帖作者"), nil
		}
	default: // submitted/approved
		// submitted/approved 帖 → PublishRolesFrom 非空 + AssertCommunitiesScope（数据范围覆盖帖小区）
		roles, err := scope.PublishRolesFrom(l.ctx, l.svcCtx.PermissionClient, userID)
		if err != nil {
			l.Errorf("UpdateContentPost: publish roles failed: %v", err)
			return nil, err
		}
		if len(roles) == 0 {
			return l.err(CodePublishDenied, "无发布权限"), nil
		}
		communities, err := l.svcCtx.ContentPostScopeModel.FindCommunityIdsByPostId(l.ctx, post.Id)
		if err != nil {
			l.Errorf("UpdateContentPost: find post scope failed: %v", err)
			return nil, err
		}
		if len(communities) == 0 {
			return l.err(scope.CodePublishScopeDenied, "目标小区超出发布者数据范围"), nil // 帖无 scope（数据异常）fail-closed
		}
		if err := scope.AssertCommunitiesScope(l.ctx, l.svcCtx.PermissionClient, userID, communities); err != nil {
			if scope.IsPublishScopeDenied(err) {
				return &communityv1.UpdateContentPostResponse{Base: scope.DenyBase()}, nil
			}
			l.Errorf("UpdateContentPost: assert publish scope failed: %v", err)
			return nil, err
		}
	}

	// 置顶/取消置顶一律走 UpdateIsPinned（独立列更新，不碰 title/text/section_code，评审 data-model v4 M1）
	isPinned := int32(0)
	if in.IsPinned != nil && *in.IsPinned {
		isPinned = 1
	}
	if err := l.svcCtx.ContentPostModel.UpdateIsPinned(l.ctx, post.Id, isPinned); err != nil {
		l.Errorf("UpdateContentPost: update is_pinned failed: %v", err)
		return nil, err
	}
	l.Infof("UpdateContentPost is_pinned: id=%d pinned=%d", post.Id, isPinned)
	return &communityv1.UpdateContentPostResponse{Base: responsex.NewBaseResp()}, nil
}

// applyContentEdit (a) 内容/附件/scope 编辑路径 + submit 动作。
func (l *UpdateContentPostLogic) applyContentEdit(post *model.ContentPost, in *communityv1.UpdateContentPostRequest, userID int64) (*communityv1.UpdateContentPostResponse, error) {
	// 作者校验（(a) 分支先行使 is_pinned 操作者路径不受影响，评审 data-model v3 M1）
	if post.PublisherId == nil || *post.PublisherId != userID {
		return l.err(CodePublishDenied, "非帖作者"), nil
	}
	// status 值语义（action：0=编辑 / 1=submit；其他 → 080005）
	if in.GetStatus() != 0 && in.GetStatus() != 1 {
		return l.err(scope.CodeInvalidParam, "status 非法"), nil
	}

	// submit 动作（status==1，仅 draft 可提交）
	if in.GetStatus() == 1 {
		if post.Status != model.StatusDraft {
			return l.err(scope.CodeInvalidParam, "仅 draft 可提交"), nil
		}
		now := time.Now()
		err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			return l.svcCtx.ContentPostModel.UpdateStatusAndPublishTx(ctx, session, post.Id, model.StatusApproved, now)
		})
		if err != nil {
			l.Errorf("UpdateContentPost: submit failed: %v", err)
			return nil, err
		}
		// 事务提交成功后 Producer.Push（先提交后推送，提交失败不推送，评审 data-model v3 S1）
		if l.svcCtx.KafkaProducer != nil {
			atts, err := l.svcCtx.ContentPostAttachmentModel.FindByPostId(l.ctx, post.Id)
			if err != nil {
				l.Errorf("UpdateContentPost: find attachments failed: %v", err)
			} else if err := l.svcCtx.KafkaProducer.Push(l.ctx, post, atts); err != nil {
				l.Errorf("UpdateContentPost: kafka push failed post=%d: %v", post.Id, err)
			}
		}
		l.Infof("UpdateContentPost submit: id=%d", post.Id)
		return &communityv1.UpdateContentPostResponse{Base: responsex.NewBaseResp()}, nil
	}

	// 非 draft 不可内容编辑（(a) 分支非 submit 路径）
	if post.Status != model.StatusDraft {
		return l.err(scope.CodeInvalidParam, "仅 draft 可编辑"), nil
	}

	// 内容字段全量替换语义（V5 显式声明，评审 MUST 1(c)）：present 字段覆盖、未携带保持现值
	title, text, sectionCode := post.Title, post.Text, post.SectionCode
	if in.Title != nil {
		if *in.Title == "" {
			return l.err(scope.CodeInvalidParam, "标题不能为空"), nil
		}
		title = *in.Title
	}
	if in.Text != nil {
		if *in.Text == "" {
			return l.err(scope.CodeInvalidParam, "内容不能为空"), nil
		}
		text = *in.Text
	}
	if in.SectionCode != nil {
		if *in.SectionCode == "" || *in.SectionCode != SectionCodeNotice {
			return l.err(scope.CodeInvalidParam, "板块不支持"), nil
		}
		sectionCode = *in.SectionCode
	}

	// 附件集合变更（HasAttachmentChange==true）→ 全量替换 + 复跑完整绑定校验（REQ-CPB-6）→ 080005 超限整体拒绝
	var newAttachments []*model.ContentPostAttachment
	if in.HasAttachmentChange {
		atts, err := bindAttachments(l.ctx, l.svcCtx.FileClient, userID, in.AttachmentIds)
		if err != nil {
			if base := baseFromError(err); base != nil {
				return &communityv1.UpdateContentPostResponse{Base: base}, nil
			}
			l.Errorf("UpdateContentPost: bind attachments failed: %v", err)
			return nil, err
		}
		newAttachments = atts
	}

	// scope 变更（HasScopeChange==true）→ 复跑 AssertCommunitiesScope（新目标集；空集 → 080005）+ 重写 scope 行
	var newTargets []int64
	if in.HasScopeChange {
		newTargets = dedupeInt64(in.CommunityIds)
		if len(newTargets) == 0 {
			return l.err(scope.CodeInvalidParam, "发布范围不能为空"), nil
		}
		if err := scope.AssertCommunitiesScope(l.ctx, l.svcCtx.PermissionClient, userID, newTargets); err != nil {
			if scope.IsPublishScopeDenied(err) {
				return &communityv1.UpdateContentPostResponse{Base: scope.DenyBase()}, nil
			}
			l.Errorf("UpdateContentPost: assert publish scope failed: %v", err)
			return nil, err
		}
	}

	// 单事务 all-or-nothing：正文 + 附件集合重写 + attachment_count 重算 + scope 重写
	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if in.Title != nil || in.Text != nil || in.SectionCode != nil {
			if err := l.svcCtx.ContentPostModel.UpdateContentTx(ctx, session, post.Id, title, text, sectionCode); err != nil {
				return err
			}
		}
		if in.HasAttachmentChange {
			if err := l.svcCtx.ContentPostAttachmentModel.DeleteByPostIdTx(ctx, session, post.Id); err != nil {
				return err
			}
			for _, a := range newAttachments {
				a.PostId = post.Id
			}
			if err := l.svcCtx.ContentPostAttachmentModel.InsertBatchTx(ctx, session, newAttachments); err != nil {
				return err
			}
			// 同事务重算 attachment_count（新绑定数；空集=0，D19 不变量可归零——评审 MUST 1(b)）
			if err := l.svcCtx.ContentPostModel.UpdateAttachmentCountTx(ctx, session, post.Id, int64(len(newAttachments))); err != nil {
				return err
			}
		}
		if in.HasScopeChange {
			if err := l.svcCtx.ContentPostScopeModel.DeleteByPostIdTx(ctx, session, post.Id); err != nil {
				return err
			}
			if err := l.svcCtx.ContentPostScopeModel.InsertBatchTx(ctx, session, post.Id, newTargets); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		l.Errorf("UpdateContentPost: tx update failed: %v", err)
		return nil, err
	}

	l.Infof("UpdateContentPost success: id=%d", post.Id)
	return &communityv1.UpdateContentPostResponse{Base: responsex.NewBaseResp()}, nil
}

func (l *UpdateContentPostLogic) err(code int32, msg string) *communityv1.UpdateContentPostResponse {
	return &communityv1.UpdateContentPostResponse{Base: responsex.NewBaseRespWithError(code, msg)}
}
