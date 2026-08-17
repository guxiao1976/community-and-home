package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-common/v2/pkg/snowflake"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type JoinCommunityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinCommunityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinCommunityLogic {
	return &JoinCommunityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// JoinCommunity 加入小区（per 设计文档 3.2）
func (l *JoinCommunityLogic) JoinCommunity(in *userv1.JoinCommunityRequest) (*userv1.JoinCommunityResponse, error) {
	// 0. 房号/权属全有或全无（用户拍板：加入小区=建 membership，填写房号=独立步骤）
	//   - 全提供（ownership∈{OWNED,RENTED} 且 building/unit/room>0）→ 带房号 membership + 自动授权 owner/tenant
	//   - 全不提供（无房号+无权属）→ 仅建 membership，不自动授权（网格员/物业管理员等后续 applyRole）
	//   - 部分提供（只给了房号没给权属，或反之）→ 10040（要么全有要么全无）
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]] — 无房号不自动授权，避免未认证 owner/tenant grant 立即生效
	hasResidence, roleCode, ok := joinResidenceProvided(in.Ownership, int64(in.Building), int64(in.Unit), int64(in.Room))
	if !ok {
		return &userv1.JoinCommunityResponse{
			Base: responsex.NewBaseRespWithError(10040, "房号与权属须同时提供或同时不提供"),
		}, nil
	}

	// 无房号 join → membership 地址落 0（后续业主/租户走 bindResidence + applyRole 补房号）
	resBuilding, resUnit, resRoom := 0, 0, 0
	if hasResidence {
		resBuilding, resUnit, resRoom = int(in.Building), int(in.Unit), int(in.Room)
	}

	// 0.05. 房屋地址校验（仅带房号时）：区间 + 房号格式 → 10040（权威校验，防前端绕过）
	// SEE: [[api-required-field-marked-optional]]
	if hasResidence {
		if in.Building > 200 {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10040, "楼号不能超过 200"),
			}, nil
		}
		if in.Unit > 6 {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10040, "单元号不能超过 6"),
			}, nil
		}
		// 房号：3或4位；楼号=除后2位外的前1-2位（1-55层），门牌号=后2位（01-04）
		if roomStr := fmt.Sprintf("%d", in.Room); len(roomStr) < 3 || len(roomStr) > 4 {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10040, "房号须为3或4位数字（如 502 或 1102）"),
			}, nil
		} else {
			door := roomStr[len(roomStr)-2:]
			floor, _ := strconv.Atoi(roomStr[:len(roomStr)-2])
			if floor < 1 || floor > 55 {
				return &userv1.JoinCommunityResponse{
					Base: responsex.NewBaseRespWithError(10040, "楼层不能超过 55 层"),
				}, nil
			}
			if door != "01" && door != "02" && door != "03" && door != "04" {
				return &userv1.JoinCommunityResponse{
					Base: responsex.NewBaseRespWithError(10040, "门牌号须为 01-04"),
				}, nil
			}
		}
	}

	// 0.1. 校验用户存在
	if _, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, in.UserId); err != nil {
		if err == model.ErrNotFound {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10001, "用户不存在"),
			}, nil
		}
		return nil, err
	}

	// 1. 校验：最多加入的小区数（可通过 sysconfig 动态配置）
	maxCommunities := int64(model.MaxCommunities)
	if l.svcCtx.SysConfig != nil {
		if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.max_community_join_count"); err == nil {
			maxCommunities = int64(v)
		}
	}
	count, err := l.svcCtx.UserCommunityMembershipModel.CountActiveByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("count active memberships error: %v", err)
		return nil, err
	}
	if count >= maxCommunities {
		return &userv1.JoinCommunityResponse{
			Base: responsex.NewBaseRespWithError(10006, "最多加入 3 个小区"),
		}, nil
	}

	// 2. 检查是否已加入（uk_user_community 保证不重复）
	existing, err := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("check membership error: %v", err)
		return nil, err
	}
	if existing != nil && existing.BindStatus == model.MembershipBindStatusActive {
		return &userv1.JoinCommunityResponse{
			Base: responsex.NewBaseRespWithError(10007, "不能重复加入同一个小区"),
		}, nil
	}

	// 2.5. 频次限制（首次加入新小区）：
	//   - 每年限制（10012）仅对非认证用户生效（STAGE3-1 per-community 粒度）
	//   - 终身限制（10013）对全部用户生效（对齐 spec）
	isFirstJoin := existing == nil
	if isFirstJoin && !l.isVerifiedOwnerOrTenant(in.UserId, in.CommunityId) {
		yearStart := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.Local)
		yearCount, err := l.svcCtx.UserCommunityMembershipModel.CountDistinctCommunitiesThisYear(l.ctx, in.UserId, yearStart)
		if err != nil {
			l.Errorf("count distinct communities this year error: %v", err)
			return nil, err
		}
		maxNewPerYear := int64(model.MaxNewCommunitiesPerYear)
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.max_new_communities_per_year"); err == nil {
				maxNewPerYear = int64(v)
			}
		}
		if yearCount >= maxNewPerYear {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10012, "每年最多加入 3 个新小区"),
			}, nil
		}
	}

	// 2.6. 终身限制（10013）：对所有用户生效，仅首次加入新小区时校验
	if isFirstJoin {
		totalCount, err := l.svcCtx.UserCommunityMembershipModel.CountDistinctCommunities(l.ctx, in.UserId)
		if err != nil {
			l.Errorf("count distinct communities error: %v", err)
			return nil, err
		}
		maxTotalLifetime := int64(model.MaxTotalCommunitiesLifetime)
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.max_total_communities_lifetime"); err == nil {
				maxTotalLifetime = int64(v)
			}
		}
		if totalCount >= maxTotalLifetime {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10013, "总计最多加入 12 个不同小区"),
			}, nil
		}
	}

	// 3. 每户人数校验（仅带房号时）：同小区同楼/单元/房号 active 成员 < maxHouseMembers（默认 6）→ 超限 10014
	// SEE: [[auto-grant-unverified-grant-confers-scope-level0]] — 加入即授权，房屋上限防反复退出重加入绕过
	if hasResidence {
		maxHouseMembers := int64(model.MaxHouseMembers)
		if l.svcCtx.SysConfig != nil {
			if v, err := l.svcCtx.SysConfig.GetInt(l.ctx, "user.max_house_members"); err == nil {
				maxHouseMembers = int64(v)
			}
		}
		excludeUserId := int64(0)
		if existing != nil {
			// 重新激活场景：排除当前用户自身（旧 membership 已退出，不计入）
			excludeUserId = in.UserId
		}
		houseCount, err := l.svcCtx.UserCommunityMembershipModel.CountActiveByAddress(
			l.ctx, in.CommunityId, resBuilding, resUnit, resRoom, excludeUserId)
		if err != nil {
			l.Errorf("count active by address error: %v", err)
			return nil, err
		}
		if houseCount >= maxHouseMembers {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10014, "该房屋已满员"),
			}, nil
		}
	}

	// 3. 如果之前退出过，重新激活；否则插入新记录
	now := time.Now()
	if existing != nil {
		err = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, existing.Id, model.MembershipBindStatusActive, now)
		if err != nil {
			l.Errorf("re-activate membership error: %v", err)
			return nil, err
		}
		existing.BindStatus = model.MembershipBindStatusActive
		// 更新地址信息（无房号 join → 清空为 0）
		_ = l.svcCtx.UserCommunityMembershipModel.UpdateAddress(l.ctx, existing.Id, resBuilding, resUnit, resRoom)
		existing.Building = resBuilding
		existing.Unit = resUnit
		existing.Room = resRoom
		existing.LeaveTime = sql.NullTime{}
		existing.UpdatedTime = now

		// 带房号：同步自动授权（ownership → owner/tenant）；失败补偿恢复为 left 并返回失败
		if hasResidence {
			if err := l.assignCommunityRole(in.UserId, in.CommunityId, roleCode); err != nil {
				_ = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, existing.Id, model.MembershipBindStatusLeft, now)
				return nil, err
			}
		}

		// 更新 preferences
		l.updateDefaultCommunity(in.UserId, in.CommunityId)

		return &userv1.JoinCommunityResponse{
			Base:       responsex.NewBaseResp(),
			Membership: toProtoMembership(existing),
		}, nil
	}

	membership := &model.UserCommunityMembership{
		Id:          snowflake.NextID(),
		UserId:      in.UserId,
		CommunityId: in.CommunityId,
		BindStatus:  model.MembershipBindStatusActive,
		JoinTime:    now,
		CreatedTime: now,
		UpdatedTime: now,
		Building:    resBuilding,
		Unit:        resUnit,
		Room:        resRoom,
	}

	_, err = l.svcCtx.UserCommunityMembershipModel.Insert(l.ctx, membership)
	if err != nil {
		l.Errorf("insert membership error: %v", err)
		return nil, err
	}

	// 带房号：同步自动授权（ownership → owner/tenant）；失败补偿恢复 membership 并返回失败（不留「有成员无 scope」）
	if hasResidence {
		if err := l.assignCommunityRole(in.UserId, in.CommunityId, roleCode); err != nil {
			_ = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, membership.Id, model.MembershipBindStatusLeft, time.Now())
			return nil, err
		}
	}

	// 4. 首次加入小区，设置 default_community_id
	l.updateDefaultCommunity(in.UserId, in.CommunityId)

	// 重新查询获取 ID
	created, _ := l.svcCtx.UserCommunityMembershipModel.FindByUserAndCommunity(l.ctx, in.UserId, in.CommunityId)

	l.Infof("JoinCommunity success, userId=%d, communityId=%d", in.UserId, in.CommunityId)
	return &userv1.JoinCommunityResponse{
		Base:       responsex.NewBaseResp(),
		Membership: toProtoMembership(created),
	}, nil
}

// isVerifiedOwnerOrTenant 检查用户在目标小区是否有已认证（status=approved）的业主或租户角色。
// 认证粒度 per-community（STAGE3-1）：仅校验目标小区 community_id 的认证状态，不从全局判定。
func (l *JoinCommunityLogic) isVerifiedOwnerOrTenant(userId, targetCommunityId int64) bool {
	if l.svcCtx.PermissionClient == nil {
		return false
	}
	resp, err := l.svcCtx.PermissionClient.GetUserRoles(l.ctx, &permissionv1.GetUserRolesRequest{UserId: userId})
	if err != nil {
		return false
	}
	for _, r := range resp.Roles {
		if r.Status == model.RoleVerfStatusApproved &&
			(r.Role.Code == model.RoleCodeOwner || r.Role.Code == model.RoleCodeTenant) &&
			r.ScopeId == targetCommunityId {
			return true
		}
	}
	return false
}

// joinResidenceProvided 判定加入小区请求中房号/权属的提供情况（全有或全无，用户拍板模型）。
// 返回 hasResidence（是否带房号）/ roleCode（权属映射的自动授权角色）/ ok（是否非「部分提供」）。
//   - 全提供：ownership∈{OWNED,RENTED} 且 building/unit/room>0 → (true, owner|tenant, true)
//   - 全不提供：无房号且无权属 → (false, "", true)（网格员/物业管理员等后续 applyRole，不自动授权）
//   - 部分提供：只给了房号没给权属，或反之 → (false, "", false) → 调用方返回 10040
func joinResidenceProvided(o userv1.CommunityOwnership, building, unit, room int64) (hasResidence bool, roleCode string, ok bool) {
	ownershipProvided := o == userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED ||
		o == userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_RENTED
	buildingProvided := building > 0
	unitProvided := unit > 0
	roomProvided := room > 0

	allProvided := ownershipProvided && buildingProvided && unitProvided && roomProvided
	noneProvided := !ownershipProvided && !buildingProvided && !unitProvided && !roomProvided

	switch {
	case allProvided:
		return true, ownershipRoleCode(o), true
	case noneProvided:
		return false, "", true
	default:
		return false, "", false
	}
}

// ownershipRoleCode 映射权属枚举到角色编码（自有→owner，租住→tenant）
func ownershipRoleCode(o userv1.CommunityOwnership) string {
	if o == userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_RENTED {
		return model.RoleCodeTenant
	}
	return model.RoleCodeOwner
}

// assignCommunityRole 加入小区后同步自动授权（owner/tenant, scope_type='community', scope_id=community_id, status=0）。
// role_id 经 roleMapper 解析；失败由调用方补偿 membership。
func (l *JoinCommunityLogic) assignCommunityRole(userId, communityId int64, roleCode string) error {
	roleID, ok := roleIDByCode(l.ctx, l.svcCtx, l.Logger, roleCode)
	if !ok {
		l.Errorf("JoinCommunity: role code %s not found in permission-service", roleCode)
		return fmt.Errorf("role code %s not found", roleCode)
	}
	if err := assignRoleToUser(l.ctx, l.svcCtx, l.Logger, userId, roleID,
		model.ScopeTypeCommunity, communityId, int32Ptr(0)); err != nil {
		return err
	}
	l.Infof("JoinCommunity: role %s granted, userId=%d communityId=%d roleId=%d", roleCode, userId, communityId, roleID)
	return nil
}

// updateDefaultCommunity 首次加入小区时设置默认小区偏好
func (l *JoinCommunityLogic) updateDefaultCommunity(userId, communityId int64) {
	user, err := l.svcCtx.UserBaseModel.FindOne(l.ctx, userId)
	if err != nil {
		return
	}

	// 如果已有 default_community_id，不覆盖
	prefs := user.Preferences.String
	if prefs != "" {
		var m map[string]interface{}
		if json.Unmarshal([]byte(prefs), &m) == nil {
			if _, ok := m["default_community_id"]; ok {
				return
			}
		}
	}

	// 设置 default_community_id
	newPrefs := map[string]interface{}{"default_community_id": communityId}
	b, _ := json.Marshal(newPrefs)
	user.Preferences = sql.NullString{String: string(b), Valid: true}
	_ = l.svcCtx.UserBaseModel.Update(l.ctx, user)
}
