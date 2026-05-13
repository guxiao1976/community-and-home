package logic

import (
	"context"
	"fmt"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckPermissionLogic) CheckPermission(in *pb.CheckPermissionReq) (*pb.CheckPermissionResp, error) {
	// Get user's active roles
	userRoles, err := l.svcCtx.AuthUserRoleModel.FindActiveByUserId(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("Failed to get user roles: %v", err)
		return &pb.CheckPermissionResp{Allowed: false}, nil
	}

	// System roles bypass check
	for _, ur := range userRoles {
		if ur.IsSystem == 1 {
			return &pb.CheckPermissionResp{Allowed: true}, nil
		}
	}

	// Collect all permission codes for the user's roles
	permCodes := make(map[string]struct{})
	for _, ur := range userRoles {
		rolePerms, err := l.svcCtx.AuthRolePermissionModel.FindByRoleId(l.ctx, ur.RoleId)
		if err != nil {
			continue
		}
		permIds := make([]int64, 0, len(rolePerms))
		for _, rp := range rolePerms {
			permIds = append(permIds, rp.PermissionId)
		}
		if len(permIds) == 0 {
			continue
		}
		permissions, err := l.svcCtx.AuthPermissionModel.FindByIds(l.ctx, permIds)
		if err != nil {
			continue
		}
		for _, p := range permissions {
			if p.Status == 1 {
				permCodes[p.Code] = struct{}{}
			}
		}
	}

	// Check if the requested action+resource matches any permission code
	requestKey := fmt.Sprintf("%s:%s", in.Action, in.Resource)
	_, allowed := permCodes[requestKey]

	return &pb.CheckPermissionResp{Allowed: allowed}, nil
}
