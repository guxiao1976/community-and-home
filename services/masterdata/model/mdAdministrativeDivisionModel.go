package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdAdministrativeDivisionModel = (*customMdAdministrativeDivisionModel)(nil)

type (
	// MdAdministrativeDivisionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdAdministrativeDivisionModel.
	MdAdministrativeDivisionModel interface {
		mdAdministrativeDivisionModel
		FindChildren(ctx context.Context, parentId int64) ([]*MdAdministrativeDivision, error)
		FindChildrenWithFilter(ctx context.Context, parentId int64, level *int64, submissionStatus *int64) ([]*MdAdministrativeDivision, error)
		FindByLevel(ctx context.Context, level int64) ([]*MdAdministrativeDivision, error)
		FindByParentCode(ctx context.Context, parentCode string) ([]*MdAdministrativeDivision, error)
		FindAll(ctx context.Context, page, pageSize int64) ([]*MdAdministrativeDivision, error)
		FindAllWithFilter(ctx context.Context, level *int64, minLevel *int64, submissionStatus *int64, page, pageSize int64) ([]*MdAdministrativeDivision, error)
		CountAllWithFilter(ctx context.Context, level *int64, minLevel *int64, submissionStatus *int64) (int64, error)
		CountChildren(ctx context.Context, parentId int64, level *int64, submissionStatus *int64) (int64, error)
		FindRootDivisions(ctx context.Context) ([]*MdAdministrativeDivision, error)
		CountByParentId(ctx context.Context, parentId int64) (int64, error)
		UpdatePathForDescendants(ctx context.Context, oldPath string, newPath string) error
		SoftDelete(ctx context.Context, id int64) error
		Delete(ctx context.Context, id int64) error
		CountBySubmissionStatus(ctx context.Context, status int64) (int64, error)
		FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdAdministrativeDivision, error)
		UpdateStatusAndType(ctx context.Context, id int64, submissionStatus int64, submissionType int64) error
		UpdateStatus(ctx context.Context, id int64, submissionStatus int64) error
			Restore(ctx context.Context, id int64) error
			CountDeleted(ctx context.Context) (int64, error)
			FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdAdministrativeDivision, int64, error)
	}

	customMdAdministrativeDivisionModel struct {
		*defaultMdAdministrativeDivisionModel
	}
)

// NewMdAdministrativeDivisionModel returns a model for the database table.
func NewMdAdministrativeDivisionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdAdministrativeDivisionModel {
	return &customMdAdministrativeDivisionModel{
		defaultMdAdministrativeDivisionModel: newMdAdministrativeDivisionModel(conn, c, opts...),
	}
}

func (m *customMdAdministrativeDivisionModel) FindChildren(ctx context.Context, parentId int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where parent_id = ? and delete_time is null and submission_status != 4 order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, parentId)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindChildrenWithFilter(ctx context.Context, parentId int64, level *int64, submissionStatus *int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "parent_id = ?")
	args = append(args, parentId)
	if level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, *level)
	}
	if submissionStatus != nil {
		conditions = append(conditions, "submission_status = ?")
		args = append(args, *submissionStatus)
	}
	conditions = append(conditions, "delete_time is null")
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s order by sort_order, id", mdAdministrativeDivisionRows, m.table, where)
	err := m.QueryRowsNoCache(&divisions, query, args...)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindByLevel(ctx context.Context, level int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where level = ? and delete_time is null and submission_status != 4 order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, level)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindByParentCode(ctx context.Context, parentCode string) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where parent_code = ? and delete_time is null order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, parentCode)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindAll(ctx context.Context, page, pageSize int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is null and submission_status != 4 order by level, sort_order, id limit ?, ?", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindAllWithFilter(ctx context.Context, level *int64, minLevel *int64, submissionStatus *int64, page, pageSize int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "delete_time is null")
	if level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, *level)
	}
	if minLevel != nil {
		conditions = append(conditions, "level >= ?")
		args = append(args, *minLevel)
	}
	if submissionStatus != nil {
		conditions = append(conditions, "submission_status = ?")
		args = append(args, *submissionStatus)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s order by level, sort_order, id limit ? offset ?", mdAdministrativeDivisionRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCache(&divisions, query, args...)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindRootDivisions(ctx context.Context) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where parent_id is null and delete_time is null order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) CountByParentId(ctx context.Context, parentId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where parent_id = ? and delete_time is null", m.table)
	err := m.QueryRowNoCache(&count, query, parentId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdAdministrativeDivisionModel) UpdatePathForDescendants(ctx context.Context, oldPath string, newPath string) error {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where path like ?", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, oldPath+"%")
	if err != nil {
		return err
	}

	for _, d := range divisions {
		if d.Path == oldPath {
			continue
		}
		newDivisionPath := strings.Replace(d.Path, oldPath, newPath, 1)
		d.Path = newDivisionPath
		if err := m.Update(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func (m *customMdAdministrativeDivisionModel) SoftDelete(ctx context.Context, id int64) error {
	mdAdministrativeDivisionIdKey := fmt.Sprintf("%s%v", cacheMdAdministrativeDivisionIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set delete_time = now() where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdAdministrativeDivisionIdKey)
	return err
}

func (m *customMdAdministrativeDivisionModel) Delete(ctx context.Context, id int64) error {
	mdAdministrativeDivisionIdKey := fmt.Sprintf("%s%v", cacheMdAdministrativeDivisionIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdAdministrativeDivisionIdKey)
	return err
}

func (m *customMdAdministrativeDivisionModel) CountBySubmissionStatus(ctx context.Context, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where submission_status = ? and delete_time is null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdAdministrativeDivisionModel) FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
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
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdAdministrativeDivisionRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &divisions, query, args...)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) CountAllWithFilter(ctx context.Context, level *int64, minLevel *int64, submissionStatus *int64) (int64, error) {
	var count int64
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "delete_time is null")
	if level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, *level)
	}
	if minLevel != nil {
		conditions = append(conditions, "level >= ?")
		args = append(args, *minLevel)
	}
	if submissionStatus != nil {
		conditions = append(conditions, "submission_status = ?")
		args = append(args, *submissionStatus)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdAdministrativeDivisionModel) CountChildren(ctx context.Context, parentId int64, level *int64, submissionStatus *int64) (int64, error) {
	var count int64
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "parent_id = ?")
	args = append(args, parentId)
	if level != nil {
		conditions = append(conditions, "level = ?")
		args = append(args, *level)
	}
	if submissionStatus != nil {
		conditions = append(conditions, "submission_status = ?")
		args = append(args, *submissionStatus)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select count(*) from %s where %s and delete_time is null", m.table, where)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdAdministrativeDivisionModel) UpdateStatusAndType(ctx context.Context, id int64, submissionStatus int64, submissionType int64) error {
	mdAdministrativeDivisionIdKey := fmt.Sprintf("%s%v", cacheMdAdministrativeDivisionIdPrefix, id)
	query := fmt.Sprintf("update %s set submission_status = ?, submission_type = ? where `id` = ?", m.table)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, submissionStatus, submissionType, id)
	}, mdAdministrativeDivisionIdKey)
	return err
}

func (m *customMdAdministrativeDivisionModel) UpdateStatus(ctx context.Context, id int64, submissionStatus int64) error {
	mdAdministrativeDivisionIdKey := fmt.Sprintf("%s%v", cacheMdAdministrativeDivisionIdPrefix, id)
	query := fmt.Sprintf("update %s set submission_status = ? where `id` = ?", m.table)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, submissionStatus, id)
	}, mdAdministrativeDivisionIdKey)
	return err
}

func (m *customMdAdministrativeDivisionModel) Restore(ctx context.Context, id int64) error {
	mdAdministrativeDivisionIdKey := fmt.Sprintf("%s%v", cacheMdAdministrativeDivisionIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set delete_time = null, submission_status = 2 where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdAdministrativeDivisionIdKey)
	return err
}

func (m *customMdAdministrativeDivisionModel) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	return count, err
}

func (m *customMdAdministrativeDivisionModel) FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdAdministrativeDivision, int64, error) {
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}
	var divisions []*MdAdministrativeDivision
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is not null order by delete_time desc limit ? offset ?", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &divisions, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return divisions, total, nil
}
