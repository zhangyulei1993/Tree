package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "tree/backend/internal/common/errors"
)

const SuccessCode = 0
const SuccessMessage = "success"

type Context = *gin.Context

type APIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type PageResponse[T any] struct {
	Records  []T   `json:"records"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

func Success[T any](ctx Context, httpStatus int, data T) {
	ctx.JSON(httpStatus, APIResponse[T]{
		Code:    SuccessCode,
		Message: SuccessMessage,
		Data:    data,
	})
}

func Fail(ctx Context, httpStatus int, code int, message string) {
	ctx.JSON(httpStatus, APIResponse[any]{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func OK[T any](ctx Context, data T) {
	Success(ctx, http.StatusOK, data)
}

func Created[T any](ctx Context, data T) {
	Success(ctx, http.StatusCreated, data)
}

func Error(ctx Context, httpStatus int, err *apperrors.BusinessError) {
	ctx.JSON(httpStatus, APIResponse[any]{
		Code:    int(err.Code),
		Message: err.Message,
		Data:    err.Data,
	})
}

func Abort(ctx Context, httpStatus int, code apperrors.Code) {
	ctx.AbortWithStatusJSON(httpStatus, APIResponse[any]{
		Code:    int(code),
		Message: apperrors.Message(code),
		Data:    nil,
	})
}
