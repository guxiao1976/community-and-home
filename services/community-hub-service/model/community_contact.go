package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CommunityContact 便民联络
type CommunityContact struct {
	Id          int64     `db:"id"`
	CommunityId int64     `db:"community_id"`
	Category    string    `db:"category"`
	Name        string    `db:"name"`
	Phone       string    `db:"phone"`
	SortOrder   int32     `db:"sort_order"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// CommunityContactModel 便民联络数据访问层
type CommunityContactModel interface {
	Insert(ctx context.Context, c *CommunityContact) (int64, error)
	FindByCommunityId(ctx context.Context, communityId int64) ([]*CommunityContact, error)
	DeleteByCommunityId(ctx context.Context, communityId int64) error
}

type defaultCommunityContactModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewCommunityContactModel(conn sqlx.SqlConn) CommunityContactModel {
	return &defaultCommunityContactModel{conn: conn, table: "`community_contacts`"}
}

func (m *defaultCommunityContactModel) Insert(ctx context.Context, c *CommunityContact) (int64, error) {
	query := `insert into ` + m.table + `
		(id, community_id, category, name, phone, sort_order)
		values (?, ?, ?, ?, ?, ?)`
	res, err := m.conn.ExecCtx(ctx, query,
		c.Id, c.CommunityId, c.Category, c.Name, c.Phone, c.SortOrder)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultCommunityContactModel) FindByCommunityId(ctx context.Context, communityId int64) ([]*CommunityContact, error) {
	var list []*CommunityContact
	query := `select * from ` + m.table + ` where community_id = ? order by sort_order asc`
	err := m.conn.QueryRowsCtx(ctx, &list, query, communityId)
	return list, err
}

func (m *defaultCommunityContactModel) DeleteByCommunityId(ctx context.Context, communityId int64) error {
	query := `delete from ` + m.table + ` where community_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, communityId)
	return err
}

var _ = time.Now
