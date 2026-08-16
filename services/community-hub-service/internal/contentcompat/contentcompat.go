// Package contentcompat 承载详情 community_id 兼容回退的共享实现（R2，评审 interface v3 MUST 2 修复）。
//
// 为什么独立包：`rpc/internal/...` 受 Go internal 规则限制，REST API 层（api/...）无法导入；
// 而兼容回退按设计只落 REST 薄代理层（RPC 层 GetContentPost 保持 community_id 严格必填）。
// 故将核心逻辑收敛到 `github.com/guxiao1976/community-hub/internal/...`（模块根级 internal，
// 父目录为模块根，api/ 与 rpc/ 均可导入），由 API 层注入 scope 过滤回调，避免双实现漂移。
package contentcompat

import (
	"context"
	"database/sql"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-hub/model"
)

// 错误码（与 rpc scope 包对齐；避免跨层依赖 rpc/internal 常量）。
const (
	CodePostNotFound = 80001 // 内容帖不存在
	CodeInvalidParam = 80005 // 参数无效（帖无任何 scope 小区，数据异常）
)

// ErrorInvalidParam 构造 080005 业务错误（供 API 薄代理层透传）。
func ErrorInvalidParam(msg string) error {
	return errx.NewCodeError(CodeInvalidParam, msg)
}

// ScopeFilter 读范围判定函数（GLOBAL/LIMITED/EMPTY 语义），由调用方注入：
//   - RPC 侧经 rpc/internal/logic/scope.FilterAllowed（GetDataScopes）；
//   - API 侧经本包 FilterAllowed 或调用方等价实现。
type ScopeFilter func(ctx context.Context, userID, communityID int64) (bool, error)

// ResolveReadableCommunityForCompat 详情 community_id 兼容回退（R2，评审 interface v3 MUST 2 修复——
// 取消 grant 唯一假设，改 scope 反查）。
//
//   - FindOneReviewComplete(postId) 未找到 → 080001；
//   - ContentPostScopeModel.FindCommunityIdsByPostId(postId) 取帖所属小区集；
//   - 对每小区 filter(userID, community_id) 任一允许即放行（多小区 grid_worker / 多房产业主迁移后详情仍可用）；
//   - 帖有 scope 但全部不可读 → 080001（V5 消歧：与 RPC 层 scope 外统一 080001 一致，不泄露）；
//   - 帖无任何 scope 小区（数据异常）→ 080005。
//
// RPC 层 GetContentPost 保持 community_id 必填不变（新消费方走严格契约），回退只落 REST 薄代理层。
func ResolveReadableCommunityForCompat(ctx context.Context, postModel model.ContentPostModel, scopeModel model.ContentPostScopeModel,
	userID, postID int64, filter ScopeFilter) (int64, error) {

	if _, err := postModel.FindOneReviewComplete(ctx, postID); err != nil {
		if err == sql.ErrNoRows {
			return 0, errx.NewCodeError(CodePostNotFound, "内容帖不存在")
		}
		return 0, err
	}

	communities, err := scopeModel.FindCommunityIdsByPostId(ctx, postID)
	if err != nil {
		return 0, err
	}
	if len(communities) == 0 {
		return 0, errx.NewCodeError(CodeInvalidParam, "内容帖无范围小区")
	}
	for _, cid := range communities {
		allowed, err := filter(ctx, userID, cid)
		if err != nil {
			return 0, err
		}
		if allowed {
			return cid, nil
		}
	}
	return 0, errx.NewCodeError(CodePostNotFound, "内容帖不存在")
}
