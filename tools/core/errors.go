package core

import (
	"fmt"
)

// 错误代码常量
const (
	ErrCodeInvalidParams    = "INVALID_PARAMS"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeToolNotFound     = "TOOL_NOT_FOUND"
	ErrCodeExecutionFailed  = "EXECUTION_FAILED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeCancelled        = "CANCELLED"
	ErrCodeNotImplemented   = "NOT_IMPLEMENTED"
	ErrCodeInternalError    = "INTERNAL_ERROR"
)

// ToolError 工具错误
type ToolError struct {
	Code      string         // 错误代码
	Message   string         // 错误消息
	Tool      string         // 工具名称
	Params    map[string]any // 相关参数
	Cause     error          // 原因错误
	Retryable bool           // 是否可重试
}

// Error 实现 error 接口
func (e *ToolError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Code, e.Tool, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Tool, e.Message)
}

// Unwrap 实现错误解包
func (e *ToolError) Unwrap() error {
	return e.Cause
}

// NewToolError 创建新的工具错误
func NewToolError(code, tool, message string) *ToolError {
	return &ToolError{
		Code:    code,
		Tool:    tool,
		Message: message,
	}
}

// WithCause 设置原因错误
func (e *ToolError) WithCause(cause error) *ToolError {
	e.Cause = cause
	return e
}

// WithParams 设置相关参数
func (e *ToolError) WithParams(params map[string]any) *ToolError {
	e.Params = params
	return e
}

// WithRetryable 设置是否可重试
func (e *ToolError) WithRetryable(retryable bool) *ToolError {
	e.Retryable = retryable
	return e
}

// 常用错误构造函数

// ErrInvalidParams 创建参数无效错误
func ErrInvalidParams(tool, message string) *ToolError {
	return NewToolError(ErrCodeInvalidParams, tool, message)
}

// ErrPermissionDenied 创建权限拒绝错误
func ErrPermissionDenied(tool, action string) *ToolError {
	return NewToolError(ErrCodePermissionDenied, tool, fmt.Sprintf("permission denied for action: %s", action))
}

// ErrToolNotFound 创建工具未找到错误
func ErrToolNotFound(name string) *ToolError {
	return NewToolError(ErrCodeToolNotFound, name, fmt.Sprintf("tool not found: %s", name))
}

// ErrExecutionFailed 创建执行失败错误
func ErrExecutionFailed(tool, message string) *ToolError {
	return NewToolError(ErrCodeExecutionFailed, tool, message)
}

// ErrTimeout 创建超时错误
func ErrTimeout(tool string) *ToolError {
	return NewToolError(ErrCodeTimeout, tool, "execution timeout").WithRetryable(true)
}

// ErrCancelled 创建取消错误
func ErrCancelled(tool string) *ToolError {
	return NewToolError(ErrCodeCancelled, tool, "execution cancelled")
}

// ErrNotImplemented 创建未实现错误
func ErrNotImplemented(tool, feature string) *ToolError {
	return NewToolError(ErrCodeNotImplemented, tool, fmt.Sprintf("feature not implemented: %s", feature))
}

// ErrInternalError 创建内部错误
func ErrInternalError(tool, message string) *ToolError {
	return NewToolError(ErrCodeInternalError, tool, message)
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	if toolErr, ok := err.(*ToolError); ok {
		return toolErr.Retryable
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if toolErr, ok := err.(*ToolError); ok {
		return toolErr.Code
	}
	return ErrCodeInternalError
}