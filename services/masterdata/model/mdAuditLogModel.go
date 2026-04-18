package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdAuditLogModel = (*customMdAuditLogModel)(nil)

type (
	// MdAuditLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customMdAuditLogModel.
	MdAuditLogModel interface {
		mdAuditLogModel
	}

	customMdAuditLogModel struct {
		*defaultMdAuditLogModel
	}
)

// NewMdAuditLogModel returns a model for the database table.
func NewMdAuditLogModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdAuditLogModel {
	return &customMdAuditLogModel{
		defaultMdAuditLogModel: newMdAuditLogModel(conn, c, opts...),
	}
}
