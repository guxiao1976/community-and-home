// Package scope 承载 community-hub 数据权限校验与读过滤逻辑。
//
// 校验委托 permission-service 权威计算（AssertPublishScope / GetDataScopes），
// 本包只负责：构造 ScopeRef target、解析 JWT 身份（gRPC metadata 注入）、
// 错误码映射（permission 060007 → community-hub 080006）、GLOBAL/LIMITED/EMPTY 过滤语义。
//
// 依赖方向（access-control-design §8）：community-hub → permission-service → master-data；
// 本包不直连 master-data 做 scope 解析，祖先链仅经 permission ResolveScopeAncestors 消费。
package scope

import (
	"context"

	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

// CodePublishScopeDenied 数据权限拒绝（API 面 080006）：目标小区超出发布者数据范围。
// 对应 permission-service 的 060007（RPC 面），消费方统一映射为 080006。
// 与 080002（功能权限：无发布权限）分层：080002 由 PermMiddleware/CheckPermission 产出，
// 080006 由 AssertPublishScope 数据权限产出。
const CodePublishScopeDenied = 80006

// ScopeTypeCommunity 是本变更唯一支持的 target 作用域类型。
const ScopeTypeCommunity = "community"

// AssertCommunityScope 校验 userID 是否对目标小区有发布数据权限。
//
// 构造 [{scope_type:'community', scope_id: communityID}] 调 permission AssertPublishScope；
// allowed=false（含 EMPTY/祖先链未覆盖/未知节点安全拒绝，permission 侧 060007）→ 映射为 080006。
// gRPC 传输错误原样返回（fail-closed 由调用方决定）。
//
// 校验顺序（T4.2）：功能权限（PermMiddleware，REST 中间件链先于 handler）→
// 数据权限（本函数，落库前）→ 落库。
//
// SEE: [[grpc-timeout-layers]] — AssertPublishScope 内嵌 master-data ResolveScopeAncestors，三层超时对齐
func AssertCommunityScope(ctx context.Context, client permissionv1.PermissionServiceClient, userID, communityID int64) error {
	resp, err := client.AssertPublishScope(ctx, &permissionv1.AssertPublishScopeRequest{
		UserId: userID,
		Targets: []*permissionv1.ScopeRef{
			{ScopeType: ScopeTypeCommunity, ScopeId: communityID},
		},
	})
	if err != nil {
		return err
	}
	if !resp.GetAllowed() {
		return errx.NewCodeError(CodePublishScopeDenied, "目标小区超出发布者数据范围")
	}
	return nil
}

// IsPublishScopeDenied 判断 err 是否为 080006 数据权限拒绝（区别于 gRPC 传输错误）。
// 调用方据此决定返回业务响应还是传播传输错误。
func IsPublishScopeDenied(err error) bool {
	if err == nil {
		return false
	}
	ce := errx.FromError(err)
	return ce != nil && ce.Code == CodePublishScopeDenied
}

// DenyBase 构造 080006 数据权限拒绝的 BaseResp（供各 RPC 写逻辑直接嵌入响应）。
func DenyBase() *commonv1.BaseResp {
	return responsex.NewBaseRespWithError(CodePublishScopeDenied, "目标小区超出发布者数据范围")
}

// CheckPublishScope 是各写接口（Create/Update/Delete/Resolve/Upsert）落库前的统一数据权限校验。
//
// 返回约定：
//   - (nil, nil)      → 允许，继续落库；
//   - (denyResp, nil) → 拒绝（080006），调用方返回该响应（err=nil，交由 API 层转错误）；
//   - (nil, err)      → 身份解析失败或 permission gRPC 传输错误，调用方返回 nil, err。
//
// 身份解析：仅信任入站 gRPC metadata（API 层经 util.WithUserID 注入的 JWT 身份）。
// 不信任请求 body 携带的 publisher_id（安全考虑 #2：publisher_id/userId 伪造 → 一律取 JWT）。
// 无身份（0）→ fail-closed 拒绝。
//
// SEE: [[is-system-no-permission-shortcut]] — 拒绝走统一判定路径，无字段短路
func CheckPublishScope(ctx context.Context, client permissionv1.PermissionServiceClient, communityID int64) (*commonv1.BaseResp, error) {
	userID := UserIDFromCtx(ctx)
	if userID == 0 {
		return DenyBase(), nil
	}
	return checkScope(ctx, client, userID, communityID)
}

// CheckSystemPublishScope 以系统审核身份（SystemUserID=0，global scope）校验目标小区。
//
// 供 moderation 服务间回调使用（T4.5 / S4）：回调无用户 JWT，reverse-lookup 出内容
// community_id 后，以系统身份执行数据权限校验（不按内容作者 scope 判定，服务身份 global 放行）。
// 系统身份是合法身份（0 不代表「无身份」），必须直接走 AssertCommunityScope，不能经
// CheckPublishScope 的「userID==0 → 拒绝」分支。
//
// SEE: [[is-system-no-permission-shortcut]] — 系统身份建模为 global grant，无字段短路
func CheckSystemPublishScope(ctx context.Context, client permissionv1.PermissionServiceClient, communityID int64) (*commonv1.BaseResp, error) {
	return checkScope(ctx, client, SystemUserID, communityID)
}

func checkScope(ctx context.Context, client permissionv1.PermissionServiceClient, userID, communityID int64) (*commonv1.BaseResp, error) {
	if err := AssertCommunityScope(ctx, client, userID, communityID); err != nil {
		if IsPublishScopeDenied(err) {
			return DenyBase(), nil
		}
		return nil, err
	}
	return nil, nil
}
