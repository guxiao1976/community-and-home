package residentialarea

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guxiao/community-and-home/common/errorx"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/svc"
	"github.com/guxiao/community-and-home/services/masterdata/api/internal/types"
	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateResidentialAreaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateResidentialAreaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateResidentialAreaLogic {
	return &CreateResidentialAreaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateResidentialAreaLogic) CreateResidentialArea(req *types.CreateResidentialAreaReq) (resp *types.CreateResidentialAreaResp, err error) {
	// 自动生成编码
	code, err := l.generateCode(req.CountyId)
	if err != nil {
		return nil, errorx.NewDefaultError("生成小区编码失败: " + err.Error())
	}

	// 检查同区县下小区名称唯一性
	existing, err := l.svcCtx.MdResidentialAreaModel.FindByNameAndCountyId(l.ctx, req.Name, req.CountyId)
	if err == nil && existing != nil {
		return nil, errorx.NewDefaultError("该区县下已存在同名小区，请修改名称")
	}

	now := time.Now()

	result, err := l.svcCtx.MdResidentialAreaModel.Insert(l.ctx, &model.MdResidentialArea{
		CountyId:         sql.NullInt64{Int64: req.CountyId, Valid: true},
		StreetId:         toNullInt64(req.StreetId),
		CommunityDivId:   toNullInt64(req.CommunityDivId),
		Code:             sql.NullString{String: code, Valid: true},
		Name:             req.Name,
		Address:          req.Address,
		CommunityType:    int64(req.CommunityType),
		SubmissionType:    sql.NullInt64{Int64: 1, Valid: true},
		SubmissionStatus: 0, // 待提交
		SubmitterId:      0,
		SubmitTime:       sql.NullTime{},
		CreatedTime:      now,
		UpdatedTime:      now,
	})
	if err != nil {
		return nil, errorx.NewDefaultError("创建住宅小区失败: " + err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errorx.NewDefaultError("获取ID失败")
	}

	return &types.CreateResidentialAreaResp{Id: id}, nil
}

func (l *CreateResidentialAreaLogic) generateCode(countyId int64) (string, error) {
	// 查区县 code
	county, err := l.svcCtx.MdAdministrativeDivisionModel.FindOne(l.ctx, countyId)
	if err != nil {
		return "", fmt.Errorf("查询区县信息失败: %w", err)
	}
	countyCode := county.Code

	return l.generateNextCode(countyId, countyCode)
}

func (l *CreateResidentialAreaLogic) generateNextCode(countyId int64, countyCode string) (string, error) {
	prefix := countyCode // 6 digits, we want countyCode + 4-digit seq = 10 digits

	for attempt := 0; attempt < 3; attempt++ {
		maxCode, err := l.svcCtx.MdResidentialAreaModel.GetMaxCodeByCountyId(l.ctx, countyId, prefix)
		if err != nil {
			return "", err
		}

		nextSeq := 1
		if maxCode != "" && len(maxCode) >= 10 {
			// Extract last 4 digits from the 10-digit code (county 6 + seq 4)
			seqStr := maxCode[6:10]
			var seq int
			_, err := fmt.Sscanf(seqStr, "%d", &seq)
			if err == nil && seq > 0 {
				nextSeq = seq + 1
			}
		}

		newCode := fmt.Sprintf("%s%04d", countyCode, nextSeq)

		// Check uniqueness
		existing, err := l.svcCtx.MdResidentialAreaModel.FindByCode(l.ctx, newCode)
		if err != nil || existing == nil {
			return newCode, nil
		}
	}

	return "", fmt.Errorf("生成唯一编码失败，请重试")
}

func toNullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}
