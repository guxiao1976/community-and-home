package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type LeaveCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveCommunityLogic {
	return &LeaveCommunityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LeaveCommunity 退出小区（per 设计文档 3.7）
func (l *LeaveCommunityLogic) LeaveCommunity(in *userv1.LeaveCommunityRequest) (*userv1.LeaveCommunityResponse, error) {
	// 1. 查 membership
	membership, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.LeaveCommunityResponse{
				Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
			}, nil
		}
		l.Errorf("find membership error: %v", err)
		return nil, err
	}
	if membership.BindStatus != model.MembershipBindStatusActive {
		return &userv1.LeaveCommunityResponse{
			Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
		}, nil
	}

	// 2. 更新 membership 状态
	now := time.Now()
	err = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, membership.Id, model.MembershipBindStatusLeft, now)
	if err != nil {
		l.Errorf("update membership status error: %v", err)
		return nil, err
	}

	// 3. 如果 preferences.default_community_id 是此小区，清空或更新为其他小区
	l.updateDefaultCommunityOnLeave(in.UserId, in.CommunityId)

	l.Infof("LeaveCommunity success, userId=%d, communityId=%d", in.UserId, in.CommunityId)
	return &userv1.LeaveCommunityResponse{
		Base: responsex.NewBaseResp(),
	}, nil
}

func (l *LeaveCommunityLogic) updateDefaultCommunityOnLeave(userId, communityId int64) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, userId)
	if err != nil {
		return
	}

	prefs := user.Preferences.String
	if prefs == "" {
		return
	}

	var m map[string]interface{}
	if json.Unmarshal([]byte(prefs), &m) != nil {
		return
	}

	dcid, ok := m["default_community_id"]
	if !ok {
		return
	}

	// 类型断言
	var currentDefault int64
	switch v := dcid.(type) {
	case float64:
		currentDefault = int64(v)
	case int64:
		currentDefault = v
	default:
		return
	}

	if currentDefault != communityId {
		return
	}

	// 查找用户其他有效 membership
	activeMemberships, err := l.svcCtx.UserCommunityMembershipModel.FindByUserId(l.ctx, userId)
	if err != nil || len(activeMemberships) == 0 {
		// 没有其他小区，置 NULL
		user.Preferences = sql.NullString{Valid: false}
	} else {
		// 更新为第一个其他小区
		newDefault := activeMemberships[0].CommunityId
		newPrefs := map[string]interface{}{"default_community_id": newDefault}
		b, _ := json.Marshal(newPrefs)
		user.Preferences = sql.NullString{String: string(b), Valid: true}
	}
	_ = l.svcCtx.UserBaseModel.Update(l.ctx, user)
}
