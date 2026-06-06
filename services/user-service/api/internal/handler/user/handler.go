package user

import (
	"net/http"

	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/api/internal/logic/user"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListUsersReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewListUsersLogic(r.Context(), svcCtx)
		resp, err := l.ListUsers(&req)
		responsex.Response(w, resp, err)
	}
}

func CreateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewCreateUserLogic(r.Context(), svcCtx)
		resp, err := l.CreateUser(&req)
		responsex.Response(w, resp, err)
	}
}

func GetUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUserReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewGetUserLogic(r.Context(), svcCtx)
		resp, err := l.GetUser(&req)
		responsex.Response(w, resp, err)
	}
}

func UpdateUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewUpdateUserLogic(r.Context(), svcCtx)
		resp, err := l.UpdateUser(&req)
		responsex.Response(w, resp, err)
	}
}

func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetProfile()
		responsex.Response(w, resp, err)
	}
}

func DeleteUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteUserReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.Response(w, nil, err)
			return
		}
		l := user.NewDeleteUserLogic(r.Context(), svcCtx)
		resp, err := l.DeleteUser(&req)
		responsex.Response(w, resp, err)
	}
}
