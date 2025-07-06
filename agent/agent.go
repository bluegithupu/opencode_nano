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

你是一个智能体 - 请继续工作直到用户的查询完全解决，然后再结束你的回合并交还给用户。只有当你确定问题已解决时才终止你的回合。如果你对与用户请求相关的文件内容或代码库结构不确定，请使用你的工具来读取文件并收集相关信息：不要猜测或编造答案。

请通过编辑和测试当前代码执行会话中的代码文件来解决用户的任务。你是一个已部署的编程智能体。你的会话允许你修改和运行代码。仓库已经克隆到你的工作目录中，你必须完全解决问题才能被认为是正确的答案。

重要：为了提高效率和减少用户交互：
- 在开始任务前，先制定完整的执行计划
- 尽可能批量处理相关操作，减少往返确认
- 优先使用读取操作了解项目结构，再进行修改
- 将相关的文件操作组合在一起执行
- 遇到错误时，自主分析并尝试解决，而不是立即询问用户
- 完成任务后，主动验证结果的正确性

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
5. 总结阶段：简要报告完成的工作，清理已完成的 todo

记住：尽量减少询问用户，通过仔细分析和规划来自主完成任务。对于复杂的多步任务，主动使用 todo 工具进行任务管理。

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

// RunOnce 执行单次对话（用于命令行参数模式）
func (a *Agent) RunOnce(ctx context.Context, prompt string) error {
	fmt.Printf("🤖 OpenCode Nano is thinking...\n\n")
	
	// 添加用户消息
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	
	messages := append(a.conversation, userMsg)
	
	// 流式响应处理
	err := a.provider.StreamResponse(
		ctx,
		messages,
		func(delta string) {
			fmt.Print(delta)
		},
		func(toolCall openai.ToolCall) (string, error) {
			fmt.Printf("\n🔧 Executing tool: %s\n", toolCall.Function.Name)
			return "", nil
		},
	)
	
	if err != nil {
		return fmt.Errorf("failed to get response: %v", err)
	}
	
	fmt.Printf("\n\n✅ Task completed!\n")
	return nil
}

// RunInteractive 执行交互式对话（保持对话历史）
func (a *Agent) RunInteractive(ctx context.Context, prompt string) error {
	fmt.Printf("\n🤖 Assistant: ")
	
	// 添加用户消息到对话历史
	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	a.conversation = append(a.conversation, userMsg)
	
	// 收集助手的完整响应
	var assistantResponse string
	var toolResults []string
	
	// 流式响应处理
	err := a.provider.StreamResponseWithHistory(
		ctx,
		a.conversation,
		func(delta string) {
			fmt.Print(delta)
			assistantResponse += delta
		},
		func(toolCall openai.ToolCall, result string) {
			fmt.Printf("\n🔧 Tool %s executed\n", toolCall.Function.Name)
			if result != "" {
				fmt.Printf("📝 Result: %s\n", result)
				toolResults = append(toolResults, result)
			}
			fmt.Print("🤖 Assistant: ")
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
	
	// 如果有工具执行结果，也添加到对话历史
	for _, result := range toolResults {
		toolMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Tool execution result: %s", result),
		}
		a.conversation = append(a.conversation, toolMsg)
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