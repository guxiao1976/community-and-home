package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 附件级审核状态（对齐 Migration 003 content_post_attachments.review_status 列：0=pending 1=approved 2=rejected）。
// 与 ContentPost.Status 枚举不同域——本枚举为附件级审核结论，draft/withdrawn 等正文态不适用。
const (
	AttachmentReviewPending  int64 = 0
	AttachmentReviewApproved int64 = 1 // 本期默认 approved（D14）
	AttachmentReviewRejected int64 = 2
)

// ContentPostAttachment 内容帖附件（原 notice_attachments RENAME，Migration 003）。
// file_url 为存量回退用 stored URL（新行写占位空串 ”，file_id 为权威重生载体，D14/REUSE:notice-D24）。
type ContentPostAttachment struct {
	Id           int64     `db:"id"`
	PostId       int64     `db:"post_id"` // 关联 content_posts.id（原 notice_id 改名，post_id 全链一致）
	FileName     string    `db:"file_name"`
	FileUrl      string    `db:"file_url"`
	FileSize     int64     `db:"file_size"`
	ReviewStatus int64     `db:"review_status"` // 附件级审核：0=pending 1=approved 2=rejected（本期默认 approved，D14）
	FileId       int64     `db:"file_id"`       // file-service 文件ID（重生预签名 URL 权威载体）；兼容期存量行 0
	FileType     *string   `db:"file_type"`     // 白名单校验通过的文件类型（扩展名，自 FileInfo 回读）
	CreatedAt    time.Time `db:"created_at"`
}

// ContentPostAttachmentModel 内容帖附件数据访问层
type ContentPostAttachmentModel interface {
	// InsertBatch 批量插入（Create 单事务落库）。
	InsertBatch(ctx context.Context, attachments []*ContentPostAttachment) error
	// InsertBatchTx 事务内批量插入（Create 单事务落库经共享 session）。
	InsertBatchTx(ctx context.Context, session sqlx.Session, attachments []*ContentPostAttachment) error
	// FindByPostId 按帖查附件（详情组装）。
	FindByPostId(ctx context.Context, postId int64) ([]*ContentPostAttachment, error)
	// DeleteByPostId 按帖删除（draft 编辑附件集合全量重写）。
	DeleteByPostId(ctx context.Context, postId int64) error
	// DeleteByPostIdTx 事务内按帖删除（draft 编辑附件重写经共享 session）。
	DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error
}

type defaultContentPostAttachmentModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewContentPostAttachmentModel(conn sqlx.SqlConn) ContentPostAttachmentModel {
	return &defaultContentPostAttachmentModel{conn: conn, table: "`content_post_attachments`"}
}

func (m *defaultContentPostAttachmentModel) InsertBatch(ctx context.Context, attachments []*ContentPostAttachment) error {
	return m.insertBatch(ctx, m.conn, attachments)
}

func (m *defaultContentPostAttachmentModel) InsertBatchTx(ctx context.Context, session sqlx.Session, attachments []*ContentPostAttachment) error {
	return m.insertBatch(ctx, session, attachments)
}

func (m *defaultContentPostAttachmentModel) insertBatch(ctx context.Context, e execer, attachments []*ContentPostAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	query := `insert into ` + m.table + `
		(id, post_id, file_name, file_url, file_size, review_status, file_id, file_type)
		values (?, ?, ?, ?, ?, ?, ?, ?)`
	for _, a := range attachments {
		var fileType interface{}
		if a.FileType != nil {
			fileType = *a.FileType
		}
		if _, err := e.ExecCtx(ctx, query, a.Id, a.PostId, a.FileName, a.FileUrl, a.FileSize, a.ReviewStatus, a.FileId, fileType); err != nil {
			return err
		}
	}
	return nil
}

func (m *defaultContentPostAttachmentModel) FindByPostId(ctx context.Context, postId int64) ([]*ContentPostAttachment, error) {
	var list []*ContentPostAttachment
	query := `select * from ` + m.table + ` where post_id = ?`
	err := m.conn.QueryRowsCtx(ctx, &list, query, postId)
	return list, err
}

func (m *defaultContentPostAttachmentModel) DeleteByPostId(ctx context.Context, postId int64) error {
	return m.deleteByPostId(ctx, m.conn, postId)
}

func (m *defaultContentPostAttachmentModel) DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error {
	return m.deleteByPostId(ctx, session, postId)
}

func (m *defaultContentPostAttachmentModel) deleteByPostId(ctx context.Context, e execer, postId int64) error {
	query := `delete from ` + m.table + ` where post_id = ?`
	_, err := e.ExecCtx(ctx, query, postId)
	return err
}

var _ = sql.ErrNoRows
