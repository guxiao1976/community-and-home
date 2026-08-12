package permission

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeRoleLogic {
	return &RevokeRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// RevokeRole 撤销角色 → 删除 rel_user_role → 失效缓存
func (l *RevokeRoleLogic) RevokeRole(in *permissionv1.RevokeRoleRequest) (*permissionv1.RevokeRoleResponse, error) {
	scopeType := "community"
	scopeId := int64(0)
	if in.ScopeType != nil {
		scopeType = *in.ScopeType
	}
	if in.ScopeId != nil {
		scopeId = *in.ScopeId
	}

	err := l.svcCtx.UserRoleModel.DeleteByUserIdAndRoleId(l.ctx, in.UserId, in.RoleId, scopeType, scopeId)
	if err != nil {
		l.Errorf("RevokeRole: delete failed: %v", err)
		return nil, err
	}

	// 失效缓存（收敛到本处理器，不依赖调用方）
	invalidateUserCaches(l.ctx, l.svcCtx.RedisClient, in.UserId)

	l.Infof("RevokeRole success: userId=%d, roleId=%d, scope=%s:%d", in.UserId, in.RoleId, scopeType, scopeId)

	return &permissionv1.RevokeRoleResponse{Base: responsex.NewBaseResp()}, nil
}
