// Package kafkapush 承载 content-review Kafka 推送（at-least-once 待推标记 + 定时重推）。
package kafkapush

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/config"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// ContentReviewVersion 消息契约版本（变更即 bump，供未来消费者协商，REVISION）。
const ContentReviewVersion = 1

// ContentReviewMessage content-review 消息契约（REQ-CPM-2 单源权威 payload 定义）。
// 附件 file_url 为可再生预签名 URL（发布时经 GetFileUrl 生成，勿当永久链接）。
// 无附件帖推 `attachments: []`（空数组非 null）。
// 消费者须按 post_id 为幂等键去重（并发 submit / 重推扫描可产生同 post_id 重复消息，at-least-once 容忍）。
type ContentReviewMessage struct {
	Version     int32                     `json:"version"`
	PostID      string                    `json:"post_id"`
	SectionCode string                    `json:"section_code"`
	Text        string                    `json:"text"`
	PublisherID string                    `json:"publisher_id"`
	Attachments []ContentReviewAttachment `json:"attachments"`
}

// ContentReviewAttachment 附件级审核快照（review_status 为推送时刻默认值 approved，非审核结论）。
type ContentReviewAttachment struct {
	FileID       string `json:"file_id"`
	FileType     string `json:"file_type"`
	ReviewStatus int32  `json:"review_status"`
	FileURL      string `json:"file_url"`
}

// Writer 可 mock 的 kafka Writer 接口（测试注入）。
type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// Pusher 发布侧推送抽象（ServiceContext.KafkaProducer 以接口承载，供逻辑测试注入 mock）。
// *Producer 实现该接口。
type Pusher interface {
	Push(ctx context.Context, post *model.ContentPost, attachments []*model.ContentPostAttachment) error
}

// Producer content-review 生产者：事务提交成功后推送（先提交后推送，提交失败不推送）。
// 成功 → UpdateKafkaPushStatus(2, pushed_at)；失败 → 置 pending(1) + last_error 落库（D20，不阻塞发布）。
//
// SEE: [[best-effort-compensation-must-log]] — 推送失败补偿不可静默丢弃，last_error 落库 + 日志
type Producer struct {
	writer     Writer
	topic      string
	postModel  model.ContentPostModel
	fileClient filev1.FileServiceClient
}

// NewProducer 以真实 kafka-go Writer 初始化（acks=all，at-least-once 语义）。
func NewProducer(cfg config.KafkaConf, postModel model.ContentPostModel, fileClient filev1.FileServiceClient) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	return &Producer{writer: writer, topic: cfg.Topic, postModel: postModel, fileClient: fileClient}
}

// newProducerWithWriter 测试注入可 mock Writer。
func newProducerWithWriter(w Writer, topic string, m model.ContentPostModel, fc filev1.FileServiceClient) *Producer {
	return &Producer{writer: w, topic: topic, postModel: m, fileClient: fc}
}

// pushTimeout 推送独立超时：Kafka 不可用时 kafka-go WriteMessages 重试不得耗尽 RPC 请求
// deadline 导致客户端收到 DeadlineExceeded——帖已提交成功，推送是尽力而为（D20）。
// 网络写/URL 再生走 pushCtx；DB 落标（markPending/ack）走请求 ctx（本地写，实时性不受限）。
const pushTimeout = 3 * time.Second

// Push 推送 content-review 消息（含附件 file_url 经 GetFileUrl 重生）。
func (p *Producer) Push(ctx context.Context, post *model.ContentPost, attachments []*model.ContentPostAttachment) error {
	// 不阻塞发布（D20）：脱离请求 deadline 用独立短超时——Kafka 不可用时 3s 内快速失败 → markPending 待重推
	pushCtx, cancel := context.WithTimeout(context.Background(), pushTimeout)
	defer cancel()

	msg := buildMessage(post, attachments)
	// 重生附件预签名 URL（file_id 权威载体；兼容期 file_id=0 回退 stored file_url）
	for i := range msg.Attachments {
		if i < len(attachments) && attachments[i].FileId > 0 {
			if url, err := p.regenerateURL(pushCtx, attachments[i].FileId); err == nil && url != "" {
				msg.Attachments[i].FileURL = url
			}
		}
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		p.markPending(ctx, post, err)
		return err
	}

	if err := p.writer.WriteMessages(pushCtx, kafka.Message{Value: payload}); err != nil {
		logx.WithContext(ctx).Errorf("kafkapush: push content-review failed post=%d: %v", post.Id, err)
		p.markPending(ctx, post, err)
		return err
	}

	// 成功 ack
	if err := p.postModel.UpdateKafkaPushStatus(ctx, post.Id, model.KafkaPushDone, post.KafkaPushRetries,
		sql.NullString{}, sql.NullTime{Valid: true, Time: time.Now()}); err != nil {
		logx.WithContext(ctx).Errorf("kafkapush: ack kafka_push_status failed post=%d: %v", post.Id, err)
	}
	logx.WithContext(ctx).Infof("kafkapush: content-review pushed post=%d", post.Id)
	return nil
}

// markPending 推送失败：保留 pending + retries+1 + last_error 落库（不静默丢弃，D20）。
func (p *Producer) markPending(ctx context.Context, post *model.ContentPost, pushErr error) {
	if err := p.postModel.UpdateKafkaPushStatus(ctx, post.Id, model.KafkaPushPending, post.KafkaPushRetries+1,
		sql.NullString{String: pushErr.Error(), Valid: true}, sql.NullTime{}); err != nil {
		logx.WithContext(ctx).Errorf("kafkapush: mark pending failed post=%d: %v", post.Id, err)
	}
}

func (p *Producer) regenerateURL(ctx context.Context, fileID int64) (string, error) {
	resp, err := p.fileClient.GetFileUrl(ctx, &filev1.GetFileUrlRequest{FileId: fileID})
	if err != nil {
		return "", err
	}
	return resp.GetDownloadUrl(), nil
}

// buildMessage 组装 content-review 契约（version/post_id string/section_code/text/publisher_id/attachments）。
func buildMessage(post *model.ContentPost, attachments []*model.ContentPostAttachment) ContentReviewMessage {
	msg := ContentReviewMessage{
		Version:     ContentReviewVersion,
		PostID:      int64String(post.Id),
		SectionCode: post.SectionCode,
		Text:        post.Text,
		PublisherID: int64String(derefInt64(post.PublisherId)),
		Attachments: []ContentReviewAttachment{}, // 空数组非 null
	}
	for _, a := range attachments {
		msg.Attachments = append(msg.Attachments, ContentReviewAttachment{
			FileID:       int64String(a.FileId),
			FileType:     derefStr(a.FileType),
			ReviewStatus: int32(a.ReviewStatus),
			FileURL:      a.FileUrl,
		})
	}
	return msg
}

func int64String(v int64) string {
	return strconv.FormatInt(v, 10)
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
