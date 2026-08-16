package logic

// file-service API 层错误码定义
const (
	// ErrCodeFileNotFound 文件不存在
	ErrCodeFileNotFound = 70001
	// ErrCodeFileAccessDenied 文件访问被拒绝
	ErrCodeFileAccessDenied = 70002
	// ErrCodeFileOperationFailed 文件操作失败
	ErrCodeFileOperationFailed = 70003
	// ErrCodeUnsupportedFileType 不支持的文件类型
	ErrCodeUnsupportedFileType = 70004
	// ErrCodeFileSizeExceeded 文件大小超限
	ErrCodeFileSizeExceeded = 70005
)
