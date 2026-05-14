package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmPromptTemplateModel = (*customAmPromptTemplateModel)(nil)

type (
	// AmPromptTemplateModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmPromptTemplateModel.
	AmPromptTemplateModel interface {
		amPromptTemplateModel
	}

	customAmPromptTemplateModel struct {
		*defaultAmPromptTemplateModel
	}
)

// NewAmPromptTemplateModel returns a model for the database table.
func NewAmPromptTemplateModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmPromptTemplateModel {
	return &customAmPromptTemplateModel{
		defaultAmPromptTemplateModel: newAmPromptTemplateModel(conn, c, opts...),
	}
}
