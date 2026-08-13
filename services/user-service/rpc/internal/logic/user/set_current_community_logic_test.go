package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.3: GetAppState / SetCurrentCommunity
// =============================================================================
func TestSetCurrentCommunity_GLOBAL_Allowed(t *testing.T) {
	svc, permMock := certTestSvc(t)
	permMock.EXPECT().GetDataScopes(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetDataScopesResponse{State: permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL}, nil)

	logic := NewSetCurrentCommunityLogic(context.Background(), svc)
	resp, err := logic.SetCurrentCommunity(&userv1.SetCurrentCommunityRequest{UserId: 9101, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	s, _ := appStateModel(svc).FindOne(context.Background(), 9101)
	require.NotNil(t, s)
	assert.Equal(t, int64(2001), s.CurrentCommunityId)
}

func TestSetCurrentCommunity_EMPTY_Returns10015(t *testing.T) {
	svc, permMock := certTestSvc(t)
	permMock.EXPECT().GetDataScopes(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetDataScopesResponse{State: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY}, nil)

	logic := NewSetCurrentCommunityLogic(context.Background(), svc)
	resp, err := logic.SetCurrentCommunity(&userv1.SetCurrentCommunityRequest{UserId: 9102, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(10015), resp.Base.Code)
}

func TestSetCurrentCommunity_LIMITED_Hit_Allowed(t *testing.T) {
	svc, permMock := certTestSvc(t)
	permMock.EXPECT().GetDataScopes(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetDataScopesResponse{
			State:    permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED,
			ScopeIds: []int64{2001, 2002},
		}, nil)

	logic := NewSetCurrentCommunityLogic(context.Background(), svc)
	resp, err := logic.SetCurrentCommunity(&userv1.SetCurrentCommunityRequest{UserId: 9103, CommunityId: 2002})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	s, _ := appStateModel(svc).FindOne(context.Background(), 9103)
	assert.Equal(t, int64(2002), s.CurrentCommunityId)
}

func TestSetCurrentCommunity_LIMITED_Miss_Returns10015(t *testing.T) {
	svc, permMock := certTestSvc(t)
	permMock.EXPECT().GetDataScopes(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetDataScopesResponse{
			State:    permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED,
			ScopeIds: []int64{2001},
		}, nil)

	logic := NewSetCurrentCommunityLogic(context.Background(), svc)
	resp, err := logic.SetCurrentCommunity(&userv1.SetCurrentCommunityRequest{UserId: 9104, CommunityId: 2003})
	require.NoError(t, err)
	assert.Equal(t, int32(10015), resp.Base.Code)
}

func TestSetCurrentCommunity_GetDataScopesError_Propagated(t *testing.T) {
	svc, permMock := certTestSvc(t)
	permMock.EXPECT().GetDataScopes(gomock.Any(), gomock.Any()).Return(nil, errors.New("permission unavailable"))

	logic := NewSetCurrentCommunityLogic(context.Background(), svc)
	_, err := logic.SetCurrentCommunity(&userv1.SetCurrentCommunityRequest{UserId: 9105, CommunityId: 2001})
	require.Error(t, err)
}
