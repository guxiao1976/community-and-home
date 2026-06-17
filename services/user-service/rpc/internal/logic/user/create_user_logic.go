package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateUser 创建用户（供 auth-service 注册时调用）
// 保留兼容 auth-service 传入的 user_type, scope_id（忽略）
func (l *CreateUserLogic) CreateUser(in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	// 1. AES 加密手机号
	encryptedPhone, err := crypto.AESEncrypt(in.Phone)
	if err != nil {
		l.Errorf("encrypt phone error: %v", err)
		return nil, err
	}

	// 2. 检查手机号是否已注册
	existing, err := l.svcCtx.UserBaseModel.FindOneByPhone(l.ctx, encryptedPhone)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("check phone exists error: %v", err)
		return nil, err
	}
	if existing != nil {
		return &userv1.CreateUserResponse{
			Base: responsex.NewBaseRespWithError(10002, "手机号已注册"),
		}, nil
	}

	// 3. 生成雪花 ID
	userId := snowflake.NextID()

	// 4. 构建用户数据
	nickname := sql.NullString{Valid: false}
	if in.Nickname != "" {
		nickname = sql.NullString{String: in.Nickname, Valid: true}
	}

	now := time.Now()
	user := &model.UserBase{
		Id:          userId,
		Phone:       encryptedPhone,
		Nickname:    nickname,
		Status:      model.UserStatusActive,
		CreditScore: 100,
		CreatedTime: now,
		UpdatedTime: now,
	}

	// 5. 插入数据库
	_, err = l.svcCtx.UserBaseModel.Insert(l.ctx, user)
	if err != nil {
		l.Errorf("insert user error: %v", err)
		return nil, err
	}

	l.Infof("CreateUser success, userId=%d, phone=%s", userId, in.Phone)

	// 6. Submit nickname to moderation (if provided and moderation client configured)
	if in.Nickname != "" && l.svcCtx.ModerationClient != nil {
		contentSummary := in.Nickname
		if len([]rune(contentSummary)) > 100 {
			contentSummary = string([]rune(contentSummary)[:100])
		}
		auditResp, err := l.svcCtx.ModerationClient.CreateAuditLog(l.ctx, &moderationv1.CreateAuditLogRequest{
			ContentType:    "text",
			ContentSummary: contentSummary,
			RiskLevel:      "low",
			Pass:           false,
			SourceType:     "nickname",
			SourceId:       userId,
			UserId:         userId,
		})
		if err != nil {
			l.Errorf("CreateAuditLog for nickname failed: %v", err)
		} else {
			taskMsg := map[string]interface{}{
				"task_id":      fmt.Sprintf("nickname_%d_%d", userId, time.Now().UnixNano()),
				"audit_log_id": auditResp.Id,
				"source_type":  "nickname",
				"source_id":    userId,
				"action":       "create",
				"items": []map[string]string{
					{"type": "text", "content": in.Nickname, "field": "nickname"},
				},
			}
			body, _ := json.Marshal(taskMsg)
			if _, err := l.svcCtx.RedisClient.LpushCtx(l.ctx, "moderation:task:queue", string(body)); err != nil {
				l.Errorf("enqueue moderation task failed: %v", err)
			}
		}
	}

	return &userv1.CreateUserResponse{
		Base:   responsex.NewBaseResp(),
		UserId: userId,
	}, nil
}
