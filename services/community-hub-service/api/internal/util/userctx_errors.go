package util

import "fmt"

// errMissingIdentity indicates no JWT user identity was found in the context.
type errMissingIdentity struct{}

func (errMissingIdentity) Error() string {
	return "未找到认证身份：JWT claims 中缺少 user_id"
}

func newErrMissingIdentity() error { return errMissingIdentity{} }

// errInvalidIdentity indicates the JWT user identity was present but not a
// usable non-zero numeric ID.
type errInvalidIdentity struct {
	value interface{}
}

func (e errInvalidIdentity) Error() string {
	return fmt.Sprintf("无效的认证身份 user_id=%v", e.value)
}

func newErrInvalidIdentity(v interface{}) error { return errInvalidIdentity{value: v} }
