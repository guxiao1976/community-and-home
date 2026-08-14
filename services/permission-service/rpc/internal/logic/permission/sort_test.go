package permission

import (
	"testing"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/stretchr/testify/assert"
)

// TestValidateSort — validateSort 纯函数：白名单 7 字段 + 大小写不敏感 + 空方向默认 asc + 非法返回 99400
// RED: sort.go 尚未创建，本测试引用 validateSort → 编译失败（RED）
func TestValidateSort(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		order       string
		wantField   string
		wantOrder   string
		wantErrCode int
	}{
		{name: "合法字段+desc", field: "role_name", order: "desc", wantField: "role_name", wantOrder: "desc"},
		{name: "字段大小写不敏感 ROLE_CODE→role_code", field: "ROLE_CODE", order: "asc", wantField: "role_code", wantOrder: "asc"},
		{name: "方向大小写不敏感 DESC→desc", field: "role_name", order: "DESC", wantField: "role_name", wantOrder: "desc"},
		{name: "空方向默认 asc", field: "status", order: "", wantField: "status", wantOrder: "asc"},
		{name: "白名单含 id", field: "id", order: "", wantField: "id", wantOrder: "asc"},
		{name: "白名单含 created_at", field: "created_at", order: "desc", wantField: "created_at", wantOrder: "desc"},
		{name: "空字段+空方向 → 未指定排序", field: "", order: "", wantField: "", wantOrder: ""},
		{name: "空字段+非空方向不报错", field: "", order: "desc", wantField: "", wantOrder: ""},
		{name: "非法字段 → 99400", field: "evil", order: "", wantErrCode: errx.CodeInvalidParam},
		{name: "注入载荷被拒 → 99400", field: "role_name; drop table sys_role", order: "", wantErrCode: errx.CodeInvalidParam},
		{name: "非法方向 → 99400", field: "role_name", order: "random", wantErrCode: errx.CodeInvalidParam},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, order, err := validateSort(tt.field, tt.order)
			if tt.wantErrCode != 0 {
				assert.Error(t, err, "非法输入应返回 error")
				ce := errx.FromError(err)
				assert.NotNil(t, ce, "error 应为 errx.CodeError")
				assert.Equal(t, tt.wantErrCode, ce.Code, "错误码应等于 CodeInvalidParam(99400)")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantField, field, "规范化后的排序字段")
			assert.Equal(t, tt.wantOrder, order, "规范化后的排序方向")
		})
	}
}
