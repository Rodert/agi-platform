package errors

import "fmt"

// 错误码定义
const (
	// 通用错误 1xxx
	ErrCodeBadRequest     = 1001
	ErrCodeUnauthorized   = 1002
	ErrCodeForbidden      = 1003
	ErrCodeNotFound       = 1004
	ErrCodeInternalServer = 1005

	// 用户相关 2xxx
	ErrCodeUserNotFound      = 2001
	ErrCodeUserExists        = 2002
	ErrCodeInvalidPassword   = 2003
	ErrCodeInvalidEmail      = 2004
	ErrCodeInvalidCode       = 2005
	ErrCodeCodeExpired       = 2006
	ErrCodeCodeUsed          = 2007

	// 任务相关 3xxx
	ErrCodeTaskNotFound        = 3001
	ErrCodeTaskCreateFailed    = 3002
	ErrCodeInsufficientBalance = 3003
	ErrCodeMaxConcurrentTasks  = 3004

	// 作品相关 4xxx
	ErrCodeWorkNotFound     = 4001
	ErrCodeWorkAlreadyLiked = 4002
	ErrCodeWorkNotLiked     = 4003

	// 积分相关 5xxx
	ErrCodeInsufficientCredit = 5001
	ErrCodeInvalidRedeemCode  = 5002
	ErrCodeRedeemCodeUsed     = 5003
	ErrCodeRedeemCodeExpired  = 5004
	ErrCodeAlreadyCheckedIn   = 5005

	// 支付相关 6xxx
	ErrCodePaymentFailed       = 6001
	ErrCodeInvalidPaymentOrder = 6002
	ErrCodePaymentTimeout      = 6003

	// 上传相关 7xxx
	ErrCodeFileTooLarge      = 7001
	ErrCodeInvalidFileType   = 7002
	ErrCodeUploadFailed      = 7003

	// 邀请相关 8xxx
	ErrCodeInvalidInviteCode = 8001
	ErrCodeInviteCodeUsed    = 8002
)

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建新错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWithDetails 创建带详情的错误
func NewWithDetails(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 预定义错误
var (
	// 通用
	ErrBadRequest     = New(ErrCodeBadRequest, "请求参数错误")
	ErrUnauthorized   = New(ErrCodeUnauthorized, "未授权")
	ErrForbidden      = New(ErrCodeForbidden, "无权限访问")
	ErrNotFound       = New(ErrCodeNotFound, "资源不存在")
	ErrInternalServer = New(ErrCodeInternalServer, "服务器内部错误")

	// 用户
	ErrUserNotFound    = New(ErrCodeUserNotFound, "用户不存在")
	ErrUserExists      = New(ErrCodeUserExists, "用户已存在")
	ErrInvalidPassword = New(ErrCodeInvalidPassword, "密码错误")
	ErrInvalidEmail    = New(ErrCodeInvalidEmail, "邮箱格式错误")
	ErrInvalidCode     = New(ErrCodeInvalidCode, "验证码错误")
	ErrCodeExpired     = New(ErrCodeCodeExpired, "验证码已过期")
	ErrCodeUsed        = New(ErrCodeCodeUsed, "验证码已使用")

	// 任务
	ErrTaskNotFound        = New(ErrCodeTaskNotFound, "任务不存在")
	ErrInsufficientBalance = New(ErrCodeInsufficientBalance, "灵感值不足")
	ErrMaxConcurrentTasks  = New(ErrCodeMaxConcurrentTasks, "超过最大并发任务数")

	// 作品
	ErrWorkNotFound     = New(ErrCodeWorkNotFound, "作品不存在")
	ErrWorkAlreadyLiked = New(ErrCodeWorkAlreadyLiked, "已点赞")
	ErrWorkNotLiked     = New(ErrCodeWorkNotLiked, "未点赞")

	// 积分
	ErrInsufficientCredit = New(ErrCodeInsufficientCredit, "灵感值不足")
	ErrInvalidRedeemCode  = New(ErrCodeInvalidRedeemCode, "兑换码无效")
	ErrRedeemCodeUsed     = New(ErrCodeRedeemCodeUsed, "兑换码已使用")
	ErrRedeemCodeExpired  = New(ErrCodeRedeemCodeExpired, "兑换码已过期")
	ErrAlreadyCheckedIn   = New(ErrCodeAlreadyCheckedIn, "今天已签到")

	// 支付
	ErrPaymentFailed       = New(ErrCodePaymentFailed, "支付失败")
	ErrInvalidPaymentOrder = New(ErrCodeInvalidPaymentOrder, "订单不存在")

	// 上传
	ErrFileTooLarge    = New(ErrCodeFileTooLarge, "文件大小超过限制")
	ErrInvalidFileType = New(ErrCodeInvalidFileType, "不支持的文件类型")
	ErrUploadFailed    = New(ErrCodeUploadFailed, "文件上传失败")

	// 邀请
	ErrInvalidInviteCode = New(ErrCodeInvalidInviteCode, "邀请码无效")
	ErrInviteCodeUsed    = New(ErrCodeInviteCodeUsed, "邀请码已使用")
)
