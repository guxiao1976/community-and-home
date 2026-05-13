package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionsLogic {
	return &GetUserPermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserPermissionsLogic) GetUserPermissions(in *pb.GetUserPermissionsReq) (*pb.GetUserPermissionsResp, error) {
	userRoles, err := l.svcCtx.AuthUserRoleModel.FindActiveByUserId(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("Failed to get user roles: %v", err)
		return nil, err
	}

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

	codes := make([]string, 0, len(permCodes))
	for code := range permCodes {
		codes = append(codes, code)
	}

	return &pb.GetUserPermissionsResp{Permissions: codes}, nil
}
