package core

import (
	"context"
)

// Tool 主工具接口
type Tool interface {
	// Info 返回工具的基本信息
	Info() ToolInfo

	// Execute 执行工具
	Execute(ctx context.Context, params Parameters) (Result, error)

	// Schema 返回参数 schema
	Schema() ParameterSchema
}

// ToolInfo 工具信息
type ToolInfo struct {
	Name         string   // 工具名称
	Category     string   // 工具分类
	Description  string   // 工具描述
	RequiresPerm bool     // 是否需要权限
	Tags         []string // 标签
}

// Parameters 参数接口
type Parameters interface {
	// Get 获取参数值
	Get(key string) (any, error)
	
	// GetString 获取字符串参数
	GetString(key string) (string, error)
	
	// GetInt 获取整数参数
	GetInt(key string) (int, error)
	
	// GetBool 获取布尔参数
	GetBool(key string) (bool, error)
	
	// GetStringSlice 获取字符串数组参数
	GetStringSlice(key string) ([]string, error)
	
	// Has 检查参数是否存在
	Has(key string) bool
	
	// Validate 验证参数
	Validate(schema ParameterSchema) error
	
	// Raw 获取原始 map
	Raw() map[string]any
}

// Result 结果接口
type Result interface {
	// String 返回字符串表示
	String() string
	
	// Data 返回原始数据
	Data() any
	
	// Error 返回错误（如果有）
	Error() error
	
	// Metadata 返回元数据
	Metadata() map[string]any
	
	// Success 是否成功
	Success() bool
}

// ParameterSchema 参数 schema
type ParameterSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]PropertySchema  `json:"properties"`
	Required   []string                   `json:"required"`
}

// PropertySchema 属性 schema
type PropertySchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	MinLength   int      `json:"minLength,omitempty"`
	MaxLength   int      `json:"maxLength,omitempty"`
}

// AsyncTool 异步工具接口
type AsyncTool interface {
	Tool
	ExecuteAsync(ctx context.Context, params Parameters) <-chan Result
}

// PermissionChecker 权限检查器接口
type PermissionChecker interface {
	// Check 检查单个工具的权限
	Check(tool Tool, params Parameters) error
	
	// RequestBatch 批量请求权限
	RequestBatch(requests []PermissionRequest) error
}

// PermissionRequest 权限请求
type PermissionRequest struct {
	Tool        Tool
	Action      string
	Description string
	Params      Parameters
}

// Registry 工具注册表接口
type Registry interface {
	// Register 注册工具
	Register(tool Tool, aliases ...string) error
	
	// Get 获取工具
	Get(name string) (Tool, error)
	
	// Find 查找工具
	Find(query string) []Tool
	
	// GetByCategory 按分类获取工具
	GetByCategory(category string) []Tool
	
	// GetByTags 按标签获取工具
	GetByTags(tags ...string) []Tool
	
	// All 获取所有工具
	All() []Tool
	
	// Categories 获取所有分类
	Categories() []string
}

// Pipeline 工具管道接口
type Pipeline interface {
	// Add 添加工具到管道
	Add(tool Tool, params Parameters) Pipeline
	
	// Execute 执行管道
	Execute(ctx context.Context) ([]Result, error)
	
	// ExecuteAsync 异步执行管道
	ExecuteAsync(ctx context.Context) <-chan Result
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// Session 会话接口
type Session interface {
	// ID 获取会话 ID
	ID() string
	
	// Get 获取会话数据
	Get(key string) (any, bool)
	
	// Set 设置会话数据
	Set(key string, value any)
	
	// Delete 删除会话数据
	Delete(key string)
}

// ToolContext 工具上下文
type ToolContext struct {
	context.Context
	Session     Session
	Permissions PermissionChecker
	Logger      Logger
	Registry    Registry
}