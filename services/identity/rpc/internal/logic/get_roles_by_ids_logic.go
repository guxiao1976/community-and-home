package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRolesByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRolesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolesByIdsLogic {
	return &GetRolesByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRolesByIdsLogic) GetRolesByIds(in *pb.GetRolesByIdsReq) (*pb.GetRolesByIdsResp, error) {
	if len(in.Ids) == 0 {
		return &pb.GetRolesByIdsResp{Roles: []*pb.Role{}}, nil
	}

	// Get roles by IDs
	roles, err := l.svcCtx.AuthRoleModel.FindByIds(l.ctx, in.Ids)
	if err != nil {
		logx.Errorf("Failed to get roles by ids: %v", err)
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

	return &pb.GetRolesByIdsResp{Roles: pbRoles}, nil
}

