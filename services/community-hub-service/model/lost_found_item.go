package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// LostFoundItem 寻失互助
type LostFoundItem struct {
	Id               int64        `db:"id"`
	CommunityId      int64        `db:"community_id"`
	Type             string       `db:"type"`
	Title            string       `db:"title"`
	Description      string       `db:"description"`
	ImageUrls        string       `db:"image_urls"` // JSON string
	ContactPhone     string       `db:"contact_phone"`
	Status           string       `db:"status"`
	PublisherId      int64        `db:"publisher_id"`
	ModerationStatus int64        `db:"moderation_status"`
	ModerationTime   sql.NullTime `db:"moderation_time"`
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
	DeletedAt        *time.Time   `db:"deleted_at"`
}

// LostFoundItemModel 寻失互助数据访问层
type LostFoundItemModel interface {
	Insert(ctx context.Context, item *LostFoundItem) (int64, error)
	FindOne(ctx context.Context, id int64) (*LostFoundItem, error)
	// FindOnePublished 仅返回 moderation_status=通过 的内容（读路径可见性门禁）。
	FindOnePublished(ctx context.Context, id int64) (*LostFoundItem, error)
	FindList(ctx context.Context, communityId int64, typ string, offset, limit int64) ([]*LostFoundItem, int64, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateModerationStatus(ctx context.Context, id int64, status int64) error
	// CountQuotaOccupied 统计占配额内容数（用户×小区×板块三维计数，Task 4.1）。
	CountQuotaOccupied(ctx context.Context, publisherId, communityId int64, typ string) (int64, error)
}

type defaultLostFoundItemModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewLostFoundItemModel(conn sqlx.SqlConn) LostFoundItemModel {
	return &defaultLostFoundItemModel{conn: conn, table: "`lost_found_items`"}
}

func (m *defaultLostFoundItemModel) Insert(ctx context.Context, item *LostFoundItem) (int64, error) {
	query := `insert into ` + m.table + `
		(id, community_id, type, title, description, image_urls, contact_phone, status, publisher_id, moderation_status, moderation_time)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := m.conn.ExecCtx(ctx, query,
		item.Id, item.CommunityId, item.Type, item.Title,
		item.Description, item.ImageUrls, item.ContactPhone,
		item.Status, item.PublisherId, item.ModerationStatus, item.ModerationTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultLostFoundItemModel) FindOne(ctx context.Context, id int64) (*LostFoundItem, error) {
	var v LostFoundItem
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null`
	err := m.conn.QueryRowCtx(ctx, &v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindOnePublished 读路径专用：仅返回审核通过（moderation_status=1）的内容。
// 待审核(0)/拒绝(2)的内容对普通用户读路径不可见（审核可见性门禁）。
func (m *defaultLostFoundItemModel) FindOnePublished(ctx context.Context, id int64) (*LostFoundItem, error) {
	var v LostFoundItem
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null and moderation_status = 1`
	err := m.conn.QueryRowCtx(ctx, &v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (m *defaultLostFoundItemModel) FindList(ctx context.Context, communityId int64, typ string, offset, limit int64) ([]*LostFoundItem, int64, error) {
	var list []*LostFoundItem
	var total int64

	countQuery := `select count(*) from ` + m.table + ` where community_id = ? and deleted_at is null and moderation_status = 1`
	args := []interface{}{communityId}

	if typ != "" {
		countQuery += ` and type = ?`
		args = append(args, typ)
	}

	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := `select * from ` + m.table + ` where community_id = ? and deleted_at is null and moderation_status = 1`
	queryArgs := []interface{}{communityId}
	if typ != "" {
		query += ` and type = ?`
		queryArgs = append(queryArgs, typ)
	}
	query += ` order by created_at desc limit ?, ?`
	queryArgs = append(queryArgs, offset, limit)

	err := m.conn.QueryRowsCtx(ctx, &list, query, queryArgs...)
	return list, total, err
}

// CountQuotaOccupied 统计占配额内容数（用户×小区×板块三维，Task 4.1 / design §7）。
//
// 谓词：deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)。
// 即：待审(0)+active、通过(1)+active 同占配额；驳回(2)、已解决(status=resolved)、
// 已删除(deleted_at 非空) 释放。状态机依据 STAGE3-2：内容创建即 status='active'。
//
// 口径：个人（publisher_id）×小区（community_id）×板块。typ 为板块标识
// （sys_section_quota.section_type，当前仅 lost_found）。lost_found_items 表即 lost_found
// 板块的唯一承载（type 列仅区分 lost/found 子类，二者同属 lost_found 板块共同占配额），
// 故不按 type 过滤，按「板块=lost_found」统计整板占配额内容；typ 保留为多板块扩展位。
//
// SEE: [[moderation-status-write-without-read-gating]] — 计数须覆盖待审(0)，防「发→删→重发」刷配额
func (m *defaultLostFoundItemModel) CountQuotaOccupied(ctx context.Context, publisherId, communityId int64, typ string) (int64, error) {
	var count int64
	query := `select count(*) from ` + m.table + ` where publisher_id = ? and community_id = ? and deleted_at is null and status = 'active' and moderation_status in (0, 1)`
	if err := m.conn.QueryRowCtx(ctx, &count, query, publisherId, communityId); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultLostFoundItemModel) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `update ` + m.table + ` set status = ? where id = ? and deleted_at is null`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

func (m *defaultLostFoundItemModel) UpdateModerationStatus(ctx context.Context, id int64, status int64) error {
	query := `update ` + m.table + ` set moderation_status = ?, moderation_time = NOW() where id = ? and deleted_at is null`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

var _ = time.Now
