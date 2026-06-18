package permission

import (
	"context"
	"fmt"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-permission/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDataScopesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDataScopesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDataScopesLogic {
	return &GetDataScopesLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// GetDataScopes 获取数据范围（spec/permission.md 核心逻辑流 2）
//   查 rel_user_role → 根据 user_id + scope_type 返回 scope_ids
//   先查 Redis 缓存，未命中则查 DB
func (l *GetDataScopesLogic) GetDataScopes(in *permissionv1.GetDataScopesRequest) (*permissionv1.GetDataScopesResponse, error) {
	// 查 DB
	scopeIds, err := l.svcCtx.UserRoleModel.FindScopesByUserId(l.ctx, in.UserId, in.ScopeType)
	if err != nil || len(scopeIds) == 0 {
		return &permissionv1.GetDataScopesResponse{
			Base:     responsex.NewBaseResp(),
			ScopeIds: []int64{},
		}, nil
	}

	// 回填 Redis 缓存
	cacheKey := fmt.Sprintf("perm:scopes:%d:%s", in.UserId, in.ScopeType)
	for _, id := range scopeIds {
		l.svcCtx.RedisClient.SAdd(l.ctx, cacheKey, fmt.Sprintf("%d", id))
	}
	l.svcCtx.RedisClient.Expire(l.ctx, cacheKey, 30*60*1e9)

	return &permissionv1.GetDataScopesResponse{
		Base:     responsex.NewBaseResp(),
		ScopeIds: scopeIds,
	}, nil
}
