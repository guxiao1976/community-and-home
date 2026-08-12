package scope

import (
	"context"
	"strconv"

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
