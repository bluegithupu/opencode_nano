package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"

	"opencode_nano/config"
	"opencode_nano/tools"
)

type Agent struct {
	provider     *Provider
	conversation []openai.ChatCompletionMessage
}

const systemPrompt = `你是 OpenCode Nano，一个乐于助人的 AI 编程助手。你可以通过读取和写入文件以及在必要时执行 bash 命令来帮助用户完成编程任务。

**重要：你是一个自主工作的智能体**
- 你应该持续工作直到用户的任务完全解决
- 当你使用工具后，会立即收到工具执行的结果
- 基于工具执行结果，你应该继续分析并执行下一步操作
- 不要在执行一个工具后就停止，而应该根据结果继续工作
- 只有当任务真正完成时，才结束工作

请通过编辑和测试当前代码执行会话中的代码文件来解决用户的任务。你是一个已部署的编程智能体。你的会话允许你修改和运行代码。仓库已经克隆到你的工作目录中，你必须完全解决问题才能被认为是正确的答案。

**工作流程示例：**
1. 用户："创建一个简单的 web 服务器"
2. 你应该：
   - 使用 todo 工具规划任务步骤
   - 读取项目结构了解现状
   - 创建必要的代码文件
   - 运行测试验证功能
   - 更新 todo 状态标记完成
   - 报告任务完成

**多步任务管理：**
- 当遇到需要多步完成的复杂任务时，请使用 todo 工具来规划和跟踪进度
- 在开始复杂任务前，先使用 todo 工具创建任务列表
- 完成每个子任务后，及时更新 todo 状态
- 这样可以确保不遗漏任何重要步骤，并给用户清晰的进度反馈

可用工具：
- read_file：读取文件内容（无需权限，可自由使用）
- write_file：写入文件内容（需要权限，建议批量规划）
- bash：执行 bash 命令（需要权限，建议批量规划）
- todo：管理会话 todo 列表（无需权限，推荐用于复杂任务规划）

执行策略：
1. 分析阶段：先通过 read_file 充分了解代码结构和需求
2. 规划阶段：对于复杂任务，使用 todo 工具创建任务列表；制定详细的修改计划
3. 执行阶段：按计划执行，遇到问题自主调整，及时更新 todo 状态
4. 验证阶段：通过测试命令验证修改是否正确
5. 总结阶段：简要报告完成的工作，确认任务已完成

记住：
- 你会自动接收工具执行结果，应基于结果继续工作
- 不要在执行一个操作后就停下来等待
- 持续工作直到任务完全解决
- 主动使用 todo 工具进行任务管理

当前工作目录：%s`

func New(cfg *config.Config, toolSet []tools.Tool) (*Agent, error) {
	provider := NewProvider(cfg, toolSet)
	
	// 获取当前工作目录
	cwd, _ := os.Getwd()
	
	// 初始化对话历史
	conversation := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(systemPrompt, cwd),
		},
	}
	
	return &Agent{
		provider:     provider,
		conversation: conversation,
	}, nil
}

// RunOnce 执行单次对话（用于命令行参数模式）- 支持多轮自主对话
func (a *Agent) RunOnce(ctx context.Context, prompt string) error {
	fmt.Printf("🤖 OpenCode Nano is thinking...\n\n")
	
	// 添加用户消息
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	
	messages := append(a.conversation, userMsg)
	
	// 最大轮次限制，防止无限循环
	maxRounds := 10
	
	for round := 0; round < maxRounds; round++ {
		var assistantResponse string
		var toolCalls []openai.ToolCall
		hasToolCalls := false
		
		// 流式响应处理
		err := a.provider.StreamResponseWithTools(
			ctx,
			messages,
			func(delta string) {
				fmt.Print(delta)
				assistantResponse += delta
			},
			func(toolCall openai.ToolCall) {
				toolCalls = append(toolCalls, toolCall)
				hasToolCalls = true
			},
		)
		
		if err != nil {
			return fmt.Errorf("failed to get response: %v", err)
		}
		
		// 添加助手响应到消息历史
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: assistantResponse,
		})
		
		// 如果没有工具调用，说明任务完成
		if !hasToolCalls {
			break
		}
		
		// 执行所有工具调用
		fmt.Printf("\n")
		for _, toolCall := range toolCalls {
			fmt.Printf("🔧 Executing tool: %s\n", toolCall.Function.Name)
			result, err := a.provider.ExecuteToolCall(toolCall)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %v", err)
			}
			
			// 将工具结果作为用户消息添加到历史
			toolResultMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Tool [%s] result:\n%s", toolCall.Function.Name, result),
			}
			messages = append(messages, toolResultMsg)
			
			// 显示工具结果
			fmt.Printf("📝 Result: %s\n", result)
		}
		
		// 继续下一轮对话
		fmt.Printf("\n🤖 Assistant: ")
	}
	
	fmt.Printf("\n\n✅ Task completed!\n")
	return nil
}

// RunInteractive 执行交互式对话（保持对话历史）- 支持多轮自主对话
func (a *Agent) RunInteractive(ctx context.Context, prompt string) error {
	fmt.Printf("\n🤖 Assistant: ")
	
	// 添加用户消息到对话历史
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	a.conversation = append(a.conversation, userMsg)
	
	// 最大轮次限制
	maxRounds := 5 // 交互模式下轮次少一些
	
	for round := 0; round < maxRounds; round++ {
		var assistantResponse string
		var toolCalls []openai.ToolCall
		hasToolCalls := false
		
		// 流式响应处理
		err := a.provider.StreamResponseWithTools(
			ctx,
			a.conversation,
			func(delta string) {
				fmt.Print(delta)
				assistantResponse += delta
			},
			func(toolCall openai.ToolCall) {
				toolCalls = append(toolCalls, toolCall)
				hasToolCalls = true
			},
		)
		
		if err != nil {
			return fmt.Errorf("failed to get response: %v", err)
		}
		
		// 添加助手响应到对话历史
		assistantMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: assistantResponse,
		}
		a.conversation = append(a.conversation, assistantMsg)
		
		// 如果没有工具调用，结束本次交互
		if !hasToolCalls {
			break
		}
		
		// 执行所有工具调用
		fmt.Printf("\n")
		for _, toolCall := range toolCalls {
			fmt.Printf("🔧 Executing tool: %s\n", toolCall.Function.Name)
			result, err := a.provider.ExecuteToolCall(toolCall)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %v", err)
			}
			
			// 将工具结果作为用户消息添加到历史
			toolResultMsg := openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Tool [%s] result:\n%s", toolCall.Function.Name, result),
			}
			a.conversation = append(a.conversation, toolResultMsg)
			
			// 显示工具结果
			fmt.Printf("📝 Result: %s\n", result)
		}
		
		// 如果还有轮次，继续对话
		if round < maxRounds-1 {
			fmt.Printf("\n🤖 Assistant: ")
		}
	}
	
	return nil
}

// ClearConversation 清除对话历史
func (a *Agent) ClearConversation() {
	// 保留系统消息，清除其他消息
	if len(a.conversation) > 0 && a.conversation[0].Role == openai.ChatMessageRoleSystem {
		a.conversation = a.conversation[:1]
	} else {
		// 重新创建系统消息
		cwd, _ := os.Getwd()
		a.conversation = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(systemPrompt, cwd),
			},
		}
	}
}