package file

import (
	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/api/internal/types"
)

// toFileInfo 将 proto FileInfo 转换为 API 类型
func toFileInfo(pb *filev1.FileInfo) *types.FileInfo {
	if pb == nil {
		return nil
	}
	return &types.FileInfo{
		Id:         pb.Id,
		UserId:     pb.UserId,
		EntityType: pb.EntityType,
		EntityId:   pb.EntityId,
		FileName:   pb.FileName,
		FilePath:   pb.FilePath,
		FileSize:   pb.FileSize,
		MimeType:   pb.MimeType,
		BucketName: pb.BucketName,
		UploadTime: pb.UploadTime,
	}
}
