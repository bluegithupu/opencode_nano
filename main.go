package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"opencode_nano/agent"
	"opencode_nano/config"
	"opencode_nano/permission"
	"opencode_nano/tools"
)

func main() {
	fmt.Println("🤖 OpenCode Nano - Interactive AI Programming Assistant")
	fmt.Println("Type 'exit' or 'quit' to exit, Ctrl+C to interrupt")
	fmt.Println(strings.Repeat("=", 50))

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 创建权限管理器
	perm := permission.New()

	// 创建工具集
	toolSet := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(perm),
		tools.NewBashTool(perm),
	}

	// 创建代理
	ag, err := agent.New(cfg, toolSet)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// 设置信号处理
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n👋 Goodbye!")
		cancel()
		os.Exit(0)
	}()

	// 如果有命令行参数，执行单次对话模式
	if len(os.Args) > 1 {
		prompt := strings.Join(os.Args[1:], " ")
		err := ag.RunOnce(ctx, prompt)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 交互式模式
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n💬 You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if input == "clear" {
			ag.ClearConversation()
			fmt.Println("🧹 Conversation cleared!")
			continue
		}

		if input == "help" {
			printHelp()
			continue
		}

		// 处理用户输入
		err := ag.RunInteractive(ctx, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

func printHelp() {
	fmt.Println(`
📖 Available commands:
  • Just type your request to chat with the AI
  • 'clear' - Clear conversation history
  • 'help' - Show this help message  
  • 'exit' or 'quit' - Exit the program
  • Ctrl+C - Interrupt current operation

🔧 Available tools:
  • read_file - Read file contents
  • write_file - Write to files (requires permission)
  • bash - Execute shell commands (requires permission)

💡 Example prompts:
  • "Create a hello world Go program"
  • "Read the contents of README.md"
  • "List files in the current directory"
  • "Help me debug this code"
`)
}