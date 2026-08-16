package notice

import (
	"context"
	"errors"
	"testing"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// =============================================================================
// Task 1.12 DeleteContentPost
// =============================================================================

func TestDeleteContentPost_AuthorWithdraw(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewDeleteContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.DeleteContentPost(&communityv1.DeleteContentPostRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.Equal(t, int64(1), pm.withdrawn, "发布者本人撤回")
}

func TestDeleteContentPost_NotAuthor(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 200)} // 作者 200，调用者 100
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewDeleteContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.DeleteContentPost(&communityv1.DeleteContentPostRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80002), resp.GetBase().GetCode(), "非发布者 080002")
	assert.Equal(t, int64(0), pm.withdrawn, "非作者不得撤回")
}

func TestDeleteContentPost_NotFound(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewDeleteContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.DeleteContentPost(&communityv1.DeleteContentPostRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode())
}

func TestDeleteContentPost_WithdrawError_NoPartialState(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 100), withdrawErr: errors.New("db error")}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewDeleteContentPostLogic(ctxWithUserID(t, 100), sc)
	_, err := l.DeleteContentPost(&communityv1.DeleteContentPostRequest{Id: 1})
	require.Error(t, err, "Withdraw 报错 → 整体失败无半态")
}

// =============================================================================
// Task 1.13 ListContentPosts
// =============================================================================

func TestListContentPosts_FilterDenied_EmptyList(t *testing.T) {
	perm := &fakePerm{
		dataFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
			return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY}, nil
		},
	}
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewListContentPostsLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.ListContentPosts(&communityv1.ListContentPostsRequest{CommunityId: 9999})
	require.NoError(t, err)
	assert.Empty(t, resp.Posts, "越权读空列表不泄露")
	assert.Equal(t, int64(0), resp.Total)
}

func TestListContentPosts_CommunityIDInjected(t *testing.T) {
	pm := &fakeContentPostModel{
		listItems: []*model.ContentPost{draftPost(1, 100), draftPost(2, 100)},
		listTotal: 2,
	}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewListContentPostsLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.ListContentPosts(&communityv1.ListContentPostsRequest{CommunityId: 2001})
	require.NoError(t, err)
	require.Len(t, resp.Posts, 2)
	for _, p := range resp.Posts {
		assert.Equal(t, int64(2001), p.CommunityId, "community_id=请求小区（scope 派生，不读弃用列）")
	}
}

// =============================================================================
// Task 1.14 GetContentPost
// =============================================================================

func TestGetContentPost_MissingCommunityID(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode(), "community_id RPC 必填")
}

func TestGetContentPost_NotFound(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode(), "FindOneReviewComplete 未找到 → 080001")
}

func TestGetContentPost_ScopeMismatch(t *testing.T) {
	pm := &fakeContentPostModel{findReviewItem: approvedPost(1, 100)}
	sm := &fakeScopeModel{scopeCommunities: []int64{2001}} // 帖 scope 仅 2001，请求 9999
	sc := noticeSvcCtx(pm, sm, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1, CommunityId: 9999})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode(), "scope 匹配缺失 → 080001")
}

func TestGetContentPost_ReadDenied(t *testing.T) {
	pm := &fakeContentPostModel{findReviewItem: approvedPost(1, 100)}
	sm := &fakeScopeModel{scopeCommunities: []int64{2001}}
	perm := &fakePerm{
		dataFn: func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
			return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY}, nil
		},
	}
	sc := noticeSvcCtx(pm, sm, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode(), "FilterAllowed false → 080001 不泄露")
}

func TestGetContentPost_Success_AttachmentRegenerated(t *testing.T) {
	pm := &fakeContentPostModel{findReviewItem: approvedPost(1, 100)}
	sm := &fakeScopeModel{scopeCommunities: []int64{2001}}
	ft := "pdf"
	am := &fakeAttachmentModel{findAtts: []*model.ContentPostAttachment{
		{Id: 11, PostId: 1, FileName: "a.pdf", FileUrl: "stored", FileSize: 1024, ReviewStatus: model.AttachmentReviewApproved, FileId: 5001, FileType: &ft},
	}}
	sc := noticeSvcCtx(pm, sm, am, permAllowAll(), &fakeMD{}, &fakeFile{url: "https://presigned/new"}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.Equal(t, int64(2001), resp.ContentPost.CommunityId, "community_id=请求小区")
	require.Len(t, resp.ContentPost.Attachments, 1)
	a := resp.ContentPost.Attachments[0]
	assert.Equal(t, "https://presigned/new", a.FileUrl, "附件 file_url 经 GetFileUrl 重生")
	assert.Equal(t, int64(5001), a.FileId)
	assert.Equal(t, int32(model.AttachmentReviewApproved), a.ReviewStatus)
	assert.Equal(t, "pdf", a.FileType)
}

func TestGetContentPost_FileIdZero_FallbackStoredURL(t *testing.T) {
	pm := &fakeContentPostModel{findReviewItem: approvedPost(1, 100)}
	sm := &fakeScopeModel{scopeCommunities: []int64{2001}}
	am := &fakeAttachmentModel{findAtts: []*model.ContentPostAttachment{
		{Id: 11, PostId: 1, FileName: "a.pdf", FileUrl: "stored-url", FileSize: 1024, ReviewStatus: model.AttachmentReviewApproved, FileId: 0},
	}}
	sc := noticeSvcCtx(pm, sm, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetContentPost(&communityv1.GetContentPostRequest{Id: 1, CommunityId: 2001})
	require.NoError(t, err)
	assert.Equal(t, "stored-url", resp.ContentPost.Attachments[0].FileUrl, "file_id=0 回退 stored file_url")
}

// =============================================================================
// Task 1.15 GetMarqueeNotices
// =============================================================================

func TestGetMarqueeNotices_MissingCommunityID(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetMarqueeNoticesLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetMarqueeNotices(&communityv1.GetMarqueeNoticesRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode())
}

func TestGetMarqueeNotices_Success(t *testing.T) {
	pm := &fakeContentPostModel{marqueeItems: []*model.ContentPost{
		{Id: 1, Title: "置顶", IsPinned: 1},
		{Id: 2, Title: "普通"},
	}}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetMarqueeNoticesLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.GetMarqueeNotices(&communityv1.GetMarqueeNoticesRequest{CommunityId: 2001})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, int64(1), resp.Items[0].Id)
	assert.Equal(t, "置顶", resp.Items[0].Title)
}

// =============================================================================
// Task 1.16 GetPublishPermission
// =============================================================================

func TestGetPublishPermission(t *testing.T) {
	now := time.Now().Unix()
	verified := func(code string, status int32, verifiedAt, expiresAt int64) *permissionv1.UserRoleInfo {
		return &permissionv1.UserRoleInfo{
			Role:       &permissionv1.Role{Code: code},
			ScopeType:  scope.ScopeTypeCommunity,
			ScopeId:    1001,
			Status:     status,
			VerifiedAt: verifiedAt,
			ExpiresAt:  expiresAt,
		}
	}

	tests := []struct {
		name        string
		roles       []*permissionv1.UserRoleInfo
		wantPublish bool
		wantRoles   []int32
	}{
		{"网格员已认证通过", []*permissionv1.UserRoleInfo{verified(scope.RoleGridWorker, scope.UserRoleStatusVerified, now, 0)}, true, []int32{4}},
		{"property_admin 通过（D6）", []*permissionv1.UserRoleInfo{verified(scope.RolePropertyAdmin, scope.UserRoleStatusVerified, now, 0)}, true, []int32{3}},
		{"committee 通过", []*permissionv1.UserRoleInfo{verified(scope.RoleCommittee, scope.UserRoleStatusVerified, now, 0)}, true, []int32{2}},
		{"多角色按固定序", []*permissionv1.UserRoleInfo{
			verified(scope.RolePropertyAdmin, scope.UserRoleStatusVerified, now, 0),
			verified(scope.RoleGridWorker, scope.UserRoleStatusVerified, now, 0),
		}, true, []int32{4, 3}},
		{"status=2 但 verified_at=0 拒绝", []*permissionv1.UserRoleInfo{verified(scope.RoleGridWorker, scope.UserRoleStatusVerified, 0, 0)}, false, nil},
		{"角色过期拒绝", []*permissionv1.UserRoleInfo{verified(scope.RoleGridWorker, scope.UserRoleStatusVerified, now, now-1000)}, false, nil},
		{"owner/tenant 拒绝", []*permissionv1.UserRoleInfo{verified("owner", scope.UserRoleStatusVerified, now, 0)}, false, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			perm := &fakePerm{
				rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
					return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: tc.roles}, nil
				},
			}
			sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
			l := NewGetPublishPermissionLogic(ctxWithUserID(t, 100), sc)
			resp, err := l.GetPublishPermission(&communityv1.GetPublishPermissionRequest{})
			require.NoError(t, err)
			assert.Equal(t, tc.wantPublish, resp.CanPublish)
			assert.Equal(t, tc.wantRoles, toInt32Slice(resp.PublishableRoles))
		})
	}
}

func TestGetPublishPermission_NoIdentity(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetPublishPermissionLogic(context.Background(), sc) // 无 metadata → userID=0
	resp, err := l.GetPublishPermission(&communityv1.GetPublishPermissionRequest{})
	require.NoError(t, err)
	assert.False(t, resp.CanPublish, "无身份 → can_publish=false")
}

func TestGetPublishPermission_TransportError(t *testing.T) {
	perm := &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return nil, errors.New("permission unavailable")
		},
	}
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewGetPublishPermissionLogic(ctxWithUserID(t, 100), sc)
	_, err := l.GetPublishPermission(&communityv1.GetPublishPermissionRequest{})
	require.Error(t, err, "GetUserRoles 传输错误 fail-closed")
}

func toInt32Slice(roles []communityv1.ContentPostRole) []int32 {
	if len(roles) == 0 {
		return nil
	}
	out := make([]int32, 0, len(roles))
	for _, r := range roles {
		out = append(out, int32(r))
	}
	return out
}
