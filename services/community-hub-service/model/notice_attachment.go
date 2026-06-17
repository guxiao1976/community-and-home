package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// NoticeAttachment 通知附件
type NoticeAttachment struct {
	Id        int64     `db:"id"`
	NoticeId  int64     `db:"notice_id"`
	FileName  string    `db:"file_name"`
	FileUrl   string    `db:"file_url"`
	FileSize  int64     `db:"file_size"`
	CreatedAt time.Time `db:"created_at"`
}

// NoticeAttachmentModel 通知附件数据访问层
type NoticeAttachmentModel interface {
	Insert(ctx context.Context, a *NoticeAttachment) (int64, error)
	FindByNoticeId(ctx context.Context, noticeId int64) ([]*NoticeAttachment, error)
}

type defaultNoticeAttachmentModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewNoticeAttachmentModel(conn sqlx.SqlConn) NoticeAttachmentModel {
	return &defaultNoticeAttachmentModel{conn: conn, table: "`notice_attachments`"}
}

func (m *defaultNoticeAttachmentModel) Insert(ctx context.Context, a *NoticeAttachment) (int64, error) {
	query := `insert into ` + m.table + `
		(id, notice_id, file_name, file_url, file_size)
		values (?, ?, ?, ?, ?)`
	res, err := m.conn.ExecCtx(ctx, query, a.Id, a.NoticeId, a.FileName, a.FileUrl, a.FileSize)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultNoticeAttachmentModel) FindByNoticeId(ctx context.Context, noticeId int64) ([]*NoticeAttachment, error) {
	var list []*NoticeAttachment
	query := `select * from ` + m.table + ` where notice_id = ?`
	err := m.conn.QueryRowsCtx(ctx, &list, query, noticeId)
	return list, err
}

var _ = time.Now
