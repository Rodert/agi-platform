package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List       interface{} `json:"list"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var statusCode int
	var errorInfo *ErrorInfo

	// 判断是否是自定义错误
	if e, ok := err.(*errors.AppError); ok {
		statusCode = getStatusCodeFromErrorCode(e.Code)
		errorInfo = &ErrorInfo{
			Code:    e.Code,
			Message: e.Message,
			Details: e.Details,
		}
	} else {
		// 未知错误
		statusCode = http.StatusInternalServerError
		errorInfo = &ErrorInfo{
			Code:    errors.ErrCodeInternalServer,
			Message: "服务器内部错误",
			Details: err.Error(),
		}
	}

	c.JSON(statusCode, Response{
		Success:   false,
		Error:     errorInfo,
		Timestamp: time.Now().Unix(),
	})
}

// ErrorWithCode 指定状态码的错误响应
func ErrorWithCode(c *gin.Context, statusCode int, err *errors.AppError) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
		Timestamp: time.Now().Unix(),
	})
}

// Page 分页响应
func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, PageData{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// getStatusCodeFromErrorCode 根据错误码获取 HTTP 状态码
func getStatusCodeFromErrorCode(code int) int {
	switch {
	case code >= 1000 && code < 2000:
		// 通用错误
		switch code {
		case errors.ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case errors.ErrCodeForbidden:
			return http.StatusForbidden
		case errors.ErrCodeNotFound:
			return http.StatusNotFound
		default:
			return http.StatusBadRequest
		}
	case code >= 2000 && code < 3000:
		// 用户相关错误
		return http.StatusBadRequest
	case code >= 3000 && code < 4000:
		// 任务相关错误
		return http.StatusBadRequest
	case code >= 4000 && code < 5000:
		// 作品相关错误
		return http.StatusBadRequest
	case code >= 5000 && code < 6000:
		// 积分相关错误
		return http.StatusBadRequest
	case code >= 6000 && code < 7000:
		// 支付相关错误
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
