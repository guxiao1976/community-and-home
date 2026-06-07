package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type GetUserByPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByPhoneLogic {
	return &GetUserByPhoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserByPhone 根据手机号查询用户（供 auth-service 登录时调用）
func (l *GetUserByPhoneLogic) GetUserByPhone(in *userv1.GetUserByPhoneRequest) (*userv1.GetUserResponse, error) {
	// AES 加密手机号用于查询
	encryptedPhone, err := crypto.AESEncrypt(in.Phone)
	if err != nil {
		l.Errorf("encrypt phone error: %v", err)
		return nil, err
	}

	user, err := l.svcCtx.UserBaseModel.FindOneByPhone(l.ctx, encryptedPhone)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.GetUserResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		l.Errorf("find user by phone error: %v", err)
		return nil, err
	}

	return &userv1.GetUserResponse{
		Base: responsex.NewBaseResp(),
		User: toProtoUser(user),
	}, nil
}
