package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserAppState 用户应用状态表（账号级当前小区）
type UserAppState struct {
	UserId             int64     `db:"user_id"`
	CurrentCommunityId int64     `db:"current_community_id"`
	CreatedTime        time.Time `db:"created_at"`
	UpdatedTime        time.Time `db:"updated_at"`
}

// UserAppStateModel 用户应用状态数据访问接口
type UserAppStateModel interface {
	// FindOne 按 user_id 查询；无记录返回 ErrNotFound
	FindOne(ctx context.Context, userId int64) (*UserAppState, error)
	// Upsert 插入或更新当前小区（INSERT ... ON DUPLICATE KEY UPDATE）
	Upsert(ctx context.Context, userId, communityId int64) error
}

type defaultUserAppStateModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserAppStateModel(conn sqlx.SqlConn) UserAppStateModel {
	return &defaultUserAppStateModel{
		conn:  conn,
		table: "user_app_state",
	}
}

func (m *defaultUserAppStateModel) FindOne(ctx context.Context, userId int64) (*UserAppState, error) {
	query := fmt.Sprintf(`SELECT user_id, current_community_id, created_at, updated_at FROM %s WHERE user_id = ?`, m.table)
	var resp UserAppState
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserAppStateModel) Upsert(ctx context.Context, userId, communityId int64) error {
	query := fmt.Sprintf(`INSERT INTO %s (user_id, current_community_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE current_community_id = ?, updated_at = NOW()`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, communityId, communityId)
	return err
}
