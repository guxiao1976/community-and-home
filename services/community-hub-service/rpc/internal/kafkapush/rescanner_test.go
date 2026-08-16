package kafkapush

import (
	"context"
	"errors"
	"testing"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePendingModel 覆盖 FindPendingPush（返回待推帖）并委托 baseModel 的 ack 记录。
type fakePendingModel struct {
	*fakePostModel
	pendingPosts []*model.ContentPost
	findErr      error
}

func (f *fakePendingModel) FindPendingPush(ctx context.Context, limit int64) ([]*model.ContentPost, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.pendingPosts, nil
}

// fakeAttachmentModel 覆盖 FindByPostId。
type fakeAttachmentModel struct {
	model.ContentPostAttachmentModel
	atts []*model.ContentPostAttachment
	err  error
}

func (f *fakeAttachmentModel) FindByPostId(ctx context.Context, postId int64) ([]*model.ContentPostAttachment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.atts, nil
}

func TestRescanner_ScanOnce_PendingPushedAndAcked(t *testing.T) {
	w := &fakeWriter{}
	pm := &fakePostModel{}
	base := &fakePendingModel{fakePostModel: pm, pendingPosts: []*model.ContentPost{testPost(1001), testPost(1002)}}
	am := &fakeAttachmentModel{}
	r := NewRescanner(newProducerWithWriter(w, "content-review", base, &fakeFile{url: "regenerated"}), base, am, 60)

	err := r.ScanOnce(context.Background())
	require.NoError(t, err)

	// pending→重推→ack 置 2（两条各成功）
	assert.Len(t, w.msgs, 2)
	assert.Equal(t, []int64{model.KafkaPushDone, model.KafkaPushDone}, pm.pushStatusCalls)
	assert.Equal(t, int64(2), r.PendingCount())
}

func TestRescanner_ScanOnce_PushFailureKeepsPending(t *testing.T) {
	w := &fakeWriter{fail: true}
	pm := &fakePostModel{}
	base := &fakePendingModel{fakePostModel: pm, pendingPosts: []*model.ContentPost{testPost(1001)}}
	am := &fakeAttachmentModel{}
	r := NewRescanner(newProducerWithWriter(w, "content-review", base, &fakeFile{}), base, am, 60)

	err := r.ScanOnce(context.Background())
	require.NoError(t, err)

	// 失败 → 保留 pending(1) + retries+1 + last_error
	assert.Equal(t, []int64{model.KafkaPushPending}, pm.pushStatusCalls)
	assert.Equal(t, []int64{1}, pm.retriesCalls)
	assert.True(t, pm.lastErr[0].Valid)
}

func TestRescanner_ScanOnce_ExceedThresholdKeepsPending(t *testing.T) {
	w := &fakeWriter{fail: true}
	pm := &fakePostModel{}
	post := testPost(1001)
	post.KafkaPushRetries = 3 // 已达阈值
	base := &fakePendingModel{fakePostModel: pm, pendingPosts: []*model.ContentPost{post}}
	am := &fakeAttachmentModel{}
	r := NewRescanner(newProducerWithWriter(w, "content-review", base, &fakeFile{}), base, am, 60)
	r.maxRetries = MaxPushRetries

	err := r.ScanOnce(context.Background())
	require.NoError(t, err)
	// 超阈值仍保留 pending（不静默丢弃，log quarantine）
	assert.Equal(t, []int64{model.KafkaPushPending}, pm.pushStatusCalls)
}

func TestRescanner_ScanOnce_FindErrorPropagates(t *testing.T) {
	base := &fakePendingModel{findErr: errors.New("db unavailable")}
	r := NewRescanner(nil, base, &fakeAttachmentModel{}, 60)
	err := r.ScanOnce(context.Background())
	require.Error(t, err)
}

var _ = filev1.FileInfo{}
