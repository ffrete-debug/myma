package utils

// ErrorCode represents a structured API error
type ErrorCode string

const (
	ErrBadRequest       ErrorCode = "BAD_REQUEST"
	ErrUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrConflict         ErrorCode = "CONFLICT"
	ErrInternal         ErrorCode = "INTERNAL_ERROR"
	ErrRateLimited      ErrorCode = "RATE_LIMITED"
	ErrTimeout          ErrorCode = "TIMEOUT"
	ErrValidation       ErrorCode = "VALIDATION_ERROR"
	ErrForbidden        ErrorCode = "FORBIDDEN"
)

// ErrorResponseV2 is a structured API error response with error code
type ErrorResponseV2 struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
