package guard

import (
	"testing"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 2.2 ValidateFileName — 白名单/禁止集/zip/rar/大小写/无扩展名
// =============================================================================

func TestValidateFileName_WhitelistPass(t *testing.T) {
	for _, name := range []string{"a.png", "b.jpg", "c.jpeg", "d.gif", "e.pdf", "f.doc", "g.docx"} {
		t.Run(name, func(t *testing.T) {
			ext, err := ValidateFileName(name)
			require.NoError(t, err)
			assert.NotEmpty(t, ext)
		})
	}
}

func TestValidateFileName_ForbiddenRejected(t *testing.T) {
	for _, name := range []string{"a.exe", "b.sh", "c.js", "d.bat", "e.msi", "f.apk", "g.php", "h.py"} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateFileName(name)
			require.Error(t, err)
			assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "禁止集 → 070004: %v", err)
		})
	}
}

func TestValidateFileName_ZipRarRejected(t *testing.T) {
	for _, name := range []string{"a.zip", "b.rar"} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateFileName(name)
			require.Error(t, err)
			assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "zip/rar → 070004: %v", err)
		})
	}
}

func TestValidateFileName_CaseInsensitive(t *testing.T) {
	ext, err := ValidateFileName("PHOTO.JPG")
	require.NoError(t, err)
	assert.Equal(t, "jpg", ext, "扩展名大小写不敏感，返回规范小写")

	_, err = ValidateFileName("EVIL.EXE")
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "大写 EXE 仍拒绝 → 070004")
}

func TestValidateFileName_NoExtensionOrDotfileRejected(t *testing.T) {
	for _, name := range []string{"README", ".gitignore", "archive", "file."} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateFileName(name)
			require.Error(t, err)
			assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "无扩展名/点文件 → 070004: %v", err)
		})
	}
}

func TestValidateFileName_UnknownExtensionRejected(t *testing.T) {
	for _, name := range []string{"a.svg", "b.webp", "c.txt", "d.json"} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateFileName(name)
			require.Error(t, err)
			assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "非白名单 → 070004: %v", err)
		})
	}
}

// =============================================================================
// Task 2.2 ValidateFileSize — 10MB 硬上限（=10MB 放行）
// =============================================================================

func TestValidateFileSize_Exceeded(t *testing.T) {
	err := ValidateFileSize(10*1024*1024 + 1)
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeFileSizeExceeded), ">10MB → 070005")
}

func TestValidateFileSize_ExactlyLimitPass(t *testing.T) {
	err := ValidateFileSize(10 * 1024 * 1024)
	assert.NoError(t, err, "=10MB 放行")
}

func TestValidateFileSize_SmallPass(t *testing.T) {
	assert.NoError(t, ValidateFileSize(1024))
}

// =============================================================================
// Task 2.6 entity_type 覆盖机制（REQ-CAS-4：基线上叠加，禁止集/10MB 不可弱化）
// =============================================================================

// registerOverride 注册 entity_type 覆盖并在测试后清理（避免污染其他用例）。
func registerOverride(t *testing.T, entityType string, cfg Override) {
	t.Helper()
	RegisterEntityTypeOverride(entityType, cfg)
	t.Cleanup(func() {
		overrideMu.Lock()
		delete(overrides, entityType)
		overrideMu.Unlock()
	})
}

func TestOverride_AppendLegalType_OnlyForThatEntity(t *testing.T) {
	registerOverride(t, "id_card", Override{Extensions: []string{"webp", "svg"}})

	// override 生效：该 entity_type 下新类型放行
	ext, err := ValidateFileNameForEntityType("card.webp", "id_card")
	require.NoError(t, err)
	assert.Equal(t, "webp", ext)

	// 全局基线不受影响：无 override 时新类型仍拒绝
	_, err = ValidateFileName("card.webp")
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "全局基线不因 override 扩大")

	// 其他 entity_type 不受影响（基线默认）
	_, err = ValidateFileNameForEntityType("card.webp", "verification")
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType))
}

func TestOverride_CannotRelaxSizeHardCap(t *testing.T) {
	registerOverride(t, "video_upload", Override{Extensions: []string{"mp4"}})
	// 10MB 硬上限不可放宽：即使存在 override，>10MB 仍 070005
	err := ValidateFileSize(10*1024*1024 + 1)
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeFileSizeExceeded), "override 不可放宽 10MB → 070005")
}

func TestOverride_CannotAllowForbiddenExt(t *testing.T) {
	registerOverride(t, "evil_upload", Override{Extensions: []string{"exe", "msi", "sh"}})
	// 禁止集不可放行：即使 override 显式加入，仍 070004
	for _, name := range []string{"a.exe", "b.msi", "c.sh"} {
		_, err := ValidateFileNameForEntityType(name, "evil_upload")
		require.Error(t, err)
		assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "禁止集不可被 override 放行: %s", name)
	}
}

func TestOverride_BaselineDefaults_ExistingUploadTypesNoRegression(t *testing.T) {
	// 既有 avatar/verification/lostfound/contacts 上传不回归（无 override → 全局基线）
	for _, entityType := range []string{"avatar", "verification", "lostfound", "contacts"} {
		ext, err := ValidateFileNameForEntityType("photo.png", entityType)
		require.NoError(t, err)
		assert.Equal(t, "png", ext)
		_, err = ValidateFileNameForEntityType("evil.exe", entityType)
		require.Error(t, err)
	}
}
