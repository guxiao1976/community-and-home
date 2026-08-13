package notice

import (
	"context"
	"database/sql"
	"testing"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAttachmentModel struct {
	model.NoticeAttachmentModel
	findCalled bool
}

func (f *fakeAttachmentModel) FindByNoticeId(ctx context.Context, noticeId int64) ([]*model.NoticeAttachment, error) {
	f.findCalled = true
	return nil, nil
}

// 越权 Get 回归（评审 CRITICAL：T4.6 只覆盖 List，Get-by-ID 漏网 → 任意已登录用户按 ID 越权读取）
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 行为型断言 RED 摘录留档
// SEE: [[testing-discipline]] — 同一读路径（Get 与 List）控制必须一致（REQ-1.6）
func TestGetNotice_FilterByScope(t *testing.T) {
	tests := []struct {
		name        string
		perm        *fakeListPerm
		userID      int64 // 0 → 不注入身份（fail-closed 路径）
		noticeComID int64 // 内容所在小区
		wantAllowed bool  // true → 返回内容；false → 080006 + 空
	}{
		{
			name:        "LIMITED 不含目标小区 → 080006 拒绝（不返回越权内容）",
			perm:        listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      42,
			noticeComID: 200,
			wantAllowed: false,
		},
		{
			name:        "EMPTY 空范围 → 080006 拒绝",
			perm:        listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_EMPTY),
			userID:      42,
			noticeComID: 100,
			wantAllowed: false,
		},
		{
			name:        "无身份（metadata 缺失）→ fail-closed 拒绝",
			perm:        listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      0,
			noticeComID: 100,
			wantAllowed: false,
		},
		{
			name:        "LIMITED 含目标小区 → 返回内容",
			perm:        listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_LIMITED, 100),
			userID:      42,
			noticeComID: 100,
			wantAllowed: true,
		},
		{
			name:        "GLOBAL → 返回内容",
			perm:        listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL),
			userID:      42,
			noticeComID: 999,
			wantAllowed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mdl := &fakeNoticeModel{findItem: noticeItem(1, tc.noticeComID)}
			att := &fakeAttachmentModel{}
			sc := &svc.ServiceContext{NoticeModel: mdl, PermissionClient: tc.perm, NoticeAttachmentModel: att}

			ctx := context.Background()
			if tc.userID != 0 {
				ctx = noticeCtxWithUserID(t, tc.userID)
			}
			l := NewGetNoticeLogic(ctx, sc)
			resp, err := l.GetNotice(&communityv1.GetNoticeRequest{Id: 1})
			require.NoError(t, err)

			if tc.wantAllowed {
				assert.Equal(t, int32(0), resp.GetBase().GetCode())
				require.NotNil(t, resp.GetNotice(), "允许读取时返回内容")
				assert.True(t, att.findCalled, "允许读取时查询附件")
			} else {
				assert.Equal(t, int32(80006), resp.GetBase().GetCode(), "越权 Get → 080006")
				assert.Nil(t, resp.GetNotice(), "拒绝时不得返回内容")
				assert.False(t, att.findCalled, "拒绝时不得查询附件")
			}
		})
	}
}

// 审核可见性门禁（最小实现）：Get 读路径仅返回 moderation_status=通过 的内容。
// FindOne 仍被写接口 / 审核回调使用（不过滤），读路径改用 FindOnePublished。
// 本测试证明 GetNotice 走 FindOnePublished：内容 status=0（待审核）时 FindOne 能查到、
// 但 FindOnePublished 过滤后 not found → 返回「不存在」。
//
// SEE: [[tdd-red-evidence-requires-fail-excerpt]]
func TestGetNotice_FilterByModerationStatus(t *testing.T) {
	mdl := &fakeNoticeModel{
		findItem:         noticeItem(1, 100), // FindOne 不过滤：可查到（ModerationStatus 默认 0）
		findPublishedErr: sql.ErrNoRows,      // FindOnePublished 过滤后：待审核 → not found
	}
	sc := &svc.ServiceContext{
		NoticeModel:           mdl,
		PermissionClient:      listPerm(permissionv1.DataScopeState_DATA_SCOPE_STATE_GLOBAL),
		NoticeAttachmentModel: &fakeAttachmentModel{},
	}

	l := NewGetNoticeLogic(noticeCtxWithUserID(t, 42), sc)
	resp, err := l.GetNotice(&communityv1.GetNoticeRequest{Id: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(80001), resp.GetBase().GetCode(), "待审核内容对读路径不可见 → 80001 不存在")
	assert.Nil(t, resp.GetNotice(), "拒绝时不得返回内容")
}
