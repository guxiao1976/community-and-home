package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdSensitiveWordModel = (*customMdSensitiveWordModel)(nil)

type (
	MdSensitiveWordModel interface {
		mdSensitiveWordModel
		FindOneByWord(ctx context.Context, word string) (*MdSensitiveWord, error)
		FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdSensitiveWord, int64, error)
		FindWithFilters(ctx context.Context, category string, severity *int64, status *int64, limit, offset int) ([]*MdSensitiveWord, int64, error)
		CountBySubmissionStatus(ctx context.Context, status int64) (int64, error)
		FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdSensitiveWord, error)
		CountDeleted(ctx context.Context) (int64, error)
		FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdSensitiveWord, int64, error)
		Restore(ctx context.Context, id int64) error
	}

	customMdSensitiveWordModel struct {
		*defaultMdSensitiveWordModel
	}
)

// NewMdSensitiveWordModel returns a model for the database table.
func NewMdSensitiveWordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdSensitiveWordModel {
	return &customMdSensitiveWordModel{
		defaultMdSensitiveWordModel: newMdSensitiveWordModel(conn, c, opts...),
	}
}

func (m *customMdSensitiveWordModel) FindOneByWord(ctx context.Context, word string) (*MdSensitiveWord, error) {
	var resp MdSensitiveWord
	query := fmt.Sprintf("select %s from %s where `word` = ? and delete_time is null limit 1", mdSensitiveWordRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, word)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMdSensitiveWordModel) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdSensitiveWord, int64, error) {
	var resp []*MdSensitiveWord
	var total int64

	query := fmt.Sprintf("select %s from %s where delete_time is null", mdSensitiveWordRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is null", m.table)

	if category != "" {
		query += " and `category` = ?"
		countQuery += " and `category` = ?"
	}

	query += " order by id desc limit ? offset ?"

	var err error
	if category != "" {
		err = m.QueryRowsNoCacheCtx(ctx, &resp, query, category, limit, offset)
		err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery, category)
		if err2 != nil {
			return nil, 0, err2
		}
	} else {
		err = m.QueryRowsNoCacheCtx(ctx, &resp, query, limit, offset)
		err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery)
		if err2 != nil {
			return nil, 0, err2
		}
	}

	switch err {
	case nil:
		return resp, total, nil
	case sqlx.ErrNotFound:
		return nil, 0, nil
	default:
		return nil, 0, err
	}
}

func (m *customMdSensitiveWordModel) FindWithFilters(ctx context.Context, category string, severity *int64, status *int64, limit, offset int) ([]*MdSensitiveWord, int64, error) {
	var resp []*MdSensitiveWord
	var total int64

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "delete_time is null")
	conditions = append(conditions, "submission_status != 4")

	if category != "" {
		conditions = append(conditions, "`category` = ?")
		args = append(args, category)
	}
	if severity != nil {
		conditions = append(conditions, "`severity` = ?")
		args = append(args, *severity)
	}
	if status != nil {
		conditions = append(conditions, "`status` = ?")
		args = append(args, *status)
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf("SELECT %s FROM %s%s ORDER BY id DESC LIMIT ? OFFSET ?",
		mdSensitiveWordRows, m.table, whereClause)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", m.table, whereClause)

	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	queryArgs := append(args, limit, offset)
	err = m.QueryRowsNoCacheCtx(ctx, &resp, query, queryArgs...)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	return resp, total, nil
}

func (m *customMdSensitiveWordModel) CountBySubmissionStatus(ctx context.Context, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where submission_status = ? and delete_time is null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdSensitiveWordModel) FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdSensitiveWord, error) {
	var words []*MdSensitiveWord
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "submission_status = ?")
	args = append(args, status)
	if submissionType != nil {
		conditions = append(conditions, "submission_type = ?")
		args = append(args, *submissionType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdSensitiveWordRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &words, query, args...)
	if err != nil {
		return nil, err
	}
	return words, nil
}

func (m *customMdSensitiveWordModel) Restore(ctx context.Context, id int64) error {
	mdSensitiveWordIdKey := fmt.Sprintf("%s%v", cacheMdSensitiveWordIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set delete_time = null, submission_status = 2 where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdSensitiveWordIdKey)
	return err
}

func (m *customMdSensitiveWordModel) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	return count, err
}
func (m *customMdSensitiveWordModel) FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdSensitiveWord, int64, error) {

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}
	var words []*MdSensitiveWord
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is not null order by delete_time desc limit ? offset ?", mdSensitiveWordRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &words, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return words, total, nil
}
