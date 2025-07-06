package core

import (
	"context"
	"fmt"
	"strconv"
)

// BaseTool 提供工具的基础实现
type BaseTool struct {
	name        string
	category    string
	description string
	requiresPerm bool
	tags        []string
	schema      ParameterSchema
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, category, description string) *BaseTool {
	return &BaseTool{
		name:        name,
		category:    category,
		description: description,
		tags:        []string{},
	}
}

// Info 实现 Tool 接口
func (t *BaseTool) Info() ToolInfo {
	return ToolInfo{
		Name:         t.name,
		Category:     t.category,
		Description:  t.description,
		RequiresPerm: t.requiresPerm,
		Tags:         t.tags,
	}
}

// Schema 实现 Tool 接口
func (t *BaseTool) Schema() ParameterSchema {
	return t.schema
}

// SetRequiresPerm 设置是否需要权限
func (t *BaseTool) SetRequiresPerm(requires bool) *BaseTool {
	t.requiresPerm = requires
	return t
}

// SetTags 设置标签
func (t *BaseTool) SetTags(tags ...string) *BaseTool {
	t.tags = tags
	return t
}

// SetSchema 设置参数 schema
func (t *BaseTool) SetSchema(schema ParameterSchema) *BaseTool {
	t.schema = schema
	return t
}

// MapParameters 实现 Parameters 接口的 map 版本
type MapParameters struct {
	data map[string]any
}

// NewMapParameters 创建新的 MapParameters
func NewMapParameters(data map[string]any) *MapParameters {
	if data == nil {
		data = make(map[string]any)
	}
	return &MapParameters{data: data}
}

// Get 获取参数值
func (p *MapParameters) Get(key string) (any, error) {
	value, exists := p.data[key]
	if !exists {
		return nil, fmt.Errorf("parameter %s not found", key)
	}
	return value, nil
}

// GetString 获取字符串参数
func (p *MapParameters) GetString(key string) (string, error) {
	value, err := p.Get(key)
	if err != nil {
		return "", err
	}
	
	switch v := value.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// GetInt 获取整数参数
func (p *MapParameters) GetInt(key string) (int, error) {
	value, err := p.Get(key)
	if err != nil {
		return 0, err
	}
	
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("parameter %s is not an integer", key)
	}
}

// GetBool 获取布尔参数
func (p *MapParameters) GetBool(key string) (bool, error) {
	value, err := p.Get(key)
	if err != nil {
		return false, err
	}
	
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("parameter %s is not a boolean", key)
	}
}

// GetStringSlice 获取字符串数组参数
func (p *MapParameters) GetStringSlice(key string) ([]string, error) {
	value, err := p.Get(key)
	if err != nil {
		return nil, err
	}
	
	switch v := value.(type) {
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("parameter %s is not a string array", key)
	}
}

// Has 检查参数是否存在
func (p *MapParameters) Has(key string) bool {
	_, exists := p.data[key]
	return exists
}

// Set 设置参数值
func (p *MapParameters) Set(key string, value any) {
	p.data[key] = value
}

// Validate 验证参数
func (p *MapParameters) Validate(schema ParameterSchema) error {
	// 检查必需参数
	for _, required := range schema.Required {
		if !p.Has(required) {
			return fmt.Errorf("required parameter %s is missing", required)
		}
	}
	
	// 验证参数类型和约束
	for key, propSchema := range schema.Properties {
		if !p.Has(key) {
			continue
		}
		
		value, _ := p.Get(key)
		
		// 验证枚举值
		if len(propSchema.Enum) > 0 {
			strValue := fmt.Sprintf("%v", value)
			found := false
			for _, enum := range propSchema.Enum {
				if strValue == enum {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("parameter %s must be one of %v", key, propSchema.Enum)
			}
		}
		
		// 验证字符串长度
		if propSchema.Type == "string" {
			strValue, _ := p.GetString(key)
			if propSchema.MinLength > 0 && len(strValue) < propSchema.MinLength {
				return fmt.Errorf("parameter %s must be at least %d characters", key, propSchema.MinLength)
			}
			if propSchema.MaxLength > 0 && len(strValue) > propSchema.MaxLength {
				return fmt.Errorf("parameter %s must be at most %d characters", key, propSchema.MaxLength)
			}
		}
	}
	
	return nil
}

// Raw 获取原始 map
func (p *MapParameters) Raw() map[string]any {
	return p.data
}

// SimpleResult 实现 Result 接口的简单版本
type SimpleResult struct {
	data     any
	err      error
	metadata map[string]any
}

// NewSimpleResult 创建简单结果
func NewSimpleResult(data any) *SimpleResult {
	return &SimpleResult{
		data:     data,
		metadata: make(map[string]any),
	}
}

// NewErrorResult 创建错误结果
func NewErrorResult(err error) *SimpleResult {
	return &SimpleResult{
		err:      err,
		metadata: make(map[string]any),
	}
}

// String 返回字符串表示
func (r *SimpleResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("Error: %v", r.err)
	}
	return fmt.Sprintf("%v", r.data)
}

// Data 返回原始数据
func (r *SimpleResult) Data() any {
	return r.data
}

// Error 返回错误
func (r *SimpleResult) Error() error {
	return r.err
}

// Metadata 返回元数据
func (r *SimpleResult) Metadata() map[string]any {
	return r.metadata
}

// Success 是否成功
func (r *SimpleResult) Success() bool {
	return r.err == nil
}

// WithMetadata 添加元数据
func (r *SimpleResult) WithMetadata(key string, value any) *SimpleResult {
	r.metadata[key] = value
	return r
}

// WithError 设置错误
func (r *SimpleResult) WithError(err error) *SimpleResult {
	r.err = err
	return r
}

// DefaultLogger 默认日志器（简单实现）
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, fields ...any) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields ...any) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...any) {
	fmt.Printf("[WARN] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...any) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

// SimpleSession 简单会话实现
type SimpleSession struct {
	id   string
	data map[string]any
}

// NewSimpleSession 创建简单会话
func NewSimpleSession(id string) *SimpleSession {
	return &SimpleSession{
		id:   id,
		data: make(map[string]any),
	}
}

func (s *SimpleSession) ID() string {
	return s.id
}

func (s *SimpleSession) Get(key string) (any, bool) {
	value, exists := s.data[key]
	return value, exists
}

func (s *SimpleSession) Set(key string, value any) {
	s.data[key] = value
}

func (s *SimpleSession) Delete(key string) {
	delete(s.data, key)
}

// NewToolContext 创建工具上下文
func NewToolContext(ctx context.Context, session Session, perms PermissionChecker, logger Logger, registry Registry) *ToolContext {
	return &ToolContext{
		Context:     ctx,
		Session:     session,
		Permissions: perms,
		Logger:      logger,
		Registry:    registry,
	}
}