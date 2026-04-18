package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AuthUploadedFileModel = (*customAuthUploadedFileModel)(nil)

type (
	// AuthUploadedFileModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAuthUploadedFileModel.
	AuthUploadedFileModel interface {
		authUploadedFileModel
		BatchInsert(ctx context.Context, files []*AuthUploadedFile) error
		FindByEntityId(ctx context.Context, entityType string, entityId int64) ([]*AuthUploadedFile, error)
	}

	customAuthUploadedFileModel struct {
		*defaultAuthUploadedFileModel
	}
)

// NewAuthUploadedFileModel returns a model for the database table.
func NewAuthUploadedFileModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthUploadedFileModel {
	return &customAuthUploadedFileModel{
		defaultAuthUploadedFileModel: newAuthUploadedFileModel(conn, c, opts...),
	}
}

func (m *customAuthUploadedFileModel) BatchInsert(ctx context.Context, files []*AuthUploadedFile) error {
	if len(files) == 0 {
		return nil
	}

	var valueStrings []string
	var valueArgs []interface{}

	for _, file := range files {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs, file.UserId, file.EntityType, file.EntityId, file.FileName, file.FilePath, file.FileSize, file.FileType, file.BucketName, time.Now())
	}

	query := fmt.Sprintf("INSERT INTO %s (user_id, entity_type, entity_id, file_name, file_path, file_size, file_type, bucket_name, created_time) VALUES %s",
		m.table, strings.Join(valueStrings, ","))

	_, err := m.ExecNoCacheCtx(ctx, query, valueArgs...)
	return err
}

func (m *customAuthUploadedFileModel) FindByEntityId(ctx context.Context, entityType string, entityId int64) ([]*AuthUploadedFile, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE entity_type = ? AND entity_id = ? AND delete_time IS NULL ORDER BY created_time ASC", authUploadedFileRows, m.table)
	var resp []*AuthUploadedFile
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, entityType, entityId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
