package notice

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
)

// beginCommitConn 返回带 ExpectBegin/ExpectCommit 的 sqlmock Conn（供 Create 单事务成功路径）。
func beginCommitConn(t *testing.T) (sqlx.SqlConn, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	mock.ExpectBegin()
	mock.ExpectCommit()
	return sqlx.NewSqlConnFromDB(db), mock
}

func gridWorkerPerm() *fakePerm {
	return &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return verifiedRoles(scope.RoleGridWorker), nil
		},
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Allowed: true}, nil
		},
	}
}

func TestCreateContentPost_Validation(t *testing.T) {
	base := func() *communityv1.CreateContentPostRequest {
		return &communityv1.CreateContentPostRequest{
			SectionCode:  SectionCodeNotice,
			Title:        "停水通知",
			Text:         "内容",
			CommunityIds: []int64{2001},
		}
	}

	tests := []struct {
		name     string
		mutate   func(in *communityv1.CreateContentPostRequest)
		wantCode int32
	}{
		{"section_code 非法 → 080005", func(in *communityv1.CreateContentPostRequest) { in.SectionCode = "repair" }, 80005},
		{"title 空 → 080005", func(in *communityv1.CreateContentPostRequest) { in.Title = "" }, 80005},
		{"text 空 → 080005", func(in *communityv1.CreateContentPostRequest) { in.Text = "" }, 80005},
		{"空范围 → 080005", func(in *communityv1.CreateContentPostRequest) { in.CommunityIds = nil }, 80005},
		{"entry_status 非法 → 080005", func(in *communityv1.CreateContentPostRequest) { in.EntryStatus = 9 }, 80005},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := base()
			tc.mutate(in)
			sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, gridWorkerPerm(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
			l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
			resp, err := l.CreateContentPost(in)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, resp.GetBase().GetCode())
		})
	}
}

func TestCreateContentPost_ScopeDenied(t *testing.T) {
	in := &communityv1.CreateContentPostRequest{SectionCode: SectionCodeNotice, Title: "t", Text: "c", CommunityIds: []int64{9999}}
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permDenyAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(in)
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "越权目标 → 080006")
}

// TestCreateContentPost_DraftSuccess 正常 draft 落库：单事务 InsertTx + scope + attachments，
// role/publisher 由身份派生（伪造 body 被纠正），community_id 不写、kafka_push_status=0。
func TestCreateContentPost_DraftSuccess(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{}
	sm := &fakeScopeModel{}
	am := &fakeAttachmentModel{}
	sc := noticeSvcCtx(pm, sm, am, gridWorkerPerm(), &fakeMD{}, &fakeFile{}, &fakeUser{realName: "张三"}, nil)
	sc.Conn = conn

	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(&communityv1.CreateContentPostRequest{
		SectionCode:  SectionCodeNotice,
		Title:        "停水通知",
		Text:         "内容",
		CommunityIds: []int64{2001, 2001}, // 去重
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())

	require.NotNil(t, pm.insertedTx, "InsertTx 必须被调用")
	assert.Equal(t, model.StatusDraft, pm.insertedTx.Status, "draft 入口 status=0")
	assert.Nil(t, pm.insertedTx.CommunityId, "community_id 不写入（scope 单源）")
	assert.Equal(t, "grid_officer", pm.insertedTx.Role, "role 由 PublishRolesFrom 派生映射")
	assert.Equal(t, "张三", pm.insertedTx.Publisher, "publisher 取真实档案（禁请求体信任）")
	assert.Equal(t, int64(100), *pm.insertedTx.PublisherId, "publisher_id=JWT")
	assert.Equal(t, model.KafkaPushNone, pm.insertedTx.KafkaPushStatus, "draft 无待推标记")
	assert.Equal(t, []int64{2001}, sm.insertedScope, "scope 去重后插入")
}

// TestCreateContentPost_SubmittedImmediate 入口 submitted：status=approved + published_at=NOW +
// kafka_push_status=1 + 提交成功后 Producer.Push。
func TestCreateContentPost_SubmittedImmediate(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{}
	sm := &fakeScopeModel{}
	am := &fakeAttachmentModel{}
	pusher := &fakePusher{}
	sc := noticeSvcCtx(pm, sm, am, gridWorkerPerm(), &fakeMD{}, &fakeFile{}, &fakeUser{realName: "张三"}, pusher)
	sc.Conn = conn

	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(&communityv1.CreateContentPostRequest{
		SectionCode:  SectionCodeNotice,
		Title:        "停水通知",
		Text:         "内容",
		EntryStatus:  1, // submitted 立即提交（隐式通过 D16）
		CommunityIds: []int64{2001},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())

	require.NotNil(t, pm.insertedTx)
	assert.Equal(t, model.StatusApproved, pm.insertedTx.Status, "submitted 隐式通过 status=2")
	assert.True(t, pm.insertedTx.PublishedAt.Valid, "published_at=NOW()")
	assert.Equal(t, model.KafkaPushPending, pm.insertedTx.KafkaPushStatus, "submitted 待推标记")
	assert.Equal(t, []int64{pm.insertedTx.Id}, pusher.pushed, "事务提交成功后 Producer.Push")
}

// TestCreateContentPost_CommunityAdminExpand 社区管理员 division 展开：targets=展开小区。
func TestCreateContentPost_CommunityAdminExpand(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{}
	sm := &fakeScopeModel{}
	am := &fakeAttachmentModel{}
	perm := &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return verifiedRoles(scope.RoleCommunityAdmin), nil
		},
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Allowed: true}, nil
		},
	}
	sc := noticeSvcCtx(pm, sm, am, perm, &fakeMD{}, &fakeFile{}, &fakeUser{realName: "李四"}, nil)
	sc.Conn = conn

	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(&communityv1.CreateContentPostRequest{
		SectionCode:  SectionCodeNotice,
		Title:        "t",
		Text:         "c",
		CommunityIds: []int64{1001},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.Equal(t, []int64{2001, 2002}, sm.insertedScope, "community_admin 展开为其 division 下 approved 小区")
	assert.Equal(t, "community", pm.insertedTx.Role)
}

// TestCreateContentPost_SanitizesText XSS 净化（REQ-XSS-1）：正文落库前白名单净化。
// 注入 payload（<script>/<img onerror>）落库前被剥离；非空校验（080005）以原始正文先行（语义不变）。
func TestCreateContentPost_SanitizesText(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, gridWorkerPerm(), &fakeMD{}, &fakeFile{}, &fakeUser{realName: "张三"}, nil)
	sc.Conn = conn

	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(&communityv1.CreateContentPostRequest{
		SectionCode:  SectionCodeNotice,
		Title:        "停水通知",
		Text:         `<script>alert(document.cookie)</script><img src=x onerror=alert(1)>安全文本`,
		CommunityIds: []int64{2001},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode(), "原始正文非空 → 通过 080005，净化在非空校验后执行")
	require.NotNil(t, pm.insertedTx, "InsertTx 必须被调用")
	assert.Equal(t, "安全文本", pm.insertedTx.Text, "落库正文为净化后内容（script/img 剥离）")
	assert.NotContains(t, pm.insertedTx.Text, "<script", "落库正文不得残留 <script")
	assert.NotContains(t, pm.insertedTx.Text, "onerror=", "落库正文不得残留 onerror=")
}

// TestCreateContentPost_AttachmentOverLimit 附件 >10 个 → 080005。
func TestCreateContentPost_AttachmentOverLimit(t *testing.T) {
	ids := make([]int64, 11)
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, gridWorkerPerm(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewCreateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateContentPost(&communityv1.CreateContentPostRequest{
		SectionCode:   SectionCodeNotice,
		Title:         "t",
		Text:          "c",
		CommunityIds:  []int64{2001},
		AttachmentIds: ids,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode(), "附件超量 → 080005")
}
