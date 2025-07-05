package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"opencode_nano/config"
	"opencode_nano/tools"
)

type Provider struct {
	client *openai.Client
	tools  []tools.Tool
}

func NewProvider(cfg *config.Config, toolSet []tools.Tool) *Provider {
	clientConfig := openai.DefaultConfig(cfg.OpenAIAPIKey)
	clientConfig.BaseURL = cfg.OpenAIBaseURL
	client := openai.NewClientWithConfig(clientConfig)
	return &Provider{
		client: client,
		tools:  toolSet,
	}
}

// StreamResponse 发送消息并处理流式响应
func (p *Provider) StreamResponse(ctx context.Context, messages []openai.ChatCompletionMessage, onDelta func(string), onToolCall func(openai.ToolCall) (string, error)) error {
	// 准备工具定义
	var toolDefinitions []openai.Tool
	for _, tool := range p.tools {
		toolDef := openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
		toolDefinitions = append(toolDefinitions, toolDef)
	}

	req := openai.ChatCompletionRequest{
		Model:    "gpt-4.1-mini",
		Messages: messages,
		Tools:    toolDefinitions,
		Stream:   true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}
	defer stream.Close()

	var currentToolCall *openai.ToolCall

	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("stream error: %v", err)
		}

		if len(response.Choices) == 0 {
			continue
		}

		delta := response.Choices[0].Delta

		// 处理文本内容
		if delta.Content != "" {
			onDelta(delta.Content)
		}

		// 处理工具调用
		if len(delta.ToolCalls) > 0 {
			for _, toolCall := range delta.ToolCalls {
				if toolCall.ID != "" {
					// 新的工具调用
					if currentToolCall != nil {
						// 执行之前的工具调用
						result, err := p.executeToolCall(*currentToolCall)
						if err != nil {
							onDelta(fmt.Sprintf("\nTool execution error: %v\n", err))
						} else {
							onDelta(fmt.Sprintf("\nTool result: %s\n", result))
						}
					}
					currentToolCall = &openai.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
						Function: openai.FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					}
				} else if currentToolCall != nil {
					// 继续构建当前工具调用
					currentToolCall.Function.Arguments += toolCall.Function.Arguments
				}
			}
		}
	}

	// 执行最后一个工具调用
	if currentToolCall != nil {
		result, err := p.executeToolCall(*currentToolCall)
		if err != nil {
			onDelta(fmt.Sprintf("\nTool execution error: %v\n", err))
		} else {
			onDelta(fmt.Sprintf("\nTool result: %s\n", result))
		}
	}

	return nil
}

// StreamResponseWithHistory 支持历史对话的流式响应
func (p *Provider) StreamResponseWithHistory(ctx context.Context, messages []openai.ChatCompletionMessage, onDelta func(string), onToolResult func(openai.ToolCall, string)) error {
	// 准备工具定义
	var toolDefinitions []openai.Tool
	for _, tool := range p.tools {
		toolDef := openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		}
		toolDefinitions = append(toolDefinitions, toolDef)
	}

	req := openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: messages,
		Tools:    toolDefinitions,
		Stream:   true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}
	defer stream.Close()

	var currentToolCall *openai.ToolCall
	var toolCalls []openai.ToolCall

	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("stream error: %v", err)
		}

		if len(response.Choices) == 0 {
			continue
		}

		delta := response.Choices[0].Delta

		// 处理文本内容
		if delta.Content != "" {
			onDelta(delta.Content)
		}

		// 处理工具调用
		if len(delta.ToolCalls) > 0 {
			for _, toolCall := range delta.ToolCalls {
				if toolCall.ID != "" {
					// 新的工具调用
					currentToolCall = &openai.ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
						Function: openai.FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					}
					toolCalls = append(toolCalls, *currentToolCall)
				} else if currentToolCall != nil {
					// 继续构建当前工具调用
					currentToolCall.Function.Arguments += toolCall.Function.Arguments
					// 更新最后一个工具调用
					if len(toolCalls) > 0 {
						toolCalls[len(toolCalls)-1].Function.Arguments = currentToolCall.Function.Arguments
					}
				}
			}
		}
	}

	// 执行所有工具调用
	for _, toolCall := range toolCalls {
		result, err := p.executeToolCall(toolCall)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}
		onToolResult(toolCall, result)
	}

	return nil
}

func (p *Provider) executeToolCall(toolCall openai.ToolCall) (string, error) {
	// 找到对应的工具
	var targetTool tools.Tool
	for _, tool := range p.tools {
		if tool.Name() == toolCall.Function.Name {
			targetTool = tool
			break
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	// 解析参数
	var params map[string]any
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		return "", fmt.Errorf("failed to parse tool arguments: %v", err)
	}

	// 执行工具
	return targetTool.Execute(params)
}