package auth

import (
	"context"

	permissionv1 "github.com/guxiao1976/api-proto/gen/go/permission/v1"
	"github.com/guxiao1976/community-auth/rpc/internal/svc"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/zeromicro/go-zero/core/logx"
)

// 端限制错误码（access-control design §0 STAGE3-3）
// SEE: [[is-system-no-permission-shortcut]] — platforms 为配置属性，不得用 is_system 短路
const codePlatformDenied = 50007

// classifyDeviceType 将设备类型归类为端类别：
//   - web/admin → pc
//   - ios/android/miniapp → mobile
//   - 空/未知 → ""（fail-open 放行）
func classifyDeviceType(deviceType string) string {
	switch deviceType {
	case "web", "admin":
		return "pc"
	case "ios", "android", "miniapp":
		return "mobile"
	default:
		return ""
	}
}

// roleAllows 判断某角色的 platforms 是否含指定端类别。
func roleAllows(platforms []string, deviceClass string) bool {
	for _, p := range platforms {
		if p == deviceClass {
			return true
		}
	}
	return false
}

// checkPlatformAccess 端准入判定（UX 引导，非安全边界，fail-open）。
//
//  1. 归类 deviceType：空/未知 → 放行
//  2. PermissionClient.GetUserRoles(userId) 失败 → 放行（失败不锁人）并 Infof
//  3. 零角色 → 放行（fail-open，注册新用户无角色）
//  4. 遍历 roles：任一角色 platforms 为空 → 放行；任一含当前端 → 放行
//  5. 全部不满足 → 返回 50007
func checkPlatformAccess(ctx context.Context, svcCtx *svc.ServiceContext, userID int64, deviceType string) error {
	deviceClass := classifyDeviceType(deviceType)
	if deviceClass == "" {
		// 空/未知端 → fail-open 放行
		return nil
	}

	resp, err := svcCtx.PermissionClient.GetUserRoles(ctx, &permissionv1.GetUserRolesRequest{UserId: userID})
	if err != nil {
		logx.WithContext(ctx).Infof("checkPlatformAccess: GetUserRoles failed for user=%d, fail-open: %v", userID, err)
		return nil
	}
	if resp == nil || resp.GetBase().GetCode() != 0 {
		logx.WithContext(ctx).Infof("checkPlatformAccess: GetUserRoles business error for user=%d, fail-open", userID)
		return nil
	}

	roles := resp.GetRoles()
	if len(roles) == 0 {
		// 零角色（如注册新用户）→ fail-open 放行
		return nil
	}

	for _, info := range roles {
		if info == nil || info.GetRole() == nil {
			// 无角色信息 → fail-open 放行
			return nil
		}
		platforms := info.GetRole().GetPlatforms()
		if len(platforms) == 0 || roleAllows(platforms, deviceClass) {
			// 空 platforms（fail-open）或含当前端 → 放行
			return nil
		}
	}

	return errx.NewCodeError(codePlatformDenied, "该账号为移动端用户，请使用移动端 APP")
}
