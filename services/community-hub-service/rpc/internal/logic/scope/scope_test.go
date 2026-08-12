package scope

import (
	"context"
	"errors"
	"testing"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakePermClient 仅覆盖 AssertPublishScope，其余方法嵌入不调用
type fakePermClient struct {
	permissionv1.PermissionServiceClient
	assertFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
}

func (f *fakePermClient) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	return f.assertFn(ctx, in, opts...)
}

// SEE: [[testing-discipline]] — 统一判据消费方映射：permission 060007 → community-hub 080006
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言失败摘录留档
func TestAssertCommunityScope(t *testing.T) {
	allowedResp := func() *permissionv1.AssertPublishScopeResponse {
		return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}
	}
	deniedResp := func() *permissionv1.AssertPublishScopeResponse {
		return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"), Allowed: false}
	}

	tests := []struct {
		name        string
		respFn      func() *permissionv1.AssertPublishScopeResponse
		callErr     error
		userID      int64
		targetID    int64
		wantErr     bool
		wantErrCode int
		wantSent    bool // 是否断言请求已正确构造
	}{
		{
			name:     "owner@A 发 A✅ 允许",
			respFn:   allowedResp,
			userID:   100,
			targetID: 200,
			wantErr:  false,
			wantSent: true,
		},
		{
			name:        "owner@A 发 B❌ 拒绝映射 080006",
			respFn:      deniedResp,
			userID:      100,
			targetID:    999,
			wantErr:     true,
			wantErrCode: CodePublishScopeDenied,
			wantSent:    true,
		},
		{
			name:     "global 审核员任意✅ 允许",
			respFn:   allowedResp,
			userID:   8,
			targetID: 12345,
			wantErr:  false,
			wantSent: true,
		},
		{
			name:        "registered_user EMPTY 拒绝映射 080006",
			respFn:      deniedResp,
			userID:      9,
			targetID:    200,
			wantErr:     true,
			wantErrCode: CodePublishScopeDenied,
			wantSent:    true,
		},
		{
			name:     "gRPC 传输错误 → 原样传播",
			respFn:   allowedResp,
			callErr:  errors.New("permission rpc unavailable"),
			userID:   100,
			targetID: 200,
			wantErr:  true,
			wantSent: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var sentUserID, sentTargetID int64
			var sentScopeType string
			client := &fakePermClient{
				assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
					sentUserID = in.GetUserId()
					if len(in.GetTargets()) > 0 {
						sentTargetID = in.GetTargets()[0].GetScopeId()
						sentScopeType = in.GetTargets()[0].GetScopeType()
					}
					if tc.callErr != nil {
						return nil, tc.callErr
					}
					return tc.respFn(), nil
				},
			}

			err := AssertCommunityScope(context.Background(), client, tc.userID, tc.targetID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantErrCode != 0 {
					ce, ok := err.(*errx.CodeError)
					require.True(t, ok, "应为 CodeError，实际 %T", err)
					assert.Equal(t, tc.wantErrCode, ce.Code)
				}
			} else {
				require.NoError(t, err)
			}

			if tc.wantSent {
				assert.Equal(t, tc.userID, sentUserID, "AssertPublishScope 必须携带调用方 JWT 身份")
				assert.Equal(t, tc.targetID, sentTargetID, "target 必须是目标小区 community_id")
				assert.Equal(t, "community", sentScopeType)
			}
		})
	}
}
