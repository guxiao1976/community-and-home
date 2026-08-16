package notice

import (
	"context"
	"database/sql"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/internal/contentcompat"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetContentPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContentPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContentPostLogic {
	return &GetContentPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetContentPost 详情读路径（community_id RPC 契约必填）。
//
// FindOneReviewComplete(id)（审核完整性谓词）未找到 → 080001；
// content_post_scope 匹配 (id, community_id) 缺失 → 080001；
// FilterAllowed(userID, community_id) false → 080001（scope 外/不存在/未完整统一 080001，不泄露）。
// 附件 file_url 按 file_id 重生（GetFileUrl）；兼容期 file_id=0 回退 stored file_url。
func (l *GetContentPostLogic) GetContentPost(in *communityv1.GetContentPostRequest) (*communityv1.GetContentPostResponse, error) {
	// community_id RPC 必填（缺失/空 → 080005；REST 兼容回退只落薄代理层，RPC 严格必填——R2）
	if in.CommunityId <= 0 {
		return l.err(scope.CodeInvalidParam, "缺少小区上下文"), nil
	}

	post, err := l.svcCtx.ContentPostModel.FindOneReviewComplete(l.ctx, in.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return l.err(CodePostNotFound, "内容帖不存在"), nil
		}
		l.Errorf("GetContentPost: find failed: %v", err)
		return nil, err
	}

	// scope 匹配 (id, community_id)
	communities, err := l.svcCtx.ContentPostScopeModel.FindCommunityIdsByPostId(l.ctx, in.Id)
	if err != nil {
		l.Errorf("GetContentPost: find scope failed: %v", err)
		return nil, err
	}
	if !containsInt64(communities, in.CommunityId) {
		return l.err(CodePostNotFound, "内容帖不存在"), nil
	}

	// 读数据范围（GLOBAL 放行 / LIMITED IN / EMPTY 拒绝 → 080001 不泄露）
	userID := scope.UserIDFromCtx(l.ctx)
	allowed, err := scope.FilterAllowed(l.ctx, l.svcCtx.PermissionClient, userID, in.CommunityId)
	if err != nil {
		l.Errorf("GetContentPost: filter by scope failed: %v", err)
		return nil, err
	}
	if !allowed {
		return l.err(CodePostNotFound, "内容帖不存在"), nil
	}

	// 附件 file_url 重生（file_id 权威载体；兼容期 file_id=0 回退 stored file_url）
	attachments, err := l.svcCtx.ContentPostAttachmentModel.FindByPostId(l.ctx, in.Id)
	if err != nil {
		l.Errorf("GetContentPost: find attachments failed: %v", err)
		return nil, err
	}
	pbAttachments, err := l.toProtoAttachments(attachments)
	if err != nil {
		l.Errorf("GetContentPost: regenerate attachment urls failed: %v", err)
		return nil, err
	}

	pb := toProtoContentPost(post, in.CommunityId)
	pb.Attachments = pbAttachments
	return &communityv1.GetContentPostResponse{
		Base:        responsex.NewBaseResp(),
		ContentPost: pb,
	}, nil
}

// toProtoAttachments 组装附件（含 file_url 重生）。
func (l *GetContentPostLogic) toProtoAttachments(atts []*model.ContentPostAttachment) ([]*communityv1.ContentPostAttachment, error) {
	out := make([]*communityv1.ContentPostAttachment, 0, len(atts))
	for _, a := range atts {
		fileURL := a.FileUrl // 兼容期回退 stored file_url
		if a.FileId > 0 {
			resp, err := l.svcCtx.FileClient.GetFileUrl(l.ctx, &filev1.GetFileUrlRequest{FileId: a.FileId})
			if err != nil {
				return nil, err
			}
			if resp.GetDownloadUrl() != "" {
				fileURL = resp.GetDownloadUrl()
			}
		}
		out = append(out, &communityv1.ContentPostAttachment{
			Id:           a.Id,
			FileName:     a.FileName,
			FileUrl:      fileURL,
			FileSize:     a.FileSize,
			FileType:     derefFileType(a.FileType),
			FileId:       a.FileId,
			ReviewStatus: int32(a.ReviewStatus),
		})
	}
	return out, nil
}

func (l *GetContentPostLogic) err(code int32, msg string) *communityv1.GetContentPostResponse {
	return &communityv1.GetContentPostResponse{Base: responsex.NewBaseRespWithError(code, msg)}
}

func containsInt64(ids []int64, target int64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func derefFileType(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ResolveReadableCommunityForCompat 详情 community_id 兼容回退（R2，评审 interface v3 MUST 2 修复——
// 取消 grant 唯一假设，改 scope 反查）。委托共享实现 internal/contentcompat（REST 薄代理层可复用），
// scope 过滤回调注入 rpc scope.FilterAllowed（GetDataScopes）。
//
// 供 REST 薄代理层在 `GET /notices/:id` 缺 community_id 时调用：
//   - 帖有 scope 但全部不可读 → 080001（不泄露，与 RPC 层 scope 外统一 080001 一致）；
//   - 帖无任何 scope 小区（数据异常）→ 080005。
//
// RPC 层 GetContentPost 保持 community_id 必填不变（新消费方走严格契约），回退只落 REST 薄代理层。
func ResolveReadableCommunityForCompat(ctx context.Context, postModel model.ContentPostModel, scopeModel model.ContentPostScopeModel,
	permClient permissionv1.PermissionServiceClient, userID, postID int64) (int64, error) {
	return contentcompat.ResolveReadableCommunityForCompat(ctx, postModel, scopeModel, userID, postID,
		func(ctx context.Context, uid, cid int64) (bool, error) {
			return scope.FilterAllowed(ctx, permClient, uid, cid)
		})
}
