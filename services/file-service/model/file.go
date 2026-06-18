package model

import (
	"time"

	"github.com/guxiao1976/community-common/v2/model"
)

// File 文件元数据
type File struct {
	model.BaseModel
	UserID     int64     `gorm:"column:user_id;index;not null" json:"user_id"`
	EntityType string    `gorm:"column:entity_type;size:64" json:"entity_type"`
	EntityID   int64     `gorm:"column:entity_id;index" json:"entity_id"`
	FileName   string    `gorm:"column:file_name;size:255;not null" json:"file_name"`
	FilePath   string    `gorm:"column:file_path;size:512;not null" json:"file_path"`
	FileSize   int64     `gorm:"column:file_size;not null" json:"file_size"`
	MimeType   string    `gorm:"column:mime_type;size:128" json:"mime_type"`
	BucketName string    `gorm:"column:bucket_name;size:64;not null;default:community-home" json:"bucket_name"`
	UploadTime time.Time `gorm:"column:upload_time;not null" json:"upload_time"`
}

func (File) TableName() string {
	return "uploaded_file"
}
