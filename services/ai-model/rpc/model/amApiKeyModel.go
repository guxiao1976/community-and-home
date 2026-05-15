package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmApiKeyModel = (*customAmApiKeyModel)(nil)

type (
	// AmApiKeyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmApiKeyModel.
	AmApiKeyModel interface {
		amApiKeyModel
		FindOneByModelId(ctx context.Context, modelId int64) (*AmApiKey, error)
	}

	customAmApiKeyModel struct {
		*defaultAmApiKeyModel
	}
)

// NewAmApiKeyModel returns a model for the database table.
func NewAmApiKeyModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmApiKeyModel {
	return &customAmApiKeyModel{
		defaultAmApiKeyModel: newAmApiKeyModel(conn, c, opts...),
	}
}

// FindOneByModelId 根据模型ID查找第一个可用的API Key
func (m *customAmApiKeyModel) FindOneByModelId(ctx context.Context, modelId int64) (*AmApiKey, error) {
	query := fmt.Sprintf("select %s from %s where `model_id` = ? and `status` = 1 and `delete_time` is null limit 1", amApiKeyRows, m.table)
	var resp AmApiKey
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, modelId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
