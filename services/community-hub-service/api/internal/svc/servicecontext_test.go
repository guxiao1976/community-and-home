package svc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

// SEE: [[testing-discipline]] — table-driven normal/zero/error paths
func TestCallCtx(t *testing.T) {
	s := &ServiceContext{}

	t.Run("injects JWT user_id into outgoing metadata", func(t *testing.T) {
		reqCtx := context.WithValue(context.Background(), "user_id", json.Number("4542136688377323520"))
		callCtx, uid, err := s.CallCtx(reqCtx)
		require.NoError(t, err)
		assert.Equal(t, int64(4542136688377323520), uid)

		md, ok := metadata.FromOutgoingContext(callCtx)
		require.True(t, ok, "outgoing gRPC metadata must be set")
		assert.Equal(t, []string{"4542136688377323520"}, md.Get("user_id"))
	})

	t.Run("missing identity returns error (fail-closed)", func(t *testing.T) {
		_, _, err := s.CallCtx(context.Background())
		require.Error(t, err)
	})

	t.Run("zero user_id returns error", func(t *testing.T) {
		reqCtx := context.WithValue(context.Background(), "user_id", int64(0))
		_, _, err := s.CallCtx(reqCtx)
		require.Error(t, err)
	})
}
