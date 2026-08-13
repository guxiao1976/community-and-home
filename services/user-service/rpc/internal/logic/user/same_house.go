package user

import (
	"context"

	"github.com/guxiao1976/community-user/rpc/internal/svc"
)

// maskPhone 手机号脱敏：138****1234。
// 仅对 11 位手机号脱敏；解密失败留下的密文/非手机号原样返回（兜底，避免把密文脱敏成脏数据）。
// SEE: [[phone-encryption]]、[[pii-plaintext-logging]] — 严禁明文日志，脱敏与加密存储互斥。
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// isSameHouse 判断 viewer 与 target 是否有「同小区同楼/单元/房号」的 active membership。
// 双方地址均非零（有效房屋）且 community+building+unit+room 全同 → true（同屋明文可见）。
func isSameHouse(ctx context.Context, svcCtx *svc.ServiceContext, viewerID, targetID int64) (bool, int32, int32, int32, error) {
	viewer, err := svcCtx.UserCommunityMembershipModel.FindByUserId(ctx, viewerID)
	if err != nil {
		return false, 0, 0, 0, err
	}
	target, err := svcCtx.UserCommunityMembershipModel.FindByUserId(ctx, targetID)
	if err != nil {
		return false, 0, 0, 0, err
	}

	for _, vm := range viewer {
		// 地址未采集（0）的成员不参与同屋判定，避免「两个无地址成员被误判同屋」泄露明文手机号
		if vm.Building <= 0 || vm.Unit <= 0 || vm.Room <= 0 {
			continue
		}
		for _, tm := range target {
			if tm.Building <= 0 || tm.Unit <= 0 || tm.Room <= 0 {
				continue
			}
			if vm.CommunityId == tm.CommunityId && vm.Building == tm.Building && vm.Unit == tm.Unit && vm.Room == tm.Room {
				return true, int32(vm.Building), int32(vm.Unit), int32(vm.Room), nil
			}
		}
	}
	return false, 0, 0, 0, nil
}

// ownHouseInfo 返回用户自身第一条 active membership 的楼/单元/房号（查看自身时回显）。
func ownHouseInfo(ctx context.Context, svcCtx *svc.ServiceContext, userId int64) (int32, int32, int32) {
	ms, err := svcCtx.UserCommunityMembershipModel.FindByUserId(ctx, userId)
	if err != nil || len(ms) == 0 {
		return 0, 0, 0
	}
	return int32(ms[0].Building), int32(ms[0].Unit), int32(ms[0].Room)
}
