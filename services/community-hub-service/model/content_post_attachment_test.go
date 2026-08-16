package model

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Task 1.4 测试：InsertBatch / FindByPostId / DeleteByPostId。
func TestContentPostAttachmentModel_InsertBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostAttachmentModel(conn)

	ft := "pdf"
	atts := []*ContentPostAttachment{
		{Id: 11, PostId: 1001, FileName: "a.pdf", FileUrl: "", FileSize: 1024, ReviewStatus: AttachmentReviewApproved, FileId: 5001, FileType: &ft},
		{Id: 12, PostId: 1001, FileName: "b.png", FileUrl: "", FileSize: 2048, ReviewStatus: AttachmentReviewApproved, FileId: 5002, FileType: nil},
	}

	mock.ExpectExec("insert into `content_post_attachments` \\(id, post_id, file_name, file_url, file_size, review_status, file_id, file_type\\) values \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)").
		WithArgs(int64(11), int64(1001), "a.pdf", "", int64(1024), int64(1), int64(5001), "pdf").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("insert into `content_post_attachments` \\(id, post_id, file_name, file_url, file_size, review_status, file_id, file_type\\) values \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)").
		WithArgs(int64(12), int64(1001), "b.png", "", int64(2048), int64(1), int64(5002), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = m.InsertBatch(context.Background(), atts)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostAttachmentModel_FindByPostId(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostAttachmentModel(conn)

	mock.ExpectQuery("select \\* from `content_post_attachments` where post_id = \\?").
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "post_id", "file_name", "file_url", "file_size", "review_status", "file_id", "file_type", "created_at"}).
			AddRow(11, 1001, "a.pdf", "", 1024, 1, 5001, "pdf", time.Now()))

	list, err := m.FindByPostId(context.Background(), 1001)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, int64(5001), list[0].FileId)
	assert.Equal(t, int64(1), list[0].ReviewStatus)
	require.NotNil(t, list[0].FileType)
	assert.Equal(t, "pdf", *list[0].FileType)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostAttachmentModel_DeleteByPostId(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostAttachmentModel(conn)

	mock.ExpectExec("delete from `content_post_attachments` where post_id = \\?").
		WithArgs(int64(1001)).WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.DeleteByPostId(context.Background(), 1001)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
