package guard

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
)

// =============================================================================
// 附件安全守卫（content-post-generalization）
//
// L1 扩展名/大小快速拒绝 + L2 magic-bytes 内容嗅探共用的通用校验器。
// 本包位于 internal/guard，独立于 rpc/internal 的错误码常量（Go internal 规则
// 限制 rpc/internal 仅可被 rpc/ 子树导入），故在包内以命名常量重述 070004/070005，
// 与 rpc/internal/errx、rpc/internal/logic/file/errcode.go 同值同语义（不重编号）。
// =============================================================================

// file-service 错误码（与 rpc/internal/errx 同值同语义，禁止裸数字） // SEE: [[error-code-literal-bypasses-qa-gate]]
const (
	// ErrCodeUnsupportedFileType 不支持的文件类型（070004）
	ErrCodeUnsupportedFileType = 70004
	// ErrCodeFileSizeExceeded 文件大小超限（070005）
	ErrCodeFileSizeExceeded = 70005
)

// MaxSingleFileSize 单文件硬上限（10MB，=10MB 放行，override 不可放宽） // SEE: [[frontend-business-rule-hardcode]]
const MaxSingleFileSize = 10 * 1024 * 1024

// whitelist 全局基线白名单
var whitelist = map[string]struct{}{
	"png":  {},
	"jpg":  {},
	"jpeg": {},
	"gif":  {},
	"pdf":  {},
	"doc":  {},
	"docx": {},
}

// forbidden 全局禁止集（无论 override 如何扩展均不可放行）
var forbidden = map[string]struct{}{
	"exe": {}, "bat": {}, "sh": {}, "cmd": {}, "com": {}, "msi": {}, "apk": {},
	"js": {}, "vbs": {}, "ps1": {}, "py": {}, "pl": {}, "php": {},
}

// zip/rar 扩展名层全部拒绝（REQ-CAS-1）
var archiveBlocked = map[string]struct{}{
	"zip": {},
	"rar": {},
}

// =============================================================================
// entity_type 覆盖机制（REQ-CAS-4：基线上叠加，禁止集/10MB 不可弱化）
// =============================================================================

// Override 按 entity_type 的覆盖配置：仅追加允许类型，不可放行禁止集、不可放宽 10MB。
type Override struct {
	// Extensions 追加允许的文件扩展名（在全局基线之上叠加）
	Extensions []string
}

var (
	overrideMu sync.RWMutex
	overrides  = map[string]*Override{}
)

// RegisterEntityTypeOverride 注册 entity_type 的覆盖配置（追加允许类型）。
// 禁止集不可放行、10MB 硬上限不可放宽（REQ-CAS-4 不变量）。
func RegisterEntityTypeOverride(entityType string, cfg Override) {
	overrideMu.Lock()
	defer overrideMu.Unlock()
	copied := make([]string, len(cfg.Extensions))
	copy(copied, cfg.Extensions)
	overrides[entityType] = &Override{Extensions: copied}
}

// allowedExtensions 返回全局基线（+entity_type 覆盖叠加）的允许扩展名集合。
func allowedExtensions(entityType string) map[string]struct{} {
	overrideMu.RLock()
	defer overrideMu.RUnlock()

	set := make(map[string]struct{}, len(whitelist)+8)
	for ext := range whitelist {
		set[ext] = struct{}{}
	}
	if o, ok := overrides[entityType]; ok {
		for _, ext := range o.Extensions {
			ext = strings.ToLower(strings.TrimSpace(ext))
			ext = strings.TrimPrefix(ext, ".")
			if ext == "" {
				continue
			}
			set[ext] = struct{}{}
		}
	}
	return set
}

// =============================================================================
// 校验器
// =============================================================================

// ValidateFileName 校验文件名扩展名（全局基线）：返回规范小写扩展名。
// 无扩展名/点文件/非白名单/禁止集/zip/rar → ErrCodeUnsupportedFileType。
func ValidateFileName(fileName string) (string, error) {
	return validateFileName(fileName, allowedExtensions(""))
}

// ValidateFileNameForEntityType 按 entity_type 覆盖后的扩展名校验（基线上叠加）。
func ValidateFileNameForEntityType(fileName, entityType string) (string, error) {
	return validateFileName(fileName, allowedExtensions(entityType))
}

func validateFileName(fileName string, allowed map[string]struct{}) (string, error) {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return "", errx.NewCodeError(ErrCodeUnsupportedFileType, "unsupported file type: missing extension")
	}
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	// 禁止集优先于一切（override 不可放行）
	if _, ok := forbidden[ext]; ok {
		return "", errx.NewCodeError(ErrCodeUnsupportedFileType, fmt.Sprintf("unsupported file type: .%s is forbidden", ext))
	}
	if _, ok := archiveBlocked[ext]; ok {
		return "", errx.NewCodeError(ErrCodeUnsupportedFileType, fmt.Sprintf("unsupported file type: .%s not allowed", ext))
	}
	if _, ok := allowed[ext]; !ok {
		return "", errx.NewCodeError(ErrCodeUnsupportedFileType, fmt.Sprintf("unsupported file type: .%s", ext))
	}
	return ext, nil
}

// ValidateFileSize 校验文件大小（10MB 硬上限，=10MB 放行）。override 不可放宽。
func ValidateFileSize(size int64) error {
	if size > MaxSingleFileSize {
		return errx.NewCodeError(ErrCodeFileSizeExceeded, fmt.Sprintf("file size %d exceeds %d-byte limit", size, MaxSingleFileSize))
	}
	return nil
}
