package notice

import (
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/model"
)

// roleToString 将 proto 枚举转为数据库 string
func roleToString(role communityv1.NoticeRole) string {
	switch role {
	case communityv1.NoticeRole_NOTICE_ROLE_COMMUNITY:
		return "community"
	case communityv1.NoticeRole_NOTICE_ROLE_COMMITTEE:
		return "committee"
	case communityv1.NoticeRole_NOTICE_ROLE_PROPERTY:
		return "property"
	case communityv1.NoticeRole_NOTICE_ROLE_GRID_OFFICER:
		return "grid_officer"
	default:
		return ""
	}
}

// stringToRole 将数据库 string 转为 proto 枚举
func stringToRole(s string) communityv1.NoticeRole {
	switch s {
	case "community":
		return communityv1.NoticeRole_NOTICE_ROLE_COMMUNITY
	case "committee":
		return communityv1.NoticeRole_NOTICE_ROLE_COMMITTEE
	case "property":
		return communityv1.NoticeRole_NOTICE_ROLE_PROPERTY
	case "grid_officer":
		return communityv1.NoticeRole_NOTICE_ROLE_GRID_OFFICER
	default:
		return communityv1.NoticeRole_NOTICE_ROLE_UNSPECIFIED
	}
}

func toProtoNotice(n *model.Notice) *communityv1.Notice {
	var publisherId int64
	if n.PublisherId != nil {
		publisherId = *n.PublisherId
	}
	return &communityv1.Notice{
		Id:          n.Id,
		CommunityId: n.CommunityId,
		Title:       n.Title,
		Content:     n.Content,
		Role:        stringToRole(n.Role),
		Publisher:   n.Publisher,
		PublisherId: publisherId,
		IsPinned:    n.IsPinned == 1,
		PublishedAt: n.PublishedAt.Unix(),
		CreatedAt:   n.CreatedAt.Unix(),
		UpdatedAt:   n.UpdatedAt.Unix(),
	}
}

var _ = time.Now
