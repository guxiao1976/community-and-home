package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type PermissionMiddleware struct {
	UserRoleModel model.AuthUserRoleModel
	RoleModel     model.AuthRoleModel
}

func NewPermissionMiddleware(userRoleModel model.AuthUserRoleModel, roleModel model.AuthRoleModel) *PermissionMiddleware {
	return &PermissionMiddleware{
		UserRoleModel: userRoleModel,
		RoleModel:     roleModel,
	}
}

type errorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorBody{Code: code, Message: msg})
}

func (m *PermissionMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIdVal := r.Context().Value("userId")
		if userIdVal == nil {
			writeError(w, http.StatusUnauthorized, "未授权")
			return
		}

		userId, ok := userIdVal.(int64)
		if !ok {
			writeError(w, http.StatusUnauthorized, "未授权")
			return
		}

		// Check if user has any system role (super admin bypass)
		userRoles, err := m.UserRoleModel.FindActiveByUserId(r.Context(), userId)
		if err != nil {
			logx.Errorf("Failed to get user roles for permission check: %v", err)
			writeError(w, http.StatusInternalServerError, "内部错误")
			return
		}

		for _, ur := range userRoles {
			if ur.IsSystem == 1 {
				next(w, r)
				return
			}
		}

		// Determine domain from scopeId
		scopeIdVal := r.Context().Value("scopeId")
		domain := "global"
		if scopeIdVal != nil {
			if sid, ok := scopeIdVal.(int64); ok && sid > 0 {
				domain = strconv.FormatInt(sid, 10)
			}
		}

		// Check Casbin policies for each user role
		resource := r.URL.Path
		action := r.Method

		allowed := false
		for _, ur := range userRoles {
			sub := fmt.Sprintf("role:%d", ur.RoleId)
			if m.checkCasbinPolicy(r, sub, domain, resource, action) {
				allowed = true
				break
			}
		}

		if !allowed {
			writeError(w, http.StatusForbidden, "无权限访问")
			return
		}

		next(w, r)
	}
}

// checkCasbinPolicy is a stub that will be integrated with the Casbin enforcer.
// For now it allows all requests to avoid breaking existing functionality during development.
// TODO: Replace with actual enforcer.Enforce(sub, dom, obj, act) call
func (m *PermissionMiddleware) checkCasbinPolicy(r *http.Request, sub, dom, obj, act string) bool {
	return true
}
