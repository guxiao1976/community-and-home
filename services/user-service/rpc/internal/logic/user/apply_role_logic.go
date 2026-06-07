package user

import (
	"context"
	"database/sql"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

// ApplyRole 不再直接创建 residence。
// owner/tenant 的 residence 延后到 SubmitCertification(存储房屋信息) → ReviewCertification(通过后创建)。
// 这样确保 residence 只在产权/租赁被验证后才存在。

type ApplyRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyRoleLogic {
	return &ApplyRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ApplyRole 申请角色（per 设计文档 3.3）
func (l *ApplyRoleLogic) ApplyRole(in *userv1.ApplyRoleRequest) (*userv1.ApplyRoleResponse, error) {
	// 0. 校验用户存在
	if _, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.UserId); err != nil {
		if err == model.ErrNotFound {
			return &userv1.ApplyRoleResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		return nil, err
	}

	// 商家角色特殊处理：不绑小区
	var membershipId sql.NullInt64
	var communityId int64

	if in.RoleCode == model.RoleCodeMerchant {
		communityId = 0
	} else {
		communityId = in.CommunityId

		// 查 membership
		membership, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
		if err != nil {
			if err == model.ErrNotFound {
				return &userv1.ApplyRoleResponse{
					Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
				}, nil
			}
			l.Errorf("find membership error: %v", err)
			return nil, err
		}
		if membership.BindStatus != model.MembershipBindStatusActive {
			return &userv1.ApplyRoleResponse{
				Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
			}, nil
		}
		membershipId = sql.NullInt64{Int64: membership.Id, Valid: true}
	}

	// 查是否已有该角色
	if membershipId.Valid {
		existing, err := l.svcCtx.UserMembershipRoleModel.FindByMembershipAndRole(l.ctx, membershipId.Int64, in.RoleCode)
		if err != nil && err != model.ErrNotFound {
			l.Errorf("find role error: %v", err)
			return nil, err
		}
		if existing != nil {
			return &userv1.ApplyRoleResponse{
				Base: responsex.NewBaseRespWithError(10008, "该角色已存在"),
			}, nil
		}
	}

	// 创建角色
	now := time.Now()
	roleId := snowflake.NextID()
	role := &model.UserMembershipRole{
		Id:           roleId,
		UserId:       in.UserId,
		MembershipId: membershipId,
		CommunityId:  communityId,
		RoleCode:     in.RoleCode,
		VerfStatus:   model.RoleVerfStatusUnverified,
		CreatedTime:  now,
		UpdatedTime:  now,
	}

	if _, err := l.svcCtx.UserMembershipRoleModel.Insert(l.ctx, role); err != nil {
		l.Errorf("insert role error: %v", err)
		return nil, err
	}

	// owner/tenant 的房屋记录延后到认证通过时创建
	// 房产证/租赁合同才是房屋归属的证明，未认证前不创建 residence
	// building/unit/room 信息在 SubmitCertification 时传入并暂存

	l.Infof("ApplyRole success, userId=%d, communityId=%d, roleCode=%s, roleId=%d",
		in.UserId, communityId, in.RoleCode, roleId)
	return &userv1.ApplyRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: toProtoRole(role),
	}, nil
}
