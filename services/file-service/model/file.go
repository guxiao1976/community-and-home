package model

import (
	"time"

	"github.com/guxiao1976/community-common/v2/model"
)

// File 文件元数据
type File struct {
	model.BaseModel
	UserID     int64     `gorm:"column:user_id;index;not null" json:"user_id,string"`
	EntityType string    `gorm:"column:entity_type;size:64" json:"entity_type"`
	EntityID   int64     `gorm:"column:entity_id;index" json:"entity_id,string"`
	FileName   string    `gorm:"column:file_name;size:255;not null" json:"file_name"`
	FilePath   string    `gorm:"column:file_path;size:512;not null" json:"file_path"`
	FileSize   int64     `gorm:"column:file_size;not null" json:"file_size"`
	MimeType   string    `gorm:"column:mime_type;size:128" json:"mime_type"`
	BucketName string    `gorm:"column:bucket_name;size:64;not null;default:community-home" json:"bucket_name"`
	UploadTime time.Time `gorm:"column:upload_time;not null" json:"upload_time"`
	// FileType 白名单规范类型（magic-bytes 嗅探产出，如 png/pdf/doc） // SEE: [[edit-form-data-integrity]]
	FileType string `gorm:"column:file_type;size:20" json:"file_type"`
	// Confirmed 上传流程完成标志（ConfirmUpload 成功后置 true；存量行默认 true 免嗅探）
	Confirmed bool `gorm:"column:confirmed;not null;default:true" json:"confirmed"`
}

func (File) TableName() string {
	return "uploaded_file"
}
