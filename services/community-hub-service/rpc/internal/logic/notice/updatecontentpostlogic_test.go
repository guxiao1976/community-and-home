package notice

import (
	"context"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-hub/internal/sanitize"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func int64Ptr(v int64) *int64 { return &v }

// TestUpdateContentPost_NotFound 不存在 → 080001。
func TestUpdateContentPost_NotFound(t *testing.T) {
	sc := noticeSvcCtx(&fakeContentPostModel{}, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Title: strPtr("x")})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode())
}

// TestUpdateContentPost_ContentEdit_AuthorMismatch (a) 分支非作者 → 080002。
func TestUpdateContentPost_ContentEdit_AuthorMismatch(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 200)} // 作者 200，调用者 100
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Title: strPtr("x")})
	require.NoError(t, err)
	assert.Equal(t, int32(80002), resp.GetBase().GetCode(), "(a) 分支先作者校验")
}

// TestUpdateContentPost_DraftEdit_ContentAndAttachmentsAndScope
// draft 编辑：正文全量替换 + 附件集合重写（attachment_count 重算）+ scope 重写，单事务 all-or-nothing。
func TestUpdateContentPost_DraftEdit_ContentAndAttachmentsAndScope(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sm := &fakeScopeModel{}
	am := &fakeAttachmentModel{}
	sc := noticeSvcCtx(pm, sm, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{
		Id:                  1,
		Title:               strPtr("新标题"),
		HasAttachmentChange: true,
		AttachmentIds:       []int64{5001},
		HasScopeChange:      true,
		CommunityIds:        []int64{3001},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, pm.updateContentTx, "正文列更新（UpdateContentTx）")
	assert.True(t, pm.countTxCalled, "attachment_count 同事务重算")
	assert.Equal(t, int64(1), pm.attachmentCountTx, "attachment_count=新绑定数")
	assert.Equal(t, int64(1), am.deleted, "附件旧行删除")
	require.Len(t, am.insertedAtts, 1, "附件全量替换插入")
	assert.Equal(t, int64(1), am.insertedAtts[0].PostId, "附件 post_id 全链一致")
	assert.Equal(t, int64(1), sm.deleted, "scope 旧行删除")
	assert.Equal(t, []int64{3001}, sm.insertedScope, "scope 全量替换")
}

// TestUpdateContentPost_ClearAllAttachments 清空全部附件（HasAttachmentChange=true + 空集 → attachment_count=0）。
func TestUpdateContentPost_ClearAllAttachments(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sm := &fakeScopeModel{}
	am := &fakeAttachmentModel{}
	sc := noticeSvcCtx(pm, sm, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{
		Id:                  1,
		HasAttachmentChange: true,
		AttachmentIds:       []int64{},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, pm.countTxCalled)
	assert.Equal(t, int64(0), pm.attachmentCountTx, "空集 → attachment_count=0（D19 可归零，评审 MUST 1(b)）")
	assert.Empty(t, am.insertedAtts, "附件行全删")
}

// TestUpdateContentPost_ContentEdit_NonDraft (a) 分支非 draft 内容编辑 → 080005（仅 draft 可编辑）。
func TestUpdateContentPost_ContentEdit_NonDraft(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Title: strPtr("x")})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode(), "submitted/approved 不可内容编辑")
}

// TestUpdateContentPost_TitleEmpty 携带空串 title → 080005（title/text 非空不变量，评审 MUST 1(c)）。
func TestUpdateContentPost_TitleEmpty(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Title: strPtr("")})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode())
}

// TestUpdateContentPost_ScopeEmpty 空 scope 集（HasScopeChange=true + 空 community_ids）→ 080005。
func TestUpdateContentPost_ScopeEmpty(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, HasScopeChange: true, CommunityIds: []int64{}})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode())
}

// TestUpdateContentPost_ScopeDenied scope 越权 → 080006。
func TestUpdateContentPost_ScopeDenied(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permDenyAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, HasScopeChange: true, CommunityIds: []int64{9999}})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode())
}

// TestUpdateContentPost_Submit draft submit → status=approved + kafka_push_status=1 + Producer.Push，不 LPUSH Redis。
func TestUpdateContentPost_Submit(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	am := &fakeAttachmentModel{findAtts: []*model.ContentPostAttachment{{Id: 11, PostId: 1, FileId: 5001}}}
	pusher := &fakePusher{}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, pusher)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.Equal(t, int64(model.StatusApproved), pm.statusTxCalled, "submit → UpdateStatusAndPublish(approved)")
	assert.Equal(t, []int64{1}, pusher.pushed, "提交成功后 Producer.Push")
}

// TestUpdateContentPost_ContentEdit_SanitizesText
// 内容编辑分支（text 携带）：正文落库前白名单净化（REQ-XSS-1）；净化在非空校验后、DB 落库前执行。
func TestUpdateContentPost_ContentEdit_SanitizesText(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{
		Id:   1,
		Text: strPtr(`<script>alert(1)</script><iframe src=x></iframe>净化后正文`),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, pm.updateContentTx, "正文携带 → UpdateContentTx 执行")
	assert.Equal(t, "净化后正文", pm.updateContentText, "落库正文为净化后内容")
	assert.NotContains(t, pm.updateContentText, "<script", "落库正文不得残留 <script")
	assert.NotContains(t, pm.updateContentText, "<iframe", "落库正文不得残留 <iframe")
}

// TestUpdateContentPost_ContentEdit_TextNotPresentNoResanitize
// Update 正文未携带（proto3 presence）：不重写正文、不重净化，现值保持（REQ-XSS-6/D11 残余风险）。
func TestUpdateContentPost_ContentEdit_TextNotPresentNoResanitize(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{
		Id:    1,
		Title: strPtr("仅改标题"),
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, pm.updateContentTx, "标题携带 → UpdateContentTx 执行（title/text/section_code 三列）")
	assert.Equal(t, "c", pm.updateContentText, "正文未携带 → text 保持现值，不重净化")
}

// TestUpdateContentPost_Submit_SanitizesDraftText
// submit 发布分支（REQ-XSS-6/D9）：存量 draft 正文置公开前先净化，同一事务净化后正文 + 置公开。
func TestUpdateContentPost_Submit_SanitizesDraftText(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: &model.ContentPost{
		Id: 1, PublisherId: int64Ptr(100), Status: model.StatusDraft,
		Title: "t", Text: `<script>alert(1)</script><img src=x onerror=alert(1)>存量正文`, SectionCode: "notice",
	}}
	am := &fakeAttachmentModel{findAtts: []*model.ContentPostAttachment{{Id: 11, PostId: 1, FileId: 5001}}}
	pusher := &fakePusher{}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, pusher)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, pm.updateContentTx, "置公开前先净化存量 draft 正文（UpdateContentTx）")
	assert.Equal(t, "存量正文", pm.updateContentText, "净化后正文写入同一事务")
	assert.Equal(t, int64(model.StatusApproved), pm.statusTxCalled, "随后 UpdateStatusAndPublish(approved)")
	assert.Equal(t, []int64{1}, pusher.pushed, "提交成功后 Producer.Push")
}

// TestUpdateContentPost_Submit_AlreadySanitizedNoRewrite
// 幂等（REQ-XSS-3）：既有已净化正文经 submit 不二次改写（不调用 UpdateContentTx）。
func TestUpdateContentPost_Submit_AlreadySanitizedNoRewrite(t *testing.T) {
	conn, _ := beginCommitConn(t)
	pm := &fakeContentPostModel{findItem: &model.ContentPost{
		Id: 1, PublisherId: int64Ptr(100), Status: model.StatusDraft,
		Title: "t", Text: "已净化正文", SectionCode: "notice",
	}}
	am := &fakeAttachmentModel{findAtts: []*model.ContentPostAttachment{{Id: 11, PostId: 1, FileId: 5001}}}
	pusher := &fakePusher{}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, am, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, pusher)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.False(t, pm.updateContentTx, "净化前后一致 → 不二次改写正文（幂等）")
	assert.Equal(t, int64(model.StatusApproved), pm.statusTxCalled, "仅置公开")
}

// TestUpdateContentPost_SubmitNonDraft non-draft submit → 080005。
func TestUpdateContentPost_SubmitNonDraft(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode())
}

// TestUpdateContentPost_StatusInvalid status 非法值 → 080005。
func TestUpdateContentPost_StatusInvalid(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 9})
	require.NoError(t, err)
	assert.Equal(t, int32(80005), resp.GetBase().GetCode())
}

// TestUpdateContentPost_IsPinnedOnly_DraftAuthor draft 帖发布者置顶成功（正文不变，经 UpdateIsPinned）。
func TestUpdateContentPost_IsPinnedOnly_DraftAuthor(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 100)}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, IsPinned: boolPtr(true)})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.Equal(t, int32(1), pm.isPinnedCalled, "置顶走 UpdateIsPinned 独立列")
	assert.False(t, pm.updateContentTx, "正文列不动")
}

// TestUpdateContentPost_IsPinnedOnly_OperatorPinsApproved
// 非作者操作者（持发布角色 + scope 覆盖）置顶 approved 帖成功（不 080002，REQ-CPB-9(f)）。
func TestUpdateContentPost_IsPinnedOnly_OperatorPinsApproved(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 200)} // 作者 200，操作者 100
	sm := &fakeScopeModel{scopeCommunities: []int64{2001}}
	perm := &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return verifiedRoles(scope.RoleGridWorker), nil
		},
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Allowed: true}, nil
		},
	}
	sc := noticeSvcCtx(pm, sm, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, IsPinned: boolPtr(true)})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode(), "非作者操作者置顶 approved 帖成功")
	assert.Equal(t, int32(1), pm.isPinnedCalled)
}

// TestUpdateContentPost_IsPinnedOnly_OperatorScopeNotCovered
// 非作者操作者 scope 不覆盖 → 080006（不 080002）。
func TestUpdateContentPost_IsPinnedOnly_OperatorScopeNotCovered(t *testing.T) {
	pm := &fakeContentPostModel{findItem: approvedPost(1, 200)}
	sm := &fakeScopeModel{scopeCommunities: []int64{9999}} // 帖小区超出操作者范围
	perm := &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return verifiedRoles(scope.RoleGridWorker), nil
		},
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Allowed: false}, nil
		},
	}
	sc := noticeSvcCtx(pm, sm, &fakeAttachmentModel{}, perm, &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, IsPinned: boolPtr(false)})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "(b) 分支 scope 不覆盖 → 080006")
}

// TestUpdateContentPost_MixedFields_GoesToAuthorCheck 请求含 is_pinned+内容字段 → 走 (a) 作者校验 080002。
func TestUpdateContentPost_MixedFields_GoesToAuthorCheck(t *testing.T) {
	pm := &fakeContentPostModel{findItem: draftPost(1, 200)} // 作者 200，调用者 100
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, nil)
	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, IsPinned: boolPtr(true), Title: strPtr("x")})
	require.NoError(t, err)
	assert.Equal(t, int32(80002), resp.GetBase().GetCode(), "is_pinned+内容字段 → (a) 作者校验")
}

// TestUpdateContentPost_Submit_SanitizesTextBeforeKafkaPush
// submit 发布路径：存量 draft 正文置公开前先白名单净化（REQ-XSS-6/D9），
// 且 Kafka 推送 payload 必须反映落库后的净化值（禁止复用 FindOne 载入的未净化快照）。
// SEE: [[kafka-event-payload-must-reflect-persisted-state]]
func TestUpdateContentPost_Submit_SanitizesTextBeforeKafkaPush(t *testing.T) {
	conn, _ := beginCommitConn(t)
	malicious := "<p>你好</p><img src=x onerror=alert(1)><script>alert(2)</script>"
	draft := draftPost(1, 100)
	draft.Text = malicious
	pm := &fakeContentPostModel{findItem: draft}
	pusher := &fakePusher{}
	sc := noticeSvcCtx(pm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, pusher)
	sc.Conn = conn

	l := NewUpdateContentPostLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.UpdateContentPost(&communityv1.UpdateContentPostRequest{Id: 1, Status: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())

	// 推送 payload 必须与落库净化值一致（post.Text 刷新为 sanitizedText，而非 FindOne 快照）
	require.Len(t, pusher.pushedTexts, 1, "submit 提交成功后 Producer.Push 一次")
	assert.NotContains(t, pusher.pushedTexts[0], "<img", "推送 payload 不得含未净化 img")
	assert.NotContains(t, pusher.pushedTexts[0], "onerror", "推送 payload 不得含事件属性")
	assert.NotContains(t, pusher.pushedTexts[0], "<script", "推送 payload 不得含 script")
	assert.Equal(t, sanitize.ContentPostText(malicious), pusher.pushedTexts[0], "推送 payload == 落库净化值")
}
