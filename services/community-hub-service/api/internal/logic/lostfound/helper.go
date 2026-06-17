package lostfound

import (
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/api/internal/types"
)

func toLostFoundItemInfo(it *communityv1.LostFoundItem) types.LostFoundItemInfo {
	if it == nil {
		return types.LostFoundItemInfo{}
	}
	return types.LostFoundItemInfo{
		Id:           it.Id,
		CommunityId:  it.CommunityId,
		Type:         int32(it.Type),
		Title:        it.Title,
		Description:  it.Description,
		ImageUrls:    it.ImageUrls,
		ContactPhone: it.ContactPhone,
		Status:       int32(it.Status),
		PublisherId:  it.PublisherId,
		CreatedAt:    it.CreatedAt,
	}
}
