package errorx

// Common error codes
const (
	CodeSuccess           = 0
	CodeInvalidParam      = 400
	CodeUnauthorized      = 401
	CodeForbidden         = 403
	CodeNotFound          = 404
	CodeInternalError     = 500
	CodeDatabaseError     = 501
	CodeCacheError        = 502
	CodeRPCError          = 503
)

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *CodeError {
	return e
}

func NewCodeError(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

func NewDefaultError(msg string) error {
	return NewCodeError(CodeInternalError, msg)
}

func NewInvalidParamError(msg string) error {
	return NewCodeError(CodeInvalidParam, msg)
}

func NewUnauthorizedError(msg string) error {
	return NewCodeError(CodeUnauthorized, msg)
}

func NewForbiddenError(msg string) error {
	return NewCodeError(CodeForbidden, msg)
}

func NewNotFoundError(msg string) error {
	return NewCodeError(CodeNotFound, msg)
}
