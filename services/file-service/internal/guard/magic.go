package guard

import "bytes"

// =============================================================================
// magic-bytes 内容嗅探器（L2 ConfirmUpload 回读实际对象）
//
// 规则（REQ-CAS-3）：
//   - png  `89 50 4E 47` / jpg `FF D8 FF` / gif `GIF8` / pdf `%PDF`
//   - doc  = OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）且内含 `WordDocument` 流
//           （仅 CFB 头不充分；msi/xls/ppt 无该流 → 不识别 → 070004）
//   - docx = ZIP（PK 头）且含 `word/document.xml` 部件（docx 为唯一 zip 内容特判；
//           xlsx/pptx 无该部件 → 不识别 → 070004）
//   - 其他 OLE2/OOXML/通用 zip/rar 内容 → (not recognized) → 070004
// =============================================================================

var (
	pngMagic           = []byte{0x89, 0x50, 0x4E, 0x47}
	jpgMagic           = []byte{0xFF, 0xD8, 0xFF}
	gifMagic           = []byte{'G', 'I', 'F', '8'}
	pdfMagic           = []byte{'%', 'P', 'D', 'F'}
	ole2Magic          = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	zipMagic           = []byte{0x50, 0x4B}
	wordDocumentStream = utf16LE("WordDocument")
	docxPartMarker     = []byte("word/document.xml")
)

// SniffType 根据内容前缀嗅探规范扩展名。
// 返回 (规范扩展名, 是否识别)；未识别返回 ("", false)（调用方 070004）。
func SniffType(buf []byte) (string, bool) {
	switch {
	case bytes.HasPrefix(buf, pngMagic):
		return "png", true
	case bytes.HasPrefix(buf, jpgMagic):
		return "jpg", true
	case bytes.HasPrefix(buf, gifMagic):
		return "gif", true
	case bytes.HasPrefix(buf, pdfMagic):
		return "pdf", true
	case bytes.HasPrefix(buf, ole2Magic):
		// OLE2/CFB：仅头不够，须含 WordDocument 流才映射 doc（msi/xls/ppt 无该流）
		if bytes.Contains(buf, wordDocumentStream) {
			return "doc", true
		}
		return "", false
	case bytes.HasPrefix(buf, zipMagic):
		// docx 为唯一 zip 内容特判：须含 word/document.xml 部件
		if bytes.Contains(buf, docxPartMarker) {
			return "docx", true
		}
		return "", false
	default:
		return "", false
	}
}

// utf16LE 将 ASCII 字符串编码为 UTF-16LE 字节（OLE2 目录流名存储格式）。
func utf16LE(s string) []byte {
	out := make([]byte, 0, len(s)*2)
	for _, r := range s {
		out = append(out, byte(r), byte(r>>8))
	}
	return out
}
