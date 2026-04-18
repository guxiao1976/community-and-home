package sensitiveword

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSensitiveWordLogic {
	return &CreateSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSensitiveWordLogic) CreateSensitiveWord(req *types.CreateSensitiveWordReq) (resp *types.CreateSensitiveWordResp, err error) {
	// Check if word already exists
	existing, err := l.svcCtx.MdSensitiveWordModel.FindOneByWord(l.ctx, req.Word)
	if err == nil && existing != nil {
		return nil, errorx.NewDefaultError("sensitive word already exists")
	}

	word := &model.MdSensitiveWord{
		Word:     req.Word,
		Category: req.Category,
		Severity: req.Severity,
		Action:   req.Action,
		Status:   1,
		CreatedBy: 0, // TODO: Get from context
	}

	result, err := l.svcCtx.MdSensitiveWordModel.Insert(l.ctx, word)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to create sensitive word")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errorx.NewDefaultError("failed to get insert id")
	}

	return &types.CreateSensitiveWordResp{
		Id: id,
	}, nil
}
