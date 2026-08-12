package permission

import (
	"context"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-permission/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// SEE: [[is-system-no-permission-shortcut]] — resolveUserScope 三态合并（REQ-A）
// 合并优先级：global 支配 → limited 并集(排除 scope_id=0) → empty

func TestResolveUserScope(t *testing.T) {
	community := model.ScopeTypeCommunity

	tests := []struct {
		name      string
		grants    []*model.UserRoleWithInfo
		scopeType string
		wantState permissionv1.DataScopeState
		wantIds   []int64
	}{
		{
			name:      "仅 global → GLOBAL",
			grants:    []*model.UserRoleWithInfo{{RoleId: 8, ScopeType: model.ScopeTypeGlobal, ScopeId: 0, URStatus: 2}},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL,
			wantIds:   nil,
		},
		{
			name: "global + limited → GLOBAL（支配）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 8, ScopeType: model.ScopeTypeGlobal, ScopeId: 0, URStatus: 2},
				{RoleId: 1, ScopeType: community, ScopeId: 100, URStatus: 0},
			},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL,
			wantIds:   nil,
		},
		{
			name: "多 limited → 并集（去重）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 1, ScopeType: community, ScopeId: 100, URStatus: 0},
				{RoleId: 5, ScopeType: community, ScopeId: 200, URStatus: 0},
				{RoleId: 6, ScopeType: community, ScopeId: 100, URStatus: 1}, // 重复 id
			},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED,
			wantIds:   []int64{100, 200},
		},
		{
			name: "仅 empty 行（'' scope, scope_id=0）→ EMPTY",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 9, ScopeType: model.ScopeTypeEmpty, ScopeId: 0, URStatus: 2},
			},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY,
			wantIds:   nil,
		},
		{
			name:      "无 grants → EMPTY",
			grants:    []*model.UserRoleWithInfo{},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY,
			wantIds:   nil,
		},
		{
			name: "status=3,4 排除（FindActiveRolesByUserId 已过滤，防御性断言）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 1, ScopeType: community, ScopeId: 100, URStatus: 3},
				{RoleId: 1, ScopeType: community, ScopeId: 200, URStatus: 4},
			},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY,
			wantIds:   nil,
		},
		{
			name: "其他 scopeType 的行零贡献（scope_id=0 不进并集）",
			grants: []*model.UserRoleWithInfo{
				{RoleId: 9, ScopeType: model.ScopeTypeEmpty, ScopeId: 0, URStatus: 2},
				{RoleId: 1, ScopeType: model.ScopeTypeBuilding, ScopeId: 555, URStatus: 0}, // building 非 community
			},
			scopeType: community,
			wantState: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY,
			wantIds:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRole := new(MockUserRoleModel)
			mockUserRole.On("FindActiveRolesByUserId", mock.Anything, int64(1001)).Return(tt.grants, nil)

			state, ids := resolveUserScope(context.Background(), mockUserRole, 1001, tt.scopeType)

			assert.Equal(t, tt.wantState, state, "state 应为 %v", tt.wantState)
			assert.ElementsMatch(t, tt.wantIds, ids, "scope ids 应为 %v", tt.wantIds)
			mockUserRole.AssertExpectations(t)
		})
	}
}
