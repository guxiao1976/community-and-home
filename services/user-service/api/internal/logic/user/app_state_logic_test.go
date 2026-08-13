package user

import (
	"context"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// stubUserRpc 实现 userv1.UserServiceClient，仅覆盖 GetAppState/SetCurrentCommunity 两个新 RPC。
// 嵌入 nil 接口兜底未覆盖方法（本测试不会调用到）。
type stubUserRpc struct {
	userv1.UserServiceClient
	getAppStateResp  *userv1.GetAppStateResponse
	getAppStateErr   error
	setCommunityResp *userv1.SetCurrentCommunityResponse
	setCommunityErr  error
}

func (s *stubUserRpc) GetAppState(ctx context.Context, in *userv1.GetAppStateRequest, opts ...grpc.CallOption) (*userv1.GetAppStateResponse, error) {
	return s.getAppStateResp, s.getAppStateErr
}

func (s *stubUserRpc) SetCurrentCommunity(ctx context.Context, in *userv1.SetCurrentCommunityRequest, opts ...grpc.CallOption) (*userv1.SetCurrentCommunityResponse, error) {
	return s.setCommunityResp, s.setCommunityErr
}

func TestGetAppStateLogic_Forwards(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", int64(1001))
	stub := &stubUserRpc{
		getAppStateResp: &userv1.GetAppStateResponse{
			Base:               responsex.NewBaseResp(),
			CurrentCommunityId: 2001,
			UpdatedAt:          1700000000,
		},
	}
	l := NewGetAppStateLogic(ctx, &svc.ServiceContext{UserRpc: stub})
	resp, err := l.GetAppState()
	require.NoError(t, err)
	assert.Equal(t, int64(2001), resp.CurrentCommunityId)
	assert.Equal(t, int64(1700000000), resp.UpdatedAt)
}

func TestGetAppStateLogic_NoLogin_Error(t *testing.T) {
	l := NewGetAppStateLogic(context.Background(), &svc.ServiceContext{})
	_, err := l.GetAppState()
	require.Error(t, err)
}

func TestSetCurrentCommunityLogic_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", int64(1002))
	stub := &stubUserRpc{
		setCommunityResp: &userv1.SetCurrentCommunityResponse{Base: responsex.NewBaseResp()},
	}
	l := NewSetCurrentCommunityLogic(ctx, &svc.ServiceContext{UserRpc: stub})
	_, err := l.SetCurrentCommunity(&types.SetCurrentCommunityReq{CommunityId: 2001})
	require.NoError(t, err)
}

func TestSetCurrentCommunityLogic_Surfaces10015(t *testing.T) {
	// RPC 返回 Base=10015 时，逻辑层透出 10015 错误码（非通用内部错误）
	ctx := context.WithValue(context.Background(), "user_id", int64(1003))
	stub := &stubUserRpc{
		setCommunityResp: &userv1.SetCurrentCommunityResponse{
			Base: responsex.NewBaseRespWithError(10015, "目标小区不在数据范围"),
		},
	}
	l := NewSetCurrentCommunityLogic(ctx, &svc.ServiceContext{UserRpc: stub})
	_, err := l.SetCurrentCommunity(&types.SetCurrentCommunityReq{CommunityId: 9999})
	require.Error(t, err)

	ce := errx.FromError(err)
	require.NotNil(t, ce)
	assert.Equal(t, 10015, ce.Code)
}
