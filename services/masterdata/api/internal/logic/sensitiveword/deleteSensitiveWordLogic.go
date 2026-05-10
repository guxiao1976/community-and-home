// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package sensitiveword

import (
	"context"
	"database/sql"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Delete sensitive word
func NewDeleteSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSensitiveWordLogic {
	return &DeleteSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSensitiveWordLogic) DeleteSensitiveWord(req *types.DeleteSensitiveWordReq) (resp *types.DeleteSensitiveWordResp, err error) {
	word, err := l.svcCtx.MdSensitiveWordModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.NewNotFoundError("敏感词不存在")
		}
		return nil, errorx.NewDefaultError("查询敏感词失败")
	}
	if word.DeleteTime.Valid {
		return nil, errorx.NewDefaultError("敏感词已删除")
	}

	word.SubmissionStatus = 4
	word.SubmissionType = sql.NullInt64{Int64: 3, Valid: true}
	if err := l.svcCtx.MdSensitiveWordModel.Update(l.ctx, word); err != nil {
		return nil, errorx.NewDefaultError("删除敏感词失败: " + err.Error())
	}

	return &types.DeleteSensitiveWordResp{Success: true}, nil
}
