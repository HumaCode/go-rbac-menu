package response

import (
	"github.com/nelthaarion/breeze/v2"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Success returns a standardized 2xx JSON response
func Success(ctx *breeze.Context, statusCode int, message string, data interface{}) error {
	ctx.Status(statusCode)
	return ctx.JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error returns a standardized error JSON response
func Error(ctx *breeze.Context, statusCode int, message string, errDetail interface{}) error {
	ctx.Status(statusCode)
	var errStr interface{}
	if errVal, ok := errDetail.(error); ok && errVal != nil {
		errStr = errVal.Error()
	} else {
		errStr = errDetail
	}

	return ctx.JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   errStr,
	})
}

