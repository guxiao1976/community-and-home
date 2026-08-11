package svc

import "sync"

// RouteInfo 描述一个已注册的 HTTP 路由
type RouteInfo struct {
	Method  string // GET, POST, PUT, DELETE
	Path    string // 例如 /roles, /roles/:id, /users/:userId/permissions
	Comment string // 路由描述（来自代码注释）
}

// RouteRegistry 收集所有已注册的路由，供 auto-discover 使用
type RouteRegistry struct {
	mu     sync.RWMutex
	routes []RouteInfo
}

// NewRouteRegistry 创建路由注册器
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{}
}

// Register 注册单条路由
func (r *RouteRegistry) Register(method, path, comment string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, RouteInfo{Method: method, Path: path, Comment: comment})
}

// Routes 返回所有已注册路由的快照
func (r *RouteRegistry) Routes() []RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshot := make([]RouteInfo, len(r.routes))
	copy(snapshot, r.routes)
	return snapshot
}
