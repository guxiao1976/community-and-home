package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================

func TestReviewCertification_Approved(t *testing.T) {
	// U-R-01: 审核通过 → permission UpdateUserRoleStatus(status=2)
	svc, permMock := certTestSvc(t)
	cm := certModel(svc)
	um := userBaseModel(svc)

	// 创建待审认证记录（关联 owner 角色，community=2001）
	meta := certMetadata{RoleCode: "owner", CommunityId: 2001, MembershipId: 5001,
		RealName: "张三", IdCardNumber: "encrypted", Building: "3", Unit: "2", Room: "1501"}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{
		Id: 1000, UserId: 7001, RoleId: 100,
		DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true},
		Status:       model.CertStatusPending,
	}
	cm.data[cert.Id] = cert
	// 审核人存在
	um.data[9001] = &model.UserBase{Id: 9001, Status: model.UserStatusActive}

	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.UpdateUserRoleStatusRequest, _ ...interface{}) (*permissionv1.UpdateUserRoleStatusResponse, error) {
			assert.Equal(t, int32(2), req.Status)
			assert.Equal(t, int64(2001), req.ScopeId)
			return &permissionv1.UpdateUserRoleStatusResponse{}, nil
		})

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1000,
		ReviewerId:      9001,
		Result:          int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestReviewCertification_Rejected(t *testing.T) {
	// U-R-02: 审核驳回 → permission UpdateUserRoleStatus(status=3)
	svc, permMock := certTestSvc(t)
	cm := certModel(svc)
	um := userBaseModel(svc)

	meta := certMetadata{RoleCode: "owner", CommunityId: 2001}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{
		Id: 1001, UserId: 7001, RoleId: 100,
		DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true},
		Status:       model.CertStatusPending,
	}
	cm.data[cert.Id] = cert
	um.data[9001] = &model.UserBase{Id: 9001, Status: model.UserStatusActive}

	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *permissionv1.UpdateUserRoleStatusRequest, _ ...interface{}) (*permissionv1.UpdateUserRoleStatusResponse, error) {
			assert.Equal(t, int32(3), req.Status)
			return &permissionv1.UpdateUserRoleStatusResponse{}, nil
		})

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1001,
		ReviewerId:      9001,
		Result:          int32(model.CertStatusRejected),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestReviewCertification_AlreadyReviewed(t *testing.T) {
	// U-R-03: 已审核过的认证不能重复审核
	svc, _ := certTestSvc(t)
	cm := certModel(svc)

	cert := &model.UserCertification{
		Id: 1002, UserId: 7001, RoleId: 100,
		Status: model.CertStatusApproved, // 已通过
	}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1002,
		ReviewerId:      9001,
		Result:          int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code) // 状态不允许操作
}
