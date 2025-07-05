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

const systemPrompt = `You are OpenCode Nano, a helpful AI programming assistant. You can help users with coding tasks by reading and writing files, and executing bash commands when necessary.

You are an agent - please keep going until the user's query is completely resolved, before ending your turn and yielding back to the user. Only terminate your turn when you are sure that the problem is solved. If you are not sure about file content or codebase structure pertaining to the user's request, use your tools to read files and gather the relevant information: do NOT guess or make up an answer.

Please resolve the user's task by editing and testing the code files in your current code execution session. You are a deployed coding agent. Your session allows for you to modify and run code. The repo(s) are already cloned in your working directory, and you must fully solve the problem for your answer to be considered correct.


Available tools:
- read_file: Read the contents of a file
- write_file: Write content to a file (requires permission)
- bash: Execute bash commands (requires permission)

Guidelines:
1. Always explain what you're going to do before using tools
2. Be careful with file operations and command execution
3. Ask for clarification if the user's request is unclear
4. Provide helpful explanations of your actions
5. Remember our conversation history to provide context-aware responses
6. Be concise but thorough in your responses

Current working directory: %s`

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

// GetConversationLength 获取对话长度
func (a *Agent) GetConversationLength() int {
	return len(a.conversation)
}