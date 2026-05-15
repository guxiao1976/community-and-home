package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthUserModel = (*customAuthUserModel)(nil)

type (
	AuthUserModel interface {
		authUserModel
		FindAll(ctx context.Context) ([]*AuthUser, error)
		FindByUserType(ctx context.Context, userType int64) ([]*AuthUser, error)
		CountByFilter(ctx context.Context, phone, nickname string, userType *int64, status *int64) (int64, error)
		FindPage(ctx context.Context, phone, nickname string, userType *int64, status *int64, page, pageSize int64) ([]*AuthUser, error)
		SoftDelete(ctx context.Context, id int64) error
	}

	customAuthUserModel struct {
		*defaultAuthUserModel
	}
)

func NewAuthUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthUserModel {
	return &customAuthUserModel{
		defaultAuthUserModel: newAuthUserModel(conn, c, opts...),
	}
}

func (m *customAuthUserModel) FindAll(ctx context.Context) ([]*AuthUser, error) {
	var users []*AuthUser
	query := fmt.Sprintf("select %s from %s where delete_time is null order by id desc", authUserRows, m.table)
	err := m.QueryRowsNoCache(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *customAuthUserModel) FindByUserType(ctx context.Context, userType int64) ([]*AuthUser, error) {
	var users []*AuthUser
	query := fmt.Sprintf("select %s from %s where user_type = ? and delete_time is null order by id desc", authUserRows, m.table)
	err := m.QueryRowsNoCache(&users, query, userType)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *customAuthUserModel) CountByFilter(ctx context.Context, phone, nickname string, userType *int64, status *int64) (int64, error) {
	var where []string
	var args []interface{}
	where = append(where, "delete_time is null")
	if phone != "" {
		where = append(where, "phone LIKE ?")
		args = append(args, "%"+phone+"%")
	}
	if nickname != "" {
		where = append(where, "nickname LIKE ?")
		args = append(args, "%"+nickname+"%")
	}
	if userType != nil {
		where = append(where, "user_type = ?")
		args = append(args, *userType)
	}
	if status != nil {
		where = append(where, "status = ?")
		args = append(args, *status)
	}

	var count int64
	query := fmt.Sprintf("select count(*) from %s where %s", m.table, joinWhere(where))
	err := m.QueryRowNoCache(&count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customAuthUserModel) FindPage(ctx context.Context, phone, nickname string, userType *int64, status *int64, page, pageSize int64) ([]*AuthUser, error) {
	var where []string
	var args []interface{}
	where = append(where, "delete_time is null")
	if phone != "" {
		where = append(where, "phone LIKE ?")
		args = append(args, "%"+phone+"%")
	}
	if nickname != "" {
		where = append(where, "nickname LIKE ?")
		args = append(args, "%"+nickname+"%")
	}
	if userType != nil {
		where = append(where, "user_type = ?")
		args = append(args, *userType)
	}
	if status != nil {
		where = append(where, "status = ?")
		args = append(args, *status)
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where %s order by id desc limit ? offset ?", authUserRows, m.table, joinWhere(where))
	args = append(args, pageSize, offset)

	var users []*AuthUser
	err := m.QueryRowsNoCache(&users, query, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (m *customAuthUserModel) SoftDelete(ctx context.Context, id int64) error {
	authUserIdKey := fmt.Sprintf("%s%v", cacheAuthUserIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set delete_time = now() where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, authUserIdKey)
	return err
}