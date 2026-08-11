package user

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permissionmocks "github.com/guxiao1976/api-proto/gen/go/permission/v1/mocks"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func certTestSvc(t *testing.T) (*svc.ServiceContext, *permissionmocks.MockPermissionServiceClient) {
	t.Helper()
	resetRoleMapper() // 重置全局角色映射缓存，避免跨测试污染
	ctrl := gomock.NewController(t)
	permMock := permissionmocks.NewMockPermissionServiceClient(ctrl)
	svc := testSvc(t)
	svc.PermissionClient = permMock
	return svc, permMock
}

// mockUserRoleResponse 构造 permission-service GetUserRoles 响应
func mockUserRoleResponse(roleID int64, roleCode string, status int32, communityID int64) []*permissionv1.UserRoleInfo {
	return []*permissionv1.UserRoleInfo{
		{
			Role: &permissionv1.Role{
				Id:   roleID,
				Code: roleCode,
			},
			ScopeType: "community",
			ScopeId:   communityID,
			Status:    status,
		},
	}
}

func TestSubmitCertification_OwnerSuccess(t *testing.T) {
	// U-S-01: 正常提交业主认证（含房产证 URL）
	svc, permMock := certTestSvc(t)
	cm := certModel(svc)

	// permission GetUserRoles 返回未认证的 owner 角色（status=0，允许提交）
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 0, 2001)}, nil)
	// 提交时更新状态为待审(status=1)
	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).Return(
		&permissionv1.UpdateUserRoleStatusResponse{}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
		Building:     "3", Unit: "2", Room: "1501",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.NotZero(t, resp.Certification.Id)

	// 验证 cert 已创建
	cert, _ := cm.FindOne(context.Background(), resp.Certification.Id)
	assert.Equal(t, int64(model.CertStatusPending), cert.Status)

	// 验证 document_urls 存储了完整 JSON（含 role_code/community_id）
	assert.True(t, cert.DocumentUrls.Valid)
	var meta certMetadata
	json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)
	assert.Equal(t, "张三", meta.RealName)
	assert.NotEqual(t, "110101199001011234", meta.IdCardNumber, "身份证号应 AES 加密")
	assert.Equal(t, "owner", meta.RoleCode)
	assert.Equal(t, "3", meta.Building)
}

func TestSubmitCertification_DuplicatePending(t *testing.T) {
	// U-S-02: 已提交待审（status=1）不能重复提交
	svc, permMock := certTestSvc(t)

	// 返回待审状态的角色（status=1），应拒绝提交
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 1, 2001)}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10003), resp.Base.Code) // 重复提交
}

func TestSubmitCertification_DuplicateApproved(t *testing.T) {
	// U-S-03: 已认证（status=2）不能再次提交
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 2, 2001)}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10003), resp.Base.Code) // 重复提交
}

func TestSubmitCertification_RoleNotFound(t *testing.T) {
	// U-S-04: 角色不存在（permission 返回空列表）应报错
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: []*permissionv1.UserRoleInfo{}}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 999,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code) // 角色不存在
}

func TestSubmitCertification_ResubmitAfterRejection(t *testing.T) {
	// U-S-05: 已驳回（status=3）允许重新提交
	svc, permMock := certTestSvc(t)
	cm := certModel(svc)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 3, 2001)}, nil)
	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).Return(
		&permissionv1.UpdateUserRoleStatusResponse{}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed2.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	cert, _ := cm.FindOne(context.Background(), resp.Certification.Id)
	assert.Equal(t, int64(model.CertStatusPending), cert.Status)
}

func TestSubmitCertification_ResubmitAfterExpired(t *testing.T) {
	// U-S-06: 已过期（status=4）允许重新提交
	svc, permMock := certTestSvc(t)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 4, 2001)}, nil)
	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).Return(
		&permissionv1.UpdateUserRoleStatusResponse{}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

func TestSubmitCertification_IDCardEncrypted(t *testing.T) {
	// U-S-07: 身份证号存储时已 AES 加密
	svc, permMock := certTestSvc(t)
	cm := certModel(svc)

	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(100, "owner", 0, 2001)}, nil)
	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).Return(
		&permissionv1.UpdateUserRoleStatusResponse{}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, _ := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 100,
		DocumentUrls: []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	cert, _ := cm.FindOne(context.Background(), resp.Certification.Id)
	var meta certMetadata
	json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)
	assert.NotEqual(t, "110101199001011234", meta.IdCardNumber)
	assert.NotEmpty(t, meta.IdCardNumber)
}

// TestSubmitCertification_Merchant_NoHouseInfo
// 商家（merchant）无需房屋信息，scope=global
func TestSubmitCertification_Merchant_NoHouseInfo(t *testing.T) {
	svc, permMock := certTestSvc(t)

	// merchant 角色：scope=global, community_id=0
	permMock.EXPECT().GetUserRoles(gomock.Any(), gomock.Any()).Return(
		&permissionv1.GetUserRolesResponse{Roles: mockUserRoleResponse(7, "merchant", 0, 0)}, nil)
	permMock.EXPECT().UpdateUserRoleStatus(gomock.Any(), gomock.Any()).Return(
		&permissionv1.UpdateUserRoleStatusResponse{}, nil)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7001, RoleId: 7,
		DocumentUrls: []string{"http://minio/license.jpg"},
		RealName:     "李四",
		IdCardNumber: "110101199001012345",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
}

// =============================================================================
// §3.4 审核认证 (ReviewCertification) 测试
