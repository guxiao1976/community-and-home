package division

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDivisionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionsLogic {
	return &GetDivisionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDivisionsLogic) GetDivisions(req *types.GetDivisionsReq) (resp *types.GetDivisionsResp, err error) {
	// Tree mode: recursively build the tree
	if req.Mode == "tree" {
		tree, err := l.buildTree(req.ParentId)
		if err != nil {
			return nil, errorx.NewDefaultError("查询行政区域树失败")
		}
		return &types.GetDivisionsResp{Tree: tree, Total: int64(len(tree))}, nil
	}

	// List mode: query with filters
	var divisions []*model.MdAdministrativeDivision
	if req.ParentId != nil {
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, *req.ParentId)
	} else if req.Level != nil {
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindByLevel(l.ctx, int64(*req.Level))
	} else {
		divisions, err = l.svcCtx.MdAdministrativeDivisionModel.FindRootDivisions(l.ctx)
	}
	if err != nil {
		return nil, errorx.NewDefaultError("查询行政区域失败")
	}

	var list []types.Division
	for _, d := range divisions {
		list = append(list, toDivisionType(d))
	}

	return &types.GetDivisionsResp{List: list, Total: int64(len(list))}, nil
}

func (l *GetDivisionsLogic) buildTree(rootParentId *int64) ([]types.DivisionTree, error) {
	var roots []*model.MdAdministrativeDivision
	var err error

	if rootParentId != nil {
		roots, err = l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, *rootParentId)
	} else {
		roots, err = l.svcCtx.MdAdministrativeDivisionModel.FindRootDivisions(l.ctx)
	}
	if err != nil {
		return nil, err
	}

	var tree []types.DivisionTree
	for _, root := range roots {
		node := types.DivisionTree{
			Division: toDivisionType(root),
		}
		children, err := l.buildTreeRecursive(root)
		if err != nil {
			return nil, err
		}
		node.Children = children
		tree = append(tree, node)
	}
	return tree, nil
}

func (l *GetDivisionsLogic) buildTreeRecursive(parent *model.MdAdministrativeDivision) ([]types.DivisionTree, error) {
	children, err := l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, parent.Id)
	if err != nil || len(children) == 0 {
		return nil, nil
	}

	var tree []types.DivisionTree
	for _, child := range children {
		node := types.DivisionTree{
			Division: toDivisionType(child),
		}
		subChildren, err := l.buildTreeRecursive(child)
		if err != nil {
			return nil, err
		}
		node.Children = subChildren
		tree = append(tree, node)
	}
	return tree, nil
}

func toDivisionType(d *model.MdAdministrativeDivision) types.Division {
	result := types.Division{
		Id:          d.Id,
		Level:       int32(d.Level),
		Name:        d.Name,
		Code:        d.Code,
		Path:        d.Path,
		SortOrder:   int32(d.SortOrder),
		Status:      int32(d.Status),
		CreatedBy:   d.CreatedBy,
		CreatedTime: d.CreatedTime.Format("2006-01-02 15:04:05"),
		UpdatedTime: d.UpdatedTime.Format("2006-01-02 15:04:05"),
	}
	if d.ParentId.Valid {
		pid := d.ParentId.Int64
		result.ParentId = &pid
	}
	return result
}