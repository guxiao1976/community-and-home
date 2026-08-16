package kafkapush

import (
	"context"
	"time"

	"github.com/guxiao1976/community-hub/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// MaxPushRetries 重推超阈值后保留 pending + 日志（quarantine 语义，BACKLOG 回填「消费者上线前
// pending-push 积压处置」）；不静默丢弃，pending-push 计数可观测。
const MaxPushRetries = 3

// Rescanner Kafka 定时重推扫描器（at-least-once，Task 1.19）。
// 复用 Producer.Push（内含 GetFileUrl 重生 file_url，避免重推携带过期预签名 URL，评审 I8）。
type Rescanner struct {
	producer         *Producer
	postModel        model.ContentPostModel
	attachmentModel  model.ContentPostAttachmentModel
	interval         time.Duration
	maxRetries       int64
	lastPendingCount int64 // 可观测：最近一次扫描的 pending-push 计数
}

// NewRescanner 初始化重推扫描器（interval 默认 60s，maxRetries 默认 MaxPushRetries）。
func NewRescanner(producer *Producer, postModel model.ContentPostModel, attachmentModel model.ContentPostAttachmentModel, interval time.Duration) *Rescanner {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	return &Rescanner{
		producer:        producer,
		postModel:       postModel,
		attachmentModel: attachmentModel,
		interval:        interval,
		maxRetries:      MaxPushRetries,
	}
}

// Start 启动定时扫描 goroutine（随 servicecontext 接线）。ctx 取消即停止。
func (r *Rescanner) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logx.WithContext(ctx).Info("kafkapush: rescanner stopped")
				return
			case <-ticker.C:
				if err := r.ScanOnce(ctx); err != nil {
					logx.WithContext(ctx).Errorf("kafkapush: rescanner scan error: %v", err)
				}
			}
		}
	}()
}

// ScanOnce 单轮扫描：FindPendingPush → 逐条复用 Producer.Push（含重生 file_url）→ ack 置 2。
func (r *Rescanner) ScanOnce(ctx context.Context) error {
	posts, err := r.postModel.FindPendingPush(ctx, 100)
	if err != nil {
		return err
	}
	r.lastPendingCount = int64(len(posts))
	for _, p := range posts {
		attachments, err := r.attachmentModel.FindByPostId(ctx, p.Id)
		if err != nil {
			logx.WithContext(ctx).Errorf("kafkapush: rescanner find attachments failed post=%d: %v", p.Id, err)
			continue
		}
		if err := r.producer.Push(ctx, p, attachments); err != nil {
			// Producer.Push 内部已置 pending + retries+1 + last_error；超阈值保留 pending + 日志
			if p.KafkaPushRetries+1 >= r.maxRetries {
				logx.WithContext(ctx).Errorf("kafkapush: rescanner quarantine post=%d retries=%d: %v", p.Id, p.KafkaPushRetries+1, err)
			}
			continue
		}
	}
	return nil
}

// PendingCount 可观测：最近一次扫描的 pending-push 计数。
func (r *Rescanner) PendingCount() int64 {
	return r.lastPendingCount
}
