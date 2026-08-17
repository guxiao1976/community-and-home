package community

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	usermocks "github.com/guxiao1976/api-proto/gen/go/user/v1/mocks"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJoinCommunityHandler_NoResidenceMinimalPayload — REST 层字段契约回归测试。
//
// 用户拍板模型：加入小区 = 建 membership，填写房号 = 独立步骤；RPC 层 joinResidenceProvided
// 已支持「无房号」join（building/unit/room=0 视为未提供，不自动授权）。
// 回归缺陷：JoinCommunityReq 的 Building/Unit/Room 未标 `,optional`（必填），httpx.Parse 在 API
// 边界对最小载荷 `{"community_id":"123"}` 返回 `"building" is not set` → 移动端
// joinCommunity('c1') 只 POST {community_id} 的无房号加入流程不可用（api_smoke 只查路由不查字段契约）。
//
// 修复：Building/Unit/Room 标记 optional（0=未提供，透传 RPC），使最小载荷在 API 边界放行。
func TestJoinCommunityHandler_NoResidenceMinimalPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userMock := usermocks.NewMockUserServiceClient(ctrl)

	// 无房号最小载荷应透传到 RPC：building/unit/room=0、ownership=UNSPECIFIED（不自动授权由 RPC 判定）
	userMock.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *userv1.JoinCommunityRequest, _ ...interface{}) (*userv1.JoinCommunityResponse, error) {
			assert.Equal(t, int64(123), req.CommunityId)
			assert.Equal(t, int32(0), req.Building, "无房号 join building 应为 0")
			assert.Equal(t, int32(0), req.Unit, "无房号 join unit 应为 0")
			assert.Equal(t, int32(0), req.Room, "无房号 join room 应为 0")
			assert.Equal(t, userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_UNSPECIFIED, req.Ownership, "无房号 join ownership 应为 UNSPECIFIED")
			return &userv1.JoinCommunityResponse{
				Base: &commonv1.BaseResp{Code: 0, Msg: "success"},
				Membership: &userv1.CommunityMembership{
					Id: 999, CommunityId: 123, Building: 0, Unit: 0, Room: 0,
				},
			}, nil
		})

	svcCtx := &svc.ServiceContext{UserRpc: userMock}
	handler := JoinCommunityHandler(svcCtx)

	// 最小载荷：仅 community_id（移动端无房号 joinCommunity('c1') 的实际请求体）
	req := httptest.NewRequest(http.MethodPost, "/api/users/communities/join", strings.NewReader(`{"community_id":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", json.Number("7001")) // 模拟 JWT
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler(w, req)

	// httpx.Parse 失败 → responsex 返回 code=500 + `"building" is not set`；修复后 code=0 且 RPC 被调用
	var body struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Code, "无房号最小载荷应在 API 边界放行并透传 RPC（当前被必填字段拦截）")
}

// TestJoinCommunityHandler_WithResidence_PayloadPassesThrough — 带房号+权属载荷正常透传（字段契约不破）。
func TestJoinCommunityHandler_WithResidence_PayloadPassesThrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userMock := usermocks.NewMockUserServiceClient(ctrl)

	userMock.EXPECT().JoinCommunity(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, req *userv1.JoinCommunityRequest, _ ...interface{}) (*userv1.JoinCommunityResponse, error) {
			assert.Equal(t, int64(123), req.CommunityId)
			assert.Equal(t, int32(5), req.Building)
			assert.Equal(t, int32(2), req.Unit)
			assert.Equal(t, int32(502), req.Room)
			assert.Equal(t, userv1.CommunityOwnership_COMMUNITY_OWNERSHIP_OWNED, req.Ownership)
			return &userv1.JoinCommunityResponse{
				Base: &commonv1.BaseResp{Code: 0, Msg: "success"},
				Membership: &userv1.CommunityMembership{
					Id: 999, CommunityId: 123, Building: 5, Unit: 2, Room: 502,
				},
			}, nil
		})

	svcCtx := &svc.ServiceContext{UserRpc: userMock}
	handler := JoinCommunityHandler(svcCtx)

	req := httptest.NewRequest(http.MethodPost, "/api/users/communities/join",
		strings.NewReader(`{"community_id":"123","building":5,"unit":2,"room":502,"ownership":1}`))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", json.Number("7001"))
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler(w, req)

	var body struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 0, body.Code)
}
