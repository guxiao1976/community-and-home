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
		FindByLevel(ctx context.Context, level int64) ([]*MdAdministrativeDivision, error)
		FindRootDivisions(ctx context.Context) ([]*MdAdministrativeDivision, error)
		CountByParentId(ctx context.Context, parentId int64) (int64, error)
		UpdatePathForDescendants(ctx context.Context, oldPath string, newPath string) error
		SoftDelete(ctx context.Context, id int64) error
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
	query := fmt.Sprintf("select %s from %s where parent_id = ? and delete_time is null order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, parentId)
	if err != nil {
		return nil, err
	}
	return divisions, nil
}

func (m *customMdAdministrativeDivisionModel) FindByLevel(ctx context.Context, level int64) ([]*MdAdministrativeDivision, error) {
	var divisions []*MdAdministrativeDivision
	query := fmt.Sprintf("select %s from %s where level = ? and delete_time is null order by sort_order, id", mdAdministrativeDivisionRows, m.table)
	err := m.QueryRowsNoCache(&divisions, query, level)
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