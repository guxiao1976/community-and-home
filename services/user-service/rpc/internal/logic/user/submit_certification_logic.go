package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

// certMetadata 认证材料元数据，存储在 document_urls 字段中
type certMetadata struct {
	Urls           []string `json:"urls"`
	RealName       string   `json:"real_name"`
	IdCardNumber   string   `json:"id_card_number"` // AES 加密
	Building       string   `json:"building,omitempty"`
	Unit           string   `json:"unit,omitempty"`
	Room           string   `json:"room,omitempty"`
}

type SubmitCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCertificationLogic {
	return &SubmitCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SubmitCertification 提交认证材料（per 设计文档 3.4）
func (l *SubmitCertificationLogic) SubmitCertification(in *userv1.SubmitCertificationRequest) (*userv1.SubmitCertificationResponse, error) {
	// 1. 查 role
	role, err := l.svcCtx.UserMembershipRoleModel.FindOne(l.ctx, in.RoleId)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.SubmitCertificationResponse{
				Base: responsex.NewBaseRespWithError(10007, "认证申请不存在或状态不允许操作"),
			}, nil
		}
		l.Errorf("find role error: %v", err)
		return nil, err
	}

	// 2. 校验 role 状态：verf_status IN (0,3,4) 允许提交
	if role.VerfStatus == model.RoleVerfStatusPending || role.VerfStatus == model.RoleVerfStatusApproved {
		return &userv1.SubmitCertificationResponse{
			Base: responsex.NewBaseRespWithError(10003, "该角色已提交认证申请，请勿重复提交"),
		}, nil
	}

	// 3. AES 加密身份证号
	encryptedIdCard, err := crypto.AESEncrypt(in.IdCardNumber)
	if err != nil {
		l.Errorf("encrypt id_card error: %v", err)
		return nil, err
	}

	// 4. 构建认证元数据（存储在 document_urls 字段中，供审核时回填使用）
	meta := certMetadata{
		Urls:         in.DocumentUrls,
		RealName:     in.RealName,
		IdCardNumber: encryptedIdCard,
	}
	// owner/tenant 角色的房屋信息也暂存在这里，认证通过后创建 residence
	if model.RolesRequiringResidence[role.RoleCode] {
		meta.Building = in.Building
		meta.Unit = in.Unit
		meta.Room = in.Room
	}
	metaJSON, _ := json.Marshal(meta)

	// 5. 插入 certification + 更新 role verf_status
	now := time.Now()
	certId := snowflake.NextID()
	cert := &model.UserCertification{
		Id:           certId,
		RoleId:       in.RoleId,
		UserId:       in.UserId,
		DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true},
		Status:       model.CertStatusPending,
		SubmitTime:   now,
	}

	_, err = l.svcCtx.UserCertificationModel.Insert(l.ctx, cert)
	if err != nil {
		l.Errorf("insert certification error: %v", err)
		return nil, err
	}

	// 6. 更新 role verf_status → 1（待审核）
	err = l.svcCtx.UserMembershipRoleModel.UpdateVerfStatusOnly(l.ctx, in.RoleId, model.RoleVerfStatusPending)
	if err != nil {
		l.Errorf("update role verf_status error: %v", err)
		return nil, err
	}

	l.Infof("SubmitCertification success, userId=%d, roleId=%d, certId=%d", in.UserId, in.RoleId, certId)
	return &userv1.SubmitCertificationResponse{
		Base:          responsex.NewBaseResp(),
		Certification: toProtoCertification(cert),
	}, nil
}
