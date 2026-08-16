package model

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// fileCols 对齐 File struct 字段序（本模型无 db tag → go-zero sqlx 按位置扫描）。
// 注意：生产表列序为 001 定义序（id,user_id,...,created_time,updated_time），
// 与 struct 字段序不同属既有历史问题，不在本次范围；测试以 struct 字段序返回以验证扫描正确。
func fileCols() []string {
	return []string{
		"id", "created_at", "updated_at", "is_deleted", // BaseModel
		"user_id", "entity_type", "entity_id", "file_name", "file_path",
		"file_size", "mime_type", "bucket_name", "upload_time", "file_type", "confirmed",
	}
}

func fileRow(id, userID int64, fileType string, confirmed bool) []driver.Value {
	now := time.Now()
	confirmedV := int64(0)
	if confirmed {
		confirmedV = int64(1)
	}
	return []driver.Value{
		id, now, now, int64(0),
		userID, "avatar", int64(0), "a.png", "uploads/100/x_a.png",
		int64(1024), "image/png", "community-home", now, fileType, confirmedV,
	}
}

func TestFileModel_Insert_IncludesFileTypeConfirmed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewFileModel(conn)

	f := &File{
		UserID:     100,
		EntityType: "avatar",
		FileName:   "a.png",
		FilePath:   "uploads/100/x_a.png",
		FileSize:   1024,
		MimeType:   "image/png",
		BucketName: "community-home",
		UploadTime: time.Now(),
		FileType:   "png",
		Confirmed:  true,
	}

	mock.ExpectExec("insert into uploaded_file \\(user_id, entity_type, entity_id, file_name, file_path, file_size, mime_type, bucket_name, upload_time, file_type, confirmed, created_at, updated_at\\) values \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, now\\(\\), now\\(\\)\\)").
		WithArgs(f.UserID, f.EntityType, f.EntityID, f.FileName, f.FilePath, f.FileSize, f.MimeType, f.BucketName, f.UploadTime, f.FileType, f.Confirmed).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := m.Insert(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileModel_FindOne_ReadsBackFileTypeConfirmed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewFileModel(conn)

	mock.ExpectQuery("select \\* from uploaded_file where id = \\? and is_deleted = 0 limit 1").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(fileCols()).AddRow(fileRow(1, 100, "png", true)...))

	f, err := m.FindOne(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, f)
	assert.Equal(t, "png", f.FileType, "Insert 后读回 file_type")
	assert.True(t, f.Confirmed, "Insert 后读回 confirmed")
	assert.Equal(t, int64(1), f.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
