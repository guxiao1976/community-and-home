package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserRolesLogic) GetUserRoles(in *pb.GetUserRolesReq) (*pb.GetUserRolesResp, error) {
	// Get user roles
	userRoles, err := l.svcCtx.AuthUserRoleModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("Failed to get user roles: %v", err)
		return nil, err
	}

	if len(userRoles) == 0 {
		return &pb.GetUserRolesResp{Roles: []*pb.Role{}}, nil
	}

	// Get role IDs
	roleIds := make([]int64, len(userRoles))
	for i, ur := range userRoles {
		roleIds[i] = ur.RoleId
	}

	// Get roles
	roles, err := l.svcCtx.AuthRoleModel.FindByIds(l.ctx, roleIds)
	if err != nil {
		logx.Errorf("Failed to get roles: %v", err)
		return nil, err
	}

	// Convert to pb.Role
	pbRoles := make([]*pb.Role, 0, len(roles))
	for _, role := range roles {
		pbRoles = append(pbRoles, &pb.Role{
			Id:          role.Id,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description.String,
			IsSystem:    int32(role.IsSystem),
			Status:      int32(role.Status),
			CreatedTime: role.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: role.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	return &pb.GetUserRolesResp{Roles: pbRoles}, nil
}

