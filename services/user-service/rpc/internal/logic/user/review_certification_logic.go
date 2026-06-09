package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
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
// 优化：owner/tenant 认证通过后创建 residence；回填 user_base 实名信息
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

	// 3. 角色状态即将变更 → 主动失效 Redis 缓存
	invalidateRolesCache(l.ctx, l.svcCtx.Redis, cert.UserId)

	// 4. 更新 certification 状态
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

	// 4. 查找 role 以确定过期策略
	role, roleErr := l.svcCtx.UserMembershipRoleModel.FindOne(l.ctx, cert.RoleId)

	if in.Result == int32(model.CertStatusApproved) {
		// ====== 审核通过 ======
		verifiedAt := sql.NullTime{Time: now, Valid: true}
		var expiresAt sql.NullTime

		if roleErr == nil {
			if in.ExpiresAt != "" {
				expiresAt = parseDate(in.ExpiresAt)
			} else if model.RolesWithExpiry[role.RoleCode] {
				expiresAt = l.defaultExpiry(role.RoleCode)
			}

			err = l.svcCtx.UserMembershipRoleModel.UpdateVerfStatus(l.ctx, cert.RoleId,
				model.RoleVerfStatusApproved, verifiedAt, expiresAt)
			if err != nil {
				l.Errorf("update role verf_status error: %v", err)
				return nil, err
			}
		}

		// 回填 user_base 实名信息（COALESCE：首次回填，已有不覆盖）
		if meta.RealName != "" && meta.IdCardNumber != "" {
			if err := l.svcCtx.UserBaseModel.UpdateRealNameAndIdCard(l.ctx, cert.UserId, meta.RealName, meta.IdCardNumber); err != nil {
				l.Errorf("backfill real_name/id_card error: %v", err)
				// 不阻塞审核流程
			}
		}

		// owner/tenant：认证通过后创建 residence（房屋归属由房产证/租赁合同证明）
		if roleErr == nil && model.RolesRequiringResidence[role.RoleCode] && role.MembershipId.Valid {
			houseId := buildHouseId(meta.Building, meta.Unit, meta.Room)
			existing, _ := l.svcCtx.UserResidenceModel.FindByMembershipAndHouse(l.ctx, role.MembershipId.Int64, houseId)
			if existing == nil {
				residence := &model.UserResidence{
					Id:           snowflake.NextID(),
					MembershipId: role.MembershipId.Int64,
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
	} else {
		// ====== 审核驳回 ======
		if roleErr == nil {
			err = l.svcCtx.UserMembershipRoleModel.UpdateVerfStatusOnly(l.ctx, cert.RoleId, model.RoleVerfStatusRejected)
			if err != nil {
				l.Errorf("update role verf_status error: %v", err)
				return nil, err
			}
		}
	}

	l.Infof("ReviewCertification success, certId=%d, result=%d, roleCode=%s",
		in.CertificationId, in.Result, func() string {
			if roleErr == nil {
				return role.RoleCode
			}
			return "unknown"
		}())
	return &userv1.ReviewCertificationResponse{Base: responsex.NewBaseResp()}, nil
}

// getRoleExpiryHours 从 sysconfig 读取角色过期时长（小时），fallback 到硬编码默认值
func getRoleExpiryHours(ctx context.Context, svcCtx *svc.ServiceContext, roleCode string) int64 {
	defaults := map[string]int64{
		"grid_worker":     8760,
		"community_admin": 17520,
		"committee":       17520,
		"property_admin":  8760,
		"tenant":          8760,
	}
	if svcCtx.SysConfig != nil {
		key := "user.role_expiry_hours." + roleCode
		if v, err := svcCtx.SysConfig.GetInt(ctx, key); err == nil {
			return int64(v)
		}
	}
	return defaults[roleCode]
}

// defaultExpiry 返回角色的默认过期时间
func (l *ReviewCertificationLogic) defaultExpiry(roleCode string) sql.NullTime {
	now := time.Now()
	expiryHours := getRoleExpiryHours(l.ctx, l.svcCtx, roleCode)
	if expiryHours == 0 {
		return sql.NullTime{} // 永久
	}
	return sql.NullTime{Time: now.Add(time.Duration(expiryHours) * time.Hour), Valid: true}
}
