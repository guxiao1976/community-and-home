package perm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoDiscoverPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAutoDiscoverPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoDiscoverPermissionsLogic {
	return &AutoDiscoverPermissionsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// sysPermRow 匹配 sys_permission 表结构（使用 sql.NullString 处理 NULL 字段）
type sysPermRow struct {
	Id        int64          `db:"id"`
	ParentId  sql.NullInt64  `db:"parent_id"`
	Name      string         `db:"name"`
	Code      string         `db:"code"`
	Type      int64          `db:"type"`
	Path      sql.NullString `db:"path"`
	SortOrder int64          `db:"sort_order"`
	Status    int64          `db:"status"`
}

// AutoDiscover 扫描 RouteRegistry 中已注册路由，对比 sys_permission 表，自动注册缺失的 API 权限
func (l *AutoDiscoverPermissionsLogic) AutoDiscover() (*types.AutoDiscoverPermissionsResp, error) {
	routes := l.svcCtx.RouteRegistry.Routes()
	if len(routes) == 0 {
		return &types.AutoDiscoverPermissionsResp{
			Added:   []types.AutoDiscoveredPerm{},
			Message: "没有可发现的路由（RouteRegistry 为空）",
		}, nil
	}

	db := l.svcCtx.DB

	// 1. 使用 SELECT * 查询全部权限（与 model 代码一致，确保 sqlx 列映射正确）
	var allRows []sysPermRow
	if err := db.QueryRowsCtx(l.ctx, &allRows,
		"SELECT * FROM sys_permission WHERE status = 1 ORDER BY sort_order",
	); err != nil {
		return nil, fmt.Errorf("查询权限表失败: %w", err)
	}

	// 2. 构建索引
	existingPaths := make(map[string]bool) // "POST:/api/perm/roles"
	var maxID, type3Count int64
	for _, row := range allRows {
		if row.Id > maxID {
			maxID = row.Id
		}
		if row.Type == 3 {
			type3Count++
			if row.Path.Valid && row.Path.String != "" {
				existingPaths[strings.ToUpper(row.Path.String)] = true
			}
		}
	}
	_ = type3Count // 供调试使用

	// 3. 查找缺失路由
	var added []types.AutoDiscoveredPerm
	now := time.Now()

	for _, route := range routes {
		// 路径参数统一归一化为 :id（如 :userId → :id），与 PermMiddleware 的 normalizePathParams 一致
		normalizedPath := normalizeRouteParams(route.Path)
		pathKey := route.Method + ":" + normalizedPath
		if existingPaths[strings.ToUpper(pathKey)] {
			continue
		}

		parentID := findParent(route.Path, route.Method, allRows)
		code := genCode(route.Method, route.Path)
		name := route.Comment
		if name == "" {
			name = route.Method + " " + route.Path
		}
		maxID++

		_, err := db.ExecCtx(l.ctx,
			`INSERT INTO sys_permission (id, parent_id, name, code, type, path, sort_order, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, 3, ?, 99, 1, ?, ?)`,
			maxID, parentID, name, code, pathKey, now, now,
		)
		if err != nil {
			l.Logger.Errorf("auto-discover: 插入失败 path=%s code=%s err=%v", pathKey, code, err)
			maxID-- // 回退 ID
			continue
		}

		added = append(added, types.AutoDiscoveredPerm{
			Id:       maxID,
			ParentId: parentID,
			Name:     name,
			Code:     code,
			Path:     pathKey,
		})
	}

	msg := fmt.Sprintf("扫描 %d 条路由，新增 %d 条权限", len(routes), len(added))
	if len(added) == 0 {
		msg = fmt.Sprintf("扫描 %d 条路由，权限均已存在，无新增", len(routes))
	}

	return &types.AutoDiscoverPermissionsResp{
		Added:   added,
		Total:   len(added),
		Message: msg,
	}, nil
}

// resourceToMenu 资源标识 → 菜单 path 映射
var resourceToMenu = map[string]string{
	"roles":       "/roles",
	"permissions": "/permissions",
	"user-roles":  "/users",
	"users":       "/users",
}

// specialParent 针对特殊路由的父节点覆盖（path:method → parent code）
// 例如 DELETE /api/perm/user-roles 的语义是"撤销角色"而非"删除用户"
var specialParent = map[string]string{
	"/api/perm/user-roles:DELETE": "user:assign-role",
}

// normalizeRouteParams 将路由路径中的命名参数统一归一化为 :id
// 与 PermMiddleware 的 normalizePathParams（数字段 → :id）保持一致
// 例: /api/perm/users/:userId/roles → /api/perm/users/:id/roles
func normalizeRouteParams(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") && seg != ":id" {
			segments[i] = ":id"
		}
	}
	return strings.Join(segments, "/")
}

func findParent(path, method string, allRows []sysPermRow) int64 {
	// 检查特殊父节点映射
	if code, ok := specialParent[path+":"+strings.ToUpper(method)]; ok {
		for _, row := range allRows {
			if row.Type == 2 && row.Code == code {
				return row.Id
			}
		}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0
	}
	resource := parts[2]
	if resource == "perm" && len(parts) >= 4 {
		resource = parts[3]
	}

	menuPath, ok := resourceToMenu[resource]
	if !ok {
		menuPath = "/" + resource
	}
	trimmed := strings.Trim(menuPath, "/")

	// 找 type=1 menu 节点
	var menuID int64
	for _, row := range allRows {
		if row.Type == 1 && row.Path.Valid && strings.Trim(row.Path.String, "/") == trimmed {
			menuID = row.Id
			break
		}
	}
	if menuID == 0 {
		return 0
	}

	// 找 type=2 button 节点（按 HTTP 方法匹配 action 后缀）
	suffix := methodAction(method)
	for _, row := range allRows {
		if row.Type == 2 && row.ParentId.Valid && row.ParentId.Int64 == menuID &&
			strings.HasSuffix(row.Code, ":"+suffix) {
			return row.Id
		}
	}
	// fallback: menu 下第一个 button
	for _, row := range allRows {
		if row.Type == 2 && row.ParentId.Valid && row.ParentId.Int64 == menuID {
			return row.Id
		}
	}
	return 0
}

func methodAction(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func genCode(method, path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	resource := "api"
	if len(parts) >= 3 {
		resource = parts[2]
	}
	if len(parts) >= 4 && !strings.HasPrefix(parts[3], ":") {
		resource += ":" + parts[3]
	}
	if len(parts) >= 5 && !strings.HasPrefix(parts[4], ":") {
		resource += ":" + parts[4]
	}
	return resource + ":" + methodAction(method) + "-api"
}
