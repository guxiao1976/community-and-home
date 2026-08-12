package notice

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	moderationv1 "github.com/guxiao1976/api-proto/gen/go/moderation/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakePerm struct {
	permissionv1.PermissionServiceClient
	assertFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
}

func (f *fakePerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	return f.assertFn(ctx, in, opts...)
}

type fakeNoticeModel struct {
	model.NoticeModel
	inserted        *model.Notice
	findItem        *model.Notice
	updateCalled    bool
	deleteCalled    bool
	modStatusCalled bool
	modStatusSetTo  int64
	findListCalled  bool
}

func (f *fakeNoticeModel) Insert(ctx context.Context, n *model.Notice) (int64, error) {
	f.inserted = n
	return n.Id, nil
}

func (f *fakeNoticeModel) FindOne(ctx context.Context, id int64) (*model.Notice, error) {
	if f.findItem == nil {
		return nil, sql.ErrNoRows
	}
	return f.findItem, nil
}

func (f *fakeNoticeModel) Update(ctx context.Context, id int64, title, content string, isPinned int32) error {
	f.updateCalled = true
	return nil
}

func (f *fakeNoticeModel) SoftDelete(ctx context.Context, id int64) error {
	f.deleteCalled = true
	return nil
}

func (f *fakeNoticeModel) UpdateModerationStatus(ctx context.Context, id int64, status int64) error {
	f.modStatusCalled = true
	f.modStatusSetTo = status
	return nil
}

func (f *fakeNoticeModel) FindList(ctx context.Context, communityId int64, role string, offset, limit int64) ([]*model.Notice, int64, error) {
	f.findListCalled = true
	return nil, 0, nil
}

type fakeModeration struct {
	moderationv1.ModerationServiceClient
}

func (f *fakeModeration) CreateAuditLog(ctx context.Context, in *moderationv1.CreateAuditLogRequest, opts ...grpc.CallOption) (*moderationv1.CreateAuditLogResponse, error) {
	return &moderationv1.CreateAuditLogResponse{Id: 1}, nil
}

func noticeCtxWithUserID(t *testing.T, uid int64) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("user_id", fmt.Sprintf("%d", uid)))
}

func noticeDenyAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{
				Base:    responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"),
				Allowed: false,
			}, nil
		},
	}
}

func noticeAllowAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
}

func noticeItem(id, communityID int64) *model.Notice {
	return &model.Notice{Id: id, CommunityId: communityID}
}

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func noticeSvcCtx(mdl *fakeNoticeModel, perm *fakePerm) *svc.ServiceContext {
	return &svc.ServiceContext{
		NoticeModel:      mdl,
		PermissionClient: perm,
		ModerationClient: &fakeModeration{},
		RedisClient:      redis.New("127.0.0.1:6379"),
	}
}

func TestUpdateNotice_ScopeDenied_OutOfScope(t *testing.T) {
	mdl := &fakeNoticeModel{findItem: noticeItem(1, 200)}
	sc := noticeSvcCtx(mdl, noticeDenyAll())

	l := NewUpdateNoticeLogic(noticeCtxWithUserID(t, 100), sc)
	resp, err := l.UpdateNotice(&communityv1.UpdateNoticeRequest{Id: 1, Title: "x", Content: "y"})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "更新超范围 → 080006")
	assert.False(t, mdl.updateCalled, "数据权限拒绝后不得更新")
}

func TestUpdateNotice_ScopeAllowed(t *testing.T) {
	mdl := &fakeNoticeModel{findItem: noticeItem(1, 100)}
	sc := noticeSvcCtx(mdl, noticeAllowAll())

	l := NewUpdateNoticeLogic(noticeCtxWithUserID(t, 100), sc)
	resp, err := l.UpdateNotice(&communityv1.UpdateNoticeRequest{Id: 1, Title: "x", Content: "y"})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, mdl.updateCalled, "数据权限允许后更新")
}

func TestDeleteNotice_ScopeDenied_NotDeleted(t *testing.T) {
	mdl := &fakeNoticeModel{findItem: noticeItem(1, 200)}
	sc := noticeSvcCtx(mdl, noticeDenyAll())

	l := NewDeleteNoticeLogic(noticeCtxWithUserID(t, 100), sc)
	resp, err := l.DeleteNotice(&communityv1.DeleteNoticeRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "删除超范围 → 080006")
	assert.False(t, mdl.deleteCalled, "数据权限拒绝后不得删除")
}

func TestDeleteNotice_ScopeAllowed(t *testing.T) {
	mdl := &fakeNoticeModel{findItem: noticeItem(1, 100)}
	sc := noticeSvcCtx(mdl, noticeAllowAll())

	l := NewDeleteNoticeLogic(noticeCtxWithUserID(t, 100), sc)
	resp, err := l.DeleteNotice(&communityv1.DeleteNoticeRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, mdl.deleteCalled, "数据权限允许后删除")
}

func TestCreateNotice_ScopeDenied(t *testing.T) {
	mdl := &fakeNoticeModel{}
	sc := noticeSvcCtx(mdl, noticeDenyAll())

	l := NewCreateNoticeLogic(noticeCtxWithUserID(t, 100), sc)
	resp, err := l.CreateNotice(&communityv1.CreateNoticeRequest{
		CommunityId: 200,
		Title:       "停水通知",
		Content:     "x",
		PublisherId: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "owner@A 发 B 通知 → 080006")
	assert.Nil(t, mdl.inserted, "数据权限拒绝后不得落库")
}
