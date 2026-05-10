// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"context"
	"database/sql"
	"time"

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

// Create new sensitive word
func NewCreateSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSensitiveWordLogic {
	return &CreateSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSensitiveWordLogic) CreateSensitiveWord(req *types.CreateSensitiveWordReq) (resp *types.CreateSensitiveWordResp, err error) {
	// 1. Validate word doesn't already exist
	existing, err := l.svcCtx.MdSensitiveWordModel.FindOneByWord(l.ctx, req.Word)
	if err == nil && existing != nil {
		return nil, errorx.NewInvalidParamError("敏感词已存在")
	}

	// 2. Validate severity (1-3)
	if req.Severity < 1 || req.Severity > 3 {
		return nil, errorx.NewInvalidParamError("严重程度必须在1-3之间")
	}

	// 3. Validate action (1-3)
	if req.Action < 1 || req.Action > 3 {
		return nil, errorx.NewInvalidParamError("处理动作必须在1-3之间")
	}

	// 4. Create model
	data := &model.MdSensitiveWord{
		Word:             req.Word,
		Category:         req.Category,
		Severity:         int64(req.Severity),
		Action:           int64(req.Action),
		Status:           1,
		SubmissionType:    sql.NullInt64{Int64: 1, Valid: true},
		SubmissionStatus: 0,
		CreatedBy:        0, // TODO: Get from JWT context
		CreatedTime:      time.Now(),
		UpdatedTime:      time.Now(),
	}

	res, err := l.svcCtx.MdSensitiveWordModel.Insert(l.ctx, data)
	if err != nil {
		return nil, errorx.NewDefaultError("创建敏感词失败")
	}

	id, _ := res.LastInsertId()
	return &types.CreateSensitiveWordResp{Id: id}, nil
}
