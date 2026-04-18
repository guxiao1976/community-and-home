package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdCommunityModel = (*customMdCommunityModel)(nil)

type (
	MdCommunityModel interface {
		mdCommunityModel
		FindByDivisionId(ctx context.Context, divisionId int64) ([]*MdCommunity, error)
		FindAll(ctx context.Context) ([]*MdCommunity, error)
	}

	customMdCommunityModel struct {
		*defaultMdCommunityModel
	}
)

func NewMdCommunityModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdCommunityModel {
	return &customMdCommunityModel{
		defaultMdCommunityModel: newMdCommunityModel(conn, c, opts...),
	}
}

func (m *customMdCommunityModel) FindByDivisionId(ctx context.Context, divisionId int64) ([]*MdCommunity, error) {
	var communities []*MdCommunity
	query := fmt.Sprintf("select %s from %s where division_id = ? and delete_time is null order by id desc", mdCommunityRows, m.table)
	err := m.QueryRowsNoCache(&communities, query, divisionId)
	if err != nil {
		return nil, err
	}
	return communities, nil
}

func (m *customMdCommunityModel) FindAll(ctx context.Context) ([]*MdCommunity, error) {
	var communities []*MdCommunity
	query := fmt.Sprintf("select %s from %s where delete_time is null order by id desc", mdCommunityRows, m.table)
	err := m.QueryRowsNoCache(&communities, query)
	if err != nil {
		return nil, err
	}
	return communities, nil
}