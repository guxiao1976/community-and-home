package logic

import (
	"context"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDivisionTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionTreeLogic {
	return &GetDivisionTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDivisionTreeLogic) GetDivisionTree(in *pb.GetDivisionTreeReq) (*pb.GetDivisionTreeResp, error) {
	tree, err := l.buildTree(in.ParentId)
	if err != nil {
		return nil, err
	}
	return &pb.GetDivisionTreeResp{Tree: tree}, nil
}

func (l *GetDivisionTreeLogic) buildTree(parentId int64) ([]*pb.DivisionTree, error) {
	children, err := l.svcCtx.MdAdministrativeDivisionModel.FindChildren(l.ctx, parentId)
	if err != nil || len(children) == 0 {
		return nil, nil
	}

	var tree []*pb.DivisionTree
	for _, child := range children {
		node := &pb.DivisionTree{
			Division: modelDivisionToPb(child),
		}
		subTree, err := l.buildTree(child.Id)
		if err != nil {
			return nil, err
		}
		node.Children = subTree
		tree = append(tree, node)
	}
	return tree, nil
}