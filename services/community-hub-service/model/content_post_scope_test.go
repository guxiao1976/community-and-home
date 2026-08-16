package model

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Task 1.3 测试：InsertBatch（正常 + 空列表 + 去重后单行）+ FindCommunityIdsByPostId + DeleteByPostId。
// SEE: [[snake-camel-field-mismatch]] — db tag 与 Go 字段 snake_case 对齐
func TestContentPostScopeModel_InsertBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostScopeModel(conn)

	mock.ExpectExec("insert into `content_post_scope` \\(post_id, community_id\\) values \\(\\?, \\?\\)").
		WithArgs(int64(1001), int64(2001)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("insert into `content_post_scope` \\(post_id, community_id\\) values \\(\\?, \\?\\)").
		WithArgs(int64(1001), int64(2002)).WillReturnResult(sqlmock.NewResult(1, 1))

	err = m.InsertBatch(context.Background(), 1001, []int64{2001, 2002})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostScopeModel_InsertBatch_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostScopeModel(conn)

	// 空列表为 no-op，不执行任何 SQL
	err = m.InsertBatch(context.Background(), 1001, nil)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostScopeModel_FindCommunityIdsByPostId(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostScopeModel(conn)

	mock.ExpectQuery("select community_id from `content_post_scope` where post_id = \\?").
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"community_id"}).AddRow(2001).AddRow(2002))

	ids, err := m.FindCommunityIdsByPostId(context.Background(), 1001)
	require.NoError(t, err)
	assert.Equal(t, []int64{2001, 2002}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostScopeModel_DeleteByPostId(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostScopeModel(conn)

	mock.ExpectExec("delete from `content_post_scope` where post_id = \\?").
		WithArgs(int64(1001)).WillReturnResult(sqlmock.NewResult(0, 2))

	err = m.DeleteByPostId(context.Background(), 1001)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
