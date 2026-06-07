package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type JoinCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinCommunityLogic {
	return &JoinCommunityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// JoinCommunity 加入小区（per 设计文档 3.2）
func (l *JoinCommunityLogic) JoinCommunity(in *userv1.JoinCommunityRequest) (*userv1.JoinCommunityResponse, error) {
	// 0. 校验用户存在
	if _, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.UserId); err != nil {
		if err == model.ErrNotFound {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		return nil, err
	}

	// 1. 校验：最多加入 5 个小区
	count, err := l.svcCtx.UserCommunityMembershipModel.CountActiveByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("count active memberships error: %v", err)
		return nil, err
	}
	if count >= model.MaxCommunities {
		return &userv1.JoinCommunityResponse{
			Base: responsex.NewBaseRespWithError(10006, "最多加入 5 个小区"),
		}, nil
	}

	// 2. 检查是否已加入（uk_user_community 保证不重复）
	existing, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("check membership error: %v", err)
		return nil, err
	}
	if existing != nil && existing.BindStatus == model.MembershipBindStatusActive {
		// 已加入，不能重复加入
		return &userv1.JoinCommunityResponse{
			Base: responsex.NewBaseRespWithError(10007, "不能重复加入同一个小区"),
		}, nil
	}

	// 2.5. 校验同小区同地址唯一性
	if in.Building > 0 && in.Room > 0 {
		addrExisting, err := l.svcCtx.UserCommunityMembershipModel.FindByAddress(
			l.ctx, in.CommunityId, int(in.Building), int(in.Unit), int(in.Room))
		if err != nil && err != model.ErrNotFound {
			l.Errorf("check address uniqueness error: %v", err)
			return nil, err
		}
		if addrExisting != nil {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10011, "该地址已有人加入"),
			}, nil
		}
	}

	// 3. 如果之前退出过，重新激活；否则插入新记录
	now := time.Now()
	if existing != nil {
		err = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, existing.Id, model.MembershipBindStatusActive, now)
		if err != nil {
			l.Errorf("re-activate membership error: %v", err)
			return nil, err
		}
		existing.BindStatus = model.MembershipBindStatusActive
		// 更新地址信息
		_ = l.svcCtx.UserCommunityMembershipModel.UpdateAddress(l.ctx, existing.Id, int(in.Building), int(in.Unit), int(in.Room))
		existing.LeaveTime = sql.NullTime{}
		existing.UpdatedTime = now

		// 更新 preferences
		l.updateDefaultCommunity(in.UserId, in.CommunityId)

		return &userv1.JoinCommunityResponse{
			Base:       responsex.NewBaseResp(),
			Membership: toProtoMembership(existing),
		}, nil
	}

	membership := &model.UserCommunityMembership{
		Id:          snowflake.NextID(),
		UserId:      in.UserId,
		CommunityId: in.CommunityId,
		BindStatus:  model.MembershipBindStatusActive,
		JoinTime:    now,
		CreatedTime: now,
		UpdatedTime: now,
		Building:    int(in.Building),
		Unit:        int(in.Unit),
		Room:        int(in.Room),
	}

	_, err = l.svcCtx.UserCommunityMembershipModel.Insert(l.ctx, membership)
	if err != nil {
		l.Errorf("insert membership error: %v", err)
		return nil, err
	}

	// 4. 首次加入小区，设置 default_community_id
	l.updateDefaultCommunity(in.UserId, in.CommunityId)

	// 重新查询获取 ID
	created, _ := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)

	l.Infof("JoinCommunity success, userId=%d, communityId=%d", in.UserId, in.CommunityId)
	return &userv1.JoinCommunityResponse{
		Base:       responsex.NewBaseResp(),
		Membership: toProtoMembership(created),
	}, nil
}

// updateDefaultCommunity 首次加入小区时设置默认小区偏好
func (l *JoinCommunityLogic) updateDefaultCommunity(userId, communityId int64) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, userId)
	if err != nil {
		return
	}

	// 如果已有 default_community_id，不覆盖
	prefs := user.Preferences.String
	if prefs != "" {
		var m map[string]interface{}
		if json.Unmarshal([]byte(prefs), &m) == nil {
			if _, ok := m["default_community_id"]; ok {
				return
			}
		}
	}

	// 设置 default_community_id
	newPrefs := map[string]interface{}{"default_community_id": communityId}
	b, _ := json.Marshal(newPrefs)
	user.Preferences = sql.NullString{String: string(b), Valid: true}
	_ = l.svcCtx.UserBaseModel.Update(l.ctx, user)
}
