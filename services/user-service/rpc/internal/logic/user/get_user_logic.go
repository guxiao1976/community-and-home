package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.GetUserResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		l.Errorf("find user error: %v", err)
		return nil, err
	}

	// toProtoUser 已解密手机号为明文；若解密失败保留原值（兜底）
	resp := &userv1.GetUserResponse{
		Base: responsex.NewBaseResp(),
		User: toProtoUser(user),
	}

	viewerId := in.GetViewerId()
	switch {
	case viewerId == 0:
		// 无查看上下文：默认脱敏，不返回房屋号（默认安全）
		resp.User.Phone = maskPhone(resp.User.Phone)
	case viewerId == in.Id:
		// 查看自身：明文 + 自身房屋号
		b, u, r := ownHouseInfo(l.ctx, l.svcCtx, in.Id)
		resp.SameHouse = &userv1.SameHouseInfo{SameHouse: true, Building: b, Unit: u, Room: r}
	default:
		// 他人查看：同屋判定决定明文/脱敏 + 房屋号
		same, b, u, r, herr := isSameHouse(l.ctx, l.svcCtx, viewerId, in.Id)
		if herr != nil {
			// 判定失败兜底：脱敏（默认安全），不返回房屋号
			l.Errorf("isSameHouse error viewerId=%d targetId=%d: %v", viewerId, in.Id, herr)
			resp.User.Phone = maskPhone(resp.User.Phone)
			resp.SameHouse = &userv1.SameHouseInfo{SameHouse: false}
		} else if same {
			resp.SameHouse = &userv1.SameHouseInfo{SameHouse: true, Building: b, Unit: u, Room: r}
		} else {
			resp.User.Phone = maskPhone(resp.User.Phone)
			resp.SameHouse = &userv1.SameHouseInfo{SameHouse: false}
		}
	}

	return resp, nil
}
