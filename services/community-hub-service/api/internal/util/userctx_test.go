package util

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// SEE: [[testing-discipline]] — table-driven, covers normal/zero/type-variant paths
// SEE: [[verify-api-before-calling]] — claims 结构已确认（user_id 优先 + userId 兜底，json.Number 形态）
func TestJWTUserID(t *testing.T) {
	cases := []struct {
		name    string
		ctx     context.Context
		want    int64
		wantErr bool
	}{
		{
			name:    "json.Number under user_id (go-zero rest.WithJwt real shape)",
			ctx:     context.WithValue(context.Background(), "user_id", json.Number("4542136688377323520")),
			want:    4542136688377323520,
			wantErr: false,
		},
		{
			name:    "int64 under user_id",
			ctx:     context.WithValue(context.Background(), "user_id", int64(7)),
			want:    7,
			wantErr: false,
		},
		{
			name:    "float64 under user_id (legacy)",
			ctx:     context.WithValue(context.Background(), "user_id", float64(7)),
			want:    7,
			wantErr: false,
		},
		{
			name:    "fallback to legacy userId key",
			ctx:     context.WithValue(context.Background(), "userId", int64(9)),
			want:    9,
			wantErr: false,
		},
		{
			name:    "user_id takes precedence over userId",
			ctx:     withBoth(userIDKey(), int64(11), "userId", int64(9)),
			want:    11,
			wantErr: false,
		},
		{
			name:    "missing claims returns error (fail-closed for writes)",
			ctx:     context.Background(),
			want:    0,
			wantErr: true,
		},
		{
			name:    "non-numeric claim returns error",
			ctx:     context.WithValue(context.Background(), "user_id", "not-a-number"),
			want:    0,
			wantErr: true,
		},
		{
			name:    "zero-valued user_id returns error",
			ctx:     context.WithValue(context.Background(), "user_id", int64(0)),
			want:    0,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := JWTUserID(tc.ctx)
			assert.Equal(t, tc.want, got)
			if tc.wantErr {
				require.Error(t, err, "expected error for %s", tc.name)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// userIDKey returns the JWT claim key so the test does not hardcode a diverging literal.
func userIDKey() string {
	return UserIDClaimKey
}

func withBoth(k1 string, v1 interface{}, k2 string, v2 interface{}) context.Context {
	ctx := context.WithValue(context.Background(), k2, v2)
	return context.WithValue(ctx, k1, v1)
}

func TestWithUserID(t *testing.T) {
	ctx := WithUserID(context.Background(), 42)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok, "outgoing metadata must be present")
	assert.Equal(t, []string{"42"}, md.Get(UserIDClaimKey), "user_id must be attached as outgoing gRPC metadata")
}

func TestWithUserID_PropagatesOverChained(t *testing.T) {
	// Simulates a request already carrying outgoing metadata, then re-wrapping
	base := context.Background()
	first := WithUserID(base, 1)
	second := WithUserID(first, 2)
	md, ok := metadata.FromOutgoingContext(second)
	assert.True(t, ok)
	assert.Contains(t, md.Get(UserIDClaimKey), "2", "last wrap must win/append user_id")
}
