package lostfound

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
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

// fakePerm 仅覆盖数据权限相关方法，其余嵌入不调用
type fakePerm struct {
	permissionv1.PermissionServiceClient
	assertFn func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error)
}

func (f *fakePerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	return f.assertFn(ctx, in, opts...)
}

// fakeMasterData 仅覆盖 GetSectionQuota；未显式设置 quotaFn 时默认「未配置=不限」（放行）。
type fakeMasterData struct {
	masterdatav1.MasterdataServiceClient
	quotaFn func(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error)
}

func (f *fakeMasterData) GetSectionQuota(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error) {
	if f.quotaFn != nil {
		return f.quotaFn(ctx, in, opts...)
	}
	return &masterdatav1.GetSectionQuotaResp{Base: responsex.NewBaseResp(), Configured: false}, nil
}

// fakeLostFoundModel 记录 Insert / UpdateStatus / FindList 调用
type fakeLostFoundModel struct {
	model.LostFoundItemModel
	inserted          *model.LostFoundItem
	findItem          *model.LostFoundItem
	findPublishedItem *model.LostFoundItem
	findPublishedErr  error
	updateCalled      bool
	updatedStatus     string
	modStatusCalled   bool
	modStatusSetTo    int64
	findListCalled    bool
	quotaCount        int64
	quotaErr          error
	quotaCalled       bool
	quotaGotPublisher int64
	quotaGotCommunity int64
	quotaGotTyp       string
}

func (f *fakeLostFoundModel) UpdateModerationStatus(ctx context.Context, id int64, status int64) error {
	f.modStatusCalled = true
	f.modStatusSetTo = status
	return nil
}

func (f *fakeLostFoundModel) Insert(ctx context.Context, item *model.LostFoundItem) (int64, error) {
	f.inserted = item
	return item.Id, nil
}

func (f *fakeLostFoundModel) FindOne(ctx context.Context, id int64) (*model.LostFoundItem, error) {
	if f.findItem == nil {
		return nil, sql.ErrNoRows
	}
	return f.findItem, nil
}

func (f *fakeLostFoundModel) UpdateStatus(ctx context.Context, id int64, status string) error {
	f.updateCalled = true
	f.updatedStatus = status
	return nil
}

func (f *fakeLostFoundModel) FindList(ctx context.Context, communityId int64, typ string, offset, limit int64) ([]*model.LostFoundItem, int64, error) {
	f.findListCalled = true
	return nil, 0, nil
}

func (f *fakeLostFoundModel) CountQuotaOccupied(ctx context.Context, publisherId, communityId int64, typ string) (int64, error) {
	f.quotaCalled = true
	f.quotaGotPublisher = publisherId
	f.quotaGotCommunity = communityId
	f.quotaGotTyp = typ
	return f.quotaCount, f.quotaErr
}

func (f *fakeLostFoundModel) FindOnePublished(ctx context.Context, id int64) (*model.LostFoundItem, error) {
	if f.findPublishedErr != nil {
		return nil, f.findPublishedErr
	}
	if f.findPublishedItem != nil {
		return f.findPublishedItem, nil
	}
	// 未显式设置 published 行为时，向后兼容 fallback 到 FindOne
	return f.FindOne(ctx, id)
}

type fakeModeration struct {
	moderationv1.ModerationServiceClient
}

func (f *fakeModeration) CreateAuditLog(ctx context.Context, in *moderationv1.CreateAuditLogRequest, opts ...grpc.CallOption) (*moderationv1.CreateAuditLogResponse, error) {
	return &moderationv1.CreateAuditLogResponse{Id: 1}, nil
}

func ctxWithUserID(t *testing.T, uid int64) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("user_id", stringInt(uid)))
}

func stringInt(n int64) string {
	// 简化：仅测试用，直接 int64→str 通过 fmt
	return fmt.Sprintf("%d", n)
}

func allowAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
		},
	}
}

func denyAll() *fakePerm {
	return &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			return &permissionv1.AssertPublishScopeResponse{
				Base:    responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"),
				Allowed: false,
			}, nil
		},
	}
}

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[is-system-no-permission-shortcut]] — 身份取 metadata（JWT），不信任客户端 body
func TestCreateLostFound_ScopeDenied_WhenJWTOutOfScope(t *testing.T) {
	// 攻击者（JWT=1001）伪造 body publisher_id=999999（管理员 id）→ 必须按 JWT 自身 scope 拒绝
	var gotUserID, gotTargetID int64
	perm := &fakePerm{
		assertFn: func(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
			gotUserID = in.GetUserId()
			gotTargetID = in.GetTargets()[0].GetScopeId()
			return &permissionv1.AssertPublishScopeResponse{
				Base:    responsex.NewBaseRespWithError(60007, "目标小区超出发布者数据范围"),
				Allowed: false,
			}, nil
		},
	}
	mdl := &fakeLostFoundModel{}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   perm,
		ModerationClient:   &fakeModeration{},
		RedisClient:        redis.New("127.0.0.1:6379"),
	}

	l := NewCreateLostFoundLogic(ctxWithUserID(t, 1001), sc)
	resp, err := l.CreateLostFound(&communityv1.CreateLostFoundRequest{
		CommunityId: 200, // 攻击者无数据范围的小区 B
		Type:        communityv1.LostFoundType_LOST_FOUND_TYPE_LOST,
		Title:       "寻物",
		PublisherId: 999999, // 伪造的管理员 id，必须被忽略
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "目标小区超出发布者数据范围 → 080006")
	assert.Nil(t, mdl.inserted, "数据权限拒绝后不得落库")

	// 身份必须是 JWT（metadata）而非 body 伪造值
	assert.Equal(t, int64(1001), gotUserID, "AssertPublishScope 必须用 JWT 身份")
	assert.Equal(t, int64(200), gotTargetID, "target 必须是目标小区")
}

func TestCreateLostFound_ScopeAllowed_WithinScope(t *testing.T) {
	mdl := &fakeLostFoundModel{}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   allowAll(),
		MasterDataClient:   &fakeMasterData{}, // 默认未配置=不限
		ModerationClient:   &fakeModeration{},
		RedisClient:        redis.New("127.0.0.1:6379"),
	}

	l := NewCreateLostFoundLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateLostFound(&communityv1.CreateLostFoundRequest{
		CommunityId: 100, // 业主@A 发 A
		Type:        communityv1.LostFoundType_LOST_FOUND_TYPE_LOST,
		Title:       "寻物",
		PublisherId: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	require.NotNil(t, mdl.inserted, "数据权限允许后落库")
	assert.Equal(t, int64(100), mdl.inserted.CommunityId)
}

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录：`expected: 80007, actual: 0`（达上限未拦截，直接落库）
// SEE: [[grpc-only-comms]] — 配额经 master-data GetSectionQuota 读取，不直连 DB
func TestCreateLostFound_QuotaExceeded(t *testing.T) {
	// 板块已配置 lost_found=5 且发布者已占 5 条 → 再发被 80007 拦截，不得落库
	mdl := &fakeLostFoundModel{quotaCount: 5}
	md := &fakeMasterData{
		quotaFn: func(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error) {
			assert.Equal(t, "lost_found", in.GetSectionType())
			return &masterdatav1.GetSectionQuotaResp{Base: responsex.NewBaseResp(), Configured: true, MaxCount: 5}, nil
		},
	}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   allowAll(),
		MasterDataClient:   md,
		ModerationClient:   &fakeModeration{},
		RedisClient:        redis.New("127.0.0.1:6379"),
	}

	l := NewCreateLostFoundLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.CreateLostFound(&communityv1.CreateLostFoundRequest{
		CommunityId: 100,
		Type:        communityv1.LostFoundType_LOST_FOUND_TYPE_LOST,
		Title:       "寻物",
		PublisherId: 100,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80007), resp.GetBase().GetCode(), "达板块上限 → 080007")
	assert.Nil(t, mdl.inserted, "配额超限后不得落库")

	// 计数口径：用户×小区×板块
	assert.True(t, mdl.quotaCalled, "配额校验必须触发占配额计数")
	assert.Equal(t, int64(100), mdl.quotaGotPublisher)
	assert.Equal(t, int64(100), mdl.quotaGotCommunity)
	assert.Equal(t, "lost_found", mdl.quotaGotTyp)
}

func TestCreateLostFound_ScopeDenied_NoIdentity(t *testing.T) {
	// 无 JWT 身份注入（gRPC metadata 缺失）→ fail-closed 拒绝
	mdl := &fakeLostFoundModel{}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   allowAll(), // 即使 permission 允许，无身份也必须拒绝
		ModerationClient:   &fakeModeration{},
		RedisClient:        redis.New("127.0.0.1:6379"),
	}

	l := NewCreateLostFoundLogic(context.Background(), sc)
	resp, err := l.CreateLostFound(&communityv1.CreateLostFoundRequest{
		CommunityId: 100,
		Type:        communityv1.LostFoundType_LOST_FOUND_TYPE_LOST,
		Title:       "寻物",
		PublisherId: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode())
	assert.Nil(t, mdl.inserted)
}

func TestResolveLostFound_ScopeDenied_OutOfScope(t *testing.T) {
	mdl := &fakeLostFoundModel{findItem: &model.LostFoundItem{Id: 1, CommunityId: 200}}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   denyAll(),
	}

	l := NewResolveLostFoundLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.ResolveLostFound(&communityv1.ResolveLostFoundRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "owner@A 解决 B 内容 → 080006")
	assert.False(t, mdl.updateCalled, "数据权限拒绝后不得更新状态")
}

func TestResolveLostFound_ScopeAllowed(t *testing.T) {
	mdl := &fakeLostFoundModel{findItem: &model.LostFoundItem{Id: 1, CommunityId: 100}}
	sc := &svc.ServiceContext{
		LostFoundItemModel: mdl,
		PermissionClient:   allowAll(),
	}

	l := NewResolveLostFoundLogic(ctxWithUserID(t, 100), sc)
	resp, err := l.ResolveLostFound(&communityv1.ResolveLostFoundRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode())
	assert.True(t, mdl.updateCalled, "数据权限允许后更新状态")
	assert.Equal(t, "resolved", mdl.updatedStatus)
}

// --- T4.5 moderation 回调 ---

// capturingPerm 记录 AssertPublishScope 收到的身份，供系统身份断言
type capturingPerm struct {
	fakePerm
	gotUserID   int64
	gotTargetID int64
}

func (c *capturingPerm) AssertPublishScope(ctx context.Context, in *permissionv1.AssertPublishScopeRequest, opts ...grpc.CallOption) (*permissionv1.AssertPublishScopeResponse, error) {
	c.gotUserID = in.GetUserId()
	c.gotTargetID = in.GetTargets()[0].GetScopeId()
	return &permissionv1.AssertPublishScopeResponse{Base: responsex.NewBaseResp(), Allowed: true}, nil
}

// SEE: [[is-system-no-permission-shortcut]] — moderation 系统身份走 grant 判定
func TestUpdateLostFoundModerationStatus_ContentExists_SystemIdentityAllowed(t *testing.T) {
	mdl := &fakeLostFoundModel{findItem: &model.LostFoundItem{Id: 1, CommunityId: 100}}
	perm := &capturingPerm{}
	sc := &svc.ServiceContext{LostFoundItemModel: mdl, PermissionClient: perm}

	l := NewUpdateLostFoundModerationStatusLogic(context.Background(), sc)
	resp, err := l.UpdateLostFoundModerationStatus(&communityv1.UpdateModerationStatusRequest{
		Id:               1,
		ModerationStatus: 3,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode(), "服务回调存在内容 → 放行")
	assert.True(t, mdl.modStatusCalled, "放行后更新审核状态")
	assert.Equal(t, int64(3), mdl.modStatusSetTo)
	assert.Equal(t, int64(0), perm.gotUserID, "服务回调必须以 system_user_id=0 校验")
	assert.Equal(t, int64(100), perm.gotTargetID)
}

func TestUpdateLostFoundModerationStatus_ContentMissing_Rejected(t *testing.T) {
	mdl := &fakeLostFoundModel{} // findItem=nil
	perm := &capturingPerm{}
	sc := &svc.ServiceContext{LostFoundItemModel: mdl, PermissionClient: perm}

	l := NewUpdateLostFoundModerationStatusLogic(context.Background(), sc)
	resp, err := l.UpdateLostFoundModerationStatus(&communityv1.UpdateModerationStatusRequest{
		Id:               999,
		ModerationStatus: 3,
	})
	require.NoError(t, err)
	assert.NotEqual(t, int32(0), resp.GetBase().GetCode(), "内容不存在 → 拒绝")
	assert.False(t, mdl.modStatusCalled)
	assert.Zero(t, perm.gotUserID, "内容不存在时不得发起数据权限校验")
}
