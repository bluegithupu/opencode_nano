# OpenCode Nano 使用示例

## 基本使用

```bash
# 设置 API Key
export OPENAI_API_KEY="your-openai-api-key"

# 编译
go build -o opencode_nano

# 交互式模式（推荐）
./opencode_nano

# 单次命令模式
./opencode_nano "帮我创建一个简单的 Hello World Go 程序"
```

## 交互式对话示例

### 启动交互式模式
```bash
$ ./opencode_nano
🤖 OpenCode Nano - Interactive AI Programming Assistant
Type 'exit' or 'quit' to exit, Ctrl+C to interrupt
==================================================

💬 You: 帮我创建一个简单的 Go HTTP 服务器

🤖 Assistant: 我来帮你创建一个简单的 Go HTTP 服务器...
[AI 响应和文件创建过程]

💬 You: 现在帮我添加一个 JSON API 端点

🤖 Assistant: 好的，我来为你的 HTTP 服务器添加一个 JSON API 端点...
[基于之前的对话上下文继续工作]

💬 You: clear
🧹 Conversation cleared!

💬 You: exit
👋 Goodbye!
```

## 示例对话

### 1. 创建文件
```bash
./opencode_nano "创建一个名为 hello.go 的文件，包含一个简单的 Hello World 程序"
```

**预期行为**:
1. AI 会解释要做什么
2. 使用 `write_file` 工具创建文件
3. 系统会请求权限确认
4. 创建成功后会显示结果

### 2. 读取和修改文件
```bash
./opencode_nano "读取 hello.go 文件，然后修改它以接受命令行参数"
```

**预期行为**:
1. 使用 `read_file` 工具读取现有文件
2. 分析代码内容
3. 使用 `write_file` 工具写入修改后的版本
4. 解释所做的更改

### 3. 执行命令
```bash
./opencode_nano "编译并运行我的 Go 程序"
```

**预期行为**:
1. 使用 `bash` 工具执行 `go build`
2. 请求权限确认
3. 执行编译后的程序
4. 显示输出结果

## 权限系统演示

当 AI 尝试执行危险操作时，系统会请求确认：

```
🔐 Permission required:
Action: write_file
Description: Write to file: hello.go
Allow? [y/N]: y
```

用户需要输入 `y` 或 `yes` 来确认操作。

## 安全特性

- 禁止执行危险命令（如 `rm -rf`, `sudo` 等）
- 所有文件写入操作都需要权限确认
- 所有 bash 命令都需要权限确认
- 简单但有效的命令过滤机制

## 扩展示例

### 创建完整项目
```bash
./opencode_nano "帮我创建一个简单的 HTTP 服务器项目，包括 main.go 和 README.md"
```

### 代码分析
```bash
./opencode_nano "分析当前目录下的 Go 代码，找出潜在的问题"
```

### 测试执行
```bash
./opencode_nano "为我的代码创建单元测试并运行"
```

这些示例展示了 OpenCode Nano 如何通过简单但强大的工具系统来协助编程任务。