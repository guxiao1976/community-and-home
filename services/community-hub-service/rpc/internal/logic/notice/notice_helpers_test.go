package notice

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// fakeContentPostModel 覆盖 ContentPostModel 被调用方法（嵌入 nil 接口兜底其余）。
type fakeContentPostModel struct {
	model.ContentPostModel
	findItem          *model.ContentPost
	findReviewItem    *model.ContentPost
	listItems         []*model.ContentPost
	listTotal         int64
	marqueeItems      []*model.ContentPost
	insertedTx        *model.ContentPost
	updateContentTx   bool
	updateContentText string // XSS 净化：捕获 UpdateContentTx 落库正文
	updateIsPinnedVal int32
	isPinnedCalled    int32
	statusTxCalled    int64
	attachmentCountTx int64
	countTxCalled     bool
	withdrawn         int64
	withdrawErr       error
	listOpts          []model.ContentPostListOption // Task 1.4：捕获时间窗口选项
}

func (f *fakeContentPostModel) FindListByCommunity(ctx context.Context, communityId int64, sectionCode, role string, offset, limit int64, opts ...model.ContentPostListOption) ([]*model.ContentPost, int64, error) {
	f.listOpts = opts
	return f.listItems, f.listTotal, nil
}

func (f *fakeContentPostModel) FindMarquee(ctx context.Context, communityId int64, since time.Time, limit int64) ([]*model.ContentPost, error) {
	return f.marqueeItems, nil
}

func (f *fakeContentPostModel) InsertTx(ctx context.Context, session sqlx.Session, n *model.ContentPost) (int64, error) {
	f.insertedTx = n
	return n.Id, nil
}

func (f *fakeContentPostModel) FindOne(ctx context.Context, id int64) (*model.ContentPost, error) {
	if f.findItem == nil {
		return nil, sql.ErrNoRows
	}
	return f.findItem, nil
}

func (f *fakeContentPostModel) FindOneReviewComplete(ctx context.Context, id int64) (*model.ContentPost, error) {
	if f.findReviewItem == nil {
		return nil, sql.ErrNoRows
	}
	return f.findReviewItem, nil
}

func (f *fakeContentPostModel) UpdateContentTx(ctx context.Context, session sqlx.Session, id int64, title, text, sectionCode string) error {
	f.updateContentTx = true
	f.updateContentText = text
	return nil
}

func (f *fakeContentPostModel) UpdateIsPinned(ctx context.Context, id int64, isPinned int32) error {
	f.isPinnedCalled = isPinned
	return nil
}

func (f *fakeContentPostModel) UpdateStatusAndPublishTx(ctx context.Context, session sqlx.Session, id int64, status int64, publishedAt time.Time) error {
	f.statusTxCalled = status
	return nil
}

func (f *fakeContentPostModel) UpdateAttachmentCountTx(ctx context.Context, session sqlx.Session, id int64, count int64) error {
	f.attachmentCountTx = count
	f.countTxCalled = true
	return nil
}

func (f *fakeContentPostModel) Withdraw(ctx context.Context, id int64) error {
	if f.withdrawErr != nil {
		return f.withdrawErr
	}
	f.withdrawn = id
	return nil
}

// fakeScopeModel 覆盖 scope 模型。
type fakeScopeModel struct {
	model.ContentPostScopeModel
	scopeCommunities []int64
	insertedScope    []int64
	deleted          int64
}

func (f *fakeScopeModel) InsertBatchTx(ctx context.Context, session sqlx.Session, postId int64, communityIds []int64) error {
	f.insertedScope = communityIds
	return nil
}

func (f *fakeScopeModel) FindCommunityIdsByPostId(ctx context.Context, postId int64) ([]int64, error) {
	return f.scopeCommunities, nil
}

func (f *fakeScopeModel) DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error {
	f.deleted = postId
	return nil
}

// fakeAttachmentModel 覆盖附件模型。
type fakeAttachmentModel struct {
	model.ContentPostAttachmentModel
	insertedAtts []*model.ContentPostAttachment
	findAtts     []*model.ContentPostAttachment
	deleted      int64
}

func (f *fakeAttachmentModel) InsertBatchTx(ctx context.Context, session sqlx.Session, attachments []*model.ContentPostAttachment) error {
	f.insertedAtts = attachments
	return nil
}

func (f *fakeAttachmentModel) FindByPostId(ctx context.Context, postId int64) ([]*model.ContentPostAttachment, error) {
	return f.findAtts, nil
}

func (f *fakeAttachmentModel) DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error {
	f.deleted = postId
	return nil
}

// fakePerm 覆盖 GetUserRoles + AssertPublishScope + GetDataScopes。
type fakePerm struct {
	permissionv1.PermissionServiceClient
	rolesFn func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error)
	scopeFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
	dataFn  func(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error)
}

func (f *fakePerm) GetDataScopes(ctx context.Context, in *permissionv1.GetDataScopesRequest, opts ...grpc.CallOption) (*permissionv1.GetDataScopesResponse, error) {
	if f.dataFn == nil {
		return &permissionv1.GetDataScopesResponse{Base: responsex.NewBaseResp(), State: permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, ScopeIds: []int64{2001}}, nil
	}
	return f.dataFn(ctx, in, opts...)
}

func (f *fakePerm) GetUserRoles(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
	if f.rolesFn == nil {
		return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp()}, nil
	}
	return f.rolesFn(ctx, in, opts...)
}

func (f *fakePerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	if f.scopeFn == nil {
		return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
	}
	return f.scopeFn(ctx, in, opts...)
}

func permAllowAll() *fakePerm {
	return &fakePerm{
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
}

func permDenyAll() *fakePerm {
	return &fakePerm{
		rolesFn: func(ctx context.Context, in *permissionv1.GetUserRolesRequest, opts ...grpc.CallOption) (*permissionv1.GetUserRolesResponse, error) {
			return verifiedRoles(scope.RoleGridWorker), nil
		},
		scopeFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseRespWithError(60007, "denied"), Allowed: false}, nil
		},
	}
}

// fakeMD 覆盖 masterdata division 接口。
type fakeMD struct {
	masterdatav1.MasterdataServiceClient
	areaFn  func(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error)
	byDivFn func(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error)
}

func (f *fakeMD) GetResidentialArea(ctx context.Context, in *masterdatav1.GetResidentialAreaReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreaResp, error) {
	if f.areaFn == nil {
		return &masterdatav1.GetResidentialAreaResp{Base: responsex.NewBaseResp(), ResidentialArea: &masterdatav1.ResidentialArea{Id: in.Id, CommunityDivId: 90}}, nil
	}
	return f.areaFn(ctx, in, opts...)
}

func (f *fakeMD) GetResidentialAreasByDivision(ctx context.Context, in *masterdatav1.GetResidentialAreasByDivisionReq, opts ...grpc.CallOption) (*masterdatav1.GetResidentialAreasByDivisionResp, error) {
	if f.byDivFn == nil {
		return &masterdatav1.GetResidentialAreasByDivisionResp{Base: responsex.NewBaseResp(), ResidentialAreas: []*masterdatav1.ResidentialArea{{Id: 2001}, {Id: 2002}}}, nil
	}
	return f.byDivFn(ctx, in, opts...)
}

// fakeFile 覆盖 GetFileUrl。
type fakeFile struct {
	filev1.FileServiceClient
	url string
	err error
}

func (f *fakeFile) GetFileUrl(ctx context.Context, in *filev1.GetFileUrlRequest, opts ...grpc.CallOption) (*filev1.GetFileUrlResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &filev1.GetFileUrlResponse{
		Base:        responsex.NewBaseResp(),
		DownloadUrl: f.url,
		File:        &filev1.FileInfo{Id: in.FileId, UserId: 100, FileName: "a.pdf", FileSize: 1024, FileType: "pdf", Confirmed: true},
	}, nil
}

// fakeUser 覆盖 GetUsersByIds。
type fakeUser struct {
	userv1.UserServiceClient
	realName string
	nickname string
}

func (f *fakeUser) GetUsersByIds(ctx context.Context, in *userv1.GetUsersByIdsRequest, opts ...grpc.CallOption) (*userv1.GetUsersByIdsResponse, error) {
	return &userv1.GetUsersByIdsResponse{
		Base:  responsex.NewBaseResp(),
		Users: []*userv1.User{{Id: 100, RealName: f.realName, Nickname: f.nickname}},
	}, nil
}

// fakePusher 断言 Producer.Push 被调用；pushedTexts 捕获完整消息 payload（Text），
// 用于「推送内容 == 落库最终值」断言（SEE: [[kafka-event-payload-must-reflect-persisted-state]]）。
type fakePusher struct {
	pushed      []int64
	pushedTexts []string
	pushErr     error
}

func (f *fakePusher) Push(ctx context.Context, post *model.ContentPost, attachments []*model.ContentPostAttachment) error {
	f.pushed = append(f.pushed, post.Id)
	f.pushedTexts = append(f.pushedTexts, post.Text)
	return f.pushErr
}

// noticeSvcCtx 构造测试 ServiceContext。
func noticeSvcCtx(pm model.ContentPostModel, sm model.ContentPostScopeModel, am model.ContentPostAttachmentModel,
	perm *fakePerm, md *fakeMD, file *fakeFile, user *fakeUser, pusher *fakePusher) *svc.ServiceContext {
	return &svc.ServiceContext{
		ContentPostModel:           pm,
		ContentPostScopeModel:      sm,
		ContentPostAttachmentModel: am,
		PermissionClient:           perm,
		MasterDataClient:           md,
		FileClient:                 file,
		UserClient:                 user,
		KafkaProducer:              pusher,
		RedisClient:                redis.New("127.0.0.1:6379"),
	}
}

func ctxWithUserID(t *testing.T, uid int64) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(scope.UserIDMetadataKey, fmt.Sprintf("%d", uid)))
}

func verifiedRoles(codes ...string) *permissionv1.GetUserRolesResponse {
	now := time.Now().Unix()
	roles := make([]*permissionv1.UserRoleInfo, 0, len(codes))
	for _, c := range codes {
		roles = append(roles, &permissionv1.UserRoleInfo{
			Role:       &permissionv1.Role{Code: c},
			ScopeType:  scope.ScopeTypeCommunity,
			ScopeId:    1001,
			Status:     scope.UserRoleStatusVerified,
			VerifiedAt: now,
			ExpiresAt:  0,
		})
	}
	return &permissionv1.GetUserRolesResponse{Base: responsex.NewBaseResp(), Roles: roles}
}

func draftPost(id, publisherID int64) *model.ContentPost {
	return &model.ContentPost{Id: id, PublisherId: &publisherID, Status: model.StatusDraft, Title: "t", Text: "c", SectionCode: "notice"}
}

func approvedPost(id, publisherID int64) *model.ContentPost {
	return &model.ContentPost{Id: id, PublisherId: &publisherID, Status: model.StatusApproved, Title: "t", Text: "c", SectionCode: "notice"}
}
