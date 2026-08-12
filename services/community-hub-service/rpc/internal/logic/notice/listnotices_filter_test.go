package notice

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

type fakeListPerm struct {
	fakePerm
	scopesFn func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error)
}

func (f *fakeListPerm) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	return f.scopesFn(ctx, in, opts...)
}

func listPerm(state permissionv1.DataScopeState, ids ...int64) *fakeListPerm {
	return &fakeListPerm{
		scopesFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
			return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: state, ScopeIds: ids}, nil
		},
	}
}

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — 读过滤三态语义（REQ-1.6）
func TestListNotices_FilterByScope(t *testing.T) {
	tests := []struct {
		name         string
		perm         *fakeListPerm
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
			mdl := &fakeNoticeModel{}
			sc := &svc.ServiceContext{NoticeModel: mdl, PermissionClient: tc.perm}

			l := NewListNoticesLogic(noticeCtxWithUserID(t, 42), sc)
			resp, err := l.ListNotices(&communityv1.ListNoticesRequest{
				CommunityId: tc.requestComID,
				Page:        1,
				PageSize:    10,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.wantQuery, mdl.findListCalled)
			if !tc.wantQuery {
				assert.Empty(t, resp.GetNotices(), "被过滤时返回空列表")
				assert.Equal(t, int64(0), resp.GetTotal())
			}
		})
	}
}
