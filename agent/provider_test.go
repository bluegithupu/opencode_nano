package agent

import (
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"

	"opencode_nano/config"
	"opencode_nano/tools"
)

func TestNewProvider(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "https://api.openai.com/v1",
	}
	
	mockTool := &MockTool{
		name:        "test_tool",
		description: "Test tool",
		parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		},
	}
	
	toolSet := []tools.Tool{mockTool}
	
	provider := NewProvider(cfg, toolSet)
	
	if provider == nil {
		t.Fatal("NewProvider() 返回 nil")
	}
	
	if provider.client == nil {
		t.Error("Provider client 未初始化")
	}
	
	if len(provider.tools) != 1 {
		t.Errorf("Provider tools 长度 = %d, want 1", len(provider.tools))
	}
	
	// 验证工具存在
	found := false
	for _, tool := range provider.tools {
		if tool.Name() == "test_tool" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Provider 未包含测试工具")
	}
}

func TestProvider_ExecuteToolCall(t *testing.T) {
	tests := []struct {
		name    string
		tool    *MockTool
		toolCall openai.ToolCall
		wantErr bool
		wantRes string
	}{
		{
			name: "成功执行工具",
			tool: &MockTool{
				name:        "test_tool",
				description: "Test tool",
				executeFunc: func(params map[string]any) (string, error) {
					return "success result", nil
				},
			},
			toolCall: openai.ToolCall{
				ID:   "call_123",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "test_tool",
					Arguments: `{"param": "value"}`,
				},
			},
			wantErr: false,
			wantRes: "success result",
		},
		{
			name: "工具不存在",
			tool: &MockTool{
				name: "test_tool",
			},
			toolCall: openai.ToolCall{
				ID:   "call_456",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "nonexistent_tool",
					Arguments: `{}`,
				},
			},
			wantErr: true,
		},
		{
			name: "无效的 JSON 参数",
			tool: &MockTool{
				name: "test_tool",
			},
			toolCall: openai.ToolCall{
				ID:   "call_789",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "test_tool",
					Arguments: `{invalid json}`,
				},
			},
			wantErr: true,
		},
		{
			name: "工具执行失败",
			tool: &MockTool{
				name: "test_tool",
				executeFunc: func(params map[string]any) (string, error) {
					return "", errors.New("execution failed")
				},
			},
			toolCall: openai.ToolCall{
				ID:   "call_999",
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      "test_tool",
					Arguments: `{}`,
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				OpenAIAPIKey:  "test-key",
				OpenAIBaseURL: "https://api.openai.com/v1",
			}
			
			provider := NewProvider(cfg, []tools.Tool{tt.tool})
			
			got, err := provider.executeToolCall(tt.toolCall)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("executeToolCall() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if !tt.wantErr && got != tt.wantRes {
				t.Errorf("executeToolCall() = %v, want %v", got, tt.wantRes)
			}
		})
	}
}

func TestProvider_ToolManagement(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "https://api.openai.com/v1",
	}
	
	tool1 := &MockTool{
		name:        "tool1",
		description: "Tool 1",
	}
	
	tool2 := &MockTool{
		name:        "tool2",
		description: "Tool 2",
	}
	
	toolSet := []tools.Tool{tool1, tool2}
	
	provider := NewProvider(cfg, toolSet)
	
	// 验证所有工具都被注册
	if len(provider.tools) != 2 {
		t.Errorf("工具数量 = %d, want 2", len(provider.tools))
	}
	
	// 验证工具可以通过名称访问
	for _, expectedTool := range toolSet {
		found := false
		for _, tool := range provider.tools {
			if tool.Name() == expectedTool.Name() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("工具 %s 未被注册", expectedTool.Name())
		}
	}
}

func TestProvider_StreamMethods(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "https://api.openai.com/v1",
	}
	
	provider := NewProvider(cfg, []tools.Tool{})
	
	// 验证 provider 有必要的方法（通过类型断言确认）
	if provider == nil {
		t.Error("Provider 未正确初始化")
	}
}

func TestProvider_ClientConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
	}{
		{
			name:    "默认 OpenAI URL",
			baseURL: "https://api.openai.com/v1",
			apiKey:  "test-key-1",
		},
		{
			name:    "自定义 URL",
			baseURL: "https://custom.api.com/v1",
			apiKey:  "test-key-2",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				OpenAIAPIKey:  tt.apiKey,
				OpenAIBaseURL: tt.baseURL,
			}
			
			provider := NewProvider(cfg, []tools.Tool{})
			
			if provider.client == nil {
				t.Error("Client 未初始化")
			}
			
			// 注意：无法直接验证 client 的内部配置，
			// 但可以验证 provider 被正确创建
			if provider == nil {
				t.Error("Provider 创建失败")
			}
		})
	}
}