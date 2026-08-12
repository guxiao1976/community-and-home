package contact

import (
	"context"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeReadPerm 覆盖读过滤需要的 GetDataScopes（contact 包此前只有 AssertPublishScope fake）。
type fakeReadPerm struct {
	fakePerm
	scopesFn func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error)
}

func (f *fakeReadPerm) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	return f.scopesFn(ctx, in, opts...)
}

func listPerm(state permissionv1.DataScopeState, ids ...int64) *fakeReadPerm {
	return &fakeReadPerm{
		scopesFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
			return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: state, ScopeIds: ids}, nil
		},
	}
}

// ListContacts 读过滤（T4.6 / REQ-1.6）：GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表。
// 与 ListNotices（notice 包 listnotices_filter_test.go）同构——多视角评审修复：
// 工作树给 listcontactslogic.go 挂了 FilterAllowed 块，但此前无任何 ListContacts 测试引用。
//
// SEE: [[testing-discipline]] — 修改函数须有测试，同构模式照抄已验证模板
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func TestListContacts_FilterByScope(t *testing.T) {
	tests := []struct {
		name         string
		perm         *fakeReadPerm
		requestComID int64
		wantQuery    bool // 是否应真正查询（allowed）
	}{
		{
			name:         "空范围 → 空列表（不查库）",
			perm:         listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY),
			requestComID: 100,
			wantQuery:    false,
		},
		{
			name:         "LIMITED + 目标小区不在范围 → 空列表（不查库）",
			perm:         listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			requestComID: 200,
			wantQuery:    false,
		},
		{
			name:         "LIMITED + 目标小区在范围 → 查询",
			perm:         listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			requestComID: 100,
			wantQuery:    true,
		},
		{
			name:         "global 跨小区可见 → 查询",
			perm:         listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL),
			requestComID: 999,
			wantQuery:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdl := &fakeContactModel{}
			sc := &svc.ServiceContext{CommunityContactModel: mdl, PermissionClient: tc.perm}

			l := NewListContactsLogic(contactCtxWithUserID(t, 42), sc)
			resp, err := l.ListContacts(&communityv1.ListContactsRequest{
				CommunityId: tc.requestComID,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantQuery, mdl.findByCalled)
			if !tc.wantQuery {
				assert.Empty(t, resp.GetContacts(), "被过滤时返回空列表")
			}
		})
	}
}
