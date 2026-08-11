package user

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// ApplyRole 申请角色（角色授予迁移到 permission-service 的 rel_user_role）
//   - owner/tenant/committee/grid_worker/property_admin/community_admin → 绑定小区（scope_type=community, scope_id=communityId）
//   - merchant → 全局（scope_type=global, scope_id=0）
//   - 申请时 status=0（未认证），认证通过后 permission-service 更新为 2

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

// ApplyRole 申请角色
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

	// 确定作用域（merchant 全局，其他绑定小区）
	scopeType := "global"
	scopeId := int64(0)

	if in.RoleCode != model.RoleCodeMerchant {
		// 查 membership，校验用户是该小区成员
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
		scopeType = "community"
		scopeId = in.CommunityId
	}

	// role_code → role_id（permission-service 的 sys_role）
	roleID, ok := roleIDByCode(l.ctx, l.svcCtx, l.Logger, in.RoleCode)
	if !ok {
		l.Errorf("ApplyRole: role_code=%s not found in permission-service", in.RoleCode)
		return &userv1.ApplyRoleResponse{
			Base: responsex.NewBaseRespWithError(10008, "角色不存在"),
		}, nil
	}

	// 调用 permission-service AssignRole（status=0 未认证）
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("ApplyRole: PermissionClient is nil")
		return &userv1.ApplyRoleResponse{
			Base: responsex.NewBaseRespWithError(50000, "系统繁忙"),
		}, nil
	}
	_, err := l.svcCtx.PermissionClient.AssignRole(l.ctx, &permissionv1.AssignRoleRequest{
		UserId:    in.UserId,
		RoleId:    roleID,
		ScopeType: scopeType,
		ScopeId:   scopeId,
		Status:    int32Ptr(0), // 未认证
	})
	if err != nil {
		l.Errorf("ApplyRole: AssignRole failed userId=%d roleId=%d err=%v", in.UserId, roleID, err)
		return nil, err
	}

	l.Infof("ApplyRole success, userId=%d, roleCode=%s, roleId=%d, scope=%s:%d",
		in.UserId, in.RoleCode, roleID, scopeType, scopeId)

	return &userv1.ApplyRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: &userv1.MembershipRole{
			UserId:      in.UserId,
			RoleCode:    in.RoleCode,
			CommunityId: scopeId,
			VerfStatus:  model.RoleVerfStatusUnverified,
		},
	}, nil
}
