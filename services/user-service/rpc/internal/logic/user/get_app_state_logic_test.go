package user

import (
	"context"
	"errors"
	"testing"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.3: GetAppState / SetCurrentCommunity
// =============================================================================

func TestGetAppState_NoRecord_ReturnsZero(t *testing.T) {
	svc := testSvc(t)
	logic := NewGetAppStateLogic(context.Background(), svc)
	resp, err := logic.GetAppState(&userv1.GetAppStateRequest{UserId: 9001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(0), resp.CurrentCommunityId)
	assert.Equal(t, int64(0), resp.UpdatedAt)
}

func TestGetAppState_WithRecord_ReturnsIdAndUpdatedAt(t *testing.T) {
	svc := testSvc(t)
	am := appStateModel(svc)
	now := time.Now()
	am.data[9002] = &model.UserAppState{UserId: 9002, CurrentCommunityId: 2001, UpdatedTime: now}

	logic := NewGetAppStateLogic(context.Background(), svc)
	resp, err := logic.GetAppState(&userv1.GetAppStateRequest{UserId: 9002})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(2001), resp.CurrentCommunityId)
	assert.Equal(t, now.Unix(), resp.UpdatedAt)
}

func TestGetAppState_DBError_Propagated(t *testing.T) {
	svc := testSvc(t)
	am := appStateModel(svc)
	am.findErr = errors.New("db down") // 注入非 ErrNotFound 的 DB 错误

	logic := NewGetAppStateLogic(context.Background(), svc)
	_, err := logic.GetAppState(&userv1.GetAppStateRequest{UserId: 9003})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

