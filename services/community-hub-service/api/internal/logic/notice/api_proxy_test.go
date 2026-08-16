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
	createReq   *communityv1.CreateContentPostRequest
	updateReq   *communityv1.UpdateContentPostRequest
	getReq      *communityv1.GetContentPostRequest
	marqueeReq  *communityv1.GetMarqueeNoticesRequest
	getResp     *communityv1.GetContentPostResponse
	getErr      error
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
