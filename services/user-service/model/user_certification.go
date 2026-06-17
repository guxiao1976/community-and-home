package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserCertification 认证记录表
type UserCertification struct {
	Id           int64          `db:"id"`
	RoleId       int64          `db:"role_id"`
	UserId       int64          `db:"user_id"`
	DocumentUrls sql.NullString `db:"document_urls"`
	Status       int64          `db:"status"`
	ReviewerId   sql.NullInt64  `db:"reviewer_id"`
	ReviewTime       sql.NullTime   `db:"review_time"`
	ReviewNotes      sql.NullString `db:"review_notes"`
	ModerationStatus int64          `db:"moderation_status"`
	ModerationTime   sql.NullTime   `db:"moderation_time"`
	SubmitTime       time.Time      `db:"submit_time"`
}

type UserCertificationModel interface {
	Insert(ctx context.Context, data *UserCertification) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserCertification, error)
	FindByRoleId(ctx context.Context, roleId int64) ([]*UserCertification, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserCertification, error)
	FindPage(ctx context.Context, status *int64, userId *int64, page, pageSize int32) ([]*UserCertification, int64, error)
	Update(ctx context.Context, data *UserCertification) error
}

type defaultUserCertificationModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserCertificationModel(conn sqlx.SqlConn) UserCertificationModel {
	return &defaultUserCertificationModel{
		conn:  conn,
		table: "user_certification",
	}
}

func (m *defaultUserCertificationModel) Insert(ctx context.Context, data *UserCertification) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, moderation_status, moderation_time, submit_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.RoleId, data.UserId, data.DocumentUrls, data.Status, data.ReviewerId, data.ReviewTime, data.ReviewNotes, data.ModerationStatus, data.ModerationTime, data.SubmitTime)
}

func (m *defaultUserCertificationModel) FindOne(ctx context.Context, id int64) (*UserCertification, error) {
	query := fmt.Sprintf(`SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, moderation_status, moderation_time, submit_time FROM %s WHERE id = ?`, m.table)
	var resp UserCertification
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCertificationModel) FindByRoleId(ctx context.Context, roleId int64) ([]*UserCertification, error) {
	query := fmt.Sprintf(`SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, moderation_status, moderation_time, submit_time FROM %s WHERE role_id = ? ORDER BY submit_time DESC`, m.table)
	var resp []*UserCertification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, roleId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCertificationModel) FindByUserId(ctx context.Context, userId int64) ([]*UserCertification, error) {
	query := fmt.Sprintf(`SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, moderation_status, moderation_time, submit_time FROM %s WHERE user_id = ? ORDER BY submit_time DESC`, m.table)
	var resp []*UserCertification
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCertificationModel) FindPage(ctx context.Context, status *int64, userId *int64, page, pageSize int32) ([]*UserCertification, int64, error) {
	where := "WHERE 1=1"
	args := make([]interface{}, 0)

	if status != nil {
		where += " AND status = ?"
		args = append(args, *status)
	}
	if userId != nil {
		where += " AND user_id = ?"
		args = append(args, *userId)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, where)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT id, role_id, user_id, document_urls, status, reviewer_id, review_time, review_notes, moderation_status, moderation_time, submit_time FROM %s %s ORDER BY id DESC LIMIT ? OFFSET ?", m.table, where)
	queryArgs := append(args, pageSize, offset)
	var resp []*UserCertification
	err = m.conn.QueryRowsCtx(ctx, &resp, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	return resp, total, nil
}

func (m *defaultUserCertificationModel) Update(ctx context.Context, data *UserCertification) error {
	query := fmt.Sprintf(`UPDATE %s SET status=?, reviewer_id=?, review_notes=?, review_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Status, data.ReviewerId, data.ReviewNotes, data.ReviewTime, data.Id)
	return err
}
