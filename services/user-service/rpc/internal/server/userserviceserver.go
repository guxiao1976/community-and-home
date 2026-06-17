package server

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/logic/user"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
)

// UserServiceServer 用户中心 gRPC Server，实现 proto 定义的所有 RPC 方法
type UserServiceServer struct {
	svcCtx *svc.ServiceContext
	userv1.UnimplementedUserServiceServer
}

// NewUserServiceServer 创建 UserServiceServer
func NewUserServiceServer(svcCtx *svc.ServiceContext) *UserServiceServer {
	return &UserServiceServer{
		svcCtx: svcCtx,
	}
}

// ==================== 用户基础操作 ====================

func (s *UserServiceServer) CreateUser(ctx context.Context, in *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	return user.NewCreateUserLogic(ctx, s.svcCtx).CreateUser(in)
}

func (s *UserServiceServer) GetUser(ctx context.Context, in *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	return user.NewGetUserLogic(ctx, s.svcCtx).GetUser(in)
}

func (s *UserServiceServer) GetUserByPhone(ctx context.Context, in *userv1.GetUserByPhoneRequest) (*userv1.GetUserResponse, error) {
	return user.NewGetUserByPhoneLogic(ctx, s.svcCtx).GetUserByPhone(in)
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, in *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	return user.NewUpdateUserLogic(ctx, s.svcCtx).UpdateUser(in)
}

func (s *UserServiceServer) ListUsers(ctx context.Context, in *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	return user.NewListUsersLogic(ctx, s.svcCtx).ListUsers(in)
}

func (s *UserServiceServer) GetUsersByIds(ctx context.Context, in *userv1.GetUsersByIdsRequest) (*userv1.GetUsersByIdsResponse, error) {
	return user.NewGetUsersByIdsLogic(ctx, s.svcCtx).GetUsersByIds(in)
}

// ==================== 小区成员 ====================

func (s *UserServiceServer) JoinCommunity(ctx context.Context, in *userv1.JoinCommunityRequest) (*userv1.JoinCommunityResponse, error) {
	return user.NewJoinCommunityLogic(ctx, s.svcCtx).JoinCommunity(in)
}

func (s *UserServiceServer) LeaveCommunity(ctx context.Context, in *userv1.LeaveCommunityRequest) (*userv1.LeaveCommunityResponse, error) {
	return user.NewLeaveCommunityLogic(ctx, s.svcCtx).LeaveCommunity(in)
}

func (s *UserServiceServer) GetUserMemberships(ctx context.Context, in *userv1.GetUserMembershipsRequest) (*userv1.GetUserMembershipsResponse, error) {
	return user.NewGetUserMembershipsLogic(ctx, s.svcCtx).GetUserMemberships(in)
}

// ==================== 角色 ====================

func (s *UserServiceServer) ApplyRole(ctx context.Context, in *userv1.ApplyRoleRequest) (*userv1.ApplyRoleResponse, error) {
	return user.NewApplyRoleLogic(ctx, s.svcCtx).ApplyRole(in)
}

func (s *UserServiceServer) GetUserRoles(ctx context.Context, in *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	return user.NewGetUserRolesLogic(ctx, s.svcCtx).GetUserRoles(in)
}

func (s *UserServiceServer) CheckAccess(ctx context.Context, in *userv1.CheckAccessRequest) (*userv1.CheckAccessResponse, error) {
	return user.NewCheckAccessLogic(ctx, s.svcCtx).CheckAccess(in)
}

// ==================== 认证（统一流程） ====================

func (s *UserServiceServer) SubmitCertification(ctx context.Context, in *userv1.SubmitCertificationRequest) (*userv1.SubmitCertificationResponse, error) {
	return user.NewSubmitCertificationLogic(ctx, s.svcCtx).SubmitCertification(in)
}

func (s *UserServiceServer) ReviewCertification(ctx context.Context, in *userv1.ReviewCertificationRequest) (*userv1.ReviewCertificationResponse, error) {
	return user.NewReviewCertificationLogic(ctx, s.svcCtx).ReviewCertification(in)
}

func (s *UserServiceServer) ListCertifications(ctx context.Context, in *userv1.ListCertificationsRequest) (*userv1.ListCertificationsResponse, error) {
	return user.NewListCertificationsLogic(ctx, s.svcCtx).ListCertifications(in)
}

func (s *UserServiceServer) GetMyCertifications(ctx context.Context, in *userv1.GetMyCertificationsRequest) (*userv1.GetMyCertificationsResponse, error) {
	return user.NewGetMyCertificationsLogic(ctx, s.svcCtx).GetMyCertifications(in)
}

// ==================== 房屋 ====================

func (s *UserServiceServer) BindResidence(ctx context.Context, in *userv1.BindResidenceRequest) (*userv1.BindResidenceResponse, error) {
	return user.NewBindResidenceLogic(ctx, s.svcCtx).BindResidence(in)
}

func (s *UserServiceServer) GetResidences(ctx context.Context, in *userv1.GetResidencesRequest) (*userv1.GetResidencesResponse, error) {
	return user.NewGetResidencesLogic(ctx, s.svcCtx).GetResidences(in)
}

func (s *UserServiceServer) UpdateUserModerationStatus(ctx context.Context, in *userv1.UpdateModerationStatusRequest) (*userv1.UpdateModerationStatusResponse, error) {
	return user.NewUpdateUserModerationStatusLogic(ctx, s.svcCtx).UpdateUserModerationStatus(in)
}
