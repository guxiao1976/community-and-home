package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserBase 用户基础表（5 表设计）
type UserBase struct {
	Id                       int64          `db:"id"`
	Phone                    string         `db:"phone"`
	Nickname                 sql.NullString `db:"nickname"`
	AvatarUrl                sql.NullString `db:"avatar_url"`
	RealName                 sql.NullString `db:"real_name"`
	IdCardNumber             sql.NullString `db:"id_card_number"`
	Gender                   sql.NullInt64  `db:"gender"`
	BirthDate                sql.NullTime   `db:"birth_date"`
	Status                   int64          `db:"status"`
	CreditScore              int64          `db:"credit_score"`
	NicknameModerationStatus int64          `db:"nickname_moderation_status"`
	Preferences              sql.NullString `db:"preferences"`
	CreatedTime              time.Time      `db:"created_time"`
	UpdatedTime              time.Time      `db:"updated_time"`
	DeleteTime               sql.NullTime   `db:"delete_time"`
}

type UserBaseModel interface {
	Insert(ctx context.Context, data *UserBase) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserBase, error)
	FindOneByPhone(ctx context.Context, encryptedPhone string) (*UserBase, error)
	FindByIds(ctx context.Context, ids []int64) ([]*UserBase, error)
	FindPage(ctx context.Context, keyword string, status *int64, page, pageSize int32) ([]*UserBase, int64, error)
	Update(ctx context.Context, data *UserBase) error
	SoftDelete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status int64) error
	UpdateRealNameAndIdCard(ctx context.Context, id int64, realName, idCardNumber string) error
}

type defaultUserBaseModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserBaseModel(conn sqlx.SqlConn) UserBaseModel {
	return &defaultUserBaseModel{
		conn:  conn,
		table: "user_base",
	}
}

func (m *defaultUserBaseModel) Insert(ctx context.Context, data *UserBase) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, nickname_moderation_status, preferences, created_time, updated_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.Phone, data.Nickname, data.AvatarUrl, data.RealName, data.IdCardNumber, data.Gender, data.BirthDate, data.Status, data.CreditScore, data.NicknameModerationStatus, data.Preferences, data.CreatedTime, data.UpdatedTime)
}

func (m *defaultUserBaseModel) FindOne(ctx context.Context, id int64) (*UserBase, error) {
	query := fmt.Sprintf(`SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, nickname_moderation_status, preferences, created_time, updated_time, delete_time FROM %s WHERE id = ? AND delete_time IS NULL`, m.table)
	var resp UserBase
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

// FindOneByPhone 根据加密后的手机号查询用户
func (m *defaultUserBaseModel) FindOneByPhone(ctx context.Context, encryptedPhone string) (*UserBase, error) {
	query := fmt.Sprintf(`SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, nickname_moderation_status, preferences, created_time, updated_time, delete_time FROM %s WHERE phone = ? AND delete_time IS NULL`, m.table)
	var resp UserBase
	err := m.conn.QueryRowCtx(ctx, &resp, query, encryptedPhone)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserBaseModel) FindByIds(ctx context.Context, ids []int64) ([]*UserBase, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(`SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, nickname_moderation_status, preferences, created_time, updated_time, delete_time FROM %s WHERE id IN (%s) AND delete_time IS NULL`, m.table, strings.Join(placeholders, ","))
	var resp []*UserBase
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserBaseModel) FindPage(ctx context.Context, keyword string, status *int64, page, pageSize int32) ([]*UserBase, int64, error) {
	where := "WHERE delete_time IS NULL"
	args := make([]interface{}, 0)

	if keyword != "" {
		where += " AND (nickname LIKE ? OR phone LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	if status != nil {
		where += " AND status = ?"
		args = append(args, *status)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, where)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT id, phone, nickname, avatar_url, real_name, id_card_number, gender, birth_date, status, credit_score, nickname_moderation_status, preferences, created_time, updated_time, delete_time FROM %s %s ORDER BY id DESC LIMIT ? OFFSET ?", m.table, where)
	queryArgs := append(args, pageSize, offset)
	var resp []*UserBase
	err = m.conn.QueryRowsCtx(ctx, &resp, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (m *defaultUserBaseModel) Update(ctx context.Context, data *UserBase) error {
	query := fmt.Sprintf(`UPDATE %s SET nickname=?, avatar_url=?, gender=?, birth_date=?, status=?, preferences=?, updated_time=? WHERE id=? AND delete_time IS NULL`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Nickname, data.AvatarUrl, data.Gender, data.BirthDate, data.Status, data.Preferences, time.Now(), data.Id)
	return err
}

func (m *defaultUserBaseModel) SoftDelete(ctx context.Context, id int64) error {
	query := fmt.Sprintf(`UPDATE %s SET delete_time=?, updated_time=? WHERE id=? AND delete_time IS NULL`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, time.Now(), time.Now(), id)
	return err
}

func (m *defaultUserBaseModel) UpdateStatus(ctx context.Context, id int64, status int64) error {
	query := fmt.Sprintf(`UPDATE %s SET status=?, updated_time=? WHERE id=? AND delete_time IS NULL`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, time.Now(), id)
	return err
}

// UpdateRealNameAndIdCard 首次认证通过后回填，已有值不覆盖（COALESCE）
func (m *defaultUserBaseModel) UpdateRealNameAndIdCard(ctx context.Context, id int64, realName, idCardNumber string) error {
	query := fmt.Sprintf(`UPDATE %s SET real_name=COALESCE(real_name, ?), id_card_number=COALESCE(id_card_number, ?), updated_time=? WHERE id=? AND delete_time IS NULL`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, realName, idCardNumber, time.Now(), id)
	return err
}
