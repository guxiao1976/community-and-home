package guard

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ole2Header OLE2/CFB 复合文档头
var ole2Header = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

// zipHeader PK zip 头
var zipHeader = []byte{0x50, 0x4B, 0x03, 0x04}

func TestSniffType_ImagesAndPdf(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		want string
	}{
		{"png", append([]byte{0x89, 0x50, 0x4E, 0x47}, 0x0D, 0x0A, 0x1A, 0x0A), "png"},
		{"jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}, "jpg"},
		{"gif", []byte{'G', 'I', 'F', '8', '9', 'a'}, "gif"},
		{"pdf", []byte{'%', 'P', 'D', 'F', '-', '1', '.', '7'}, "pdf"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ext, ok := SniffType(tc.buf)
			require.True(t, ok, "应识别 %s", tc.name)
			assert.Equal(t, tc.want, ext)
		})
	}
}

func TestSniffType_RealDoc_Allowed(t *testing.T) {
	// CFB 头 + WordDocument 流名（UTF-16LE）→ doc
	buf := append(append([]byte{}, ole2Header...), utf16LE("WordDocument")...)
	ext, ok := SniffType(buf)
	require.True(t, ok, "OLE2 + WordDocument 流 → doc")
	assert.Equal(t, "doc", ext)
}

func TestSniffType_MSIrenamedToDoc_Rejected(t *testing.T) {
	// 仅 CFB 头，无 WordDocument 流（msi/xls/ppt 特征）→ 不识别
	ext, ok := SniffType(ole2Header)
	assert.False(t, ok, "OLE2 无 WordDocument 流（msi 改 doc）→ 拒绝")
	assert.Empty(t, ext)
}

func TestSniffType_RealDocx_Allowed(t *testing.T) {
	// ZIP 头 + word/document.xml 部件 → docx
	buf := append(append([]byte{}, zipHeader...), []byte("word/document.xml")...)
	ext, ok := SniffType(buf)
	require.True(t, ok, "ZIP + word/document.xml → docx")
	assert.Equal(t, "docx", ext)
}

func TestSniffType_XLSXrenamedToDocx_Rejected(t *testing.T) {
	// ZIP 头 + xl/workbook.xml（xlsx 部件）→ 无 word/document.xml → 不识别
	buf := append(append([]byte{}, zipHeader...), []byte("xl/workbook.xml")...)
	ext, ok := SniffType(buf)
	assert.False(t, ok, "xlsx 改 docx → 拒绝")
	assert.Empty(t, ext)
}

func TestSniffType_GenericZip_Rejected(t *testing.T) {
	// 通用 zip（无 word/document.xml）→ 拒绝
	buf := append(append([]byte{}, zipHeader...), []byte("random/entry.txt")...)
	ext, ok := SniffType(buf)
	assert.False(t, ok, "通用 zip 内容 → 拒绝")
	assert.Empty(t, ext)
}

func TestSniffType_PErenamedToPng_Rejected(t *testing.T) {
	// exe（PE MZ 头）→ 非白名单魔数 → 不识别
	buf := append([]byte{0x4D, 0x5A, 0x90, 0x00}, bytes.Repeat([]byte{0x00}, 64)...)
	ext, ok := SniffType(buf)
	assert.False(t, ok, "PE MZ 头改名 png 传 exe → 拒绝")
	assert.Empty(t, ext)
}
