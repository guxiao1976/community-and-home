package permission

import (
	"strings"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
)

// roleSortFieldWhitelist 角色排序白名单（sys_role 既有列名）。
// 与 model/permission.go 的 roleSortFieldWhitelist 需保持同步（各自独立定义，双处同步）。
// SEE: [[error-code-literal-bypasses-qa-gate]] — 校验用 errx.CodeInvalidParam 常量，禁止裸数字。
var roleSortFieldWhitelist = map[string]struct{}{
	"id":         {},
	"role_code":  {},
	"role_name":  {},
	"sort_order": {},
	"status":     {},
	"created_at": {},
	"updated_at": {},
}

// validateSort 校验并规范化排序字段/方向（纯函数，RPC 层单层权威校验）。
// 返回规范化后的 field/order（大小写已归一），供 FindList 透传。
//
// 规则：
//   - field 非空 → 转小写比对白名单；未命中 → errx.CodeInvalidParam(99400)
//   - field 为空 → 视为未指定排序，返回空（不报错，即使带方向，REQ-4）
//   - order 转小写；空默认 asc；非 asc/desc → errx.CodeInvalidParam(99400)
func validateSort(field, order string) (sortField, sortOrder string, err error) {
	if field != "" {
		lower := strings.ToLower(field)
		if _, ok := roleSortFieldWhitelist[lower]; !ok {
			return "", "", errx.NewCodeError(errx.CodeInvalidParam, "非法排序字段: "+field)
		}
		sortField = lower
	}

	// 字段为空 → 未指定排序，跳过方向校验（空字段 + 非空方向不报错）
	if sortField == "" {
		return "", "", nil
	}

	direction := strings.ToLower(order)
	if direction == "" {
		direction = "asc"
	}
	if direction != "asc" && direction != "desc" {
		return "", "", errx.NewCodeError(errx.CodeInvalidParam, "非法排序方向: "+order)
	}

	return sortField, direction, nil
}
