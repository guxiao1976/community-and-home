package notice

import (
	"context"
	"strconv"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-hub/api/internal/types"
)

// toContentPostInfo proto ContentPost → REST ContentPostInfo。
// R2 wire 兼容：REST 正文键保持 content（Go 字段 Text json:"content"，与 proto/DB text 分轨，ADR）。
func toContentPostInfo(p *communityv1.ContentPost) types.ContentPostInfo {
	if p == nil {
		return types.ContentPostInfo{}
	}
	attachments := make([]types.ContentPostAttachmentInfo, 0, len(p.Attachments))
	for _, a := range p.Attachments {
		attachments = append(attachments, types.ContentPostAttachmentInfo{
			Id:           a.Id,
			FileName:     a.FileName,
			FileUrl:      a.FileUrl,
			FileSize:     a.FileSize,
			FileType:     a.FileType,
			FileId:       a.FileId,
			ReviewStatus: a.ReviewStatus,
		})
	}
	return types.ContentPostInfo{
		Id:              p.Id,
		CommunityId:     p.CommunityId,
		Title:           p.Title,
		Text:            p.Text, // wire 键 content
		Role:            int32(p.Role),
		Publisher:       p.Publisher,
		PublisherId:     p.PublisherId,
		IsPinned:        p.IsPinned,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		SectionCode:     p.SectionCode,
		Status:          p.Status,
		AttachmentCount: p.AttachmentCount,
		Attachments:     attachments,
	}
}

// parseIDList REST []string（Snowflake ID string 形式）→ []int64（RPC JS_STRING 载体内型）。
// 含非数字 → 080005（评审 INFO 2——防静默忽略产生空范围误过校验）。
func parseIDList(in []string) ([]int64, error) {
	out := make([]int64, 0, len(in))
	for _, s := range in {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, errx.NewCodeError(types.CodeInvalidParam, "ID 参数含非数字")
		}
		out = append(out, v)
	}
	return out, nil
}

// filterAllowed 读数据范围判定（API 侧，R2 兼容回退用）。
// 语义与 rpc scope.FilterAllowed 一致：GLOBAL → true / LIMITED IN / EMPTY+无身份 → false。
func filterAllowed(ctx context.Context, permClient permissionv1.PermissionServiceClient, userID, requestedCommunityID int64) (bool, error) {
	if userID == 0 {
		return false, nil
	}
	resp, err := permClient.GetDataScopes(ctx, &permissionv1.GetDataScopesRequest{
		UserId:    userID,
		ScopeType: "community",
	})
	if err != nil {
		return false, err
	}
	switch resp.GetState() {
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL:
		return true, nil
	case permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED:
		for _, id := range resp.GetScopeIds() {
			if id == requestedCommunityID {
				return true, nil
			}
		}
		return false, nil
	default: // EMPTY / UNSPECIFIED
		return false, nil
	}
}
