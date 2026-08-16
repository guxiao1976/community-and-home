package notice

import (
	"context"
	"database/sql"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateContentPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContentPostLogic {
	return &CreateContentPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateContentPost 通用图文发布主路径（多小区 scope + 附件绑定 + 单事务 + 入口状态 draft/submitted）。
//
// 校验顺序（任一步失败整体拒绝，无部分写，design §CreateContentPost）：
//  1. section_code ∈ 板块白名单 → 080005；title/text 非空 → 080005；
//  2. community_ids 去重；空 → 080005；展开后快照 >100 → 080003；
//  3. 社区管理员自动展开（R1 grounded）：经既有 community grant 派生唯一 division → 展开 approved 小区；
//  4. 附件绑定（REQ-CPB-6 单源）：GetFileUrl → confirmed + user_id 归属 + ≤10/≤50MB → 080005；
//  5. 数据权限：单次批量 AssertCommunitiesScope（任一越权 → 080006，all-or-nothing）；
//  6. 身份派生：publisher_id=JWT、role=PublishRolesFrom 映射、publisher=真实档案（禁请求体信任）；
//  7. 单事务落库（content_posts + content_post_scope + content_post_attachments）；
//  8. entry=submitted 立即提交 → 事务提交成功后 Producer.Push（先提交后推送，提交失败不推送）。
//
// 不再 LPUSH Redis `moderation:task:queue`（D3，content_posts 只走 Kafka）。
func (l *CreateContentPostLogic) CreateContentPost(in *communityv1.CreateContentPostRequest) (*communityv1.CreateContentPostResponse, error) {
	// 1. 板块白名单 + 正文非空
	if in.SectionCode != SectionCodeNotice {
		return l.err(scope.CodeInvalidParam, "板块不支持"), nil
	}
	if in.Title == "" || in.Text == "" {
		return l.err(scope.CodeInvalidParam, "标题和内容不能为空"), nil
	}
	if in.EntryStatus != 0 && in.EntryStatus != 1 {
		return l.err(scope.CodeInvalidParam, "entry_status 非法"), nil
	}

	// 2. community_ids 去重 + 空
	targets := dedupeInt64(in.CommunityIds)
	if len(targets) == 0 {
		return l.err(scope.CodeInvalidParam, "发布范围不能为空"), nil
	}

	// 身份（JWT metadata，rpc-identity-spoofing-loopback-isolation）
	userID := scope.UserIDFromCtx(l.ctx)
	if userID == 0 {
		return l.err(CodePublishDenied, "无发布权限"), nil
	}

	// 发布角色派生（level-2 已认证过滤）
	roles, err := scope.PublishRolesFrom(l.ctx, l.svcCtx.PermissionClient, userID)
	if err != nil {
		l.Errorf("CreateContentPost: publish roles failed: %v", err)
		return nil, err
	}
	if len(roles) == 0 {
		return l.err(CodePublishDenied, "无发布权限"), nil
	}

	// 3. 目标集解析（R1）：community_admin → ResolveAdminDivision + ExpandDivisionCommunities（快照）
	if hasRole(roles, scope.RoleCommunityAdmin) {
		division, err := scope.ResolveAdminDivision(l.ctx, l.svcCtx.PermissionClient, l.svcCtx.MasterDataClient, userID)
		if err != nil {
			if base := baseFromError(err); base != nil {
				return &communityv1.CreateContentPostResponse{Base: base}, nil
			}
			l.Errorf("CreateContentPost: resolve admin division failed: %v", err)
			return nil, err
		}
		targets, err = scope.ExpandDivisionCommunities(l.ctx, l.svcCtx.MasterDataClient, division)
		if err != nil {
			if base := baseFromError(err); base != nil {
				return &communityv1.CreateContentPostResponse{Base: base}, nil
			}
			l.Errorf("CreateContentPost: expand division failed: %v", err)
			return nil, err
		}
	}
	// 展开后快照 >100 → 080003（REVISION 按展开快照计量）
	if len(targets) > MaxPublishTargets {
		return l.err(CodeOverLimit, "发布目标数量超限（≤100）"), nil
	}

	// 4. 附件绑定（REQ-CPB-6 单源）
	attachments, err := bindAttachments(l.ctx, l.svcCtx.FileClient, userID, in.AttachmentIds)
	if err != nil {
		if base := baseFromError(err); base != nil {
			return &communityv1.CreateContentPostResponse{Base: base}, nil
		}
		l.Errorf("CreateContentPost: bind attachments failed: %v", err)
		return nil, err
	}

	// 5. 数据权限（单次批量，Task 1.7）
	if err := scope.AssertCommunitiesScope(l.ctx, l.svcCtx.PermissionClient, userID, targets); err != nil {
		if scope.IsPublishScopeDenied(err) {
			return &communityv1.CreateContentPostResponse{Base: scope.DenyBase()}, nil
		}
		l.Errorf("CreateContentPost: assert publish scope failed: %v", err)
		return nil, err
	}

	// 6. 身份派生（REVISION REQ-CPB-5：JWT/RBAC/真实档案，禁请求体信任）
	roleStr := scope.PublishRoleToString(roles[0])
	publisher, err := resolvePublisher(l.ctx, l.svcCtx.UserClient, userID)
	if err != nil {
		if base := baseFromError(err); base != nil {
			return &communityv1.CreateContentPostResponse{Base: base}, nil
		}
		l.Errorf("CreateContentPost: resolve publisher failed: %v", err)
		return nil, err
	}

	// 7. 单事务落库
	id := snowflake.NextID()
	entrySubmitted := in.EntryStatus == 1
	now := time.Now()
	post := &model.ContentPost{
		Id:              id,
		Title:           in.Title,
		Text:            in.Text,
		Role:            roleStr,
		Publisher:       publisher,
		PublisherId:     &userID,
		IsPinned:        0,
		SectionCode:     in.SectionCode,
		Status:          model.StatusDraft,
		AttachmentCount: int64(len(attachments)),
		KafkaPushStatus: model.KafkaPushNone,
	}
	if entrySubmitted {
		// 隐式通过（D16）：status=approved + published_at=NOW() + 待推标记（D20）
		post.Status = model.StatusApproved
		post.PublishedAt = sql.NullTime{Valid: true, Time: now}
		post.KafkaPushStatus = model.KafkaPushPending
	}

	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := l.svcCtx.ContentPostModel.InsertTx(ctx, session, post); err != nil {
			return err
		}
		if err := l.svcCtx.ContentPostScopeModel.InsertBatchTx(ctx, session, id, targets); err != nil {
			return err
		}
		for _, a := range attachments {
			a.PostId = id
		}
		if err := l.svcCtx.ContentPostAttachmentModel.InsertBatchTx(ctx, session, attachments); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		l.Errorf("CreateContentPost: tx insert failed: %v", err)
		return nil, err
	}

	// 8. entry=submitted → 事务提交成功后 Producer.Push（先提交后推送，提交失败不推送，评审 data-model v3 S1）
	if entrySubmitted && l.svcCtx.KafkaProducer != nil {
		if err := l.svcCtx.KafkaProducer.Push(l.ctx, post, attachments); err != nil {
			l.Errorf("CreateContentPost: kafka push failed post=%d: %v", id, err)
		}
	}

	l.Infof("CreateContentPost success: id=%d, targets=%d, submitted=%v", id, len(targets), entrySubmitted)
	return &communityv1.CreateContentPostResponse{
		Base: responsex.NewBaseResp(),
		Id:   id,
	}, nil
}

func (l *CreateContentPostLogic) err(code int32, msg string) *communityv1.CreateContentPostResponse {
	return &communityv1.CreateContentPostResponse{Base: responsex.NewBaseRespWithError(code, msg)}
}

func dedupeInt64(in []int64) []int64 {
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, v := range in {
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
