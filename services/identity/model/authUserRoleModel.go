package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthUserRoleModel = (*customAuthUserRoleModel)(nil)

type (
	AuthUserRoleModel interface {
		authUserRoleModel
		FindByUserId(ctx context.Context, userId int64) ([]*AuthUserRole, error)
	}

	customAuthUserRoleModel struct {
		*defaultAuthUserRoleModel
	}
)

func NewAuthUserRoleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthUserRoleModel {
	return &customAuthUserRoleModel{
		defaultAuthUserRoleModel: newAuthUserRoleModel(conn, c, opts...),
	}
}

func (m *customAuthUserRoleModel) FindByUserId(ctx context.Context, userId int64) ([]*AuthUserRole, error) {
	var list []*AuthUserRole
	query := fmt.Sprintf("select %s from %s where user_id = ?", authUserRoleRows, m.table)
	err := m.QueryRowsNoCache(&list, query, userId)
	if err != nil {
		return nil, err
	}
	return list, nil
}