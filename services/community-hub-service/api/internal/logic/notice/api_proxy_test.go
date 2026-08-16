package notice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"strconv"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/api/internal/svc"
	"github.com/guxiao1976/community-hub/api/internal/types"
	"github.com/guxiao1976/community-hub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeContentPostRpc 记录请求并返回 canned 响应。
type fakeContentPostRpc struct {
	communityv1.ContentPostServiceClient
	createReq  *communityv1.CreateContentPostRequest
	updateReq  *communityv1.UpdateContentPostRequest
	getReq     *communityv1.GetContentPostRequest
	marqueeReq *communityv1.GetMarqueeNoticesRequest
	getResp    *communityv1.GetContentPostResponse
	getErr     error
	listReq    *communityv1.ListContentPostsRequest // Task 1.5：捕获 since_days 透传
	listResp   *communityv1.ListContentPostsResponse
	listErr    error
}

func (f *fakeContentPostRpc) CreateContentPost(ctx context.Context, in *communityv1.CreateContentPostRequest, opts ...grpc.CallOption) (*communityv1.CreateContentPostResponse, error) {
	f.createReq = in
	return &communityv1.CreateContentPostResponse{Base: responsex.NewBaseResp(), Id: 1001}, nil
}

func (f *fakeContentPostRpc) UpdateContentPost(ctx context.Context, in *communityv1.UpdateContentPostRequest, opts ...grpc.CallOption) (*communityv1.UpdateContentPostResponse, error) {
	f.updateReq = in
	return &communityv1.UpdateContentPostResponse{Base: responsex.NewBaseResp()}, nil
}

func (f *fakeContentPostRpc) GetContentPost(ctx context.Context, in *communityv1.GetContentPostRequest, opts ...grpc.CallOption) (*communityv1.GetContentPostResponse, error) {
	f.getReq = in
	if f.getErr != nil {
		return nil, f.getErr
	}
	if f.getResp != nil {
		return f.getResp, nil
	}
	return &communityv1.GetContentPostResponse{Base: responsex.NewBaseResp(), ContentPost: &communityv1.ContentPost{Id: in.Id, CommunityId: in.CommunityId, Title: "t", Text: "c"}}, nil
}

func (f *fakeContentPostRpc) GetMarqueeNotices(ctx context.Context, in *communityv1.GetMarqueeNoticesRequest, opts ...grpc.CallOption) (*communityv1.GetMarqueeNoticesResponse, error) {
	f.marqueeReq = in
	return &communityv1.GetMarqueeNoticesResponse{Base: responsex.NewBaseResp(), Items: []*communityv1.ContentPostMarqueeItem{{Id: 1, Title: "t"}}}, nil
}

func (f *fakeContentPostRpc) ListContentPosts(ctx context.Context, in *communityv1.ListContentPostsRequest, opts ...grpc.CallOption) (*communityv1.ListContentPostsResponse, error) {
	f.listReq = in
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listResp != nil {
		return f.listResp, nil
	}
	return &communityv1.ListContentPostsResponse{Base: responsex.NewBaseResp()}, nil
}

// fakePermData 覆盖 GetDataScopes。
type fakePermData struct {
	permissionv1.PermissionServiceClient
	allowedIDs []int64
	err        error
}

func (f *fakePermData) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, ScopeIds: f.allowedIDs}, nil
}

// apiCompatModels 供 compat 回退的模型。
type apiCompatModels struct {
	postModel  *apiFakePostModel
	scopeModel *apiFakeScopeModel
}

type apiFakePostModel struct {
	model.ContentPostModel
	found bool
}

func (f *apiFakePostModel) FindOneReviewComplete(ctx context.Context, id int64) (*model.ContentPost, error) {
	if !f.found {
		return nil, sql.ErrNoRows
	}
	return &model.ContentPost{Id: id}, nil
}

type apiFakeScopeModel struct {
	model.ContentPostScopeModel
	ids []int64
}

func (f *apiFakeScopeModel) FindCommunityIdsByPostId(ctx context.Context, postId int64) ([]int64, error) {
	return f.ids, nil
}

func jwtCtx(uid int64) context.Context {
	return context.WithValue(context.Background(), "user_id", json.Number(strconv.FormatInt(uid, 10)))
}

func apiSvcCtx(rpc *fakeContentPostRpc, perm *fakePermData, pm model.ContentPostModel, sm model.ContentPostScopeModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		ContentPostServiceRpc: rpc,
		PermClient:            perm,
		ContentPostModel:      pm,
		ContentPostScopeModel: sm,
	}
}

// =============================================================================
// Create：community_ids 解析 + entry_status 同号映射
// =============================================================================

func TestCreateContentPost_CommunityIDsParsed(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
	l := NewCreateContentPostLogic(jwtCtx(100), sc)

	resp, err := l.CreateContentPost(&types.CreateContentPostReq{
		SectionCode:   "notice",
		Title:         "t",
		Text:          "c",
		EntryStatus:   1,
		CommunityIds:  []string{"1001", "1002"},
		AttachmentIds: []string{"5001"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1001), resp.Id)
	require.NotNil(t, rpc.createReq)
	assert.Equal(t, []int64{1001, 1002}, rpc.createReq.CommunityIds, "[]string → []int64 解码")
	assert.Equal(t, []int64{5001}, rpc.createReq.AttachmentIds, "attachment_ids 透传")
	assert.Equal(t, int32(1), rpc.createReq.EntryStatus, "entry_status 同号映射")
}

func TestCreateContentPost_CommunityIDsNonNumeric(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
	l := NewCreateContentPostLogic(jwtCtx(100), sc)

	_, err := l.CreateContentPost(&types.CreateContentPostReq{
		CommunityIds: []string{"abc"},
	})
	require.Error(t, err)
	ce, ok := err.(*errx.CodeError)
	require.True(t, ok)
	assert.Equal(t, 80005, ce.Code, "community_ids 含非数字 → 080005")
	require.Nil(t, rpc.createReq, "解析失败不调 RPC")
}

// =============================================================================
// Update：presence 转发（V5）
// =============================================================================

func TestUpdateContentPost_PresenceForwarding(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
	l := NewUpdateContentPostLogic(jwtCtx(100), sc)

	title := "新标题"
	pinnedFalse := false
	err := l.UpdateContentPost(&types.UpdateContentPostReq{
		Id:                  1,
		Title:               &title,
		HasAttachmentChange: true,
		AttachmentIds:       []string{},
		IsPinned:            &pinnedFalse,
		HasScopeChange:      false,
		CommunityIds:        []string{"2001"},
	})
	require.NoError(t, err)
	require.NotNil(t, rpc.updateReq)
	assert.Equal(t, "新标题", *rpc.updateReq.Title, "Title=ptr → RPC 填 title")
	assert.Nil(t, rpc.updateReq.Text, "Text=nil → RPC 不填 text（presence 不丢失）")
	assert.Equal(t, false, *rpc.updateReq.IsPinned, "IsPinned=ptr(false) → RPC is_pinned=false（取消置顶语义不坍缩）")
	assert.True(t, rpc.updateReq.HasAttachmentChange, "HasAttachmentChange 直传")
	assert.Empty(t, rpc.updateReq.AttachmentIds, "true + 空数组 → RPC 空数组转发（清空语义到达）")
	assert.False(t, rpc.updateReq.HasScopeChange, "HasScopeChange=false → 不改 scope")
}

// =============================================================================
// Get：community_id 兼容回退（R2）
// =============================================================================

func TestGetContentPost_CommunityIDProvided(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
	l := NewGetContentPostLogic(jwtCtx(100), sc)

	_, err := l.GetContentPost(&types.GetContentPostReq{Id: 1, CommunityId: 456})
	require.NoError(t, err)
	require.NotNil(t, rpc.getReq)
	assert.Equal(t, int64(456), rpc.getReq.CommunityId, "form 绑定 community_id 透传")
}

func TestGetContentPost_Compat_AnyReadable(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{allowedIDs: []int64{2002}}, &apiFakePostModel{found: true}, &apiFakeScopeModel{ids: []int64{2001, 2002}})
	l := NewGetContentPostLogic(jwtCtx(100), sc)

	resp, err := l.GetContentPost(&types.GetContentPostReq{Id: 1})
	require.NoError(t, err)
	require.NotNil(t, resp.Notice)
	assert.Equal(t, int64(2002), rpc.getReq.CommunityId, "scope 反查任一可读小区注入（多小区用户不 080005）")
}

func TestGetContentPost_Compat_AllUnreadable(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{allowedIDs: []int64{9999}}, &apiFakePostModel{found: true}, &apiFakeScopeModel{ids: []int64{2001, 2002}})
	l := NewGetContentPostLogic(jwtCtx(100), sc)

	_, err := l.GetContentPost(&types.GetContentPostReq{Id: 1})
	require.Error(t, err)
	ce, ok := err.(*errx.CodeError)
	require.True(t, ok)
	assert.Equal(t, 80001, ce.Code, "全部不可读 → 080001（不泄露）")
	require.Nil(t, rpc.getReq, "未解析到可读小区不调 RPC")
}

func TestGetContentPost_Compat_NoScope(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, &apiFakePostModel{found: true}, &apiFakeScopeModel{ids: nil})
	l := NewGetContentPostLogic(jwtCtx(100), sc)

	_, err := l.GetContentPost(&types.GetContentPostReq{Id: 1})
	require.Error(t, err)
	ce, ok := err.(*errx.CodeError)
	require.True(t, ok)
	assert.Equal(t, 80005, ce.Code, "帖无 scope（数据异常）→ 080005")
}

func TestGetContentPost_Compat_PermTransportError(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{err: errors.New("permission down")}, &apiFakePostModel{found: true}, &apiFakeScopeModel{ids: []int64{2001}})
	l := NewGetContentPostLogic(jwtCtx(100), sc)

	_, err := l.GetContentPost(&types.GetContentPostReq{Id: 1})
	require.Error(t, err, "FilterAllowed 传输错误 fail-closed")
}

// =============================================================================
// Marquee：community_id 绑定透传
// =============================================================================

func TestGetMarqueeNotices_CommunityIDBound(t *testing.T) {
	rpc := &fakeContentPostRpc{}
	sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
	l := NewGetMarqueeNoticesLogic(jwtCtx(100), sc)

	resp, err := l.GetMarqueeNotices(&types.GetMarqueeNoticesReq{CommunityId: 456})
	require.NoError(t, err)
	require.NotNil(t, rpc.marqueeReq)
	assert.Equal(t, int64(456), rpc.marqueeReq.CommunityId, "marquee ?community_id 绑定")
	assert.Len(t, resp.Items, 1)
}

// =============================================================================
// ListContentPosts：since_days 透传（REVISION r2-2）+ Base 错误上抛
// =============================================================================

// TestListContentPosts_SinceDaysAndBaseError — REST 薄代理层：
//   - req.SinceDays=30 → RPC 请求 SinceDays=30（必贯通，否则移动端 30 天窗口静默失效）；
//   - RPC Base code=080005 → api 层 responsex.ToError 上抛（禁止静默吞错，与 getcontentpostlogic 一致）；
//   - RPC 成功（nil Base）→ 正常映射 res.notices + total（R2 wire 键保持）。
//
// SEE: [[verify-api-before-calling]] — 路由 GET /api/community/notices 已在 graph-context 确认
func TestListContentPosts_SinceDaysAndBaseError(t *testing.T) {
	t.Run("SinceDays=30 透传到 RPC", func(t *testing.T) {
		rpc := &fakeContentPostRpc{}
		sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
		l := NewListContentPostsLogic(jwtCtx(100), sc)

		_, err := l.ListContentPosts(&types.ListContentPostsReq{CommunityId: 456, SinceDays: 30})
		require.NoError(t, err)
		require.NotNil(t, rpc.listReq)
		assert.Equal(t, int32(30), rpc.listReq.SinceDays, "form since_days → RPC since_days 透传（缺省 0 → 不过滤）")
		assert.Equal(t, int64(456), rpc.listReq.CommunityId)
	})

	t.Run("RPC Base 080005 → api 上抛 error（不静默吞错）", func(t *testing.T) {
		rpc := &fakeContentPostRpc{listResp: &communityv1.ListContentPostsResponse{
			Base: responsex.NewBaseRespWithError(80005, "since_days 超出有效范围 1..365"),
		}}
		sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
		l := NewListContentPostsLogic(jwtCtx(100), sc)

		_, err := l.ListContentPosts(&types.ListContentPostsReq{CommunityId: 456, SinceDays: -1})
		require.Error(t, err)
		ce, ok := err.(*errx.CodeError)
		require.True(t, ok, "必须为 errx.CodeError 以便 REST 面透出 080005")
		assert.Equal(t, 80005, ce.Code)
	})

	t.Run("RPC 成功 → 映射 res.notices + total（R2 wire 键）", func(t *testing.T) {
		rpc := &fakeContentPostRpc{listResp: &communityv1.ListContentPostsResponse{
			Base: responsex.NewBaseResp(),
			Posts: []*communityv1.ContentPost{
				{Id: 1001, CommunityId: 456, Title: "t1", Text: "c1"},
			},
			Total: 1,
		}}
		sc := apiSvcCtx(rpc, &fakePermData{}, nil, nil)
		l := NewListContentPostsLogic(jwtCtx(100), sc)

		resp, err := l.ListContentPosts(&types.ListContentPostsReq{CommunityId: 456, SinceDays: 0})
		require.NoError(t, err)
		require.Len(t, resp.Notices, 1)
		assert.Equal(t, int64(1001), resp.Notices[0].Id)
		assert.Equal(t, "t1", resp.Notices[0].Title)
		assert.Equal(t, int64(1), resp.Total, "total 透传")
	})
}
