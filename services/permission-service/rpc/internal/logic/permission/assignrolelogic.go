package permission

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/model"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleLogic {
	return &AssignRoleLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// AssignRole 分配角色（spec/permission.md 核心逻辑流 1）
//   校验角色存在 → 插入 rel_user_role（幂等） → 失效 Redis 权限缓存
func (l *AssignRoleLogic) AssignRole(in *permissionv1.AssignRoleRequest) (*permissionv1.AssignRoleResponse, error) {
	// 校验角色存在
	_, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.RoleId)
	if err != nil {
		return &permissionv1.AssignRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	_, err = l.svcCtx.UserRoleModel.Insert(l.ctx, &model.RelUserRole{
		UserId:    in.UserId,
		RoleId:    in.RoleId,
		ScopeType: in.ScopeType,
		ScopeId:   in.ScopeId,
	})
	if err != nil {
		// 可能的唯一键冲突（已分配），幂等返回成功
		l.Infof("AssignRole: insert (may be duplicate) userId=%d, roleId=%d: %v", in.UserId, in.RoleId, err)
	}

	// 失效缓存
	l.invalidateCache(in.UserId)

	l.Infof("AssignRole success: userId=%d, roleId=%d, scope=%s:%d", in.UserId, in.RoleId, in.ScopeType, in.ScopeId)

	return &permissionv1.AssignRoleResponse{Base: responsex.NewBaseResp()}, nil
}

func (l *AssignRoleLogic) invalidateCache(userId int64) {
	l.svcCtx.RedisClient.Del(l.ctx,
		fmt.Sprintf("perm:user:%d", userId),
		fmt.Sprintf("perm:scopes:%d:community", userId),
		fmt.Sprintf("perm:scopes:%d:building", userId),
		fmt.Sprintf("perm:scopes:%d:unit", userId),
		fmt.Sprintf("perm:scopes:%d:grid", userId),
	)
}
