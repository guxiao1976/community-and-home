package model

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SEE: [[testing-discipline]] — Model 层边界测试：锁定配额计数谓词与参数
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 谓词回归 RED：`type=?`/缺少 moderation_status IN (0,1) 时计数失准

// TestLostFoundItemModel_CountQuotaOccupied 锁定三维计数谓词：
//   - deleted_at IS NULL（软删除不计）
//   - status='active'（resolved 不计）
//   - moderation_status IN (0,1)（驳回=2 不计，待审=0 与通过=1 同占配额）
//
// 这些语义由谓词编码，sqlmock 以精确查询串断言锁定，防止将来被误改（如退化为 moderation_status=1）。
func TestLostFoundItemModel_CountQuotaOccupied(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewLostFoundItemModel(conn)

	mock.ExpectQuery("select count\\(\\*\\) from `lost_found_items` where publisher_id = \\? and community_id = \\? and deleted_at is null and status = 'active' and moderation_status in \\(0, 1\\)").
		WithArgs(int64(100), int64(200)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	got, err := m.CountQuotaOccupied(context.Background(), 100, 200, "lost_found")
	require.NoError(t, err)
	assert.Equal(t, int64(5), got)

	require.NoError(t, mock.ExpectationsWereMet())
}

// TestLostFoundItemModel_CountQuotaOccupied_DBError DB 瞬时错误向上传播（不吞错）。
func TestLostFoundItemModel_CountQuotaOccupied_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewLostFoundItemModel(conn)

	mock.ExpectQuery("select count\\(\\*\\) from `lost_found_items`").
		WithArgs(int64(100), int64(200)).
		WillReturnError(errors.New("db unavailable"))

	_, err = m.CountQuotaOccupied(context.Background(), 100, 200, "lost_found")
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
