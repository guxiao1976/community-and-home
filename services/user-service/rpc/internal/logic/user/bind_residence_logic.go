package user

import (
	"context"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type BindResidenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindResidenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindResidenceLogic {
	return &BindResidenceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BindResidence 绑定房屋（仅认证业主/租户可操作）
// 一个小区只需一次认证，认证后可追加多套房屋
func (l *BindResidenceLogic) BindResidence(in *userv1.BindResidenceRequest) (*userv1.BindResidenceResponse, error) {
	// 1. 校验：只有认证业主/租户才能绑定房屋
	membership, err := l.svcCtx.UserCommunityMembershipModel.FindOne(l.ctx, in.MembershipId)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.BindResidenceResponse{
				Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
			}, nil
		}
		l.Errorf("find membership error: %v", err)
		return nil, err
	}
	if membership.BindStatus != model.MembershipBindStatusActive {
		return &userv1.BindResidenceResponse{
			Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
		}, nil
	}

	// 确认用户有认证通过的业主或租户角色
	ownerRole, err := l.svcCtx.UserMembershipRoleModel.FindByMembershipAndRole(l.ctx, membership.Id, model.RoleCodeOwner)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	isCertified := ownerRole != nil && ownerRole.VerfStatus == model.RoleVerfStatusApproved
	if !isCertified {
		tenantRole, _ := l.svcCtx.UserMembershipRoleModel.FindByMembershipAndRole(l.ctx, membership.Id, model.RoleCodeTenant)
		if tenantRole == nil || tenantRole.VerfStatus != model.RoleVerfStatusApproved {
			return &userv1.BindResidenceResponse{
				Base: responsex.NewBaseRespWithError(10005, "只有认证通过的业主或租户才能绑定房屋"),
			}, nil
		}
	}

	// 2. 创建或更新房屋记录
	houseId := buildHouseId(in.Building, in.Unit, in.Room)

	existing, err := l.svcCtx.UserResidenceModel.FindByMembershipAndHouse(l.ctx, in.MembershipId, houseId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("find residence error: %v", err)
		return nil, err
	}
	if existing != nil {
		existing.Building = in.Building
		existing.Unit = in.Unit
		existing.Room = in.Room
		existing.IsPrimary = int64(in.IsPrimary)
		if in.StartDate != "" {
			existing.StartDate = parseDate(in.StartDate)
		}
		if in.EndDate != "" {
			existing.EndDate = parseDate(in.EndDate)
		}
		err = l.svcCtx.UserResidenceModel.Update(l.ctx, existing)
		if err != nil {
			l.Errorf("update residence error: %v", err)
			return nil, err
		}
		return &userv1.BindResidenceResponse{
			Base:      responsex.NewBaseResp(),
			Residence: toProtoResidence(existing),
		}, nil
	}

	now := time.Now()
	residence := &model.UserResidence{
		Id:           snowflake.NextID(),
		MembershipId: in.MembershipId,
		UserId:       membership.UserId,
		HouseId:      houseId,
		Building:     in.Building,
		Unit:         in.Unit,
		Room:         in.Room,
		IsPrimary:    int64(in.IsPrimary),
		CreatedTime:  now,
		UpdatedTime:  now,
	}
	if in.StartDate != "" {
		residence.StartDate = parseDate(in.StartDate)
	}
	if in.EndDate != "" {
		residence.EndDate = parseDate(in.EndDate)
	}

	_, err = l.svcCtx.UserResidenceModel.Insert(l.ctx, residence)
	if err != nil {
		l.Errorf("insert residence error: %v", err)
		return nil, err
	}

	created, _ := l.svcCtx.UserResidenceModel.FindByMembershipAndHouse(l.ctx, in.MembershipId, houseId)

	return &userv1.BindResidenceResponse{
		Base:      responsex.NewBaseResp(),
		Residence: toProtoResidence(created),
	}, nil
}
