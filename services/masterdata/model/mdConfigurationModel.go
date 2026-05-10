package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdConfigurationModel = (*customMdConfigurationModel)(nil)

type (
	// MdConfigurationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdConfigurationModel.
	MdConfigurationModel interface {
		mdConfigurationModel
		FindByModule(ctx context.Context, module string, limit, offset int) ([]*MdConfiguration, int64, error)
		FindAll(ctx context.Context, limit, offset int) ([]*MdConfiguration, int64, error)
		FindOneByModuleConfigKey(ctx context.Context, module, configKey string) (*MdConfiguration, error)
		SoftDelete(ctx context.Context, id int64) error
		CountBySubmissionStatus(ctx context.Context, status int64) (int64, error)
		FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdConfiguration, error)
		CountDeleted(ctx context.Context) (int64, error)
		FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdConfiguration, int64, error)
		Restore(ctx context.Context, id int64) error
	}

	customMdConfigurationModel struct {
		*defaultMdConfigurationModel
	}
)

// NewMdConfigurationModel returns a model for the database table.
func NewMdConfigurationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdConfigurationModel {
	return &customMdConfigurationModel{
		defaultMdConfigurationModel: newMdConfigurationModel(conn, c, opts...),
	}
}

func (m *customMdConfigurationModel) FindOneByConfigKey(ctx context.Context, configKey string) (*MdConfiguration, error) {
	var resp MdConfiguration
	query := fmt.Sprintf("select %s from %s where `config_key` = ? and delete_time is null limit 1", mdConfigurationRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, configKey)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMdConfigurationModel) FindOneByModuleConfigKey(ctx context.Context, module, configKey string) (*MdConfiguration, error) {
	var resp MdConfiguration
	query := fmt.Sprintf("select %s from %s where `module` = ? and `config_key` = ? and delete_time is null limit 1", mdConfigurationRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, module, configKey)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customMdConfigurationModel) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s", m.table)

	if category != "" {
		query += " where `category` = ? and delete_time is null"
		countQuery += " where `category` = ? and delete_time is null"
	} else {
		query += " where delete_time is null"
		countQuery += " where delete_time is null"
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

func (m *customMdConfigurationModel) FindByModule(ctx context.Context, module string, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s where `module` = ? and delete_time is null and submission_status != 4 order by id desc limit ? offset ?", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s where `module` = ? and delete_time is null and submission_status != 4", m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, module, limit, offset)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery, module)
	if err2 != nil {
		return nil, 0, err2
	}

	return resp, total, nil
}

func (m *customMdConfigurationModel) FindAll(ctx context.Context, limit, offset int) ([]*MdConfiguration, int64, error) {
	var resp []*MdConfiguration
	var total int64

	query := fmt.Sprintf("select %s from %s where delete_time is null and submission_status != 4 order by id desc limit ? offset ?", mdConfigurationRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is null and submission_status != 4", m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, limit, offset)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, 0, err
	}

	err2 := m.QueryRowNoCacheCtx(ctx, &total, countQuery)
	if err2 != nil {
		return nil, 0, err2
	}

	return resp, total, nil
}

func (m *customMdConfigurationModel) SoftDelete(ctx context.Context, id int64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, id)
	mdConfigurationModuleConfigKeyKey := fmt.Sprintf("%s%v:%v", cacheMdConfigurationModuleConfigKeyPrefix, data.Module, data.ConfigKey)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set delete_time = now() where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdConfigurationIdKey, mdConfigurationModuleConfigKeyKey)
	return err
}

func (m *customMdConfigurationModel) CountBySubmissionStatus(ctx context.Context, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where submission_status = ? and delete_time is null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdConfigurationModel) FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdConfiguration, error) {
	var configs []*MdConfiguration
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
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdConfigurationRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &configs, query, args...)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (m *customMdConfigurationModel) Restore(ctx context.Context, id int64) error {
	mdConfigurationIdKey := fmt.Sprintf("%s%v", cacheMdConfigurationIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set delete_time = null, submission_status = 2 where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdConfigurationIdKey)
	return err
}

func (m *customMdConfigurationModel) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	return count, err
}
func (m *customMdConfigurationModel) FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdConfiguration, int64, error) {

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}
	var configs []*MdConfiguration
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is not null order by delete_time desc limit ? offset ?", mdConfigurationRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &configs, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return configs, total, nil
}
