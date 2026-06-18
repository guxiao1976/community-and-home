package handler

import (
	"net/http"
	"strings"

	"github.com/guxiao1976/community-auth/api/internal/logic"
	"github.com/guxiao1976/community-auth/api/internal/svc"
	"github.com/guxiao1976/community-auth/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// RegisterHandlers 注册所有 HTTP 路由
func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	// 公开路由（无需 JWT 认证）
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/login",
				Handler: LoginHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/login/sms",
				Handler: LoginSmsHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/register",
				Handler: RegisterHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/sms/send",
				Handler: SmsSendHandler(svcCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/token/refresh",
				Handler: RefreshTokenHandler(svcCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/public-key",
				Handler: PublicKeyHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/auth"),
	)

	// JWT 认证路由
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/logout",
				Handler: LogoutHandler(svcCtx),
			},
		},
		rest.WithPrefix("/api/auth"),
		rest.WithJwt(svcCtx.Config.JwtAuth.AccessSecret),
	)
}

// LoginHandler 密码登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		responsex.Response(w, resp, err)
	}
}

// LoginSmsHandler 短信验证码登录
func LoginSmsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginSmsReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := logic.NewLoginSmsLogic(r.Context(), svcCtx)
		resp, err := l.LoginSms(&req)
		responsex.Response(w, resp, err)
	}
}

// RegisterHandler 注册
func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := logic.NewRegisterLogic(r.Context(), svcCtx)
		resp, err := l.Register(&req)
		responsex.Response(w, resp, err)
	}
}

// SmsSendHandler 发送短信验证码
func SmsSendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SmsSendReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := logic.NewSmsSendLogic(r.Context(), svcCtx)
		err := l.SmsSend(req.Phone)
		responsex.Response(w, nil, err)
	}
}

// RefreshTokenHandler 刷新令牌
func RefreshTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := logic.NewRefreshTokenLogic(r.Context(), svcCtx)
		resp, err := l.RefreshToken(&req)
		responsex.Response(w, resp, err)
	}
}

// LogoutHandler 注销登录（需 JWT 认证）
func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogoutReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}

		// 从 Authorization header 提取 Bearer token
		accessToken := extractBearerToken(r)
		if accessToken == "" {
			responsex.Response(w, nil, nil) // JWT 中间件已验证，理论上不会为空
			return
		}

		l := logic.NewLogoutLogic(r.Context(), svcCtx)
		err := l.Logout(accessToken, &req)
		responsex.Response(w, nil, err)
	}
}

// PublicKeyHandler 获取 RSA 公钥
func PublicKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewPublicKeyLogic(r.Context(), svcCtx)
		resp, err := l.PublicKey()
		responsex.Response(w, resp, err)
	}
}

// extractBearerToken 从 Authorization header 提取 Bearer token
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
