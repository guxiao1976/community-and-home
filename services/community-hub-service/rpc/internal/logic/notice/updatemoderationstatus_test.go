package notice

import (
	"context"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/scope"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

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

// SEE: [[is-system-no-permission-shortcut]] — moderation 系统身份走 grant 判定，不按作者 scope
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
func TestUpdateNoticeModerationStatus_ContentExists_SystemIdentityAllowed(t *testing.T) {
	mdl := &fakeNoticeModel{findItem: noticeItem(1, 100)}
	perm := &capturingPerm{}
	sc := &svc.ServiceContext{NoticeModel: mdl, PermissionClient: perm}

	// 服务间回调：无用户 JWT，但 reverse-lookup 找到内容
	l := NewUpdateNoticeModerationStatusLogic(context.Background(), sc)
	resp, err := l.UpdateNoticeModerationStatus(&communityv1.UpdateModerationStatusRequest{
		Id:               1,
		ModerationStatus: 3, // human_pass
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.GetBase().GetCode(), "服务回调存在内容 → 放行")
	assert.True(t, mdl.modStatusCalled, "放行后更新审核状态")
	assert.Equal(t, int64(3), mdl.modStatusSetTo)

	// 必须以系统身份（system_user_id=0）校验，target 为内容 community_id（不按作者 scope）
	assert.Equal(t, scope.SystemUserID, perm.gotUserID, "服务回调必须以 system_user_id=0 校验")
	assert.Equal(t, int64(100), perm.gotTargetID)
}

func TestUpdateNoticeModerationStatus_ContentMissing_Rejected(t *testing.T) {
	mdl := &fakeNoticeModel{} // findItem=nil → reverse-lookup 失败
	perm := &capturingPerm{}
	sc := &svc.ServiceContext{NoticeModel: mdl, PermissionClient: perm}

	l := NewUpdateNoticeModerationStatusLogic(context.Background(), sc)
	resp, err := l.UpdateNoticeModerationStatus(&communityv1.UpdateModerationStatusRequest{
		Id:               999,
		ModerationStatus: 3,
	})
	require.NoError(t, err)
	assert.NotEqual(t, int32(0), resp.GetBase().GetCode(), "内容不存在 → 拒绝")
	assert.False(t, mdl.modStatusCalled, "内容不存在时不得更新审核状态")
	assert.Zero(t, perm.gotUserID, "内容不存在时不得发起数据权限校验（先 reverse-lookup）")
}
