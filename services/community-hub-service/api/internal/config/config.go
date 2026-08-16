package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 社区枢纽 REST API 配置
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	CommunityHubRpc zrpc.RpcClientConf
	PermissionRpc   zrpc.RpcClientConf
	// DataSource 仅用于 REST 详情 community_id 兼容回退（R2）：GET /notices/:id 缺 community_id 时
	// 按 content_post_scope 反查帖所属小区 + FilterAllowed 任一允许即放行（contentcompat）。
	// 只读查询（FindOneReviewComplete / FindCommunityIdsByPostId），不承载写路径。
	DataSource string
}

// JWT claim 契约（auth-service 签发，access-data-permission 阶段③④ 消费侧统一）：
//   - claim 键统一为 "user_id"，值经 go-zero rest.WithJwt 注入 ctx（json.Number 形态）；
//   - REST 层消费用 api/internal/util.JWTUserID(ctx) 提取；
//   - API → RPC 跨层传播复用同一键作为 gRPC outgoing metadata（util.WithUserID）。
//
// 键名常量定义在 api/internal/util.UserIDClaimKey，本文件仅记录契约约定。
// SEE: [[verify-api-before-calling]] — publisher_id/userId 取自 JWT 前先确认 claims 结构
