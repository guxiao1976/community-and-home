package types

// =============================================================================
// 获取上传 URL
// =============================================================================

// GetUploadUrlReq 获取预签名上传 URL 请求
type GetUploadUrlReq struct {
	EntityType string `json:"entityType"`
	FileName   string `json:"fileName"`
	MimeType   string `json:"mimeType"`
	FileSize   int64  `json:"fileSize,string"`
}

// GetUploadUrlResp 获取预签名上传 URL 响应
type GetUploadUrlResp struct {
	UploadUrl string `json:"uploadUrl"`
	ObjectKey string `json:"objectKey"`
	ExpireAt  int64  `json:"expireAt,string"`
}

// =============================================================================
// 确认上传
// =============================================================================

// ConfirmUploadReq 确认上传完成请求
type ConfirmUploadReq struct {
	ObjectKey  string `json:"objectKey"`
	EntityType string `json:"entityType"`
	EntityId   int64  `json:"entityId,string"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize,string"`
	MimeType   string `json:"mimeType"`
}

// =============================================================================
// 文件信息
// =============================================================================

// FileInfo 文件元数据信息
type FileInfo struct {
	Id         int64  `json:"id,string"`
	UserId     int64  `json:"userId,string"`
	EntityType string `json:"entityType"`
	EntityId   int64  `json:"entityId,string"`
	FileName   string `json:"fileName"`
	FilePath   string `json:"filePath"`
	FileSize   int64  `json:"fileSize,string"`
	MimeType   string `json:"mimeType"`
	BucketName string `json:"bucketName"`
	UploadTime int64  `json:"uploadTime,string"`
}

// ConfirmUploadResp 确认上传完成响应
type ConfirmUploadResp struct {
	File *FileInfo `json:"file"`
}

// =============================================================================
// 获取文件 URL
// =============================================================================

// GetFileUrlReq 获取文件下载 URL 请求（从路径参数解析）
type GetFileUrlReq struct {
	Id int64 `path:"id"`
}

// GetFileUrlResp 获取文件下载 URL 响应
type GetFileUrlResp struct {
	DownloadUrl string    `json:"downloadUrl"`
	File        *FileInfo `json:"file"`
}

// =============================================================================
// 删除文件
// =============================================================================

// DeleteFileReq 删除文件请求（从路径参数解析）
type DeleteFileReq struct {
	Id int64 `path:"id"`
}

// =============================================================================
// 文件列表
// =============================================================================

// ListFilesReq 文件列表请求（从查询参数解析）
type ListFilesReq struct {
	Page       int32   `form:"page,optional,default=1"`
	PageSize   int32   `form:"pageSize,optional,default=20"`
	UserId     *int64  `form:"userId,optional"`
	EntityType *string `form:"entityType,optional"`
	EntityId   *int64  `form:"entityId,optional"`
}

// PageInfo 分页信息
type PageInfo struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int32 `json:"totalPages"`
}

// ListFilesResp 文件列表响应
type ListFilesResp struct {
	Files []*FileInfo `json:"files"`
	Page  *PageInfo   `json:"page"`
}
