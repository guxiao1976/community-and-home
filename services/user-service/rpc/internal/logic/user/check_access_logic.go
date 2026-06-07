package user

import (
	"context"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
)

type CheckAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckAccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckAccessLogic {
	return &CheckAccessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckAccess 权限校验（供 API Gateway 实时调用）
// 检查用户是否拥有指定角色（任一），且认证状态为"已通过"
//
// 使用场景：
//   - Gateway 对敏感操作（审核认证、发通知）做实时鉴权
//   - 规则：role_codes 任一匹配 + community_id 匹配（或 global）+ verf_status=2
func (l *CheckAccessLogic) CheckAccess(in *userv1.CheckAccessRequest) (*userv1.CheckAccessResponse, error) {
	roles, err := l.svcCtx.UserMembershipRoleModel.FindApprovedByUser(
		l.ctx, in.UserId, in.CommunityId, in.RoleCodes)
	if err != nil {
		l.Errorf("FindApprovedByUser error: %v", err)
		return nil, err
	}

	if len(roles) == 0 {
		return &userv1.CheckAccessResponse{
			Base:    responsex.NewBaseResp(),
			Allowed: false,
		}, nil
	}

	// 命中第一个匹配的角色
	matched := roles[0]
	return &userv1.CheckAccessResponse{
		Base:                responsex.NewBaseResp(),
		Allowed:             true,
		MatchedRole:         matched.RoleCode,
		MatchedCommunityId:  matched.CommunityId,
	}, nil
}
