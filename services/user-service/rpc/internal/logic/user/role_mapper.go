package user

import (
	"context"
	"sync"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// roleMapper 维护 role_code ↔ role_id 映射（来自 permission-service 的 sys_role）
// 缓存以避免每次申请都查询角色表；角色变更频率极低，TTL 5 分钟足够
type roleMapper struct {
	mu       sync.RWMutex
	codeToID map[string]int64
	idToCode map[int64]string
	loadedAt time.Time
}

var mapper = &roleMapper{
	codeToID: make(map[string]int64),
	idToCode: make(map[int64]string),
}

const roleCacheTTL = 5 * time.Minute

// ensureLoaded 确保角色映射已加载（缓存未命中或过期时从 permission-service 拉取）
func (m *roleMapper) ensureLoaded(ctx context.Context, svcCtx *svc.ServiceContext, logger logx.Logger) {
	m.mu.RLock()
	fresh := len(m.codeToID) > 0 && time.Since(m.loadedAt) < roleCacheTTL
	m.mu.RUnlock()
	if fresh {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// double-check
	if len(m.codeToID) > 0 && time.Since(m.loadedAt) < roleCacheTTL {
		return
	}

	if svcCtx.PermissionClient == nil {
		logger.Errorf("roleMapper: PermissionClient is nil, cannot load roles")
		return
	}

	resp, err := svcCtx.PermissionClient.ListRoles(ctx, &permissionv1.ListRolesRequest{})
	if err != nil {
		logger.Errorf("roleMapper: ListRoles failed: %v", err)
		return
	}

	codeToID := make(map[string]int64)
	idToCode := make(map[int64]string)
	for _, r := range resp.Roles {
		codeToID[r.Code] = r.Id
		idToCode[r.Id] = r.Code
	}

	m.codeToID = codeToID
	m.idToCode = idToCode
	m.loadedAt = time.Now()
}

// roleIDByCode 返回 role_code 对应的 role_id（permission-service 的 sys_role.id）
func roleIDByCode(ctx context.Context, svcCtx *svc.ServiceContext, logger logx.Logger, code string) (int64, bool) {
	mapper.ensureLoaded(ctx, svcCtx, logger)
	mapper.mu.RLock()
	defer mapper.mu.RUnlock()
	id, ok := mapper.codeToID[code]
	return id, ok
}

// roleCodeByID 返回 role_id 对应的 role_code
func roleCodeByID(ctx context.Context, svcCtx *svc.ServiceContext, logger logx.Logger, roleID int64) (string, bool) {
	mapper.ensureLoaded(ctx, svcCtx, logger)
	mapper.mu.RLock()
	defer mapper.mu.RUnlock()
	code, ok := mapper.idToCode[roleID]
	return code, ok
}

// resetRoleMapper 重置全局角色映射缓存（测试用，避免跨测试状态污染）
func resetRoleMapper() {
	mapper.mu.Lock()
	defer mapper.mu.Unlock()
	mapper.codeToID = make(map[string]int64)
	mapper.idToCode = make(map[int64]string)
	mapper.loadedAt = time.Time{}
}
