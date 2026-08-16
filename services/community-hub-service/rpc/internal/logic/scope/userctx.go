package scope

import (
	"context"
	"sort"
	"strconv"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"google.golang.org/grpc/metadata"
)

// UserIDMetadataKey 是 API 层经 util.WithUserID 注入出站 gRPC 的 metadata 键，
// 与 JWT claim 键保持一致（"user_id"）。
const UserIDMetadataKey = "user_id"

// SystemUserID 系统审核身份（moderation 服务间回调，无用户 JWT）。
// permission-service 种子：rel_user_role (user_id=0, role_id=sys_admin, scope_type='global',
// scope_id=0, status=2)，校验走同一条 grant 判定路径（global → 放行），无代码级短路。
const SystemUserID = int64(0)

// UserIDFromCtx 从 gRPC 入站 metadata 提取调用方 JWT 身份（API 层注入）。
// 未注入 / 非法 → 返回 0（调用方按 fail-closed 处理）。
//
// 安全边界（评审 CRITICAL）：本函数对入站 metadata 的 user_id 是"盲信"的——它本身不校验
// 调用方是否被授权携带该身份。数据权限（AssertPublishScope 写 / GetDataScopes 读过滤）的
// 安全前提是 **网络隔离**：RPC 必须绑定回环（rpc/etc/communityhub.yaml ListenOn=127.0.0.1:8088），
// 只允许宿主机上的可信调用方（REST API 网关经 CallCtx 注入、moderation-service 系统回调无身份）
// 连通端口，阻断局域网 / Docker 桥接网络对端伪造 user_id 注入身份。
//
// 局限：仓库级模式（9 个服务均 0.0.0.0 无鉴权）尚未落地服务凭据 / mTLS + unary 拦截器校验，
// 届时本函数应叠加调用方身份校验（见 memory: rpc-identity-spoofing-loopback-isolation）。
//
// SEE: [[is-system-no-permission-shortcut]] — 身份经 metadata 传输，不信任客户端 body
// SEE: [[rpc-identity-spoofing-loopback-isolation]] — RPC 身份伪造风险 + 回环隔离缓解
func UserIDFromCtx(ctx context.Context) int64 {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0
	}
	vals := md.Get(UserIDMetadataKey)
	if len(vals) == 0 {
		return 0
	}
	id, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// IsLevel2Grant level-2 等价判定：status==2（已认证）且 verified_at>0 且未过期。
// 与 permission-service 421 min_verf_level=2 门槛同语义（SEE [[auto-grant-unverified-grant-confers-scope-level0]]），
// 基于 RPC 输出（UserRoleInfo），禁止直读 rel_user_role。供 PublishRolesFrom / GetPublishPermission / ResolveAdminDivision 共用。
func IsLevel2Grant(ur *permissionv1.UserRoleInfo) bool {
	if ur == nil || ur.GetRole() == nil {
		return false
	}
	if ur.GetStatus() != UserRoleStatusVerified {
		return false
	}
	if ur.GetVerifiedAt() <= 0 {
		return false
	}
	if ur.GetExpiresAt() != 0 && ur.GetExpiresAt() <= time.Now().Unix() {
		return false
	}
	return true
}

// publishRolePriority 发布角色优先序（D6：grid_worker > community_admin > committee > property_admin）。
var publishRolePriority = map[string]int{
	RoleGridWorker:     1,
	RoleCommunityAdmin: 2,
	RoleCommittee:      3,
	RolePropertyAdmin:  4,
}

// PublishRolesFrom 取用户实际持有的发布角色 code（level-2 已认证过滤），按发布角色优先序返回（Task 1.8）。
//
// REVISION REQ-CPB-5：JWT 仅含 user_id，角色必须显式调 permission `GetUserRoles(user_id)` 解析；
// 供 role 列映射派生与 is_pinned 操作者授权（Task 1.11 (b) 分支「PublishRolesFrom 非空」判据）。
// 无发布角色 → 空集；GetUserRoles 传输错误原样返回（fail-closed）。
//
// SEE: [[grpc-only-comms]] — 经 GetUserRoles，禁止直读 rel_user_role
func PublishRolesFrom(ctx context.Context, permClient permissionv1.PermissionServiceClient, userID int64) ([]string, error) {
	resp, err := permClient.GetUserRoles(ctx, &permissionv1.GetUserRolesRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var roles []string
	for _, ur := range resp.GetRoles() {
		if !IsLevel2Grant(ur) {
			continue
		}
		code := ur.GetRole().GetCode()
		if _, ok := publishRolePriority[code]; !ok {
			continue // 非发布角色（owner/tenant/merchant/sys_admin）
		}
		if _, dup := seen[code]; dup {
			continue // 同角色多 scope grant 去重
		}
		seen[code] = struct{}{}
		roles = append(roles, code)
	}

	// 按发布角色优先序稳定排序
	sort.SliceStable(roles, func(i, j int) bool {
		return publishRolePriority[roles[i]] < publishRolePriority[roles[j]]
	})
	return roles, nil
}

// PublishRoleToString RBAC code → DB role 列值映射（Task 1.8，D6）。
// grid_worker→grid_officer、community_admin→community、committee→committee、property_admin→property。
// 与 Task 1.13 读侧 ContentPostRoleToString 收敛同一字符串集合（评审 data-model v4 I2）。
func PublishRoleToString(roleCode string) string {
	switch roleCode {
	case RoleGridWorker:
		return "grid_officer"
	case RoleCommunityAdmin:
		return "community"
	case RoleCommittee:
		return "committee"
	case RolePropertyAdmin:
		return "property"
	default:
		return ""
	}
}
