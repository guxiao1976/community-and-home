package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetUserIdFromContext 从 go-zero JWT 中间件注入的 context 中提取 user_id
// go-zero 的 rest.WithJwt 会将 JWT claims 中的非标准字段注入到 request context
func GetUserIdFromContext(r *http.Request) (int64, error) {
	v := r.Context().Value("user_id")
	if v == nil {
		return 0, fmt.Errorf("user_id not found in context")
	}
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case json.Number:
		return val.Int64()
	case int64:
		return val, nil
	default:
		return 0, fmt.Errorf("unexpected user_id type: %T", v)
	}
}
