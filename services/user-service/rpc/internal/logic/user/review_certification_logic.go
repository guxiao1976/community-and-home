package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCertificationLogic {
	return &ReviewCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReviewCertification 审核认证（per 设计文档 3.5）
// 认证通过：调 permission-service UpdateUserRoleStatus(status=2) 激活角色 + 回填实名 + 创建 residence
// 认证驳回：调 permission-service UpdateUserRoleStatus(status=3)
func (l *ReviewCertificationLogic) ReviewCertification(in *userv1.ReviewCertificationRequest) (*userv1.ReviewCertificationResponse, error) {
	// 1. 查 certification
	cert, err := l.svcCtx.UserCertificationModel.FindOne(l.ctx, in.CertificationId)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.ReviewCertificationResponse{
				Base: responsex.NewBaseRespWithError(10007, "认证申请不存在或状态不允许操作"),
			}, nil
		}
		l.Errorf("find certification error: %v", err)
		return nil, err
	}

	if cert.Status != model.CertStatusPending {
		return &userv1.ReviewCertificationResponse{
			Base: responsex.NewBaseRespWithError(10007, "认证申请不存在或状态不允许操作"),
		}, nil
	}

	// 1.5. 校验审核人存在且未被禁用
	reviewer, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.ReviewerId)
	if err != nil || reviewer.Status != model.UserStatusActive {
		return &userv1.ReviewCertificationResponse{
			Base: responsex.NewBaseRespWithError(10007, "审核人不存在或已被禁用"),
		}, nil
	}

	// 2. 解析认证元数据（SubmitCertification 时存入的 JSON）
	var meta certMetadata
	if cert.DocumentUrls.Valid && cert.DocumentUrls.String != "" {
		json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)
	}

	now := time.Now()

	// 3. 更新 certification 状态
	cert.Status = int64(in.Result)
	cert.ReviewerId = sql.NullInt64{Int64: in.ReviewerId, Valid: true}
	cert.ReviewTime = sql.NullTime{Time: now, Valid: true}
	if in.ReviewNotes != "" {
		cert.ReviewNotes = sql.NullString{String: in.ReviewNotes, Valid: true}
	}
	err = l.svcCtx.UserCertificationModel.Update(l.ctx, cert)
	if err != nil {
		l.Errorf("update certification error: %v", err)
		return nil, err
	}

	// 4. 确定角色 scope（从 certification 元数据中的 communityId）
	//    certification.RoleId 关联 rel_user_role 的 role_id
	scopeType := "community"
	scopeId := meta.CommunityId
	if meta.CommunityId == 0 {
		scopeType = "global"
	}

	// 5. 调 permission-service 更新角色状态
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("ReviewCertification: PermissionClient is nil")
		return &userv1.ReviewCertificationResponse{
			Base: responsex.NewBaseRespWithError(50000, "系统繁忙"),
		}, nil
	}

	updateReq := &permissionv1.UpdateUserRoleStatusRequest{
		UserId:    cert.UserId,
		RoleId:    cert.RoleId,
		ScopeType: scopeType,
		ScopeId:   scopeId,
	}

	if in.Result == int32(model.CertStatusApproved) {
		// ====== 审核通过 ======
		verifiedAt := now.Unix()
		var expiresAt int64
		if in.ExpiresAt != "" {
			if t := parseDate(in.ExpiresAt); t.Valid {
				expiresAt = t.Time.Unix()
			}
		}
		// 默认有效期（从 meta 中的 role_code 推断）
		if expiresAt == 0 && meta.RoleCode != "" && model.RolesWithExpiry[meta.RoleCode] {
			if t := defaultExpiryTime(meta.RoleCode); t.Valid {
				expiresAt = t.Time.Unix()
			}
		}

		updateReq.Status = 2
		updateReq.VerifiedAt = int64Ptr(verifiedAt)
		if expiresAt > 0 {
			updateReq.ExpiresAt = int64Ptr(expiresAt)
		}
	} else {
		// ====== 审核驳回 ======
		updateReq.Status = 3
	}

	if _, err := l.svcCtx.PermissionClient.UpdateUserRoleStatus(l.ctx, updateReq); err != nil {
		l.Errorf("ReviewCertification: UpdateUserRoleStatus failed certId=%d userId=%d err=%v", in.CertificationId, cert.UserId, err)
		return nil, err
	}

	if in.Result == int32(model.CertStatusApproved) {
		// 回填 user_base 实名信息（COALESCE：首次回填，已有不覆盖）
		if meta.RealName != "" && meta.IdCardNumber != "" {
			if err := l.svcCtx.UserBaseModel.UpdateRealNameAndIdCard(l.ctx, cert.UserId, meta.RealName, meta.IdCardNumber); err != nil {
				l.Errorf("backfill real_name/id_card error: %v", err)
				// 不阻塞审核流程
			}
		}

		// owner/tenant：认证通过后创建 residence（房屋归属由房产证/租赁合同证明）
		if model.RolesRequiringResidence[meta.RoleCode] && meta.MembershipId > 0 {
			houseId := buildHouseId(meta.Building, meta.Unit, meta.Room)
			existing, _ := l.svcCtx.UserResidenceModel.FindByMembershipAndHouse(l.ctx, meta.MembershipId, houseId)
			if existing == nil {
				residence := &model.UserResidence{
					Id:           snowflake.NextID(),
					MembershipId: meta.MembershipId,
					UserId:       cert.UserId,
					HouseId:      houseId,
					Building:     meta.Building,
					Unit:         meta.Unit,
					Room:         meta.Room,
					IsPrimary:    1,
					CreatedTime:  now,
					UpdatedTime:  now,
				}
				if _, err := l.svcCtx.UserResidenceModel.Insert(l.ctx, residence); err != nil {
					l.Errorf("create residence on approval error: %v", err)
				}
			}
		}
	}

	l.Infof("ReviewCertification success, certId=%d, result=%d, roleCode=%s",
		in.CertificationId, in.Result, meta.RoleCode)
	return &userv1.ReviewCertificationResponse{Base: responsex.NewBaseResp()}, nil
}
