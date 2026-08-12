package permission

import (
	"context"
	"database/sql"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserRoleStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserRoleStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleStatusLogic {
	return &UpdateUserRoleStatusLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// UpdateUserRoleStatus 更新用户角色的生命周期状态
//
//	status: 2=已认证(通过) 3=已驳回 4=已过期
//	由 user-service 认证审核通过/驳回时调用，或过期检测触发
func (l *UpdateUserRoleStatusLogic) UpdateUserRoleStatus(in *permissionv1.UpdateUserRoleStatusRequest) (*permissionv1.UpdateUserRoleStatusResponse, error) {
	// 解析认证时间（可选）
	var verifiedAt, expiresAt sql.NullTime
	if in.VerifiedAt != nil && *in.VerifiedAt > 0 {
		verifiedAt = sql.NullTime{Time: time.Unix(*in.VerifiedAt, 0), Valid: true}
	}
	if in.ExpiresAt != nil && *in.ExpiresAt > 0 {
		expiresAt = sql.NullTime{Time: time.Unix(*in.ExpiresAt, 0), Valid: true}
	}

	// 更新状态
	err := l.svcCtx.UserRoleModel.UpdateRoleStatus(l.ctx,
		in.UserId, in.RoleId, in.ScopeType, in.ScopeId,
		int64(in.Status), verifiedAt, expiresAt)
	if err != nil {
		l.Errorf("UpdateUserRoleStatus: userId=%d roleId=%d status=%d err=%v", in.UserId, in.RoleId, in.Status, err)
		return nil, err
	}

	// 失效权限缓存（收敛到本处理器，不依赖调用方）
	invalidateUserCaches(l.ctx, l.svcCtx.RedisClient, in.UserId)

	l.Infof("UpdateUserRoleStatus: userId=%d roleId=%d status=%d", in.UserId, in.RoleId, in.Status)

	return &permissionv1.UpdateUserRoleStatusResponse{Base: responsex.NewBaseResp()}, nil
}
