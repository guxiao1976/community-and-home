package notice

import (
	"testing"
	"time"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 1.4 ListContentPosts — since_days 校验 + 窗口传参（REQ-NTW-2 / D10 / r2-5）
// =============================================================================
//
// 表驱动 5 用例：
//   - since_days=30 → fake 捕获窗口选项，下界 ≈ now-30d（WithinDuration 容忍时钟差）；
//   - since_days=0（缺省）→ 无窗口选项（PC 管理列表行为不变，REQ-NTW-2 场景 2）；
//   - since_days=-1 → Base code 080005（非法下界，REQ-NTW-2 场景 3）；
//   - since_days=366 → Base code 080005（非法上界）；
//   - since_days=365 → 边界合法，窗口选项存在。
//
// 服务端仅校验数值范围（r2-5：int32 wire 恒数字，非数字由 REST 网关解析层拒绝）。
// SEE: [[error-code-collision-and-namespace-alignment]] — 复用 080005 不新增码
func TestListContentPosts_SinceDays(t *testing.T) {
	cases := []struct {
		name          string
		sinceDays     int32
		wantBaseCode  int32
		wantWindow    bool
		wantSinceDays int
	}{
		{"since_days=30 窗口传参", 30, 0, true, 30},
		{"since_days=0 缺省不过滤", 0, 0, false, 0},
		{"since_days=-1 非法下界", -1, 80005, false, 0},
		{"since_days=366 非法上界", 366, 80005, false, 0},
		{"since_days=365 边界合法", 365, 0, true, 365},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fm := &fakeContentPostModel{listItems: []*model.ContentPost{approvedPost(1001, 100)}}
			sc := noticeSvcCtx(fm, &fakeScopeModel{}, &fakeAttachmentModel{}, permAllowAll(), &fakeMD{}, &fakeFile{}, &fakeUser{}, &fakePusher{})
			l := NewListContentPostsLogic(ctxWithUserID(t, 100), sc)

			resp, err := l.ListContentPosts(&communityv1.ListContentPostsRequest{
				CommunityId: 2001,
				Page:        1,
				PageSize:    10,
				SinceDays:   tc.sinceDays,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)

			if tc.wantBaseCode != 0 {
				assert.Equal(t, tc.wantBaseCode, resp.GetBase().GetCode(), "非法 since_days → 080005")
				assert.Empty(t, resp.GetPosts())
				assert.Equal(t, int64(0), resp.GetTotal())
				return
			}

			assert.Equal(t, int32(0), resp.GetBase().GetCode(), "合法 since_days → 成功 Base")
			if tc.wantWindow {
				require.NotEmpty(t, fm.listOpts, "since_days>0 → 必须传窗口选项")
				since := model.ContentPostListOptionSince(fm.listOpts...)
				require.NotNil(t, since, "窗口选项必须含下界")
				assert.WithinDuration(t, time.Now().AddDate(0, 0, -tc.wantSinceDays), *since, 2*time.Minute,
					"下界 ≈ now-%dd", tc.wantSinceDays)
			} else {
				assert.Nil(t, model.ContentPostListOptionSince(fm.listOpts...), "since_days=0 → 无窗口选项")
			}
		})
	}
}
