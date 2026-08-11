package community

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	usermocks "github.com/guxiao1976/api-proto/gen/go/user/v1/mocks"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCtxWithUserID 构造带 user_id 的 context（模拟 JWT 中间件注入）
func testCtxWithUserID(userID int64) context.Context {
	return context.WithValue(context.Background(), "user_id", json.Number("7001"))
}

func TestSubmitCertificationLogic(t *testing.T) {
	ctrl := gomock.NewController(t)
	userMock := usermocks.NewMockUserServiceClient(ctrl)
	svcCtx := &svc.ServiceContext{UserRpc: userMock}

	// mock SubmitCertification RPC 返回成功
	userMock.EXPECT().SubmitCertification(gomock.Any(), gomock.Any()).Return(
		&userv1.SubmitCertificationResponse{
			Certification: &userv1.Certification{Id: 100, RoleId: 5, UserId: 7001, Status: 1},
		}, nil)

	l := NewSubmitCertificationLogic(testCtxWithUserID(7001), svcCtx)
	resp, err := l.SubmitCertification(&types.SubmitCertificationReq{
		RoleId:       5,
		DocumentUrls: []string{"http://minio/contract.jpg"},
		RealName:     "张三",
		IdCardNumber: "110101199001011234",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(100), resp.Certification.Id)
	assert.Equal(t, int32(1), resp.Certification.Status)
}

func TestSubmitCertificationLogic_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	userMock := usermocks.NewMockUserServiceClient(ctrl)
	svcCtx := &svc.ServiceContext{UserRpc: userMock}

	// 无 user_id 的 context → 报错
	l := NewSubmitCertificationLogic(context.Background(), svcCtx)
	_, err := l.SubmitCertification(&types.SubmitCertificationReq{RoleId: 5})
	assert.Error(t, err)
}

func TestListCertificationsLogic(t *testing.T) {
	ctrl := gomock.NewController(t)
	userMock := usermocks.NewMockUserServiceClient(ctrl)
	svcCtx := &svc.ServiceContext{UserRpc: userMock}

	userMock.EXPECT().ListCertifications(gomock.Any(), gomock.Any()).Return(
		&userv1.ListCertificationsResponse{
			Certifications: []*userv1.Certification{{Id: 100, UserId: 7001, Status: 1}},
			Page:           &commonv1.PageResponse{Total: 1},
		}, nil)

	l := NewListCertificationsLogic(context.Background(), svcCtx)
	resp, err := l.ListCertifications(&types.ListCertificationsReq{Page: 1, PageSize: 10})

	require.NoError(t, err)
	assert.Len(t, resp.Certifications, 1)
	assert.Equal(t, int64(1), resp.Total)
}
