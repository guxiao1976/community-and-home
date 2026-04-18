package role

import (
	"context"

	"github.com/guxiao/community-and-home/services/identity/api/internal/svc"
	"github.com/guxiao/community-and-home/services/identity/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List roles
func NewGetRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRolesLogic {
	return &GetRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRolesLogic) GetRoles(req *types.GetRolesReq) (resp *types.GetRolesResp, err error) {
	// Set default pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	// Convert status to int64 pointer
	var status *int64
	if req.Status != nil {
		s := int64(*req.Status)
		status = &s
	}

	// Get roles list
	roles, err := l.svcCtx.AuthRoleModel.FindList(l.ctx, req.Page, req.PageSize, status)
	if err != nil {
		logx.Errorf("Failed to get roles list: %v", err)
		return nil, err
	}

	// Get total count
	total, err := l.svcCtx.AuthRoleModel.Count(l.ctx, status)
	if err != nil {
		logx.Errorf("Failed to count roles: %v", err)
		return nil, err
	}

	// Convert to response
	list := make([]types.Role, 0, len(roles))
	for _, role := range roles {
		list = append(list, types.Role{
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

	return &types.GetRolesResp{
		List:  list,
		Total: total,
	}, nil
}

