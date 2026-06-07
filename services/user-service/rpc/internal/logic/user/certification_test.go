package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// §3.3 提交认证材料 (SubmitCertification) 测试
// =============================================================================

func TestSubmitCertification_OwnerSuccess(t *testing.T) {
	// U-S-01: 正常提交业主认证（含房产证 URL）
	svc := testSvc(t)
	rm := roleModel(svc)
	cm := certModel(svc)

	// 先创建未认证的 owner 角色
	r := &model.UserMembershipRole{
		Id: 100, UserId: 7001, MembershipId: sql.NullInt64{Int64: 5001, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusUnverified,
	}
	rm.data[r.Id] = r

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

	// 验证 role verf_status 已更新
	assert.Equal(t, int64(model.RoleVerfStatusPending), r.VerfStatus)

	// 验证 document_urls 存储了完整 JSON
	assert.True(t, cert.DocumentUrls.Valid)
	var meta certMetadata
	json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)
	assert.Equal(t, "张三", meta.RealName)
	assert.NotEqual(t, "110101199001011234", meta.IdCardNumber, "身份证号应 AES 加密")
	assert.Equal(t, "3", meta.Building)
	assert.Equal(t, "2", meta.Unit)
	assert.Equal(t, "1501", meta.Room)
}

func TestSubmitCertification_ResubmitAfterRejection(t *testing.T) {
	// U-S-02: 驳回后重新提交
	svc := testSvc(t)
	rm := roleModel(svc)

	r := &model.UserMembershipRole{
		Id: 101, UserId: 7002, MembershipId: sql.NullInt64{Int64: 5002, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusRejected,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7002, RoleId: 101, DocumentUrls: []string{"http://minio/deed2.jpg"},
		RealName: "李四", IdCardNumber: "110101199002022345",
		Building: "5", Unit: "1", Room: "802",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(model.RoleVerfStatusPending), r.VerfStatus, "重新提交后状态应为待审核")
}

func TestSubmitCertification_ResubmitAfterExpired(t *testing.T) {
	// U-S-03: 过期后重新提交
	svc := testSvc(t)
	rm := roleModel(svc)

	r := &model.UserMembershipRole{
		Id: 102, UserId: 7003, MembershipId: sql.NullInt64{Int64: 5003, Valid: true},
		CommunityId: 2001, RoleCode: "grid_worker", VerfStatus: model.RoleVerfStatusExpired,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7003, RoleId: 102, DocumentUrls: []string{"http://minio/cert.jpg"},
		RealName: "王五", IdCardNumber: "110101199003033456",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(model.RoleVerfStatusPending), r.VerfStatus)
}

func TestSubmitCertification_DuplicatePending(t *testing.T) {
	// U-S-04: 待审核中重复提交
	svc := testSvc(t)
	rm := roleModel(svc)

	r := &model.UserMembershipRole{
		Id: 103, UserId: 7004, MembershipId: sql.NullInt64{Int64: 5004, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7004, RoleId: 103, DocumentUrls: []string{"url"}, RealName: "赵六", IdCardNumber: "110101",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10003), resp.Base.Code, "待审中重复提交应报错")
}

func TestSubmitCertification_DuplicateApproved(t *testing.T) {
	// U-S-05: 已通过重复提交
	svc := testSvc(t)
	rm := roleModel(svc)

	r := &model.UserMembershipRole{
		Id: 104, UserId: 7005, MembershipId: sql.NullInt64{Int64: 5005, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusApproved,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7005, RoleId: 104, DocumentUrls: []string{"url"}, RealName: "钱七", IdCardNumber: "110101",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10003), resp.Base.Code)
}

func TestSubmitCertification_RoleNotFound(t *testing.T) {
	// U-S-06: role 不存在
	svc := testSvc(t)

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7006, RoleId: 99999, DocumentUrls: []string{"url"}, RealName: "孙八", IdCardNumber: "110101",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code)
}

func TestSubmitCertification_IDCardEncrypted(t *testing.T) {
	// U-S-07: 身份证号 AES 加密存储
	svc := testSvc(t)
	rm := roleModel(svc)
	cm := certModel(svc)

	r := &model.UserMembershipRole{
		Id: 105, UserId: 7007, MembershipId: sql.NullInt64{Int64: 5007, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusUnverified,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7007, RoleId: 105, DocumentUrls: []string{"url"},
		RealName: "周九", IdCardNumber: "110101199004044567",
	})

	require.NoError(t, err)
	cert, _ := cm.FindOne(context.Background(), resp.Certification.Id)
	var meta certMetadata
	json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)

	// 验证 idCardNumber 不是明文
	assert.NotEqual(t, "110101199004044567", meta.IdCardNumber)
	assert.NotEmpty(t, meta.IdCardNumber)

	// 验证可以解密
	decrypted, err := crypto.AESDecrypt(meta.IdCardNumber)
	require.NoError(t, err)
	assert.Equal(t, "110101199004044567", decrypted)
}

func TestSubmitCertification_Merchant_NoHouseInfo(t *testing.T) {
	// U-S-08: 商家认证不含房屋信息
	svc := testSvc(t)
	rm := roleModel(svc)
	cm := certModel(svc)

	r := &model.UserMembershipRole{
		Id: 106, UserId: 7008,
		MembershipId: sql.NullInt64{}, // merchant: NULL
		CommunityId:  0,
		RoleCode:     "merchant",
		VerfStatus:   model.RoleVerfStatusUnverified,
	}
	rm.data[r.Id] = r

	logic := NewSubmitCertificationLogic(context.Background(), svc)
	resp, err := logic.SubmitCertification(&userv1.SubmitCertificationRequest{
		UserId: 7008, RoleId: 106, DocumentUrls: []string{"url"},
		RealName: "商家甲", IdCardNumber: "110101199005055678",
	})

	require.NoError(t, err)
	cert, _ := cm.FindOne(context.Background(), resp.Certification.Id)
	var meta certMetadata
	json.Unmarshal([]byte(cert.DocumentUrls.String), &meta)
	assert.Empty(t, meta.Building, "merchant 不应有 building")
	assert.Empty(t, meta.Unit)
	assert.Empty(t, meta.Room)
}

// =============================================================================
// §3.4 审核认证 (ReviewCertification) 测试
// =============================================================================

func TestReviewCertification_Approve_Owner(t *testing.T) {
	// U-R-01: 审核通过 owner → 创建 residence + 回填实名 + 失效缓存
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)
	resm := residenceModel(svc)

	// 创建用户和审核人
	createTestUser(t, ub, 8001, "phone_8001")
	createTestUser(t, ub, 9001, "phone_9001")

	// 创建待审核的 role
	r := &model.UserMembershipRole{
		Id: 200, UserId: 8001, MembershipId: sql.NullInt64{Int64: 5050, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	// 创建待审核的 certification（含房屋信息）
	meta := certMetadata{
		Urls:         []string{"http://minio/deed.jpg"},
		RealName:     "张三",
		IdCardNumber: "encrypted_id_card",
		Building:     "3", Unit: "2", Room: "1501",
	}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{
		Id: 1000, RoleId: 200, UserId: 8001,
		DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true},
		Status:       model.CertStatusPending,
	}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1000, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)

	// 验证 cert 状态更新
	assert.Equal(t, int64(model.CertStatusApproved), cert.Status)
	assert.Equal(t, int64(9001), cert.ReviewerId.Int64)

	// 验证 role 状态更新（已通过）
	assert.Equal(t, int64(model.RoleVerfStatusApproved), r.VerfStatus)
	assert.True(t, r.VerifiedAt.Valid, "verified_at 应设置")
	assert.False(t, r.ExpiresAt.Valid, "owner 永久有效，expires_at 为 NULL")

	// 验证实名信息回填（COALESCE）
	u, _ := ub.FindOne(context.Background(), 8001)
	assert.Equal(t, "张三", u.RealName.String)

	// 验证 residence 创建
	resList, _ := resm.FindByMembershipId(context.Background(), 5050)
	assert.Len(t, resList, 1)
	assert.Equal(t, "3-2-1501", resList[0].HouseId)
	assert.Equal(t, "3", resList[0].Building)
	assert.Equal(t, "1501", resList[0].Room)
}

func TestReviewCertification_Approve_Tenant_WithExpiry(t *testing.T) {
	// U-R-02: 审核通过 tenant（有时限）
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	createTestUser(t, ub, 8002, "phone_8002")
	createTestUser(t, ub, 9001, "phone_9001")
	r := &model.UserMembershipRole{
		Id: 201, UserId: 8002, MembershipId: sql.NullInt64{Int64: 5051, Valid: true},
		CommunityId: 2001, RoleCode: "tenant", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	meta := certMetadata{RealName: "李四", IdCardNumber: "enc", Building: "5", Unit: "1", Room: "201"}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{Id: 1001, RoleId: 201, UserId: 8002, DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true}, Status: model.CertStatusPending}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1001, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
		ExpiresAt: "2027-06-03",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.True(t, r.ExpiresAt.Valid, "tenant 应有过期时间")
}

func TestReviewCertification_Approve_Merchant(t *testing.T) {
	// U-R-05: 审核通过 merchant — 不创建 residence
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)
	resm := residenceModel(svc)

	createTestUser(t, ub, 8005, "phone_8005")
	createTestUser(t, ub, 9001, "phone_9001")
	r := &model.UserMembershipRole{
		Id: 205, UserId: 8005, MembershipId: sql.NullInt64{}, // NULL
		CommunityId: 0, RoleCode: "merchant", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	meta := certMetadata{RealName: "商家甲", IdCardNumber: "enc"}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{Id: 1005, RoleId: 205, UserId: 8005, DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true}, Status: model.CertStatusPending}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1005, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	// 不应创建 residence
	resList, _ := resm.FindByUserId(context.Background(), 8005)
	assert.Empty(t, resList, "merchant 不应创建 residence")
}

func TestReviewCertification_Reject(t *testing.T) {
	// U-R-06: 审核驳回
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	createTestUser(t, ub, 9001, "phone_9001")

	r := &model.UserMembershipRole{
		Id: 206, UserId: 8006, MembershipId: sql.NullInt64{Int64: 5056, Valid: true},
		CommunityId: 2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	cert := &model.UserCertification{Id: 1006, RoleId: 206, UserId: 8006, Status: model.CertStatusPending}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1006, ReviewerId: 9001, Result: int32(model.CertStatusRejected),
		ReviewNotes: "材料不清晰",
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.Equal(t, int64(model.CertStatusRejected), cert.Status)
	assert.Equal(t, int64(model.RoleVerfStatusRejected), r.VerfStatus)
	assert.Equal(t, "材料不清晰", cert.ReviewNotes.String)
}

func TestReviewCertification_CertNotFound(t *testing.T) {
	// U-R-07: certification 不存在
	svc := testSvc(t)

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 99999, ReviewerId: 9001, Result: 2,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code)
}

func TestReviewCertification_AlreadyReviewed(t *testing.T) {
	// U-R-08: certification 状态不是待审核
	svc := testSvc(t)
	cm := certModel(svc)

	cert := &model.UserCertification{Id: 1008, RoleId: 208, UserId: 8008, Status: model.CertStatusApproved}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1008, ReviewerId: 9001, Result: 2,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(10007), resp.Base.Code)
}

func TestReviewCertification_COALESCE_NoOverwrite(t *testing.T) {
	// U-R-10: 已存在的实名信息不被覆盖
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	u := createTestUser(t, ub, 8010, "phone_8010")
	createTestUser(t, ub, 9001, "phone_9001")
	u.RealName = sql.NullString{String: "原名", Valid: true}
	u.IdCardNumber = sql.NullString{String: "old_id_card", Valid: true}

	r := &model.UserMembershipRole{
		Id: 210, UserId: 8010, MembershipId: sql.NullInt64{Int64: 5060, Valid: true},
		CommunityId: 2001, RoleCode: "committee", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	meta := certMetadata{RealName: "新名", IdCardNumber: "new_encrypted"}
	metaJSON, _ := json.Marshal(meta)
	cert := &model.UserCertification{Id: 1010, RoleId: 210, UserId: 8010, DocumentUrls: sql.NullString{String: string(metaJSON), Valid: true}, Status: model.CertStatusPending}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	_, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1010, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
	})
	require.NoError(t, err)

	// 不应覆盖
	u2, _ := ub.FindOne(context.Background(), 8010)
	assert.Equal(t, "原名", u2.RealName.String, "已存在的 real_name 不应覆盖")
	assert.Equal(t, "old_id_card", u2.IdCardNumber.String, "已存在的 id_card 不应覆盖")
}

func TestReviewCertification_GridWorker_DefaultExpiry(t *testing.T) {
	// U-R-03: 网格员默认 1 年过期
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	createTestUser(t, ub, 8003, "phone_8003")
	createTestUser(t, ub, 9001, "phone_9001")
	r := &model.UserMembershipRole{
		Id: 203, UserId: 8003, MembershipId: sql.NullInt64{Int64: 5053, Valid: true},
		CommunityId: 2001, RoleCode: "grid_worker", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	cert := &model.UserCertification{Id: 1003, RoleId: 203, UserId: 8003, Status: model.CertStatusPending}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1003, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Base.Code)
	assert.True(t, r.ExpiresAt.Valid, "grid_worker 应有默认过期时间")
}

func TestReviewCertification_ReviewerNotFound(t *testing.T) {
	// U-R-11: 审核人不存在
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	createTestUser(t, ub, 8009, "phone_8009")
	r := &model.UserMembershipRole{
		Id: 209, UserId: 8009,
		MembershipId: sql.NullInt64{Int64: 5059, Valid: true},
		CommunityId:  2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	cert := &model.UserCertification{
		Id: 1009, RoleId: 209, UserId: 8009,
		Status: model.CertStatusPending,
	}
	cm.data[cert.Id] = cert

	// 不创建 reviewer 9001 → 应报错
	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1009, ReviewerId: 9001, Result: int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.NotEqual(t, int32(0), resp.Base.Code)
}

func TestReviewCertification_ReviewerDisabled(t *testing.T) {
	// U-R-12: 审核人已被禁用
	svc := testSvc(t)
	ub := userBaseModel(svc)
	rm := roleModel(svc)
	cm := certModel(svc)

	createTestUser(t, ub, 8011, "phone_8011")
	// 创建被禁用的审核人
	reviewer := createTestUser(t, ub, 9002, "phone_9002")
	reviewer.Status = model.UserStatusDisabled

	r := &model.UserMembershipRole{
		Id: 211, UserId: 8011,
		MembershipId: sql.NullInt64{Int64: 5061, Valid: true},
		CommunityId:  2001, RoleCode: "owner", VerfStatus: model.RoleVerfStatusPending,
	}
	rm.data[r.Id] = r

	cert := &model.UserCertification{
		Id: 1011, RoleId: 211, UserId: 8011,
		Status: model.CertStatusPending,
	}
	cm.data[cert.Id] = cert

	logic := NewReviewCertificationLogic(context.Background(), svc)
	resp, err := logic.ReviewCertification(&userv1.ReviewCertificationRequest{
		CertificationId: 1011, ReviewerId: 9002, Result: int32(model.CertStatusApproved),
	})

	require.NoError(t, err)
	assert.NotEqual(t, int32(0), resp.Base.Code)
}
