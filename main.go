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
	// 检查是否有 --auto 参数
	autoMode := false
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--auto" || arg == "-a" {
			autoMode = true
			// 从参数列表中移除 --auto
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	fmt.Println("🤖 OpenCode Nano - Interactive AI Programming Assistant")
	if autoMode {
		fmt.Println("⚡ 自动模式已启用 - 所有操作将自动批准")
		fmt.Println("⚠️  警告: 请确保您信任正在执行的任务")
	}
	fmt.Println("Type 'exit' or 'quit' to exit, Ctrl+C to interrupt")
	fmt.Println(strings.Repeat("=", 50))

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 创建权限管理器
	var perm permission.Manager
	if autoMode {
		perm = permission.NewAuto()
	} else {
		perm = permission.New()
	}

	// 创建工具集
	todoTool, err := tools.NewTodoTool()
	if err != nil {
		fmt.Printf("Warning: Failed to create todo tool: %v\n", err)
		// 不影响程序运行，继续
	}
	
	toolSet := []tools.Tool{
		tools.NewReadTool(),
		tools.NewWriteTool(perm),
		tools.NewBashTool(perm),
	}
	
	// 添加 todo 工具（如果成功创建）
	if todoTool != nil {
		toolSet = append(toolSet, todoTool)
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
	if len(args) > 0 {
		prompt := strings.Join(args, " ")
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
	fmt.Print(`
📖 可用命令:
  • 直接输入您的请求与 AI 对话
  • 'clear' - 清除对话历史
  • 'help' - 显示此帮助信息  
  • 'exit' 或 'quit' - 退出程序
  • Ctrl+C - 中断当前操作

🔧 可用工具:
  • read_file - 读取文件内容
  • write_file - 写入文件（需要权限）
  • bash - 执行 shell 命令（需要权限）
  • todo - 管理会话 todo 列表（无需权限）

⚡ 启动参数:
  • --auto 或 -a - 自动模式，批准所有操作（谨慎使用）

💡 示例提示:
  • "创建一个 Go 的 hello world 程序"
  • "读取 README.md 的内容"  
  • "列出当前目录的文件"
  • "帮我调试这段代码"
  • "添加一个 todo：实现用户认证功能"
  • "查看我的 todo 列表"

🚀 自主模式使用示例:
  • ./opencode_nano --auto "重构这个项目的错误处理"
  • ./opencode_nano -a "添加单元测试并确保通过"
`)
}