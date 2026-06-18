package server

import (
	"context"

	authv1 "github.com/guxiao1976/api-proto/gen/go/auth/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/logic/auth"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
)

// AuthServiceServer 认证中心 gRPC Server
type AuthServiceServer struct {
	svcCtx *svc.ServiceContext
	authv1.UnimplementedAuthServiceServer
}

// NewAuthServiceServer 创建 AuthServiceServer
func NewAuthServiceServer(svcCtx *svc.ServiceContext) *AuthServiceServer {
	return &AuthServiceServer{svcCtx: svcCtx}
}

// Login 账密登录（spec/auth.md 核心逻辑流 1.登录签发）
func (s *AuthServiceServer) Login(ctx context.Context, in *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	l := auth.NewLoginLogic(ctx, s.svcCtx)
	return l.Login(in)
}

// LoginSms 短信验证码登录
func (s *AuthServiceServer) LoginSms(ctx context.Context, in *authv1.LoginSmsRequest) (*authv1.LoginResponse, error) {
	l := auth.NewLoginSmsLogic(ctx, s.svcCtx)
	return l.LoginSms(in)
}

// Register 注册（调用 User Service CreateUser → 写入 auth_credential → 签发 Token）
func (s *AuthServiceServer) Register(ctx context.Context, in *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	l := auth.NewRegisterLogic(ctx, s.svcCtx)
	return l.Register(in)
}

// RefreshToken AT 过期无感刷新（spec/auth.md 核心逻辑流 2.AT 过期无感刷新）
func (s *AuthServiceServer) RefreshToken(ctx context.Context, in *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	l := auth.NewRefreshTokenLogic(ctx, s.svcCtx)
	return l.RefreshToken(in)
}

// Logout 主动注销（spec/auth.md 核心逻辑流 3.主动注销）
func (s *AuthServiceServer) Logout(ctx context.Context, in *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	l := auth.NewLogoutLogic(ctx, s.svcCtx)
	return l.Logout(in)
}

// ValidateToken 验证 AT（spec/auth.md 网关协同逻辑）
func (s *AuthServiceServer) ValidateToken(ctx context.Context, in *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	l := auth.NewValidateTokenLogic(ctx, s.svcCtx)
	return l.ValidateToken(in)
}
