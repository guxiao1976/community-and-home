package scope

import (
	"context"
	"errors"
	"testing"

	masterdatav1 "github.com/guxiao1976/api-proto/gen/go/masterdata/v1"
	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"
	"github.com/guxiao1976/community-hub/model"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// fakeMasterDataClient 仅覆盖 GetSectionQuota，其余方法嵌入不调用
type fakeMasterDataClient struct {
	masterdatav1.MasterdataServiceClient
	quotaFn func(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error)
}

func (f *fakeMasterDataClient) GetSectionQuota(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error) {
	return f.quotaFn(ctx, in, opts...)
}

// fakeQuotaModel 仅覆盖 CountQuotaOccupied
type fakeQuotaModel struct {
	model.LostFoundItemModel
	countFn func(ctx context.Context, publisherId, communityId int64, typ string) (int64, error)
}

func (f *fakeQuotaModel) CountQuotaOccupied(ctx context.Context, publisherId, communityId int64, typ string) (int64, error) {
	return f.countFn(ctx, publisherId, communityId, typ)
}

// SEE: [[testing-discipline]] — 配额判定：configured=false 不限 / count<max 放行 / count>=max → 80007 / 传输错误透传
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录：`Error: expected code 80007, actual nil`（达上限未拦截）
func TestCheckSectionQuota(t *testing.T) {
	quotaResp := func(configured bool, max int64) *masterdatav1.GetSectionQuotaResp {
		return &masterdatav1.GetSectionQuotaResp{Base: responsex.NewBaseResp(), Configured: configured, MaxCount: max}
	}

	tests := []struct {
		name        string
		resp        *masterdatav1.GetSectionQuotaResp
		quotaErr    error
		count       int64
		countErr    error
		wantErr     bool
		wantCode    int
		wantCounted bool // 是否应走到 CountQuotaOccupied
	}{
		{
			name:        "未配置=不限：configured=false → 放行（不计数）",
			resp:        quotaResp(false, 0),
			wantErr:     false,
			wantCounted: false,
		},
		{
			name:        "4/5 未达上限 → 放行",
			resp:        quotaResp(true, 5),
			count:       4,
			wantErr:     false,
			wantCounted: true,
		},
		{
			name:        "5/5 达上限 → 80007",
			resp:        quotaResp(true, 5),
			count:       5,
			wantErr:     true,
			wantCode:    CodeSectionQuotaExceeded,
			wantCounted: true,
		},
		{
			name:        "6/5 超上限 → 80007",
			resp:        quotaResp(true, 5),
			count:       6,
			wantErr:     true,
			wantCode:    CodeSectionQuotaExceeded,
			wantCounted: true,
		},
		{
			name:        "GetSectionQuota 传输错误 → 透传",
			resp:        quotaResp(true, 5),
			quotaErr:    errors.New("masterdata rpc unavailable"),
			wantErr:     true,
			wantCounted: false,
		},
		{
			name:        "CountQuotaOccupied DB 错误 → 透传",
			resp:        quotaResp(true, 5),
			countErr:    errors.New("db down"),
			wantErr:     true,
			wantCounted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var sentSectionType string
			counted := false

			md := &fakeMasterDataClient{
				quotaFn: func(ctx context.Context, in *masterdatav1.GetSectionQuotaReq, opts ...grpc.CallOption) (*masterdatav1.GetSectionQuotaResp, error) {
					sentSectionType = in.GetSectionType()
					if tc.quotaErr != nil {
						return nil, tc.quotaErr
					}
					return tc.resp, nil
				},
			}
			mdl := &fakeQuotaModel{
				countFn: func(ctx context.Context, publisherId, communityId int64, typ string) (int64, error) {
					counted = true
					if tc.countErr != nil {
						return 0, tc.countErr
					}
					return tc.count, nil
				},
			}
			sc := &svc.ServiceContext{
				MasterDataClient:   md,
				LostFoundItemModel: mdl,
			}

			err := CheckSectionQuota(context.Background(), sc, 100, 200, SectionTypeLostFound)

			assert.Equal(t, SectionTypeLostFound, sentSectionType, "GetSectionQuota 必须携带板块 section_type")
			assert.Equal(t, tc.wantCounted, counted, "CountQuotaOccupied 是否被调用须与用例一致")

			if tc.wantErr {
				require.Error(t, err)
				if tc.wantCode != 0 {
					ce := errx.FromError(err)
					require.NotNil(t, ce, "应为 CodeError，实际 %T", err)
					assert.Equal(t, tc.wantCode, ce.Code)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestQuotaAllowed 上限判定边界：count < max 放行，count == max / count > max 拒绝。
func TestQuotaAllowed(t *testing.T) {
	assert.True(t, quotaAllowed(4, 5))
	assert.False(t, quotaAllowed(5, 5))
	assert.False(t, quotaAllowed(6, 5))
}
