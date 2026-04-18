package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthFamilyModel = (*customAuthFamilyModel)(nil)

type (
	AuthFamilyModel interface {
		authFamilyModel
		FindByPropertyUnitId(ctx context.Context, propertyUnitId int64, page, pageSize int32, status *int64) ([]*AuthFamily, int64, error)
	}

	customAuthFamilyModel struct {
		*defaultAuthFamilyModel
	}
)

func NewAuthFamilyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthFamilyModel {
	return &customAuthFamilyModel{
		defaultAuthFamilyModel: newAuthFamilyModel(conn, c, opts...),
	}
}

func (m *customAuthFamilyModel) FindByPropertyUnitId(ctx context.Context, propertyUnitId int64, page, pageSize int32, status *int64) ([]*AuthFamily, int64, error) {
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("select %s from %s where property_unit_id = ?", authFamilyRows, m.table)
	countQuery := fmt.Sprintf("select count(*) from %s where property_unit_id = ?", m.table)
	args := []interface{}{propertyUnitId}

	if status != nil {
		query += " and status = ?"
		countQuery += " and status = ?"
		args = append(args, *status)
	}

	var total int64
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query += " order by created_time desc limit ? offset ?"
	queryArgs := append(args, pageSize, offset)

	var families []*AuthFamily
	err = m.QueryRowsNoCacheCtx(ctx, &families, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}

	return families, total, nil
}
