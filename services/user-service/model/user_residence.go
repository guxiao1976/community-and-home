package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserResidence 居民房屋明细表
type UserResidence struct {
	Id           int64        `db:"id"`
	MembershipId int64        `db:"membership_id"`
	UserId       int64        `db:"user_id"` // 冗余：分片键
	HouseId      string       `db:"house_id"`
	Building     string       `db:"building"`
	Unit         string       `db:"unit"`
	Room         string       `db:"room"`
	IsPrimary    int64        `db:"is_primary"`
	StartDate    sql.NullTime `db:"start_date"`
	EndDate      sql.NullTime `db:"end_date"`
	CreatedTime  time.Time    `db:"created_at"`
	UpdatedTime  time.Time    `db:"updated_at"`
}

type UserResidenceModel interface {
	Insert(ctx context.Context, data *UserResidence) (sql.Result, error)
	FindByMembershipId(ctx context.Context, membershipId int64) ([]*UserResidence, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserResidence, error)
	FindByMembershipAndHouse(ctx context.Context, membershipId int64, houseId string) (*UserResidence, error)
	Update(ctx context.Context, data *UserResidence) error
}

type defaultUserResidenceModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserResidenceModel(conn sqlx.SqlConn) UserResidenceModel {
	return &defaultUserResidenceModel{
		conn:  conn,
		table: "user_residence",
	}
}

func (m *defaultUserResidenceModel) Insert(ctx context.Context, data *UserResidence) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, membership_id, user_id, house_id, building, unit, room, is_primary, start_date, end_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.MembershipId, data.UserId, data.HouseId, data.Building, data.Unit, data.Room, data.IsPrimary, data.StartDate, data.EndDate, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserResidenceModel) FindByMembershipId(ctx context.Context, membershipId int64) ([]*UserResidence, error) {
	query := fmt.Sprintf(`SELECT id, membership_id, user_id, house_id, building, unit, room, is_primary, start_date, end_date, created_at, updated_at FROM %s WHERE membership_id = ?`, m.table)
	var resp []*UserResidence
	err := m.conn.QueryRowsCtx(ctx, &resp, query, membershipId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserResidenceModel) FindByUserId(ctx context.Context, userId int64) ([]*UserResidence, error) {
	query := fmt.Sprintf(`SELECT id, membership_id, user_id, house_id, building, unit, room, is_primary, start_date, end_date, created_at, updated_at FROM %s WHERE user_id = ?`, m.table)
	var resp []*UserResidence
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserResidenceModel) FindByMembershipAndHouse(ctx context.Context, membershipId int64, houseId string) (*UserResidence, error) {
	query := fmt.Sprintf(`SELECT id, membership_id, user_id, house_id, building, unit, room, is_primary, start_date, end_date, created_at, updated_at FROM %s WHERE membership_id = ? AND house_id = ?`, m.table)
	var resp UserResidence
	err := m.conn.QueryRowCtx(ctx, &resp, query, membershipId, houseId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserResidenceModel) Update(ctx context.Context, data *UserResidence) error {
	query := fmt.Sprintf(`UPDATE %s SET building=?, unit=?, room=?, is_primary=?, start_date=?, end_date=?, updated_at=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Building, data.Unit, data.Room, data.IsPrimary, data.StartDate, data.EndDate, time.Now(), data.Id)
	return err
}
