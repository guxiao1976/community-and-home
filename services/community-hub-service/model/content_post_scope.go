package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ContentPostScope 内容帖-小区范围关联（多小区发布单源，Migration 003 新建表）。
// 复合 PK (post_id, community_id)：同一小区同帖唯一；纯关联表仅 created_at
// （显式偏离编码规范 §3.1 时间三件套——行不可变，撤回由主表软删表达，draft 改写走物理替换）。
type ContentPostScope struct {
	PostId      int64     `db:"post_id"`
	CommunityId int64     `db:"community_id"`
	CreatedAt   time.Time `db:"created_at"`
}

// ContentPostScopeModel 内容帖-小区范围关联数据访问层
type ContentPostScopeModel interface {
	// InsertBatch 批量插入（业务层先 dedupe；空列表为 no-op 返回 nil）。
	InsertBatch(ctx context.Context, postId int64, communityIds []int64) error
	// InsertBatchTx 事务内批量插入（Create 单事务落库经共享 session）。
	InsertBatchTx(ctx context.Context, session sqlx.Session, postId int64, communityIds []int64) error
	// FindCommunityIdsByPostId 按帖查目标小区集（详情 scope 反查 / 兼容回退）。
	FindCommunityIdsByPostId(ctx context.Context, postId int64) ([]int64, error)
	// DeleteByPostId 按帖删除（draft 编辑 scope 全量重写）。
	DeleteByPostId(ctx context.Context, postId int64) error
	// DeleteByPostIdTx 事务内按帖删除（draft 编辑 scope 重写经共享 session）。
	DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error
}

type defaultContentPostScopeModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewContentPostScopeModel(conn sqlx.SqlConn) ContentPostScopeModel {
	return &defaultContentPostScopeModel{conn: conn, table: "`content_post_scope`"}
}

// execer 兼容 sqlx.SqlConn 与 sqlx.Session（均具 ExecCtx，返回 sql.Result）。
type execer interface {
	ExecCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (m *defaultContentPostScopeModel) InsertBatch(ctx context.Context, postId int64, communityIds []int64) error {
	return m.insertBatch(ctx, m.conn, postId, communityIds)
}

func (m *defaultContentPostScopeModel) InsertBatchTx(ctx context.Context, session sqlx.Session, postId int64, communityIds []int64) error {
	return m.insertBatch(ctx, session, postId, communityIds)
}

func (m *defaultContentPostScopeModel) insertBatch(ctx context.Context, e execer, postId int64, communityIds []int64) error {
	if len(communityIds) == 0 {
		return nil
	}
	query := `insert into ` + m.table + ` (post_id, community_id) values (?, ?)`
	for _, cid := range communityIds {
		if _, err := e.ExecCtx(ctx, query, postId, cid); err != nil {
			return err
		}
	}
	return nil
}

// scopeCommunityRow 仅承载 community_id 单列投影（go-zero sqlx 需与返回列一一对应）。
type scopeCommunityRow struct {
	CommunityId int64 `db:"community_id"`
}

func (m *defaultContentPostScopeModel) FindCommunityIdsByPostId(ctx context.Context, postId int64) ([]int64, error) {
	query := `select community_id from ` + m.table + ` where post_id = ?`
	var rows []scopeCommunityRow
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, postId); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.CommunityId)
	}
	return ids, nil
}

func (m *defaultContentPostScopeModel) DeleteByPostId(ctx context.Context, postId int64) error {
	return m.deleteByPostId(ctx, m.conn, postId)
}

func (m *defaultContentPostScopeModel) DeleteByPostIdTx(ctx context.Context, session sqlx.Session, postId int64) error {
	return m.deleteByPostId(ctx, session, postId)
}

func (m *defaultContentPostScopeModel) deleteByPostId(ctx context.Context, e execer, postId int64) error {
	query := `delete from ` + m.table + ` where post_id = ?`
	_, err := e.ExecCtx(ctx, query, postId)
	return err
}
