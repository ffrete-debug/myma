package utils

import "github.com/gin-gonic/gin"

// APIError wraps a structured error response
type APIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func ErrorResponse(c *gin.Context, status int, msg, detail string) {
	c.JSON(status, APIError{
		Code:    status,
		Message: msg,
		Detail:  detail,
	})
}

func BadRequest(c *gin.Context, msg, detail string) {
	ErrorResponse(c, 400, msg, detail)
}

func NotFound(c *gin.Context, msg, detail string) {
	ErrorResponse(c, 404, msg, detail)
}

func InternalError(c *gin.Context, msg, detail string) {
	ErrorResponse(c, 500, msg, detail)
}

func Unauthorized(c *gin.Context, msg, detail string) {
	ErrorResponse(c, 401, msg, detail)
}
