package user

import (
	"context"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UpdateUserModerationStatus 测试 (审核回调 RPC)
// SEE: [[change-verification-checklist]] — 每次变更后必须逐项验证
// =============================================================================

func TestUpdateModerationStatus_NicknameSuccess(t *testing.T) {
	// U-MS-01: 审核回调更新昵称审核状态
	svc := testSvc(t)
	ub := userBaseModel(svc)

	// 创建测试用户
	user := createTestUser(t, ub, 10001, "encrypted_phone_10001")
	user.NicknameModerationStatus = 0 // 初始未审核

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               10001,
		Target:           "nickname",
		ModerationStatus: 2, // machine_fail
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	// 验证状态已更新
	updated, _ := ub.FindOne(context.Background(), 10001)
	assert.Equal(t, int64(2), updated.NicknameModerationStatus, "nickname moderation status should be 2")
}

func TestUpdateModerationStatus_CertificationSuccess(t *testing.T) {
	// U-MS-02: 审核回调更新认证记录审核状态
	svc := testSvc(t)
	cm := certModel(svc)

	// 创建测试认证记录
	cert := &model.UserCertification{
		Id:               20001,
		RoleId:           100,
		UserId:           10001,
		ModerationStatus: 0,
	}
	cm.data[cert.Id] = cert

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               20001,
		Target:           "certification",
		ModerationStatus: 1, // machine_pass
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	// 验证状态已更新
	updated, _ := cm.FindOne(context.Background(), 20001)
	assert.Equal(t, int64(1), updated.ModerationStatus, "certification moderation status should be 1")
}

func TestUpdateModerationStatus_InvalidTarget(t *testing.T) {
	// U-MS-03: 无效 target 返回错误
	svc := testSvc(t)

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               99999,
		Target:           "invalid_target",
		ModerationStatus: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10005), resp.Base.Code, "invalid target should return error code 10005")
}

func TestUpdateModerationStatus_Nickname_UserNotFound(t *testing.T) {
	// U-MS-04: 更新昵称审核状态时用户不存在
	svc := testSvc(t)

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	_, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               99999,
		Target:           "nickname",
		ModerationStatus: 1,
	})

	// 用户不存在时 mock UpdateNicknameModerationStatus 返回 error
	assert.Error(t, err, "user not found should produce an error")
}

func TestUpdateModerationStatus_Certification_CertNotFound(t *testing.T) {
	// U-MS-05: 更新认证审核状态时记录不存在
	svc := testSvc(t)

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	_, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               99999,
		Target:           "certification",
		ModerationStatus: 1,
	})

	// 认证记录不存在时 mock UpdateModerationStatus 返回 ErrNotFound
	assert.Error(t, err, "certification not found should produce an error")
}

func TestUpdateModerationStatus_MachinePassStatus(t *testing.T) {
	// U-MS-06: machine_pass 状态 = 1
	svc := testSvc(t)
	cm := certModel(svc)

	cert := &model.UserCertification{
		Id:               20002,
		ModerationStatus: 0,
	}
	cm.data[cert.Id] = cert

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               20002,
		Target:           "certification",
		ModerationStatus: 1, // machine_pass
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	updated, _ := cm.FindOne(context.Background(), 20002)
	assert.Equal(t, int64(1), updated.ModerationStatus)
}

func TestUpdateModerationStatus_HumanFailStatus(t *testing.T) {
	// U-MS-07: human_fail 状态 = 4
	svc := testSvc(t)
	cm := certModel(svc)

	cert := &model.UserCertification{
		Id:               20003,
		ModerationStatus: 0,
	}
	cm.data[cert.Id] = cert

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               20003,
		Target:           "certification",
		ModerationStatus: 4, // human_fail
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	updated, _ := cm.FindOne(context.Background(), 20003)
	assert.Equal(t, int64(4), updated.ModerationStatus)
}

func TestUpdateModerationStatus_UpdateExistingNickname(t *testing.T) {
	// U-MS-08: 更新已有审核状态的昵称（覆盖原状态）
	svc := testSvc(t)
	ub := userBaseModel(svc)

	user := createTestUser(t, ub, 10008, "phone_10008")
	user.NicknameModerationStatus = 1 // 之前 machine_pass

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               10008,
		Target:           "nickname",
		ModerationStatus: 3, // 改为 human_pass
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	updated, _ := ub.FindOne(context.Background(), 10008)
	assert.Equal(t, int64(3), updated.NicknameModerationStatus, "status should be overwritten to 3")
}

func TestUpdateModerationStatus_UpdateExistingCertification(t *testing.T) {
	// U-MS-09: 更新已有审核状态的认证记录（覆盖原状态）
	svc := testSvc(t)
	cm := certModel(svc)

	cert := &model.UserCertification{
		Id:               20004,
		ModerationStatus: 2, // 之前 machine_fail
	}
	cm.data[cert.Id] = cert

	logic := NewUpdateUserModerationStatusLogic(context.Background(), svc)
	resp, err := logic.UpdateUserModerationStatus(&userv1.UpdateModerationStatusRequest{
		Id:               20004,
		Target:           "certification",
		ModerationStatus: 4, // 改为 human_fail
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	updated, _ := cm.FindOne(context.Background(), 20004)
	assert.Equal(t, int64(4), updated.ModerationStatus)
}
