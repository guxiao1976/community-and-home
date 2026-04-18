package middleware

import (
	"net/http"
	"strconv"

	"github.com/guxiao/community-and-home/common/responsex"
)

type PermissionMiddleware struct {
}

func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

func (m *PermissionMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user info from context (set by JWT middleware)
		userId := r.Context().Value("userId")
		if userId == nil {
			responsex.HttpError(r, w, responsex.NewCodeError(http.StatusUnauthorized, "未授权"))
			return
		}

		// Extract scope from context
		scopeId := r.Context().Value("scopeId")
		domain := "global"
		if scopeId != nil {
			if sid, ok := scopeId.(int64); ok && sid > 0 {
				domain = strconv.FormatInt(sid, 10)
			}
		}

		// Get resource and action from request
		resource := r.URL.Path
		action := r.Method

		// TODO: Check permission using Casbin enforcer
		// For now, allow all authenticated requests
		// enforcer.Enforce(userId, domain, resource, action)

		next(w, r)
	}
}
