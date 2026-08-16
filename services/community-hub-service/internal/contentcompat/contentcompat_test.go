package contentcompat

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/guxiao1976/community-common/v2/pkg/errx"
	"github.com/guxiao1976/community-hub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePostModel struct {
	model.ContentPostModel
	found bool
	err   error
}

func (f *fakePostModel) FindOneReviewComplete(ctx context.Context, id int64) (*model.ContentPost, error) {
	if f.err != nil {
		return nil, f.err
	}
	if !f.found {
		return nil, sql.ErrNoRows
	}
	return &model.ContentPost{Id: id}, nil
}

type fakeScopeModel struct {
	model.ContentPostScopeModel
	ids []int64
	err error
}

func (f *fakeScopeModel) FindCommunityIdsByPostId(ctx context.Context, postId int64) ([]int64, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.ids, nil
}

func TestResolveReadableCommunityForCompat(t *testing.T) {
	tests := []struct {
		name     string
		post     *fakePostModel
		scope    *fakeScopeModel
		filter   ScopeFilter
		wantErr  bool
		wantCode int
		want     int64
	}{
		{
			name:  "帖 scope 多小区中任一可读 → 该小区（多小区用户不 080005）",
			post:  &fakePostModel{found: true},
			scope: &fakeScopeModel{ids: []int64{2001, 2002}},
			filter: func(ctx context.Context, userID, communityID int64) (bool, error) {
				return communityID == 2002, nil
			},
			want: 2002,
		},
		{
			name:  "全部不可读 → 080001（V5 消歧，与 RPC 层 scope 外统一 080001 一致）",
			post:  &fakePostModel{found: true},
			scope: &fakeScopeModel{ids: []int64{2001, 2002}},
			filter: func(ctx context.Context, userID, communityID int64) (bool, error) {
				return false, nil
			},
			wantErr:  true,
			wantCode: CodePostNotFound,
		},
		{
			name:  "帖无 scope（数据异常）→ 080005",
			post:  &fakePostModel{found: true},
			scope: &fakeScopeModel{ids: nil},
			filter: func(ctx context.Context, userID, communityID int64) (bool, error) {
				return true, nil
			},
			wantErr:  true,
			wantCode: CodeInvalidParam,
		},
		{
			name:     "FindOneReviewComplete 未找到 → 080001",
			post:     &fakePostModel{found: false},
			scope:    &fakeScopeModel{ids: []int64{2001}},
			filter:   func(ctx context.Context, userID, communityID int64) (bool, error) { return true, nil },
			wantErr:  true,
			wantCode: CodePostNotFound,
		},
		{
			name:  "FilterAllowed 传输错误 fail-closed",
			post:  &fakePostModel{found: true},
			scope: &fakeScopeModel{ids: []int64{2001}},
			filter: func(ctx context.Context, userID, communityID int64) (bool, error) {
				return false, errors.New("permission unavailable")
			},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveReadableCommunityForCompat(context.Background(), tc.post, tc.scope, 100, 1, tc.filter)
			if tc.wantErr {
				require.Error(t, err)
				if tc.wantCode != 0 {
					ce, ok := err.(*errx.CodeError)
					require.True(t, ok)
					assert.Equal(t, tc.wantCode, ce.Code)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
