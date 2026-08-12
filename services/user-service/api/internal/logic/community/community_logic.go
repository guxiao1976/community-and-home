package community

import (
	"context"
	"encoding/json"
	"fmt"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

// ==================== Join Community ====================

type JoinCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinCommunityLogic {
	return &JoinCommunityLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *JoinCommunityLogic) JoinCommunity(req *types.JoinCommunityReq) (*types.JoinCommunityResp, error) {
	userId := getUserId(l.ctx)
	resp, err := l.svcCtx.UserRpc.JoinCommunity(l.ctx, &userv1.JoinCommunityRequest{
		UserId:      userId,
		CommunityId: req.CommunityId,
		Building:    req.Building,
		Unit:        req.Unit,
		Room:        req.Room,
		Ownership:   userv1.CommunityOwnership(req.Ownership),
	})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}

	return &types.JoinCommunityResp{
		Membership: toMembership(resp.Membership),
	}, nil
}

// ==================== Get Memberships ====================

type GetMembershipsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMembershipsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMembershipsLogic {
	return &GetMembershipsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetMembershipsLogic) GetMemberships() (*types.GetMembershipsResp, error) {
	userId := getUserId(l.ctx)
	resp, err := l.svcCtx.UserRpc.GetUserMemberships(l.ctx, &userv1.GetUserMembershipsRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	memberships := make([]types.CommunityMembership, 0, len(resp.Memberships))
	for _, m := range resp.Memberships {
		memberships = append(memberships, toMembership(m))
	}

	return &types.GetMembershipsResp{
		Memberships: memberships,
	}, nil
}

// ==================== Leave Community ====================

type LeaveCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveCommunityLogic {
	return &LeaveCommunityLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *LeaveCommunityLogic) LeaveCommunity(req *types.LeaveCommunityReq) (*types.LeaveCommunityResp, error) {
	userId := getUserId(l.ctx)
	resp, err := l.svcCtx.UserRpc.LeaveCommunity(l.ctx, &userv1.LeaveCommunityRequest{
		UserId:      userId,
		CommunityId: req.CommunityId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}

	return &types.LeaveCommunityResp{}, nil
}

// ==================== Helper ====================

// getUserId extracts the user ID from JWT claims stored in context.
// go-zero JWT middleware stores claims by their original key names;
// JWT field "user_id" becomes ctx.Value("user_id"), and JSON numbers are float64.
func getUserId(ctx context.Context) int64 {
	v := ctx.Value("user_id")
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		id, _ := n.Int64()
		return id
	default:
		return 0
	}
}

// ==================== Bind Residence ====================

type BindResidenceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindResidenceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindResidenceLogic {
	return &BindResidenceLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *BindResidenceLogic) BindResidence(req *types.BindResidenceReq) (*types.BindResidenceResp, error) {
	resp, err := l.svcCtx.UserRpc.BindResidence(l.ctx, &userv1.BindResidenceRequest{
		MembershipId: req.MembershipId,
		Building:     req.Building,
		Unit:         req.Unit,
		Room:         req.Room,
		IsPrimary:    req.IsPrimary,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	return &types.BindResidenceResp{Residence: toResidence(resp.Residence)}, nil
}

func toResidence(r *userv1.Residence) types.Residence {
	if r == nil {
		return types.Residence{}
	}
	return types.Residence{
		Id:           r.Id,
		MembershipId: r.MembershipId,
		HouseId:      r.HouseId,
		Building:     r.Building,
		Unit:         r.Unit,
		Room:         r.Room,
		IsPrimary:    r.IsPrimary,
	}
}

// ==================== Apply Role ====================

type ApplyRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyRoleLogic {
	return &ApplyRoleLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ApplyRoleLogic) ApplyRole(req *types.ApplyRoleReq) (*types.ApplyRoleResp, error) {
	userId := getUserId(l.ctx)
	resp, err := l.svcCtx.UserRpc.ApplyRole(l.ctx, &userv1.ApplyRoleRequest{
		UserId:      userId,
		CommunityId: req.CommunityId,
		RoleCode:    req.RoleCode,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	return &types.ApplyRoleResp{}, nil
}

// ==================== Get User Roles ====================

type GetUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRolesLogic {
	return &GetUserRolesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetUserRolesLogic) GetUserRoles() (*types.GetUserRolesResp, error) {
	userId := getUserId(l.ctx)
	resp, err := l.svcCtx.UserRpc.GetUserRoles(l.ctx, &userv1.GetUserRolesRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	roles := make([]types.RoleInfo, 0, len(resp.Roles))
	for _, r := range resp.Roles {
		roles = append(roles, types.RoleInfo{
			Id:          r.Id,
			UserId:      r.UserId,
			CommunityId: r.CommunityId,
			RoleCode:    r.RoleCode,
			VerfStatus:  r.VerfStatus,
		})
	}
	return &types.GetUserRolesResp{Roles: roles}, nil
}

func toMembership(m *userv1.CommunityMembership) types.CommunityMembership {
	if m == nil {
		return types.CommunityMembership{}
	}
	return types.CommunityMembership{
		Id:          m.Id,
		UserId:      m.UserId,
		CommunityId: m.CommunityId,
		BindStatus:  m.BindStatus,
		JoinTime:    m.JoinTime,
		LeaveTime:   m.LeaveTime,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Building:    int(m.Building),
		Unit:        int(m.Unit),
		Room:        int(m.Room),
	}
}

// =============================================================================
// 认证（Certification）
// =============================================================================

type SubmitCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitCertificationLogic {
	return &SubmitCertificationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SubmitCertificationLogic) SubmitCertification(req *types.SubmitCertificationReq) (*types.SubmitCertificationResp, error) {
	userId := getUserIdFromJwt(l.ctx)
	if userId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}
	resp, err := l.svcCtx.UserRpc.SubmitCertification(l.ctx, &userv1.SubmitCertificationRequest{
		UserId:       userId,
		RoleId:       req.RoleId,
		DocumentUrls: req.DocumentUrls,
		RealName:     req.RealName,
		IdCardNumber: req.IdCardNumber,
		Building:     req.Building,
		Unit:         req.Unit,
		Room:         req.Room,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base != nil && resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	return &types.SubmitCertificationResp{
		Certification: toCertificationInfo(resp.Certification),
	}, nil
}

type ReviewCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCertificationLogic {
	return &ReviewCertificationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ReviewCertificationLogic) ReviewCertification(certID int64, req *types.ReviewCertificationReq) (*types.ReviewCertificationResp, error) {
	reviewerId := getUserIdFromJwt(l.ctx)
	if reviewerId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}
	resp, err := l.svcCtx.UserRpc.ReviewCertification(l.ctx, &userv1.ReviewCertificationRequest{
		CertificationId: certID,
		ReviewerId:      reviewerId,
		Result:          req.Result,
		ReviewNotes:     req.ReviewNotes,
		ExpiresAt:       req.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	if resp.Base != nil && resp.Base.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.Base.GetMsg())
	}
	return &types.ReviewCertificationResp{}, nil
}

type ListCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCertificationsLogic {
	return &ListCertificationsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListCertificationsLogic) ListCertifications(req *types.ListCertificationsReq) (*types.ListCertificationsResp, error) {
	resp, err := l.svcCtx.UserRpc.ListCertifications(l.ctx, &userv1.ListCertificationsRequest{
		Page:   &commonv1.PageRequest{Page: req.Page, PageSize: req.PageSize},
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	certs := make([]types.CertificationInfo, 0, len(resp.Certifications))
	for _, c := range resp.Certifications {
		certs = append(certs, toCertificationInfo(c))
	}
	return &types.ListCertificationsResp{
		Certifications: certs,
		Total:          resp.Page.GetTotal(),
	}, nil
}

type GetMyCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCertificationsLogic {
	return &GetMyCertificationsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetMyCertificationsLogic) GetMyCertifications() (*types.GetMyCertificationsResp, error) {
	userId := getUserIdFromJwt(l.ctx)
	if userId == 0 {
		return nil, fmt.Errorf("未登录或 token 无效")
	}
	resp, err := l.svcCtx.UserRpc.GetMyCertifications(l.ctx, &userv1.GetMyCertificationsRequest{UserId: userId})
	if err != nil {
		return nil, err
	}
	certs := make([]types.CertificationInfo, 0, len(resp.Certifications))
	for _, c := range resp.Certifications {
		certs = append(certs, toCertificationInfo(c))
	}
	return &types.GetMyCertificationsResp{Certifications: certs}, nil
}

func toCertificationInfo(c *userv1.Certification) types.CertificationInfo {
	if c == nil {
		return types.CertificationInfo{}
	}
	return types.CertificationInfo{
		Id:           c.Id,
		RoleId:       c.RoleId,
		UserId:       c.UserId,
		DocumentUrls: c.DocumentUrls,
		Status:       c.Status,
		ReviewerId:   c.ReviewerId,
		ReviewNotes:  c.ReviewNotes,
		ReviewTime:   c.ReviewTime,
		SubmitTime:   c.SubmitTime,
	}
}

func getUserIdFromJwt(ctx context.Context) int64 {
	v := ctx.Value("user_id")
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		id, _ := n.Int64()
		return id
	default:
		return 0
	}
}
