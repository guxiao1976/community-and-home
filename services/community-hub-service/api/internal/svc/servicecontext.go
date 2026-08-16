package svc

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	permMiddleware "github.com/guxiao1976/community-common/v2/pkg/middleware"
	"github.com/guxiao1976/community-hub/api/internal/config"
	"github.com/guxiao1976/community-hub/api/internal/util"
	"github.com/guxiao1976/community-hub/model"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext REST API 服务上下文，持有 gRPC 客户端
type ServiceContext struct {
	Config                config.Config
	ContentPostServiceRpc communityv1.ContentPostServiceClient
	ContactServiceRpc     communityv1.ContactServiceClient
	LostFoundServiceRpc   communityv1.LostFoundServiceClient
	PermClient            permissionv1.PermissionServiceClient
	PermMiddleware        *permMiddleware.PermissionMiddleware
	// 以下仅 REST 详情 community_id 兼容回退用（R2 contentcompat 只读查询，不承载写路径）
	ContentPostModel      model.ContentPostModel
	ContentPostScopeModel model.ContentPostScopeModel
}

// NewServiceContext 创建服务上下文，初始化 gRPC 客户端
func NewServiceContext(c config.Config) *ServiceContext {
	cli := zrpc.MustNewClient(c.CommunityHubRpc)
	conn := cli.Conn()

	permCli := zrpc.MustNewClient(c.PermissionRpc)
	permClient := permissionv1.NewPermissionServiceClient(permCli.Conn())
	permMW := permMiddleware.NewPermissionMiddleware(permClient)

	// 兼容回退只读模型（R2）：content_post_scope 反查 + FindOneReviewComplete
	var postModel model.ContentPostModel
	var scopeModel model.ContentPostScopeModel
	if c.DataSource != "" {
		dbConn := sqlx.NewMysql(c.DataSource)
		postModel = model.NewContentPostModel(dbConn)
		scopeModel = model.NewContentPostScopeModel(dbConn)
	}

	return &ServiceContext{
		Config:                c,
		ContentPostServiceRpc: communityv1.NewContentPostServiceClient(conn),
		ContactServiceRpc:     communityv1.NewContactServiceClient(conn),
		LostFoundServiceRpc:   communityv1.NewLostFoundServiceClient(conn),
		PermClient:            permClient,
		PermMiddleware:        permMW,
		ContentPostModel:      postModel,
		ContentPostScopeModel: scopeModel,
	}
}

// CallCtx 返回注入了 JWT 身份 gRPC metadata 的调用上下文，以及该身份的用户 ID。
//
// 这是写接口身份注入通道（T4.0 确认）：rest.WithJwt 已将 user_id 注入请求 ctx，
// CallCtx 经 util.JWTUserID 提取并统一覆盖写入：
//   - 出站 gRPC metadata（rpc 层经 scope.UserIDFromCtx 读取，用于 AssertPublishScope / GetDataScopes 过滤）。
//
// 校验顺序（T4.2）：功能权限（PermMiddleware，在中间件链中先于 handler 执行）
// → 数据权限（rpc 层落库前 AssertPublishScope）→ 落库。REST 中间件链天然保证
// 功能权限先于数据权限执行。
//
// SEE: [[verify-api-before-calling]] — 身份取自 JWT 而非客户端 body
func (s *ServiceContext) CallCtx(ctx context.Context) (context.Context, int64, error) {
	uid, err := util.JWTUserID(ctx)
	if err != nil {
		return nil, 0, err
	}
	return util.WithUserID(ctx, uid), uid, nil
}
