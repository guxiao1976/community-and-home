package notice

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
)

// 注册板块白名单（本期仅 notice，D11）。
const SectionCodeNotice = "notice"

// 单帖附件上限（≤10 个 / ≤50MB，REQ-CPB-6 单源）。
const (
	MaxAttachmentsPerPost  = 10
	MaxTotalAttachmentSize = 50 * 1024 * 1024
)

// 发布目标数量上限（展开后快照计量，REVISION）。
const MaxPublishTargets = 100

// 业务错误码（08xxxx 命名空间，与 api/types 对齐；避免跨层依赖 rpc/internal/scope 之外常量）。
const (
	CodePostNotFound  = 80001 // 内容帖不存在
	CodePublishDenied = 80002 // 无发布权限 / 非帖作者
	CodeOverLimit     = 80003 // 发布目标数量超限
)

// ContentPostRoleToString ContentPostRole → DB role 列值映射（Task 1.13，评审 data-model v4 I2 收敛单源）。
// 与写侧 scope.PublishRoleToString 产生同一字符串集合 {community, committee, property, grid_officer}，
// 防止两份映射漂移。
func ContentPostRoleToString(role communityv1.ContentPostRole) string {
	switch role {
	case communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMUNITY:
		return "community"
	case communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMITTEE:
		return "committee"
	case communityv1.ContentPostRole_CONTENT_POST_ROLE_PROPERTY:
		return "property"
	case communityv1.ContentPostRole_CONTENT_POST_ROLE_GRID_OFFICER:
		return "grid_officer"
	default:
		return ""
	}
}

// stringToContentPostRole DB role 列值 → ContentPostRole 枚举。
func stringToContentPostRole(s string) communityv1.ContentPostRole {
	switch s {
	case "community":
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMUNITY
	case "committee":
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_COMMITTEE
	case "property":
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_PROPERTY
	case "grid_officer":
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_GRID_OFFICER
	default:
		return communityv1.ContentPostRole_CONTENT_POST_ROLE_UNSPECIFIED
	}
}

// toProtoContentPost model → proto 映射。
// communityID 由调用方以请求小区注入（scope 派生，不读弃用 community_id 列）；PublishedAt sql.NullTime null 感知。
func toProtoContentPost(n *model.ContentPost, communityID int64) *communityv1.ContentPost {
	var publisherId int64
	if n.PublisherId != nil {
		publisherId = *n.PublisherId
	}
	pb := &communityv1.ContentPost{
		Id:              n.Id,
		CommunityId:     communityID,
		Title:           n.Title,
		Text:            n.Text,
		Role:            stringToContentPostRole(n.Role),
		Publisher:       n.Publisher,
		PublisherId:     publisherId,
		IsPinned:        n.IsPinned == 1,
		SectionCode:     n.SectionCode,
		Status:          int32(n.Status),
		AttachmentCount: int32(n.AttachmentCount),
		CreatedAt:       n.CreatedAt.Unix(),
		UpdatedAt:       n.UpdatedAt.Unix(),
	}
	if n.PublishedAt.Valid {
		pb.PublishedAt = n.PublishedAt.Time.Unix()
	}
	return pb
}

// baseFromError 将 errx.CodeError 业务错误转为 BaseResp（如 080005/080002/080006）；
// 非 CodeError（gRPC/DB 传输错误）返回 nil，调用方按传输错误原样传播（fail-closed）。
func baseFromError(err error) *commonv1.BaseResp {
	ce := errx.FromError(err)
	if ce == nil {
		return nil
	}
	return responsex.NewBaseRespWithError(int32(ce.Code), ce.Msg)
}

// bindAttachments 逐 attachment_id 经 file GetFileUrl 校验并回读 FileInfo（REQ-CPB-6 单源）。
//
// 校验：FileInfo.confirmed==true 且 user_id==JWT；数量 ≤10；Σ file_size ≤ 50MB → 否则 080005。
// 回读：file_type/file_name/file_size 自 FileInfo；file_id=attachment_ids；file_url=占位空串（权威重生载体）；
// review_status=approved(1) 默认（D14）。gRPC 传输错误原样返回（fail-closed）。
//
// SEE: [[grpc-only-comms]] — 附件校验经 file GetFileUrl，不直连 uploaded_file
func bindAttachments(ctx context.Context, fileClient filev1.FileServiceClient, userID int64, attachmentIDs []int64) ([]*model.ContentPostAttachment, error) {
	if len(attachmentIDs) > MaxAttachmentsPerPost {
		return nil, errx.NewCodeError(scope.CodeInvalidParam, "附件数量超限（≤10 个）")
	}
	var totalSize int64
	atts := make([]*model.ContentPostAttachment, 0, len(attachmentIDs))
	for _, fid := range attachmentIDs {
		resp, err := fileClient.GetFileUrl(ctx, &filev1.GetFileUrlRequest{FileId: fid})
		if err != nil {
			return nil, err
		}
		f := resp.GetFile()
		if f == nil || !f.GetConfirmed() || f.GetUserId() != userID {
			return nil, errx.NewCodeError(scope.CodeInvalidParam, "附件引用无效")
		}
		totalSize += f.GetFileSize()
		if totalSize > MaxTotalAttachmentSize {
			return nil, errx.NewCodeError(scope.CodeInvalidParam, "附件总大小超限（≤50MB）")
		}
		ft := f.GetFileType()
		atts = append(atts, &model.ContentPostAttachment{
			Id:           snowflake.NextID(),
			FileName:     f.GetFileName(),
			FileUrl:      "", // 占位空串，file_id 权威重生载体
			FileSize:     f.GetFileSize(),
			ReviewStatus: model.AttachmentReviewApproved,
			FileId:       fid,
			FileType:     &ft,
		})
	}
	return atts, nil
}

// resolvePublisher 取用户真实档案展示名（经 user-service GetUsersByIds，Task 1.9 接线；禁请求体信任）。
// 展示名优先 RealName（认证回填），回退 Nickname；查无该用户 → 080005。
//
// SEE: [[grpc-only-comms]] — 档案经 gRPC 查询，禁请求体信任（堵展示名伪造向量）
func resolvePublisher(ctx context.Context, userClient userv1.UserServiceClient, userID int64) (string, error) {
	resp, err := userClient.GetUsersByIds(ctx, &userv1.GetUsersByIdsRequest{Ids: []int64{userID}})
	if err != nil {
		return "", err
	}
	for _, u := range resp.GetUsers() {
		if u.GetId() != userID {
			continue
		}
		if u.GetRealName() != "" {
			return u.GetRealName(), nil
		}
		if u.GetNickname() != "" {
			return u.GetNickname(), nil
		}
	}
	return "", errx.NewCodeError(scope.CodeInvalidParam, "用户档案不存在")
}
