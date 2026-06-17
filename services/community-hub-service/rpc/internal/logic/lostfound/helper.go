package lostfound

import (
	"encoding/json"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/model"
)

func typeToString(t communityv1.LostFoundType) string {
	switch t {
	case communityv1.LostFoundType_LOST_FOUND_TYPE_LOST:
		return "lost"
	case communityv1.LostFoundType_LOST_FOUND_TYPE_FOUND:
		return "found"
	default:
		return ""
	}
}

func stringToType(s string) communityv1.LostFoundType {
	switch s {
	case "lost":
		return communityv1.LostFoundType_LOST_FOUND_TYPE_LOST
	case "found":
		return communityv1.LostFoundType_LOST_FOUND_TYPE_FOUND
	default:
		return communityv1.LostFoundType_LOST_FOUND_TYPE_UNSPECIFIED
	}
}

func statusToString(s string) communityv1.LostFoundStatus {
	switch s {
	case "active":
		return communityv1.LostFoundStatus_LOST_FOUND_STATUS_ACTIVE
	case "resolved":
		return communityv1.LostFoundStatus_LOST_FOUND_STATUS_RESOLVED
	default:
		return communityv1.LostFoundStatus_LOST_FOUND_STATUS_UNSPECIFIED
	}
}

func toProtoLostFoundItem(it *model.LostFoundItem) *communityv1.LostFoundItem {
	var imageUrls []string
	if it.ImageUrls != "" {
		_ = json.Unmarshal([]byte(it.ImageUrls), &imageUrls)
	}
	if imageUrls == nil {
		imageUrls = []string{}
	}
	return &communityv1.LostFoundItem{
		Id:           it.Id,
		CommunityId:  it.CommunityId,
		Type:         stringToType(it.Type),
		Title:        it.Title,
		Description:  it.Description,
		ImageUrls:    imageUrls,
		ContactPhone: it.ContactPhone,
		Status:       statusToString(it.Status),
		PublisherId:  it.PublisherId,
		CreatedAt:    it.CreatedAt.Unix(),
	}
}
