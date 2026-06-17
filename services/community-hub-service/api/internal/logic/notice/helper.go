package notice

import (
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/types"
)

func toNoticeInfo(n *communityv1.Notice) types.NoticeInfo {
	if n == nil {
		return types.NoticeInfo{}
	}
	attachments := make([]types.NoticeAttachmentInfo, 0, len(n.Attachments))
	for _, a := range n.Attachments {
		attachments = append(attachments, types.NoticeAttachmentInfo{
			Id:       a.Id,
			FileName: a.FileName,
			FileUrl:  a.FileUrl,
			FileSize: a.FileSize,
		})
	}
	return types.NoticeInfo{
		Id:          n.Id,
		CommunityId: n.CommunityId,
		Title:       n.Title,
		Content:     n.Content,
		Role:        int32(n.Role),
		Publisher:   n.Publisher,
		PublisherId: n.PublisherId,
		IsPinned:    n.IsPinned,
		PublishedAt: n.PublishedAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
		Attachments: attachments,
	}
}
