package scope

import (
	"context"
	"errors"
	"testing"
	"time"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeRolesPerm 覆盖 GetUserRoles + AssertPublishScope，其余方法嵌入不调用。
type fakeRolesPerm struct {
	permissionv1.PermissionServiceClient
	rolesFn    func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error)
	assertFn   func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
}

func (f *fakeRolesPerm) GetUserRoles(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
	if f.rolesFn == nil {
		return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp()}, nil
	}
	return f.rolesFn(ctx, in, opts...)
}

func (f *fakeRolesPerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	if f.assertFn == nil {
		return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
	}
	return f.assertFn(ctx, in, opts...)
}

// fakeMD 覆盖 GetResidentialArea + GetResidentialAreasByDivision。
type fakeMD struct {
	masterdatav1.MasterdataServiceClient
	areaFn   func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error)
	byDivFn  func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error)
}

func (f *fakeMD) GetResidentialArea(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
	if f.areaFn == nil {
		return &masterdatav1.GetResidentialAreaResp{Base: responsex.NewBaseResp()}, nil
	}
	return f.areaFn(ctx, in, opts...)
}

func (f *fakeMD) GetResidentialAreasByDivision(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
	if f.byDivFn == nil {
		return &masterdatav1.GetResidentialAreasByDivisionResp{Base: responsex.NewBaseResp()}, nil
	}
	return f.byDivFn(ctx, in, opts...)
}

func verifiedGrant(roleCode string, scopeType string, scopeID int64, status int32) *permissionv1.UserRoleInfo {
	return &permissionv1.UserRoleInfo{
		Role:       &permissionv1.Role{Code: roleCode},
		ScopeType:  scopeType,
		ScopeId:    scopeID,
		Status:     status,
		VerifiedAt: time.Now().Unix(),
		ExpiresAt:  0, // 永久
	}
}

func TestExpandDivisionCommunities(t *testing.T) {
	tests := []struct {
		name       string
		divisionID int64
		byDivFn    func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error)
		callErr    error
		wantErr    bool
		wantCode   int
		want       []int64
	}{
		{
			name:       "divisionID<=0 → 080005 fail-closed",
			divisionID: 0,
			wantErr:    true,
			wantCode:   CodeInvalidParam,
		},
		{
			name:       "展开正常返回 approved 小区",
			divisionID: 90,
			byDivFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
				return &masterdatav1.GetResidentialAreasByDivisionResp{
					Base: responsex.NewBaseResp(),
					ResidentialAreas: []*masterdatav1.ResidentialArea{
						{Id: 2001, CommunityDivId: 90},
						{Id: 2002, CommunityDivId: 90},
					},
				}, nil
			},
			want: []int64{2001, 2002},
		},
		{
			name:       "展开空 → 080005",
			divisionID: 90,
			byDivFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
				return &masterdatav1.GetResidentialAreasByDivisionResp{Base: responsex.NewBaseResp()}, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name:       "传输错误 fail-closed 原样返回",
			divisionID: 90,
			callErr:    errors.New("masterdata unavailable"),
			wantErr:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			md := &fakeMD{byDivFn: tc.byDivFn}
			got, err := ExpandDivisionCommunities(context.Background(), md, tc.divisionID)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantCode != 0 {
					ce, ok := err.(*errx.CodeError)
					require.True(t, ok)
					assert.Equal(t, tc.wantCode, ce.Code)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveAdminDivision(t *testing.T) {
	permAllow := &fakeRolesPerm{}
	tests := []struct {
		name     string
		rolesFn  func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error)
		areaFn   func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error)
		wantErr  bool
		wantCode int
		want     int64
	}{
		{
			name: "单 community grant → 唯一 division",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, UserRoleStatusVerified),
				}}, nil
			},
			areaFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
				return &masterdatav1.GetResidentialAreaResp{Base: responsex.NewBaseResp(), ResidentialArea: &masterdatav1.ResidentialArea{Id: in.Id, CommunityDivId: 90}}, nil
			},
			want: 90,
		},
		{
			name: "无 community_admin grant → 080005",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommittee, ScopeTypeCommunity, 1001, UserRoleStatusVerified),
				}}, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name: "两个 community grant 映射不同 division → 080005",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, UserRoleStatusVerified),
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1002, UserRoleStatusVerified),
				}}, nil
			},
			areaFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
				div := int64(90)
				if in.Id == 1002 {
					div = 91
				}
				return &masterdatav1.GetResidentialAreaResp{Base: responsex.NewBaseResp(), ResidentialArea: &masterdatav1.ResidentialArea{Id: in.Id, CommunityDivId: div}}, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name: "两个 community grant 映射同一 division → 合并唯一放行",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, UserRoleStatusVerified),
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1002, UserRoleStatusVerified),
				}}, nil
			},
			areaFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
				return &masterdatav1.GetResidentialAreaResp{Base: responsex.NewBaseResp(), ResidentialArea: &masterdatav1.ResidentialArea{Id: in.Id, CommunityDivId: 90}}, nil
			},
			want: 90,
		},
		{
			name: "过期(4) community_admin grant 不计入 → 080005",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				expired := verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, 4)
				expired.VerifiedAt = time.Now().Unix()
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					expired,
					verifiedGrant(RoleCommittee, ScopeTypeCommunity, 2001, UserRoleStatusVerified),
				}}, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name: "驳回(3) grant 不计入 → 080005",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, 3),
				}}, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name: "GetResidentialArea 传输错误 → fail-closed",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: []*permissionv1.UserRoleInfo{
					verifiedGrant(RoleCommunityAdmin, ScopeTypeCommunity, 1001, UserRoleStatusVerified),
				}}, nil
			},
			areaFn: func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
				return nil, errors.New("masterdata unavailable")
			},
			wantErr: true,
		},
		{
			name: "GetUserRoles 传输错误 → fail-closed",
			rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
				return nil, errors.New("permission unavailable")
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perm := &fakeRolesPerm{rolesFn: tc.rolesFn}
			md := &fakeMD{areaFn: tc.areaFn}
			got, err := ResolveAdminDivision(context.Background(), perm, md, 100)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantCode != 0 {
					ce, ok := err.(*errx.CodeError)
					require.True(t, ok)
					assert.Equal(t, tc.wantCode, ce.Code)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
	_ = permAllow
}

// TestAssertCommunitiesScope 多目标批量：任一越权整体拒绝。
func TestAssertCommunitiesScope(t *testing.T) {
	tests := []struct {
		name     string
		assertFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
		targets  []int64
		wantErr  bool
		wantCode int
	}{
		{
			name: "全部目标允许",
			assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
				return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
			},
			targets: []int64{2001, 2002},
		},
		{
			name: "任一越权 → 整体 080006",
			assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
				return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseRespWithError(60007, "denied"), Allowed: false}, nil
			},
			targets:  []int64{2001, 9999},
			wantErr:  true,
			wantCode: CodePublishScopeDenied,
		},
		{
			name: "目标不存在 fail-closed → 080006",
			assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
				return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseRespWithError(60007, "not found"), Allowed: false}, nil
			},
			targets:  []int64{8888},
			wantErr:  true,
			wantCode: CodePublishScopeDenied,
		},
		{
			name: "传输错误原样返回",
			assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
				return nil, errors.New("permission unavailable")
			},
			targets: []int64{2001},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var sentTargets []int64
			var sentUserID int64
			perm := &fakeRolesPerm{
				assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
					sentUserID = in.GetUserId()
					for _, t := range in.GetTargets() {
						sentTargets = append(sentTargets, t.GetScopeId())
					}
					if tc.assertFn != nil {
						return tc.assertFn(ctx, in, opts...)
					}
					return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
				},
			}
			err := AssertCommunitiesScope(context.Background(), perm, 100, tc.targets)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantCode != 0 {
					ce, ok := err.(*errx.CodeError)
					require.True(t, ok)
					assert.Equal(t, tc.wantCode, ce.Code)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, int64(100), sentUserID)
			assert.Equal(t, tc.targets, sentTargets, "全部 target 一次携带（单次批量）")
		})
	}
}
