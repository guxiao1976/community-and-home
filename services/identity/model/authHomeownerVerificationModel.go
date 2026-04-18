package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthHomeownerVerificationModel = (*customAuthHomeownerVerificationModel)(nil)

type (
	// AuthHomeownerVerificationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthHomeownerVerificationModel.
	AuthHomeownerVerificationModel interface {
		authHomeownerVerificationModel
		FindByUserId(ctx context.Context, userId int64, page, pageSize int32, status *int32) ([]*AuthHomeownerVerification, error)
		CountByUserId(ctx context.Context, userId int64, status *int32) (int64, error)
		FindPendingList(ctx context.Context, page, pageSize int32) ([]*AuthHomeownerVerification, error)
		CountPending(ctx context.Context) (int64, error)
		FindAll(ctx context.Context, page, pageSize int32, status *int32, verificationType *int32) ([]*AuthHomeownerVerification, error)
		CountAll(ctx context.Context, status *int32, verificationType *int32) (int64, error)
		UpdateStatus(ctx context.Context, id int64, status int32, reviewerId int64, reviewNotes string) error
		FindOneByUserIdPropertyUnitId(ctx context.Context, userId, propertyUnitId int64) (*AuthHomeownerVerification, error)
	}

	customAuthHomeownerVerificationModel struct {
		*defaultAuthHomeownerVerificationModel
	}
)

// NewAuthHomeownerVerificationModel returns a model for the database table.
func NewAuthHomeownerVerificationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthHomeownerVerificationModel {
	return &customAuthHomeownerVerificationModel{
		defaultAuthHomeownerVerificationModel: newAuthHomeownerVerificationModel(conn, c, opts...),
	}
}

func (m *customAuthHomeownerVerificationModel) FindByUserId(ctx context.Context, userId int64, page, pageSize int32, status *int32) ([]*AuthHomeownerVerification, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND delete_time IS NULL", authHomeownerVerificationRows, m.table)
	args := []interface{}{userId}

	if status != nil {
		query += " AND verification_status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY created_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*AuthHomeownerVerification
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthHomeownerVerificationModel) CountByUserId(ctx context.Context, userId int64, status *int32) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE user_id = ? AND delete_time IS NULL", m.table)
	args := []interface{}{userId}

	if status != nil {
		query += " AND verification_status = ?"
		args = append(args, *status)
	}

	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customAuthHomeownerVerificationModel) FindPendingList(ctx context.Context, page, pageSize int32) ([]*AuthHomeownerVerification, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE verification_status = 0 ORDER BY submit_time DESC LIMIT ? OFFSET ?", authHomeownerVerificationRows, m.table)
	var resp []*AuthHomeownerVerification
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthHomeownerVerificationModel) CountPending(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE verification_status = 0", m.table)
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customAuthHomeownerVerificationModel) UpdateStatus(ctx context.Context, id int64, status int32, reviewerId int64, reviewNotes string) error {
	key := fmt.Sprintf("%s%v", cacheAuthHomeownerVerificationIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("UPDATE %s SET verification_status = ?, reviewer_id = ?, review_time = ?, review_notes = ?, updated_time = ? WHERE id = ?", m.table)
		return conn.ExecCtx(ctx, query, status, reviewerId, time.Now(), reviewNotes, time.Now(), id)
	}, key)
	return err
}

func (m *customAuthHomeownerVerificationModel) FindAll(ctx context.Context, page, pageSize int32, status *int32, verificationType *int32) ([]*AuthHomeownerVerification, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE 1=1", authHomeownerVerificationRows, m.table)
	args := []interface{}{}

	if status != nil {
		query += " AND verification_status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY submit_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var resp []*AuthHomeownerVerification
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customAuthHomeownerVerificationModel) CountAll(ctx context.Context, status *int32, verificationType *int32) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE 1=1", m.table)
	args := []interface{}{}

	if status != nil {
		query += " AND verification_status = ?"
		args = append(args, *status)
	}

	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customAuthHomeownerVerificationModel) FindOneByUserIdPropertyUnitId(ctx context.Context, userId, propertyUnitId int64) (*AuthHomeownerVerification, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = ? AND property_unit_id = ? AND delete_time IS NULL ORDER BY created_time DESC LIMIT 1", authHomeownerVerificationRows, m.table)
	var resp AuthHomeownerVerification
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, propertyUnitId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
