package tools

// Tool 工具接口
type Tool interface {
	Name() string                    // 工具名称
	Description() string             // 工具描述
	Parameters() map[string]any // 工具参数定义
	Execute(params map[string]any) (string, error) // 执行工具
}

// ToOpenAIFunction 将工具转换为 OpenAI 函数定义
func ToOpenAIFunction(tool Tool) map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        tool.Name(),
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		},
	}
}

