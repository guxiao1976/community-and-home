package kafkapush

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeWriter 记录写入的消息 + 注入失败。
type fakeWriter struct {
	msgs  []kafka.Message
	fail  bool
	failN int // 前 N 次失败
}

func (w *fakeWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if w.fail || w.failN > 0 {
		if w.failN > 0 {
			w.failN--
		}
		return errors.New("broker unreachable")
	}
	w.msgs = append(w.msgs, msgs...)
	return nil
}

func (w *fakeWriter) Close() error { return nil }

// fakePostModel 覆盖 UpdateKafkaPushStatus + FindPendingPush。
type fakePostModel struct {
	model.ContentPostModel
	pushStatusCalls []int64
	retriesCalls    []int64
	lastErr         []sql.NullString
}

func (f *fakePostModel) UpdateKafkaPushStatus(ctx context.Context, id int64, pushStatus, retries int64, lastErr sql.NullString, pushedAt sql.NullTime) error {
	f.pushStatusCalls = append(f.pushStatusCalls, pushStatus)
	f.retriesCalls = append(f.retriesCalls, retries)
	f.lastErr = append(f.lastErr, lastErr)
	return nil
}

type fakeFile struct {
	filev1.FileServiceClient
	url string
}

func (f *fakeFile) GetFileUrl(ctx context.Context, in *filev1.GetFileUrlRequest, opts ...grpc.CallOption) (*filev1.GetFileUrlResponse, error) {
	return &filev1.GetFileUrlResponse{Base: responsex.NewBaseResp(), DownloadUrl: f.url, File: &filev1.FileInfo{Id: in.FileId}}, nil
}

func testPost(id int64) *model.ContentPost {
	pubID := int64(456)
	return &model.ContentPost{
		Id: id, SectionCode: "notice", Text: "正文内容", PublisherId: &pubID,
		KafkaPushRetries: 0, Status: model.StatusApproved,
	}
}

func testAttachments() []*model.ContentPostAttachment {
	ft := "pdf"
	return []*model.ContentPostAttachment{
		{Id: 11, PostId: 1001, FileName: "a.pdf", FileSize: 1024, ReviewStatus: model.AttachmentReviewApproved, FileId: 5001, FileType: &ft, FileUrl: "stored-url"},
	}
}

// TestProducer_Push_Success 成功推送后 UpdateKafkaPushStatus(2) + 契约字段断言。
func TestProducer_Push_Success(t *testing.T) {
	w := &fakeWriter{}
	pm := &fakePostModel{}
	fc := &fakeFile{url: "https://presigned/regen"}
	p := newProducerWithWriter(w, "content-review", pm, fc)

	post := testPost(1001)
	err := p.Push(context.Background(), post, testAttachments())
	require.NoError(t, err)

	require.Len(t, w.msgs, 1)
	require.Len(t, pm.pushStatusCalls, 1)
	assert.Equal(t, int64(model.KafkaPushDone), pm.pushStatusCalls[0], "成功 ack 置 2")

	// 契约字段断言（REQ-CPM-2）
	msg := decodeMessage(t, w.msgs[0].Value)
	assert.Equal(t, int32(ContentReviewVersion), msg.Version)
	assert.Equal(t, "1001", msg.PostID)
	assert.Equal(t, "notice", msg.SectionCode)
	assert.Equal(t, "正文内容", msg.Text)
	assert.Equal(t, "456", msg.PublisherID)
	require.Len(t, msg.Attachments, 1)
	assert.Equal(t, "5001", msg.Attachments[0].FileID)
	assert.Equal(t, "pdf", msg.Attachments[0].FileType)
	assert.Equal(t, int32(model.AttachmentReviewApproved), msg.Attachments[0].ReviewStatus)
	// file_url 为 GetFileUrl 重生后的新预签名 URL（评审 I8）
	assert.Equal(t, "https://presigned/regen", msg.Attachments[0].FileURL)
}

// TestProducer_Push_NoAttachments 无附件帖推 attachments: []（空数组非 null）。
func TestProducer_Push_NoAttachments(t *testing.T) {
	w := &fakeWriter{}
	pm := &fakePostModel{}
	p := newProducerWithWriter(w, "content-review", pm, &fakeFile{})

	err := p.Push(context.Background(), testPost(1001), nil)
	require.NoError(t, err)

	msg := decodeMessage(t, w.msgs[0].Value)
	require.NotNil(t, msg.Attachments, "attachments 必须为 [] 而非 null")
	assert.Empty(t, msg.Attachments)
}

// TestProducer_Push_Failure 失败后置 pending + retries+1 + last_error 记录。
func TestProducer_Push_Failure(t *testing.T) {
	w := &fakeWriter{fail: true}
	pm := &fakePostModel{}
	p := newProducerWithWriter(w, "content-review", pm, &fakeFile{})

	err := p.Push(context.Background(), testPost(1001), nil)
	require.Error(t, err)

	require.Len(t, pm.pushStatusCalls, 1)
	assert.Equal(t, int64(model.KafkaPushPending), pm.pushStatusCalls[0], "失败保留 pending")
	assert.Equal(t, int64(1), pm.retriesCalls[0], "retries+1")
	assert.True(t, pm.lastErr[0].Valid, "last_error 落库")
	assert.Contains(t, pm.lastErr[0].String, "broker unreachable")
}

// TestProducer_Push_NoPanic 空附件/空文件信息 push 不 panic。
func TestProducer_Push_NoPanic(t *testing.T) {
	w := &fakeWriter{}
	pm := &fakePostModel{}
	p := newProducerWithWriter(w, "content-review", pm, &fakeFile{})

	post := testPost(1001)
	post.PublisherId = nil
	require.NotPanics(t, func() {
		_ = p.Push(context.Background(), post, nil)
	})
	require.Len(t, w.msgs, 1)
}

func decodeMessage(t *testing.T, raw []byte) ContentReviewMessage {
	t.Helper()
	var m ContentReviewMessage
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}
