package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthFamilyMemberModel = (*customAuthFamilyMemberModel)(nil)

type (
	AuthFamilyMemberModel interface {
		authFamilyMemberModel
		FindByFamilyId(ctx context.Context, familyId int64) ([]*AuthFamilyMember, error)
		CountByFamilyId(ctx context.Context, familyId int64) (int, error)
	}

	customAuthFamilyMemberModel struct {
		*defaultAuthFamilyMemberModel
	}
)

func NewAuthFamilyMemberModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthFamilyMemberModel {
	return &customAuthFamilyMemberModel{
		defaultAuthFamilyMemberModel: newAuthFamilyMemberModel(conn, c, opts...),
	}
}

func (m *customAuthFamilyMemberModel) FindByFamilyId(ctx context.Context, familyId int64) ([]*AuthFamilyMember, error) {
	query := fmt.Sprintf("select %s from %s where family_id = ? order by created_time asc", authFamilyMemberRows, m.table)
	var members []*AuthFamilyMember
	err := m.QueryRowsNoCacheCtx(ctx, &members, query, familyId)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (m *customAuthFamilyMemberModel) CountByFamilyId(ctx context.Context, familyId int64) (int, error) {
	query := fmt.Sprintf("select count(*) from %s where family_id = ?", m.table)
	var count int
	err := m.QueryRowNoCacheCtx(ctx, &count, query, familyId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
