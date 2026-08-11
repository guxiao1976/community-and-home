package user

import (
	"context"
	"database/sql"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateUser 更新用户资料（保留兼容 auth-service saga 补偿：status=3 → soft delete）
func (l *UpdateUserLogic) UpdateUser(in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	// 先查用户是否存在
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return &userv1.UpdateUserResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		l.Errorf("find user error: %v", err)
		return nil, err
	}

	// auth-service saga 补偿：status=3 → 软删除（设置  deleted_at）
	if in.Status != nil && *in.Status == 3 {
		err := l.svcCtx.UserBaseModel.SoftDelete(l.ctx, in.Id)
		if err != nil {
			l.Errorf("soft delete user error: %v", err)
			return nil, err
		}
		// 软删除也失效权限缓存 + 写禁用标记
		invalidateUserPermissionCache(l.ctx, l.svcCtx, l.Logger, in.Id, 2)
		return &userv1.UpdateUserResponse{
			Base: responsex.NewBaseResp(),
		}, nil
	}

	// 更新各字段（只更新传入的非 nil 字段）
	if in.Nickname != nil {
		user.Nickname = sql.NullString{String: *in.Nickname, Valid: *in.Nickname != ""}
	}
	if in.AvatarUrl != nil {
		user.AvatarUrl = sql.NullString{String: *in.AvatarUrl, Valid: *in.AvatarUrl != ""}
	}
	if in.Status != nil {
		user.Status = int64(*in.Status)
	}
	if in.Gender != nil {
		user.Gender = sql.NullInt64{Int64: int64(*in.Gender), Valid: true}
	}
	if in.BirthDate != nil && *in.BirthDate != "" {
		user.BirthDate = parseDate(*in.BirthDate)
	}
	if in.Preferences != nil {
		user.Preferences = sql.NullString{String: *in.Preferences, Valid: *in.Preferences != ""}
	}

	err = l.svcCtx.UserBaseModel.Update(l.ctx, user)
	if err != nil {
		l.Errorf("update user error: %v", err)
		return nil, err
	}

	// 状态变更（禁用/启用/软删）时失效权限缓存 + 写/删禁用标记
	if in.Status != nil {
		invalidateUserPermissionCache(l.ctx, l.svcCtx, l.Logger, in.Id, int64(*in.Status))
	}

	// 重新查询返回最新数据
	updated, _ := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.Id)
	return &userv1.UpdateUserResponse{
		Base: responsex.NewBaseResp(),
		User: toProtoUser(updated),
	}, nil
}
