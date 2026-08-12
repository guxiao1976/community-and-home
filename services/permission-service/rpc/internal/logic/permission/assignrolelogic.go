package permission

import (
	"context"
	"database/sql"
	"time"

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
//
//	校验角色存在 → 插入 rel_user_role（幂等） → 失效 Redis 权限缓存
func (l *AssignRoleLogic) AssignRole(in *permissionv1.AssignRoleRequest) (*permissionv1.AssignRoleResponse, error) {
	// 校验角色存在
	_, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.RoleId)
	if err != nil {
		return &permissionv1.AssignRoleResponse{
			Base: responsex.NewBaseRespWithError(60001, "角色不存在"),
		}, nil
	}

	// 解析个体角色生命周期参数（可选）
	status := int64(0) // 默认未认证
	if in.Status != nil {
		status = int64(*in.Status)
	}
	var verifiedAt, expiresAt sql.NullTime
	if in.VerifiedAt != nil && *in.VerifiedAt > 0 {
		verifiedAt = sql.NullTime{Time: time.Unix(*in.VerifiedAt, 0), Valid: true}
	}
	if in.ExpiresAt != nil && *in.ExpiresAt > 0 {
		expiresAt = sql.NullTime{Time: time.Unix(*in.ExpiresAt, 0), Valid: true}
	}

	// INSERT IGNORE 幂等：uk_user_role_scope 唯一键冲突静默跳过（不报错）
	err = l.svcCtx.UserRoleModel.InsertIgnore(l.ctx, &model.RelUserRole{
		UserId:     in.UserId,
		RoleId:     in.RoleId,
		ScopeType:  in.ScopeType,
		ScopeId:    in.ScopeId,
		Status:     status,
		VerifiedAt: verifiedAt,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		l.Errorf("AssignRole: insert failed userId=%d, roleId=%d: %v", in.UserId, in.RoleId, err)
		return nil, err
	}

	// 失效缓存（收敛到本处理器，不依赖调用方）
	invalidateUserCaches(l.ctx, l.svcCtx.RedisClient, in.UserId)

	l.Infof("AssignRole success: userId=%d, roleId=%d, scope=%s:%d", in.UserId, in.RoleId, in.ScopeType, in.ScopeId)

	return &permissionv1.AssignRoleResponse{Base: responsex.NewBaseResp()}, nil
}
