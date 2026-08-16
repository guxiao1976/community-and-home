package file

import (
	"io"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-file/internal/guard"
	"github.com/guxiao1976/community-file/model"
)

// sniffPrefixBytes 回读对象前缀字节数（须足够容纳 doc 的 WordDocument 流名与 docx 的部件路径）
const sniffPrefixBytes = 64 * 1024

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
		FileType:   f.FileType,
		Confirmed:  f.Confirmed,
	}
}

// verifySniffedContent L2 回读校验：magic-bytes 嗅探实际对象，与声明扩展名一致才放行。
// 返回嗅探的规范扩展名（用于落库 file_type）；不一致/未识别 → 070004（REQ-CAS-3 L2）。
func verifySniffedContent(declaredFileName string, r io.Reader) (string, error) {
	declaredExt, err := guard.ValidateFileName(declaredFileName)
	if err != nil {
		return "", err
	}

	prefix, err := readObjectPrefix(r)
	if err != nil {
		return "", errx.NewCodeError(ErrCodeFileOperationFailed, "read uploaded object failed")
	}

	sniffedExt, ok := guard.SniffType(prefix)
	if !ok || !extMatch(sniffedExt, declaredExt) {
		return "", errx.NewCodeError(ErrCodeUnsupportedFileType, "file content does not match declared type")
	}
	return sniffedExt, nil
}

// readObjectPrefix 读取 reader 前 sniffPrefixBytes 字节（对象不足则取全部）。
func readObjectPrefix(r io.Reader) ([]byte, error) {
	buf := make([]byte, sniffPrefixBytes)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, err
	}
	return buf[:n], nil
}

// extMatch 声明扩展名与嗅探类型是否一致（jpg/jpeg 视为等价：同为 JPEG 实际格式）。
func extMatch(a, b string) bool {
	if a == b {
		return true
	}
	return (a == "jpg" || a == "jpeg") && (b == "jpg" || b == "jpeg")
}
