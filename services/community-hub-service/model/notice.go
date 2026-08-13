package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Notice 通知公告
type Notice struct {
	Id               int64        `db:"id"`
	CommunityId      int64        `db:"community_id"`
	Title            string       `db:"title"`
	Content          string       `db:"content"`
	Role             string       `db:"role"`
	Publisher        string       `db:"publisher"`
	PublisherId      *int64       `db:"publisher_id"`
	IsPinned         int32        `db:"is_pinned"`
	PublishedAt      time.Time    `db:"published_at"`
	ModerationStatus int64        `db:"moderation_status"`
	ModerationTime   sql.NullTime `db:"moderation_time"`
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
	DeletedAt        *time.Time   `db:"deleted_at"`
}

// NoticeModel 通知公告数据访问层
type NoticeModel interface {
	Insert(ctx context.Context, n *Notice) (int64, error)
	FindOne(ctx context.Context, id int64) (*Notice, error)
	// FindOnePublished 仅返回 moderation_status=通过 的内容（读路径可见性门禁）。
	FindOnePublished(ctx context.Context, id int64) (*Notice, error)
	FindList(ctx context.Context, communityId int64, role string, offset, limit int64) ([]*Notice, int64, error)
	Update(ctx context.Context, id int64, title, content string, isPinned int32) error
	SoftDelete(ctx context.Context, id int64) error
	UpdateModerationStatus(ctx context.Context, id int64, status int64) error
}

type defaultNoticeModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewNoticeModel(conn sqlx.SqlConn) NoticeModel {
	return &defaultNoticeModel{conn: conn, table: "`notices`"}
}

func (m *defaultNoticeModel) Insert(ctx context.Context, n *Notice) (int64, error) {
	query := `insert into ` + m.table + `
		(id, community_id, title, content, role, publisher, publisher_id, is_pinned, published_at, moderation_status, moderation_time)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := m.conn.ExecCtx(ctx, query,
		n.Id, n.CommunityId, n.Title, n.Content, n.Role,
		n.Publisher, n.PublisherId, n.IsPinned, n.PublishedAt, n.ModerationStatus, n.ModerationTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultNoticeModel) FindOne(ctx context.Context, id int64) (*Notice, error) {
	var v Notice
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null`
	err := m.conn.QueryRowCtx(ctx, &v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindOnePublished 读路径专用：仅返回审核通过（moderation_status=1）的内容。
// 待审核(0)/拒绝(2)的内容对普通用户读路径不可见（审核可见性门禁）。
func (m *defaultNoticeModel) FindOnePublished(ctx context.Context, id int64) (*Notice, error) {
	var v Notice
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null and moderation_status = 1`
	err := m.conn.QueryRowCtx(ctx, &v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (m *defaultNoticeModel) FindList(ctx context.Context, communityId int64, role string, offset, limit int64) ([]*Notice, int64, error) {
	var list []*Notice
	var total int64

	countQuery := `select count(*) from ` + m.table + ` where community_id = ? and deleted_at is null and moderation_status = 1`
	args := []interface{}{communityId}

	if role != "" {
		countQuery += ` and role = ?`
		args = append(args, role)
	}

	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `select * from ` + m.table + ` where community_id = ? and deleted_at is null and moderation_status = 1`
	queryArgs := []interface{}{communityId}
	if role != "" {
		query += ` and role = ?`
		queryArgs = append(queryArgs, role)
	}
	query += ` order by is_pinned desc, published_at desc limit ?, ?`
	queryArgs = append(queryArgs, offset, limit)

	err := m.conn.QueryRowsCtx(ctx, &list, query, queryArgs...)
	return list, total, err
}

func (m *defaultNoticeModel) Update(ctx context.Context, id int64, title, content string, isPinned int32) error {
	query := `update ` + m.table + ` set title = ?, content = ?, is_pinned = ?, moderation_status = 0 where id = ? and deleted_at is null`
	_, err := m.conn.ExecCtx(ctx, query, title, content, isPinned, id)
	return err
}

func (m *defaultNoticeModel) UpdateModerationStatus(ctx context.Context, id int64, status int64) error {
	query := `update ` + m.table + ` set moderation_status = ?, moderation_time = NOW() where id = ? and deleted_at is null`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

func (m *defaultNoticeModel) SoftDelete(ctx context.Context, id int64) error {
	query := `update ` + m.table + ` set deleted_at = now() where id = ?`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

var _ = time.Now
