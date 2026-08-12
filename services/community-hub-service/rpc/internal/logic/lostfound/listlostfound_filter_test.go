package lostfound

import (
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ListLostFound 读过滤（T4.6 / REQ-1.6）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表。
// 与 ListNotices（notice 包 listnotices_filter_test.go）同构——多视角评审修复：
// 工作树给 listlostfoundlogic.go 挂了 FilterAllowed 块，但此前无任何 ListLostFound 测试引用。
//
// SEE: [[testing-discipline]] — 修改函数须有测试，同构模式照抄已验证模板
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func TestListLostFound_FilterByScope(t *testing.T) {
	tests := []struct {
		name         string
		perm         *fakeReadPerm
		requestComID int64
		wantQuery    bool // 是否应真正查询（allowed）
	}{
		{
			name:         "空范围 → 空列表（不查库）",
			perm:         readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY),
			requestComID: 100,
			wantQuery:    false,
		},
		{
			name:         "LIMITED + 目标小区不在范围 → 空列表（不查库）",
			perm:         readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			requestComID: 200,
			wantQuery:    false,
		},
		{
			name:         "LIMITED + 目标小区在范围 → 查询",
			perm:         readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			requestComID: 100,
			wantQuery:    true,
		},
		{
			name:         "global 跨小区可见 → 查询",
			perm:         readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL),
			requestComID: 999,
			wantQuery:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdl := &fakeLostFoundModel{}
			sc := &svc.ServiceContext{LostFoundItemModel: mdl, PermissionClient: tc.perm}

			l := NewListLostFoundLogic(ctxWithUserID(t, 42), sc)
			resp, err := l.ListLostFound(&communityv1.ListLostFoundRequest{
				CommunityId: tc.requestComID,
				Page:        1,
				PageSize:    10,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantQuery, mdl.findListCalled)
			if !tc.wantQuery {
				assert.Empty(t, resp.GetItems(), "被过滤时返回空列表")
				assert.Equal(t, int64(0), resp.GetTotal())
			}
		})
	}
}
