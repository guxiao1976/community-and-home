package scope

import (
	"context"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
)

// CodeSectionQuotaExceeded 超出发布配额（板块配额超限，API 面 080007）。
// 由 CheckSectionQuota 在「占配额计数 >= 板块上限」时产出。
const CodeSectionQuotaExceeded = 80007

// SectionTypeLostFound 寻失互助板块的 section_type（对应 master-data sys_section_quota 种子 lost_found）。
// 配额按板块配置（design §2.3：lost_found=5；notice/contact/second_hand 未配置=不限）。
const SectionTypeLostFound = "lost_found"

// CheckSectionQuota 校验发布者是否超出目标板块发布配额（Task 4.2 / design §3.4）。
//
// 校验顺序（design §4.3）：功能权限（PermMiddleware）→ 数据权限（AssertPublishScope）→
// 配额校验（本函数）→ 落库。故调用方须在 AssertPublishScope 之后、Insert 之前调用。
//
// 逻辑：
//  1. MasterDataClient.GetSectionQuota(sectionType)：!configured → nil（未配置=不限）；
//  2. CountQuotaOccupied(userID, communityID, sectionType) 统计占配额内容；
//  3. count >= max_count → 返回 CodeSectionQuotaExceeded(80007)；
//  4. 否则 nil（放行）。
//
// GetSectionQuota / CountQuotaOccupied 的传输/DB 错误原样返回（fail-closed 由调用方决定）。
//
// SEE: [[grpc-only-comms]] — 消费 master-data GetSectionQuota 走 gRPC，不直连 master-data DB
func CheckSectionQuota(ctx context.Context, svcCtx *svc.ServiceContext, userID, communityID int64, sectionType string) error {
	resp, err := svcCtx.MasterDataClient.GetSectionQuota(ctx, &masterdatav1.GetSectionQuotaReq{
		SectionType: sectionType,
	})
	if err != nil {
		return err
	}
	if !resp.GetConfigured() {
		return nil
	}

	count, err := svcCtx.LostFoundItemModel.CountQuotaOccupied(ctx, userID, communityID, sectionType)
	if err != nil {
		return err
	}

	if !quotaAllowed(count, resp.GetMaxCount()) {
		return errx.NewCodeError(CodeSectionQuotaExceeded, "超出发布配额")
	}
	return nil
}

// quotaAllowed 判定已占配额数是否仍低于上限（count < max 即放行）。
func quotaAllowed(count, max int64) bool {
	return count < max
}
