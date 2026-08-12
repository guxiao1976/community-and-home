package lostfound

import (
	"context"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeReadPerm 覆盖读过滤需要的 GetDataScopes（本包此前只有 AssertPublishScope fake）
type fakeReadPerm struct {
	fakePerm
	scopesFn func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error)
}

func (f *fakeReadPerm) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	return f.scopesFn(ctx, in, opts...)
}

func readPerm(state permissionv1.DataScopeState, ids ...int64) *fakeReadPerm {
	return &fakeReadPerm{
		scopesFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
			return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: state, ScopeIds: ids}, nil
		},
	}
}

// 越权 Get 回归（评审 CRITICAL：GetLostFound 未做数据范围校验，LIMITED/EMPTY 用户可按 ID 直接读取越权小区
// 的 description/contact_phone）。
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — 同一读路径（Get 与 List）控制必须一致（REQ-1.6）
func TestGetLostFound_FilterByScope(t *testing.T) {
	tests := []struct {
		name        string
		perm        *fakeReadPerm
		userID      int64 // 0 → 不注入身份（fail-closed 路径）
		itemComID   int64 // 内容所在小区
		wantAllowed bool  // true → 返回内容；false → 080006 + 空
	}{
		{
			name:        "LIMITED 不含目标小区 → 080006 拒绝",
			perm:        readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      42,
			itemComID:   200,
			wantAllowed: false,
		},
		{
			name:        "EMPTY 空范围 → 080006 拒绝",
			perm:        readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY),
			userID:      42,
			itemComID:   100,
			wantAllowed: false,
		},
		{
			name:        "无身份（metadata 缺失）→ fail-closed 拒绝",
			perm:        readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      0,
			itemComID:   100,
			wantAllowed: false,
		},
		{
			name:        "LIMITED 含目标小区 → 返回内容",
			perm:        readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      42,
			itemComID:   100,
			wantAllowed: true,
		},
		{
			name:        "GLOBAL → 返回内容",
			perm:        readPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL),
			userID:      42,
			itemComID:   999,
			wantAllowed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdl := &fakeLostFoundModel{findItem: &model.LostFoundItem{Id: 1, CommunityId: tc.itemComID}}
			sc := &svc.ServiceContext{LostFoundItemModel: mdl, PermissionClient: tc.perm}

			ctx := context.Background()
			if tc.userID != 0 {
				ctx = ctxWithUserID(t, tc.userID)
			}
			l := NewGetLostFoundLogic(ctx, sc)
			resp, err := l.GetLostFound(&communityv1.GetLostFoundRequest{Id: 1})
			require.NoError(t, err)

			if tc.wantAllowed {
				assert.Equal(t, int32(0), resp.GetBase().GetCode())
				require.NotNil(t, resp.GetItem(), "允许读取时返回内容")
			} else {
				assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "越权 Get → 080006")
				assert.Nil(t, resp.GetItem(), "拒绝时不得返回内容")
			}
		})
	}
}
