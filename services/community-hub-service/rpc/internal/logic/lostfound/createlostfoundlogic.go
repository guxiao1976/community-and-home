package lostfound

import (
	"context"
	"encoding/json"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateLostFoundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLostFoundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLostFoundLogic {
	return &CreateLostFoundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateLostFoundLogic) CreateLostFound(in *communityv1.CreateLostFoundRequest) (*communityv1.CreateLostFoundResponse, error) {
	if in.Title == "" {
		return &communityv1.CreateLostFoundResponse{
			Base: responsex.NewBaseRespWithError(80005, "标题不能为空"),
		}, nil
	}

	imageUrlsJSON := "[]"
	if len(in.ImageUrls) > 0 {
		b, _ := json.Marshal(in.ImageUrls)
		imageUrlsJSON = string(b)
	}

	id := snowflake.NextID()
	item := &model.LostFoundItem{
		Id:           id,
		CommunityId:  in.CommunityId,
		Type:         typeToString(in.Type),
		Title:        in.Title,
		Description:  in.Description,
		ImageUrls:    imageUrlsJSON,
		ContactPhone: in.ContactPhone,
		Status:       "active",
		PublisherId:  in.PublisherId,
	}

	if _, err := l.svcCtx.LostFoundItemModel.Insert(l.ctx, item); err != nil {
		l.Errorf("CreateLostFound: insert failed: %v", err)
		return nil, err
	}

	l.Infof("CreateLostFound success: id=%d, communityId=%d", id, in.CommunityId)
	return &communityv1.CreateLostFoundResponse{
		Base: responsex.NewBaseResp(),
		Id:   id,
	}, nil
}
