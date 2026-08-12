// Package util provides small shared helpers for the community-hub REST API layer.
package util

import (
	"context"
	"encoding/json"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// UserIDClaimKey is the JWT claim key signed by auth-service, and also the
// gRPC metadata key used to propagate the authenticated identity from the
// REST API layer down to the RPC layer.
//
// go-zero's rest.WithJwt injects claims into the request context under their
// original claim key and decodes numeric claims as json.Number, so the value
// arrives as ctx.Value("user_id") with type json.Number — a bare type
// assertion to int64 would panic.
const UserIDClaimKey = "user_id"

// JWTUserID extracts the authenticated user ID from the REST request context.
//
// It tries "user_id" first (the standard claim key), then falls back to
// "userId" for legacy callers, and supports int64 / float64 / json.Number so
// it is safe regardless of which middleware produced the value.
//
// Unlike a lenient extractor, it returns an error when the identity is absent,
// non-numeric, or zero — write interfaces must fail closed (no identity => no
// data scope => reject). Callers that can tolerate a missing identity should
// treat the error as "no authenticated user".
//
// SEE: [[verify-api-before-calling]] — claims 结构已确认（user_id 优先 + userId 兜底）
// SEE: [[testing-discipline]] — multi-type conversion pattern shared with master-data api/internal/util/userctx.go
func JWTUserID(ctx context.Context) (int64, error) {
	v := ctx.Value(UserIDClaimKey)
	if v == nil {
		v = ctx.Value("userId")
	}
	if v == nil {
		return 0, newErrMissingIdentity()
	}

	var id int64
	switch n := v.(type) {
	case int64:
		id = n
	case float64:
		id = int64(n)
	case json.Number:
		parsed, err := n.Int64()
		if err != nil {
			return 0, newErrInvalidIdentity(v)
		}
		id = parsed
	default:
		return 0, newErrInvalidIdentity(v)
	}

	if id == 0 {
		return 0, newErrInvalidIdentity(v)
	}
	return id, nil
}

// WithUserID returns a context that propagates the authenticated user ID to
// the community-hub RPC layer as outgoing gRPC metadata. The RPC layer
// extracts it via rpc/internal/logic/scope.UserIDFromCtx.
func WithUserID(ctx context.Context, userID int64) context.Context {
	return metadata.AppendToOutgoingContext(ctx, UserIDClaimKey, strconv.FormatInt(userID, 10))
}
