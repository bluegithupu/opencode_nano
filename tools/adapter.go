package tools

import (
	"context"
	"fmt"

	"opencode_nano/tools/core"
)

// LegacyToolAdapter 将新工具适配为旧接口
type LegacyToolAdapter struct {
	tool core.Tool
}

// NewLegacyAdapter 创建适配器
func NewLegacyAdapter(tool core.Tool) Tool {
	return &LegacyToolAdapter{tool: tool}
}

// Name 返回工具名称
func (a *LegacyToolAdapter) Name() string {
	return a.tool.Info().Name
}

// Description 返回工具描述
func (a *LegacyToolAdapter) Description() string {
	return a.tool.Info().Description
}

// Parameters 返回参数定义
func (a *LegacyToolAdapter) Parameters() map[string]any {
	schema := a.tool.Schema()
	params := make(map[string]any)
	
	// 转换为 OpenAI 函数格式的参数
	params["type"] = "object"
	params["properties"] = make(map[string]any)
	params["required"] = schema.Required
	
	properties := params["properties"].(map[string]any)
	
	// 转换参数定义
	for name, prop := range schema.Properties {
		paramDef := map[string]any{
			"type":        prop.Type,
			"description": prop.Description,
		}
		
		// 处理枚举值
		if len(prop.Enum) > 0 {
			paramDef["enum"] = prop.Enum
		}
		
		// 处理默认值
		if prop.Default != nil {
			paramDef["default"] = prop.Default
		}
		
		properties[name] = paramDef
	}
	
	return params
}

// Execute 执行工具
func (a *LegacyToolAdapter) Execute(params map[string]interface{}) (string, error) {
	// 转换参数
	coreParams := core.NewMapParameters(params)
	
	// 执行工具
	ctx := context.Background()
	result, err := a.tool.Execute(ctx, coreParams)
	if err != nil {
		return "", err
	}
	
	// 返回结果
	return fmt.Sprintf("%v", result.Data()), nil
}


// AdaptAllTools 将所有新工具适配为旧接口
func AdaptAllTools() map[string]Tool {
	if DefaultRegistry == nil {
		if _, err := InitializeRegistry(); err != nil {
			return map[string]Tool{}
		}
	}
	
	tools := make(map[string]Tool)
	for _, tool := range DefaultRegistry.All() {
		adapter := NewLegacyAdapter(tool)
		tools[tool.Info().Name] = adapter
	}
	
	return tools
}

// contains 检查切片是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}