package user

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// ApplyRole 申请角色（角色授予迁移到 permission-service 的 rel_user_role）
//   - owner/tenant/committee → 绑定小区（scope_type=community, scope_id=communityId）+ 需有效 membership
//   - grid_worker/property_admin → 绑定小区（scope_type=community, scope_id=communityId），免 membership（白名单自助申请）
//   - community_admin → 绑定小区（scope_type=community, scope_id=communityId）+ 需有效 membership
//     （security-arch CRITICAL 修复：其驱动 division 子树发布范围展开，数据范围必须绑定 membership）
//   - merchant → 全局（scope_type=global, scope_id=0），免 membership
//   - 申请时 status=0（未认证），认证通过后 permission-service 更新为 2
//
// 安全模型（用户拍板，有意反转 08-16 security-arch 回滚；SEE: [[auto-grant-unverified-grant-confers-scope-level0]]）：
//   permission-service 将 status∈{0,1,2} 的 grant 视为活跃，status=0 未认证 grant 会立即产生
//   community 数据范围 + level-0 能力（community_admin 还驱动 division 子树发布范围）。
//   grid_worker/property_admin/merchant 免 membership 自助申请是预期模型——每个申请需提交盖章文件、
//   由数据审核人员人工审核、通过后才生效；敏感权限（写/管理类）由 permission-service 的
//   min_verf_level=2 加固，未认证 grant 不能行使破坏性操作。
//   残余风险（如实披露，WARNING）：grantActive 使 status=0 未认证 grant 立即生效于
//   permission-service 的 scope 聚合（resolveUserScope/resolvePublishScope）——「审核通过后才生效」
//   仅在 min_verf_level=2 的权限层级上由代码强制，division 子树展开不校验认证状态；
//   已认证 level-2 发布角色（committee/property_admin 持 community:notice:create-api）叠加同一小区
//   未认证 community_admin grant 时，发布范围仍会按该小区所在 division 展开（越权放大边界
//   收敛为「已加入小区所在 division」，不再任意小区）。该残余风险记录于 CHANGELOG 并依赖
//   permission-service migration/004 实际执行（6 个敏感码 min_verf_level=2）后方可缓解。
//   居民角色（owner/tenant/committee）数据范围仍绑定「有效小区成员关系」——无则 10005。

type ApplyRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyRoleLogic {
	return &ApplyRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ApplyRole 申请角色
func (l *ApplyRoleLogic) ApplyRole(in *userv1.ApplyRoleRequest) (*userv1.ApplyRoleResponse, error) {
	// 0. 校验用户存在
	if _, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.UserId); err != nil {
		if err == model.ErrNotFound {
			return &userv1.ApplyRoleResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		return nil, err
	}

	// 确定作用域：
	//   - merchant → 全局（scope_type=global, scope_id=0）
	//   - 居民角色与服务角色 → 绑定小区（scope_type=community, scope_id=communityId）
	scopeType := model.ScopeTypeGlobal
	scopeId := int64(0)

	if in.RoleCode != model.RoleCodeMerchant {
		scopeType = model.ScopeTypeCommunity
		scopeId = in.CommunityId
	}

	// 居民角色（owner/tenant/committee）与 community_admin 需本小区有效 membership：
	// 数据范围绑定带房号成员关系（community_admin 驱动 division 子树发布范围展开，
	// 免 membership 自助申请对任意小区Id 即获发布放大——security-arch CRITICAL，故必须绑定）。
	// grid_worker/property_admin/merchant 免 membership——用户拍板自助申请模型：盖章文件 +
	// 数据审核人员人工审核 + 通过后生效；敏感权限由 permission-service 的 min_verf_level=2
	// 加固，未认证（status=0）grant 不能行使破坏性操作。未知角色 fail-closed（默认需 membership）。
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]]
	if needMembership(in.RoleCode) {
		membership, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
		if err != nil {
			if err == model.ErrNotFound {
				return &userv1.ApplyRoleResponse{
					Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
				}, nil
			}
			l.Errorf("find membership error: userId=%d communityId=%d roleCode=%s err=%v",
				in.UserId, in.CommunityId, in.RoleCode, err)
			return nil, err
		}
		if membership.BindStatus != model.MembershipBindStatusActive {
			return &userv1.ApplyRoleResponse{
				Base: responsex.NewBaseRespWithError(10005, "小区成员关系不存在或已退出"),
			}, nil
		}
	}

	// role_code → role_id（permission-service 的 sys_role）
	roleID, ok := roleIDByCode(l.ctx, l.svcCtx, l.Logger, in.RoleCode)
	if !ok {
		l.Errorf("ApplyRole: role_code=%s not found in permission-service", in.RoleCode)
		return &userv1.ApplyRoleResponse{
			Base: responsex.NewBaseRespWithError(10008, "角色不存在"),
		}, nil
	}

	// 调用 permission-service AssignRole（status=0 未认证）
	if l.svcCtx.PermissionClient == nil {
		l.Errorf("ApplyRole: PermissionClient is nil")
		return &userv1.ApplyRoleResponse{
			Base: responsex.NewBaseRespWithError(50000, "系统繁忙"),
		}, nil
	}
	_, err := l.svcCtx.PermissionClient.AssignRole(l.ctx, &permissionv1.AssignRoleRequest{
		UserId:    in.UserId,
		RoleId:    roleID,
		ScopeType: scopeType,
		ScopeId:   scopeId,
		Status:    int32Ptr(0), // 未认证
	})
	if err != nil {
		l.Errorf("ApplyRole: AssignRole failed userId=%d roleId=%d err=%v", in.UserId, roleID, err)
		return nil, err
	}

	l.Infof("ApplyRole success, userId=%d, roleCode=%s, roleId=%d, scope=%s:%d",
		in.UserId, in.RoleCode, roleID, scopeType, scopeId)

	return &userv1.ApplyRoleResponse{
		Base: responsex.NewBaseResp(),
		Role: &userv1.MembershipRole{
			UserId:      in.UserId,
			RoleCode:    in.RoleCode,
			CommunityId: scopeId,
			VerfStatus:  model.RoleVerfStatusUnverified,
		},
	}, nil
}

// needMembership 判断申请该角色是否需要本小区带房号 membership。
//
// 安全模型（security-arch CRITICAL 修复后，SEE: [[auto-grant-unverified-grant-confers-scope-level0]]）：
//   - 居民角色（owner/tenant/committee）数据范围绑定带房号成员关系 → true；
//   - community_admin 特权角色驱动 permission-service resolvePublishScope 的 division 子树
//     发布范围展开，且未认证(status=0) grant 亦被 grantActive 视为活跃计入展开——免 membership
//     自助申请会让任意注册用户对任意小区Id 获得发布范围放大。故其数据范围必须绑定
//     membership（禁止免 membership 自助申请）→ true；
//   - 仅显式白名单（grid_worker/property_admin/merchant）可免 membership 自助申请 → false；
//   - 未知/未来角色 fail-closed（default → true，需 membership）：任何未来在 permission-service
//     新增的特权角色（如 super_admin/moderator）若未同步此白名单，将默认要求 membership，
//     杜绝「免 membership 自助申请特权角色」提权地雷。
func needMembership(roleCode string) bool {
	switch roleCode {
	case model.RoleCodeGridWorker, model.RoleCodePropertyAdmin, model.RoleCodeMerchant:
		return false
	default:
		return true
	}
}
