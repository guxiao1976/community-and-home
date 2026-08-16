package scope

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
)

// 发布角色 code（permission-service RBAC 角色编码，D6）。
// grid_worker/community_admin/committee/property_admin 为可发布角色；owner/tenant/merchant/sys_admin 不可。
const (
	RoleGridWorker     = "grid_worker"
	RoleCommunityAdmin = "community_admin"
	RoleCommittee      = "committee"
	RolePropertyAdmin  = "property_admin"
)

// 个体角色生命周期状态（permission UserRoleInfo.Status）：0=未认证 1=待审 2=已认证 3=已驳回 4=已过期。
const (
	UserRoleStatusPending  int32 = 1
	UserRoleStatusVerified int32 = 2
)

// ExpandDivisionCommunities 将唯一管辖 division 展开为其下全部通过(approved)小区 id 快照（Task 1.7）。
//
// 先 guard `divisionID<=0 → 080005`（fail-closed，杜绝进入 masterdata 默认分支过度展开）；
// 调 masterdata `GetResidentialAreasByDivision(community_div_id=divisionID, status=1)` 返回 approved 小区；
// 展开为空 → 080005（REVISION-10/A2 语义：community_admin 唯一管辖 division 展开为空）。
// 传输错误原样返回（fail-closed 由调用方决定）。
//
// SEE: [[grpc-only-comms]] — division 展开经 masterdata gRPC，不直连 md_residential_area
// SEE: [[grpc-timeout-layers]] — GetResidentialAreasByDivision 内嵌 RPC 超时对齐
func ExpandDivisionCommunities(ctx context.Context, mdClient masterdatav1.MasterdataServiceClient, divisionID int64) ([]int64, error) {
	if divisionID <= 0 {
		return nil, errx.NewCodeError(CodeInvalidParam, "division 参数无效")
	}
	resp, err := mdClient.GetResidentialAreasByDivision(ctx, &masterdatav1.GetResidentialAreasByDivisionReq{
		CommunityDivId: divisionID,
		Status:         1, // approved only
	})
	if err != nil {
		return nil, err
	}
	areas := resp.GetResidentialAreas()
	if len(areas) == 0 {
		return nil, errx.NewCodeError(CodeInvalidParam, "管辖 division 下无通过小区")
	}
	ids := make([]int64, 0, len(areas))
	for _, a := range areas {
		if a.GetId() <= 0 {
			continue
		}
		ids = append(ids, a.GetId())
	}
	if len(ids) == 0 {
		return nil, errx.NewCodeError(CodeInvalidParam, "管辖 division 下无通过小区")
	}
	return ids, nil
}

// ResolveAdminDivision 经社区管理员的既有 `scope_type='community'` grant 派生其唯一管辖 division（R1 grounded）。
//
// 逻辑（Task 1.7，R1 重写 + 评审 M2 状态过滤）：
//  1. 经 permission `GetUserRoles(user_id)` 取 role_code=community_admin 且 scope_type='community'
//     且 scope_id!=0 且 **URStatus==2（level-2 等价：status==2 且 verified_at>0 且未过期）** 的 grant
//     （scope_id 为 communityId 集；禁止直读 rel_user_role）；
//  2. 逐条 masterdata `GetResidentialArea(scope_id).community_div_id` → 收集 distinct division 集；
//  3. **空 → 080005**（非 admin 走 community_ids 直传路径）；**>1 个 distinct division → 080005**
//     （「唯一管辖」契约守卫，fail-closed，评审 I4 语义保留）→ 返回唯一 division。
//
// 已过期(4)/已驳回(3) 的 community_admin grant 不计入（URStatus 过滤）——即使该用户另有
// level-2 发布角色（committee/grid_worker 也持 421），也不能用失效的 community_admin grant 驱动
// division 展开（评审 M2 权限提升修复）。
//
// SEE: [[grpc-only-comms]] — 经 GetUserRoles/GetResidentialArea，禁止直读 rel_user_role
// SEE: [[grpc-timeout-layers]] — GetUserRoles/GetResidentialArea 内嵌跨服务 RPC 超时对齐
func ResolveAdminDivision(ctx context.Context, permClient permissionv1.PermissionServiceClient, mdClient masterdatav1.MasterdataServiceClient, userID int64) (int64, error) {
	resp, err := permClient.GetUserRoles(ctx, &permissionv1.GetUserRolesRequest{UserId: userID})
	if err != nil {
		return 0, err
	}

	communityIDs := make([]int64, 0)
	for _, ur := range resp.GetRoles() {
		if !IsLevel2Grant(ur) {
			continue
		}
		if ur.GetRole().GetCode() != RoleCommunityAdmin {
			continue
		}
		if ur.GetScopeType() != ScopeTypeCommunity || ur.GetScopeId() == 0 {
			continue
		}
		communityIDs = append(communityIDs, ur.GetScopeId())
	}
	if len(communityIDs) == 0 {
		return 0, errx.NewCodeError(CodeInvalidParam, "非社区管理员（无有效 community 管辖）")
	}

	divisions := make(map[int64]struct{})
	for _, cid := range communityIDs {
		areaResp, err := mdClient.GetResidentialArea(ctx, &masterdatav1.GetResidentialAreaReq{Id: cid})
		if err != nil {
			return 0, err
		}
		area := areaResp.GetResidentialArea()
		if area == nil || area.GetCommunityDivId() <= 0 {
			return 0, errx.NewCodeError(CodeInvalidParam, "管辖小区无 division 归属")
		}
		divisions[area.GetCommunityDivId()] = struct{}{}
	}
	if len(divisions) > 1 {
		return 0, errx.NewCodeError(CodeInvalidParam, "社区管理员管辖多个 division")
	}
	for d := range divisions {
		return d, nil
	}
	return 0, errx.NewCodeError(CodeInvalidParam, "无管辖 division")
}
