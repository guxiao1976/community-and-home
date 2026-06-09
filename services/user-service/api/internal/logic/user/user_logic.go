package user

import (
	"context"
	"encoding/json"
	"fmt"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// ==================== List Users ====================

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListUsersLogic) ListUsers(req *types.ListUsersReq) (*types.ListUsersResp, error) {
	rpcReq := &userv1.ListUsersRequest{
		Page:   &commonv1.PageRequest{Page: req.Page, PageSize: req.PageSize},
		Status: req.Status,
	}
	if req.Keyword != "" {
		rpcReq.Keyword = &req.Keyword
	}

	resp, err := l.svcCtx.UserRpc.ListUsers(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	users := make([]types.UserInfo, 0, len(resp.Users))
	for _, u := range resp.Users {
		users = append(users, toUserInfo(u))
	}

	return &types.ListUsersResp{
		Users: users,
		Page: types.PageInfo{
			Page:       resp.Page.Page,
			PageSize:   resp.Page.PageSize,
			Total:      resp.Page.Total,
			TotalPages: resp.Page.TotalPages,
		},
	}, nil
}

// ==================== Create User ====================

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (*types.CreateUserResp, error) {
	resp, err := l.svcCtx.UserRpc.CreateUser(l.ctx, &userv1.CreateUserRequest{
		Phone:    req.Phone,
		Nickname: req.Nickname,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateUserResp{UserId: resp.UserId}, nil
}

// ==================== Get User ====================

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetUserLogic) GetUser(req *types.GetUserReq) (*types.GetUserResp, error) {
	resp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userv1.GetUserRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.GetUserResp{User: toUserInfo(resp.User)}, nil
}

// ==================== Update User ====================

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpdateUserLogic) UpdateUser(req *types.UpdateUserReq) (*types.UpdateUserResp, error) {
	rpcReq := &userv1.UpdateUserRequest{
		Id:        req.Id,
		Nickname:  req.Nickname,
		AvatarUrl: req.AvatarUrl,
		Status:    req.Status,
		Gender:    req.Gender,
		BirthDate: req.BirthDate,
	}

	resp, err := l.svcCtx.UserRpc.UpdateUser(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.UpdateUserResp{User: toUserInfo(resp.User)}, nil
}

// ==================== Delete User ====================

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *DeleteUserLogic) DeleteUser(req *types.DeleteUserReq) (*types.DeleteUserResp, error) {
	status := int32(2) // disabled
	_, err := l.svcCtx.UserRpc.UpdateUser(l.ctx, &userv1.UpdateUserRequest{
		Id:     req.Id,
		Status: &status,
	})
	if err != nil {
		return nil, err
	}
	return &types.DeleteUserResp{Success: true}, nil
}

// ==================== Get Profile (current user from JWT) ====================

type GetProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetProfileLogic) GetProfile() (*types.GetUserResp, error) {
	userId := getUserIdFromJwt(l.ctx)
	if userId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}

	resp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userv1.GetUserRequest{Id: userId})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	return &types.GetUserResp{User: toUserInfo(resp.User)}, nil
}

// ==================== Helpers ====================

// getUserIdFromJwt extracts the user ID from JWT claims in context.
func getUserIdFromJwt(ctx context.Context) int64 {
	v := ctx.Value("user_id")
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		id, _ := n.Int64()
		return id
	default:
		return 0
	}
}

func toUserInfo(u *userv1.User) types.UserInfo {
	if u == nil {
		return types.UserInfo{}
	}
	return types.UserInfo{
		Id:           u.Id,
		Phone:        u.Phone,
		Nickname:     u.Nickname,
		AvatarUrl:    u.AvatarUrl,
		RealName:     u.RealName,
		IdCardNumber: u.IdCardNumber,
		Gender:       u.Gender,
		BirthDate:    u.BirthDate,
		Status:       u.Status,
		CreditScore:  u.CreditScore,
		Preferences:  u.Preferences,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
