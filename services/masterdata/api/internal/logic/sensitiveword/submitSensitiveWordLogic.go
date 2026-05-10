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

type SubmitSensitiveWordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Submit sensitive word
func NewSubmitSensitiveWordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitSensitiveWordLogic {
	return &SubmitSensitiveWordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitSensitiveWordLogic) SubmitSensitiveWord(req *types.SubmitReq) (resp *types.SubmitResp, err error) {
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

	if word.SubmissionStatus != 0 && word.SubmissionStatus != 3 {
		return nil, errorx.NewDefaultError("当前状态不允许提交")
	}

	word.SubmissionStatus = 1
	var submitterId int64
	if uid := l.ctx.Value("userId"); uid != nil {
		submitterId = uid.(int64)
	}
	word.SubmitterId = sql.NullInt64{Int64: submitterId, Valid: true}
	word.SubmitTime = sql.NullTime{Time: time.Now(), Valid: true}
	if err := l.svcCtx.MdSensitiveWordModel.Update(l.ctx, word); err != nil {
		return nil, errorx.NewDefaultError("提交失败: " + err.Error())
	}

	subType := int64(1)
	if word.SubmissionType.Int64 > 0 {
		subType = word.SubmissionType.Int64
	}
	_, _ = l.svcCtx.SubmissionRecordModel.Insert(l.ctx, &model.SubmissionRecord{
		EntityType:     "sensitive_word",
		EntityId:       word.Id,
		EntityName:     sql.NullString{String: word.Word, Valid: true},
		SubmissionType: subType,
		SubmitterId:    submitterId,
		SubmitTime:     time.Now(),
	})

	return &types.SubmitResp{Success: true}, nil
}
