package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FileModel interface {
	Insert(ctx context.Context, f *File) (int64, error)
	FindOne(ctx context.Context, id int64) (*File, error)
	FindByIds(ctx context.Context, ids []int64) ([]*File, error)
	FindPage(ctx context.Context, userID *int64, entityType *string, entityID *int64, page, pageSize int64) ([]*File, int64, error)
	Delete(ctx context.Context, id int64) error
}

type defaultFileModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewFileModel(conn sqlx.SqlConn) FileModel {
	return &defaultFileModel{conn: conn, table: "uploaded_file"}
}

func (m *defaultFileModel) Insert(ctx context.Context, f *File) (int64, error) {
	query := fmt.Sprintf(
		"insert into %s (user_id, entity_type, entity_id, file_name, file_path, file_size, mime_type, bucket_name, upload_time, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, now(), now())",
		m.table,
	)
	res, err := m.conn.ExecCtx(ctx, query, f.UserID, f.EntityType, f.EntityID, f.FileName, f.FilePath, f.FileSize, f.MimeType, f.BucketName, f.UploadTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultFileModel) FindOne(ctx context.Context, id int64) (*File, error) {
	var f File
	query := fmt.Sprintf("select * from %s where id = ? and is_deleted = 0 limit 1", m.table)
	err := m.conn.QueryRowCtx(ctx, &f, query, id)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (m *defaultFileModel) FindByIds(ctx context.Context, ids []int64) ([]*File, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("select * from %s where id in (%s) and is_deleted = 0", m.table, strings.Join(placeholders, ","))
	var files []*File
	err := m.conn.QueryRowsCtx(ctx, &files, query, args...)
	return files, err
}

func (m *defaultFileModel) FindPage(ctx context.Context, userID *int64, entityType *string, entityID *int64, page, pageSize int64) ([]*File, int64, error) {
	where := []string{"is_deleted = 0"}
	args := make([]any, 0)

	if userID != nil {
		where = append(where, "user_id = ?")
		args = append(args, *userID)
	}
	if entityType != nil && *entityType != "" {
		where = append(where, "entity_type = ?")
		args = append(args, *entityType)
	}
	if entityID != nil {
		where = append(where, "entity_id = ?")
		args = append(args, *entityID)
	}
	whereClause := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, whereClause)
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("select * from %s where %s order by id desc limit %d offset %d", m.table, whereClause, pageSize, offset)
	var files []*File
	err := m.conn.QueryRowsCtx(ctx, &files, dataQuery, args...)
	return files, total, err
}

func (m *defaultFileModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("update %s set is_deleted = 1, updated_at = now() where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
