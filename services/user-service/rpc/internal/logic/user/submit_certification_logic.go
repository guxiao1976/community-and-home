package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// certMetadata 认证材料元数据，存储在 document_urls 字段中
type certMetadata struct {
	Urls         []string `json:"urls"`
	RealName     string   `json:"real_name"`
	IdCardNumber string   `json:"id_card_number"`                 // AES 加密
	RoleCode     string   `json:"role_code,omitempty"`            // 申请的角色编码
	MembershipId int64    `json:"membership_id,string,omitempty"` // 小区成员关系ID（merchant 为 0）— Snowflake 需 string
	CommunityId  int64    `json:"community_id,string,omitempty"`  // 小区ID（merchant 为 0）— Snowflake 需 string
	Building     string   `json:"building,omitempty"`
	Unit         string   `json:"unit,omitempty"`
	Room         string   `json:"room,omitempty"`
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
	// 1. 通过 permission-service 查询用户角色，确认该 role 存在且状态允许提交
	//    in.RoleId 是 rel_user_role 的 role_id
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("SubmitCertification: PermissionClient is nil")
		return &userv1.SubmitCertificationResponse{
			Base: responsex.NewBaseRespWithError(50000, "系统繁忙"),
		}, nil
	}

	userRolesResp, err := l.svcCtx.PermissionClient.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: in.UserId})
	if err != nil {
		l.Errorf("GetUserRoles failed: %v", err)
		return nil, err
	}

	// 找到对应 role，确定 role_code 和 scope
	var roleCode string
	var scopeType, scopeID string
	found := false
	for _, r := range userRolesResp.Roles {
		if r.Role.Id == in.RoleId {
			found = true
			roleCode = r.Role.Code
			scopeType = r.ScopeType
			scopeID = fmt.Sprintf("%d", r.ScopeId)
			// 校验状态：status IN (0,3,4) 允许提交（未认证/已驳回/已过期）
			if r.Status == 1 || r.Status == 2 {
				return &userv1.SubmitCertificationResponse{
					Base: responsex.NewBaseRespWithError(10003, "该角色已提交认证申请，请勿重复提交"),
				}, nil
			}
			break
		}
	}
	if !found {
		return &userv1.SubmitCertificationResponse{
			Base: responsex.NewBaseRespWithError(10007, "角色不存在或不属于该用户"),
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
		RoleCode:     roleCode,
	}
	if scopeType == "community" {
		meta.CommunityId, _ = strconv.ParseInt(scopeID, 10, 64)
		// 查 membership 获取 membership_id（用于认证通过后创建 residence）
		if membership, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, meta.CommunityId); err == nil {
			meta.MembershipId = membership.Id
		}
	}
	// owner/tenant 角色的房屋信息也暂存在这里，认证通过后创建 residence
	if model.RolesRequiringResidence[roleCode] {
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

	// 6. 调 permission-service 更新角色状态 → 待审核(1)
	scopeTypeVal := "community"
	scopeIDVal, _ := strconv.ParseInt(scopeID, 10, 64)
	if scopeType == "global" {
		scopeTypeVal = "global"
		scopeIDVal = 0
	}
	_, err = l.svcCtx.PermissionClient.UpdateUserRoleStatus(l.ctx, &permissionv1.UpdateUserRoleStatusRequest{
		UserId:    in.UserId,
		RoleId:    in.RoleId,
		ScopeType: scopeTypeVal,
		ScopeId:   scopeIDVal,
		Status:    1, // 待审核
	})
	if err != nil {
		l.Errorf("update role status error: %v", err)
		return nil, err
	}

	// 7. Submit real name for moderation (skip if moderation client not configured)
	if in.RealName != "" && l.svcCtx.ModerationClient != nil {
		contentSummary := in.RealName
		if len([]rune(contentSummary)) > 100 {
			contentSummary = string([]rune(contentSummary)[:100])
		}
		auditResp, err := l.svcCtx.ModerationClient.CreateAuditLog(l.ctx, &moderationv1.CreateAuditLogRequest{
			ContentType:    "text",
			ContentSummary: contentSummary,
			RiskLevel:      "low",
			Pass:           false,
			SourceType:     "certification",
			SourceId:       certId,
			UserId:         in.UserId,
		})
		if err != nil {
			l.Errorf("CreateAuditLog for certification failed: %v", err)
		} else {
			items := []map[string]string{
				{"type": "text", "content": in.RealName, "field": "real_name"},
			}
			// Document URLs as image attachments (reserved)
			for _, docUrl := range in.DocumentUrls {
				items = append(items, map[string]string{"type": "image", "content": docUrl, "field": "document"})
			}
			taskMsg := map[string]interface{}{
				"task_id":      fmt.Sprintf("cert_%d_%d", certId, time.Now().UnixNano()),
				"audit_log_id": auditResp.Id,
				"source_type":  "certification",
				"source_id":    certId,
				"action":       "create",
				"items":        items,
			}
			body, _ := json.Marshal(taskMsg)
			if _, err := l.svcCtx.RedisClient.LpushCtx(l.ctx, "moderation:task:queue", string(body)); err != nil {
				l.Errorf("enqueue moderation task failed: %v", err)
			}
		}
	}

	l.Infof("SubmitCertification success, userId=%d, roleId=%d, certId=%d", in.UserId, in.RoleId, certId)
	return &userv1.SubmitCertificationResponse{
		Base:          responsex.NewBaseResp(),
		Certification: toProtoCertification(cert),
	}, nil
}
