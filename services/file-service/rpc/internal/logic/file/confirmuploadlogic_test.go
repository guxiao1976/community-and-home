package file

import (
	"bytes"
	"context"
	"testing"
	"time"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-file/model"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// fakeFileModel — FileModel 接口假实现
// =============================================================================

type fakeFileModel struct {
	inserted *model.File
	insertID int64
	err      error
}

func (m *fakeFileModel) Insert(_ context.Context, f *model.File) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	m.inserted = f
	return m.insertID, nil
}

func (m *fakeFileModel) FindOne(_ context.Context, _ int64) (*model.File, error) { return nil, nil }
func (m *fakeFileModel) FindByIds(_ context.Context, _ []int64) ([]*model.File, error) {
	return nil, nil
}
func (m *fakeFileModel) FindPage(_ context.Context, _ *int64, _ *string, _ *int64, _, _ int64) ([]*model.File, int64, error) {
	return nil, 0, nil
}
func (m *fakeFileModel) Delete(_ context.Context, _ int64) error { return nil }

// =============================================================================

var (
	pngContent = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	exeContent = []byte{0x4D, 0x5A, 0x90, 0x00, 0x03, 0x00, 0x00, 0x00}
	docContent = append([]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, utf16LEForTest("WordDocument")...)
)

func utf16LEForTest(s string) []byte {
	out := make([]byte, 0, len(s)*2)
	for _, r := range s {
		out = append(out, byte(r), byte(r>>8))
	}
	return out
}

func confirmReq(fileName string) *filev1.ConfirmUploadRequest {
	return &filev1.ConfirmUploadRequest{
		UserId:     100,
		ObjectKey:  "uploads/100/1_a.png",
		EntityType: "avatar",
		FileName:   fileName,
		FileSize:   1024,
		MimeType:   "image/png",
	}
}

func TestConfirmUpload_DeclaredPng_ActualExe_Rejected070004(t *testing.T) {
	sc := &svc.ServiceContext{FileModel: &fakeFileModel{insertID: 1}, Bucket: "community-home"}
	l := NewConfirmUploadLogic(context.Background(), sc)
	_, err := l.confirmWithReader(confirmReq("a.png"), bytes.NewReader(exeContent))
	require.Error(t, err)
	assert.True(t, errx.IsCode(err, ErrCodeUnsupportedFileType), "声明 png 实为 exe → 070004: %v", err)
	assert.Nil(t, sc.FileModel.(*fakeFileModel).inserted, "拒绝时不落库")
}

func TestConfirmUpload_DeclaredMatchesMagic_Allowed(t *testing.T) {
	fm := &fakeFileModel{insertID: 7}
	sc := &svc.ServiceContext{FileModel: fm, Bucket: "community-home"}
	l := NewConfirmUploadLogic(context.Background(), sc)
	resp, err := l.confirmWithReader(confirmReq("a.png"), bytes.NewReader(pngContent))
	require.NoError(t, err)
	require.NotNil(t, resp.File)
	assert.Equal(t, int64(7), resp.File.Id)
	assert.Equal(t, "png", resp.File.FileType, "FileInfo.file_type = 嗅探规范扩展名")
	assert.True(t, resp.File.Confirmed, "FileInfo.confirmed = true")
	require.NotNil(t, fm.inserted)
	assert.Equal(t, "png", fm.inserted.FileType, "落库 file_type")
	assert.True(t, fm.inserted.Confirmed, "落库 confirmed")
}

func TestConfirmUpload_DocDeclared_MagicMatch_Allowed(t *testing.T) {
	fm := &fakeFileModel{insertID: 2}
	sc := &svc.ServiceContext{FileModel: fm, Bucket: "community-home"}
	l := NewConfirmUploadLogic(context.Background(), sc)
	resp, err := l.confirmWithReader(confirmReq("report.doc"), bytes.NewReader(docContent))
	require.NoError(t, err)
	assert.Equal(t, "doc", resp.File.FileType)
	assert.True(t, resp.File.Confirmed)
	assert.Equal(t, "doc", fm.inserted.FileType)
}

func TestConfirmUpload_DeclaredJpeg_SniffJpg_EquivalentAllowed(t *testing.T) {
	// jpeg/jpg 等价：声明 .jpeg + JPEG 魔数（嗅探 jpg）→ 放行，落 file_type=jpg
	fm := &fakeFileModel{insertID: 3}
	sc := &svc.ServiceContext{FileModel: fm, Bucket: "community-home"}
	l := NewConfirmUploadLogic(context.Background(), sc)
	resp, err := l.confirmWithReader(confirmReq("photo.jpeg"), bytes.NewReader([]byte{0xFF, 0xD8, 0xFF, 0xE0}))
	require.NoError(t, err)
	assert.Equal(t, "jpg", resp.File.FileType)
	assert.True(t, resp.File.Confirmed)
}

func TestToProtoFile_FileTypeConfirmed(t *testing.T) {
	f := &model.File{
		UserID:     100,
		EntityType: "avatar",
		FileName:   "a.png",
		FilePath:   "uploads/100/1_a.png",
		FileSize:   1024,
		MimeType:   "image/png",
		BucketName: "community-home",
		UploadTime: time.Now(),
		FileType:   "png",
		Confirmed:  true,
	}
	f.ID = 5
	pb := toProtoFile(f)
	assert.Equal(t, int64(5), pb.Id)
	assert.Equal(t, "png", pb.FileType)
	assert.True(t, pb.Confirmed)
}
