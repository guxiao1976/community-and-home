package scope

import (
	"context"
	"errors"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type fakeScopeClient struct {
	permissionv1.PermissionServiceClient
	scopesFn func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error)
}

func (f *fakeScopeClient) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	return f.scopesFn(ctx, in, opts...)
}

// SEE: [[testing-discipline]] — GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表（逻辑层实现，不拼空 IN）
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func TestFilterAllowed(t *testing.T) {
	limitedResp := func(ids ...int64) *permissionv1.GetDataScopesResponse {
		return &permissionv1.GetDataScopesResponse{
			Base:     responsex.NewBaseResp(),
			State:    permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED,
			ScopeIds: ids,
		}
	}

	tests := []struct {
		name        string
		resp        *permissionv1.GetDataScopesResponse
		callErr     error
		userID      int64
		communityID int64
		want        bool
		wantErr     bool
	}{
		{
			name:        "业主只见所属小区：LIMITED + 命中 → 放行",
			resp:        limitedResp(100, 101),
			userID:      42,
			communityID: 101,
			want:        true,
		},
		{
			name:        "业主只见所属小区：LIMITED + 未命中 → 过滤",
			resp:        limitedResp(100),
			userID:      42,
			communityID: 200,
			want:        false,
		},
		{
			name: "空范围空列表：EMPTY → 过滤",
			resp: &permissionv1.GetDataScopesResponse{
				Base:  responsex.NewBaseResp(),
				State: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY,
			},
			userID:      42,
			communityID: 100,
			want:        false,
		},
		{
			name: "global 跨小区可见：GLOBAL → 不过滤",
			resp: &permissionv1.GetDataScopesResponse{
				Base:  responsex.NewBaseResp(),
				State: permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL,
			},
			userID:      42,
			communityID: 999,
			want:        true,
		},
		{
			name:        "GetDataScopes 传输错误 → 传播",
			resp:        limitedResp(100),
			callErr:     errors.New("permission rpc unavailable"),
			userID:      42,
			communityID: 100,
			want:        false,
			wantErr:     true,
		},
		{
			name: "userID=0（系统身份/无身份）→ 恒过滤，不查 GetDataScopes",
			resp: &permissionv1.GetDataScopesResponse{
				Base:  responsex.NewBaseResp(),
				State: permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL, // 即使权限返回 GLOBAL 也必须拒绝
			},
			userID:      0,
			communityID: 100,
			want:        false,
		},
		{
			name: "LIMITED 且范围 id 均大于请求目标 → 未命中过滤（不等才判负）",
			resp: limitedResp(200),
			userID:      42,
			communityID: 100,
			want:        false,
		},
		{
			name: "负数 userID（非法身份）→ 照常查 GetDataScopes，不因 <= 短路拒绝",
			resp: limitedResp(100),
			userID:      -5,
			communityID: 100,
			want:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uid := tc.userID
			if uid == 0 {
				uid = 42
			}
			client := &fakeScopeClient{
				scopesFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
					assert.Equal(t, uid, in.GetUserId(), "GetDataScopes 必须携带调用方 JWT 身份")
					assert.Equal(t, "community", in.GetScopeType())
					if tc.callErr != nil {
						return nil, tc.callErr
					}
					return tc.resp, nil
				},
			}

			got, err := FilterAllowed(context.Background(), client, tc.userID, tc.communityID)
			if tc.wantErr {
				require.Error(t, err)
				assert.False(t, got, "传输错误路径必须返回 false（fail-closed 语义）")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
