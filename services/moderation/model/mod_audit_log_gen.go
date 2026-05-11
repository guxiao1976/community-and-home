package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	_ ModAuditLogModel = (*customModAuditLogModel)(nil)
)

type (
	ModAuditLog struct {
		Id            int64          `db:"id"`
		ContentType   string         `db:"content_type"`
		ContentSummary sql.NullString `db:"content_summary"`
		RiskLevel     string         `db:"risk_level"`
		Pass          int64          `db:"pass"`
		Reason        sql.NullString `db:"reason"`
		CheckLayer    sql.NullString `db:"check_layer"`
		MatchedItems  sql.NullString `db:"matched_items"`
		UserId        sql.NullInt64  `db:"user_id"`
		SourceType    sql.NullString `db:"source_type"`
		SourceId      sql.NullInt64  `db:"source_id"`
		NeedReview    int64          `db:"need_review"`
		ReviewStatus  int64          `db:"review_status"`
		ReviewerId    sql.NullInt64  `db:"reviewer_id"`
		ReviewTime    sql.NullTime   `db:"review_time"`
		CreatedTime   time.Time      `db:"created_time"`
	}

	ModAuditLogModel interface {
		Insert(ctx context.Context, data *ModAuditLog) (sql.Result, error)
	}
)

type customModAuditLogModel struct {
	*defaultModAuditLogModel
}

func NewModAuditLogModel(conn sqlx.SqlConn) ModAuditLogModel {
	return &customModAuditLogModel{
		defaultModAuditLogModel: newDefaultModAuditLogModel(conn),
	}
}

type defaultModAuditLogModel struct {
	conn     sqlx.SqlConn
	tableName string
}

func (m *defaultModAuditLogModel) getTableName() string {
	return m.tableName
}

func newDefaultModAuditLogModel(conn sqlx.SqlConn) *defaultModAuditLogModel {
	return &defaultModAuditLogModel{
		conn:     conn,
		tableName: "mod_audit_log",
	}
}

func (m *defaultModAuditLogModel) Insert(ctx context.Context, data *ModAuditLog) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (content_type, content_summary, risk_level, pass, reason, check_layer, matched_items, user_id, source_type, source_id, need_review, review_status, reviewer_id, review_time, created_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.getTableName())
	ret, err := m.conn.ExecCtx(ctx, query,
		data.ContentType, data.ContentSummary, data.RiskLevel, data.Pass,
		data.Reason, data.CheckLayer, data.MatchedItems,
		data.UserId, data.SourceType, data.SourceId,
		data.NeedReview, data.ReviewStatus,
		data.ReviewerId, data.ReviewTime, data.CreatedTime,
	)
	return ret, err
}
