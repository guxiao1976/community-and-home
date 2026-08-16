package file

import (
	"context"
	"testing"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-file/internal/guard"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uploadCtx 构造 svcCtx（RawMinio=nil → 校验通过后走 70002 分支，证明已越过 L1）。
func uploadCtx() *svc.ServiceContext {
	return &svc.ServiceContext{}
}

func TestGetUploadUrl_L1RejectsBadExtension(t *testing.T) {
	l := NewGetUploadUrlLogic(context.Background(), uploadCtx())
	_, err := l.GetUploadUrl(&filev1.GetUploadUrlRequest{
		UserId:   100,
		FileName: "evil.exe",
		FileSize: 1024,
	})
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "exe → 070004，未触碰 MinIO")
}

func TestGetUploadUrl_L1RejectsZipRar(t *testing.T) {
	for _, name := range []string{"a.zip", "b.rar"} {
		l := NewGetUploadUrlLogic(context.Background(), uploadCtx())
		_, err := l.GetUploadUrl(&filev1.GetUploadUrlRequest{UserId: 100, FileName: name, FileSize: 1024})
		require.Error(t, err)
		assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "%s → 070004", name)
	}
}

func TestGetUploadUrl_L1RejectsOversize(t *testing.T) {
	l := NewGetUploadUrlLogic(context.Background(), uploadCtx())
	_, err := l.GetUploadUrl(&filev1.GetUploadUrlRequest{
		UserId:   100,
		FileName: "big.pdf",
		FileSize: 10*1024*1024 + 1,
	})
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeFileSizeExceeded), ">10MB → 070005")
}

func TestGetUploadUrl_L1PassThenMinioNotAvailable(t *testing.T) {
	// 校验通过 → 进入 MinIO 分支（RawMinio=nil → 70002），证明 L1 放行且未在 L1 被拒
	l := NewGetUploadUrlLogic(context.Background(), uploadCtx())
	_, err := l.GetUploadUrl(&filev1.GetUploadUrlRequest{
		UserId:   100,
		FileName: "ok.png",
		FileSize: 10 * 1024 * 1024, // 恰 10MB 放行
	})
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeFileAccessDenied), "L1 放行后进入 MinIO 检查 → 70002")
}

// =============================================================================
// Task 2.6 entity_type 覆盖接线（GetUploadUrl 按 in.EntityType 查覆盖）
// =============================================================================

func TestGetUploadUrl_EntityOverrideApplied(t *testing.T) {
	// 注册 "id_card" 覆盖（本包唯一 entity_type，无其他用例依赖其缺失，无需清理）
	guard.RegisterEntityTypeOverride("id_card", guard.Override{Extensions: []string{"webp"}})

	l := NewGetUploadUrlLogic(context.Background(), uploadCtx())

	// 无 override：avatar 基线 → webp 070004
	_, err := l.GetUploadUrl(&filev1.GetUploadUrlRequest{UserId: 100, EntityType: "avatar", FileName: "card.webp", FileSize: 1024})
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "avatar 基线 → webp 070004")

	// 有 override：id_card → webp 放行，进入 MinIO 检查 → 70002（证明覆盖已应用）
	_, err = l.GetUploadUrl(&filev1.GetUploadUrlRequest{UserId: 100, EntityType: "id_card", FileName: "card.webp", FileSize: 1024})
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeFileAccessDenied), "id_card override → webp 放行后 70002")
}
