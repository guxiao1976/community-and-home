package file

import (
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/model"
)

func toProtoFile(f *model.File) *filev1.FileInfo {
	return &filev1.FileInfo{
		Id:         f.ID,
		UserId:     f.UserID,
		EntityType: f.EntityType,
		EntityId:   f.EntityID,
		FileName:   f.FileName,
		FilePath:   f.FilePath,
		FileSize:   f.FileSize,
		MimeType:   f.MimeType,
		BucketName: f.BucketName,
		UploadTime: f.UploadTime.Unix(),
	}
}
