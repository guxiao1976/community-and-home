package sensitiveword

import (
	"context"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSensitiveWordLogic {
	return &UpdateSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSensitiveWordLogic) UpdateSensitiveWord(req *types.UpdateSensitiveWordReq) (resp *types.UpdateSensitiveWordResp, err error) {
	word, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, errorx.NewDefaultError("sensitive word not found")
	}

	if req.Category != "" {
		word.Category = req.Category
	}
	if req.Severity > 0 {
		word.Severity = req.Severity
	}
	if req.Action > 0 {
		word.Action = req.Action
	}
	if req.Status > 0 {
		word.Status = req.Status
	}

	err = l.svcCtx.MdSensitiveWordModel.Update(l.ctx, word)
	if err != nil {
		return nil, errorx.NewDefaultError("failed to update sensitive word")
	}

	return &types.UpdateSensitiveWordResp{
		Success: true,
	}, nil
}
