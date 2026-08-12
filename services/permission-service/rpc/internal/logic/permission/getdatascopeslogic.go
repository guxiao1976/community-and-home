package permission

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// GetDataScopes 获取数据范围（三态重写，T1.4）
//
//	统一经 resolveUserScope 判定（global 支配 → limited 并集 → empty，REQ-A）
//	读穿缓存：perm:scopes:{userId}:{scopeType} JSON {"state","ids"}
//	  HIT 解析返回；MISS 计算后 SET + EXPIRE 30min
//
// SEE: [[redis-cache-soft-delete]] — 失效收敛到 grant 变更处理器（T1.6）
func (l *GetDataScopesLogic) GetDataScopes(in *permissionv1.GetDataScopesRequest) (*permissionv1.GetDataScopesResponse, error) {
	cacheKey := fmt.Sprintf("perm:scopes:%d:%s", in.UserId, in.ScopeType)

	// 读穿缓存 HIT → 直接返回（不查 DB）
	if raw, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result(); err == nil && raw != "" {
		var data scopeCacheData
		if json.Unmarshal([]byte(raw), &data) == nil {
			return &permissionv1.GetDataScopesResponse{
				Base:     responsex.NewBaseResp(),
				ScopeIds: data.Ids,
				State:    scopeStateFromString(data.State),
			}, nil
		}
	}

	// MISS → 计算 + 写缓存
	state, ids := resolveUserScope(l.ctx, l.svcCtx.UserRoleModel, in.UserId, in.ScopeType)
	if ids == nil {
		ids = []int64{}
	}
	if b, err := json.Marshal(scopeCacheData{State: scopeStateString(state), Ids: ids}); err == nil {
		l.svcCtx.RedisClient.Set(l.ctx, cacheKey, string(b), 30*time.Minute)
	}

	return &permissionv1.GetDataScopesResponse{
		Base:     responsex.NewBaseResp(),
		ScopeIds: ids,
		State:    state,
	}, nil
}
