package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guxiao/community-and-home/services/identity/model"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockMiddlewareUserRoleModel struct {
	findActiveByUserIdFn func(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error)
}

func (m *mockMiddlewareUserRoleModel) FindActiveByUserId(ctx context.Context, userId int64) ([]*model.UserRoleWithInfo, error) {
	return m.findActiveByUserIdFn(ctx, userId)
}
func (m *mockMiddlewareUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*model.AuthUserRole, error) {
	return nil, nil
}
func (m *mockMiddlewareUserRoleModel) FindOne(ctx context.Context, id int64) (*model.AuthUserRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockMiddlewareUserRoleModel) FindOneByUserIdRoleId(ctx context.Context, userId, roleId int64) (*model.AuthUserRole, error) {
	return nil, model.ErrNotFound
}
func (m *mockMiddlewareUserRoleModel) Insert(ctx context.Context, data *model.AuthUserRole) (sql.Result, error) {
	return nil, nil
}
func (m *mockMiddlewareUserRoleModel) Update(ctx context.Context, data *model.AuthUserRole) error { return nil }
func (m *mockMiddlewareUserRoleModel) Delete(ctx context.Context, id int64) error                   { return nil }
func (m *mockMiddlewareUserRoleModel) BatchInsertUserRoles(ctx context.Context, userId int64, roleIds []int64) error {
	return nil
}
func (m *mockMiddlewareUserRoleModel) DeleteByUserIdAndRoleId(ctx context.Context, userId, roleId int64) error {
	return nil
}

func doRequest(m *PermissionMiddleware, userId int64, scopeId int64, method, path string) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(method, path, nil)
	ctx := req.Context()
	if userId > 0 {
		ctx = context.WithValue(ctx, "userId", userId)
	}
	if scopeId > 0 {
		ctx = context.WithValue(ctx, "scopeId", scopeId)
	}
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	m.Handle(next).ServeHTTP(rec, req)
	return rec
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	assert.NoError(t, err)
	return body
}

func TestPermissionMiddleware_NoUserId(t *testing.T) {
	m := NewPermissionMiddleware(nil, nil)
	rec := doRequest(m, 0, 0, "GET", "/api/users")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	body := decodeError(t, rec)
	assert.Equal(t, float64(401), body["code"])
}

func TestPermissionMiddleware_InvalidUserIdType(t *testing.T) {
	m := NewPermissionMiddleware(nil, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/api/users", nil)
	ctx := context.WithValue(req.Context(), "userId", "not-an-int64")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	m.Handle(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPermissionMiddleware_SystemRoleBypass(t *testing.T) {
	ur := &mockMiddlewareUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 1, IsSystem: 1}}, nil
		},
	}
	m := NewPermissionMiddleware(ur, nil)
	rec := doRequest(m, 1, 0, "GET", "/api/users")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestPermissionMiddleware_NonSystemRole_CasbinStubAllowsAll(t *testing.T) {
	ur := &mockMiddlewareUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 2, IsSystem: 0}}, nil
		},
	}
	m := NewPermissionMiddleware(ur, nil)
	rec := doRequest(m, 1, 0, "GET", "/api/users")

	// Current stub allows all, so non-system roles still pass
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPermissionMiddleware_NoRoles_CasbinStubAllowsAll(t *testing.T) {
	ur := &mockMiddlewareUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, nil
		},
	}
	m := NewPermissionMiddleware(ur, nil)
	rec := doRequest(m, 1, 0, "GET", "/api/users")

	// No roles, loop doesn't execute, allowed remains false -> 403
	// But actually, the stub returns true when there ARE roles.
	// With no roles, the loop body never runs, so allowed stays false.
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestPermissionMiddleware_DBError(t *testing.T) {
	ur := &mockMiddlewareUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return nil, sql.ErrConnDone
		},
	}
	m := NewPermissionMiddleware(ur, nil)
	rec := doRequest(m, 1, 0, "GET", "/api/users")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	body := decodeError(t, rec)
	assert.Equal(t, float64(500), body["code"])
}

func TestPermissionMiddleware_ScopeIdAsDomain(t *testing.T) {
	ur := &mockMiddlewareUserRoleModel{
		findActiveByUserIdFn: func(_ context.Context, _ int64) ([]*model.UserRoleWithInfo, error) {
			return []*model.UserRoleWithInfo{{RoleId: 2, IsSystem: 0}}, nil
		},
	}
	m := NewPermissionMiddleware(ur, nil)
	// With a scopeId, the domain should be derived from it
	// Since the stub allows all, this should still pass
	rec := doRequest(m, 1, 100, "POST", "/api/communities")

	assert.Equal(t, http.StatusOK, rec.Code)
}
