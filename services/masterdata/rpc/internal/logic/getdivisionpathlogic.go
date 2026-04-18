package logic

import (
	"context"
	"strconv"
	"strings"

	"github.com/guxiao/community-and-home/services/masterdata/rpc/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDivisionPathLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDivisionPathLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDivisionPathLogic {
	return &GetDivisionPathLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDivisionPathLogic) GetDivisionPath(in *pb.GetDivisionPathReq) (*pb.GetDivisionPathResp, error) {
	d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, in.Id)
	if err != nil || d.DeleteTime.Valid {
		return &pb.GetDivisionPathResp{}, nil
	}

	// 解析路径字符串如 /1/23/456/ 并查询每个行政区域
	pathDivisions := l.resolvePath(d.Path)
	return &pb.GetDivisionPathResp{Path: pathDivisions}, nil
}

func (l *GetDivisionPathLogic) resolvePath(pathStr string) []*pb.Division {
	var result []*pb.Division
	parts := strings.Split(strings.Trim(pathStr, "/"), "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			continue
		}
		d, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, id)
		if err != nil || d.DeleteTime.Valid {
			continue
		}
		result = append(result, modelDivisionToPb(d))
	}
	return result
}
