# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `go build -o opencode_nano` - Build the main binary
- `./opencode_nano "your prompt here"` - Run the application with a prompt
- `go run main.go "your prompt here"` - Run without building binary

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output

### Requirements
- Go 1.21 or higher
- `OPENAI_API_KEY` environment variable must be set
- `OPENAI_BASE_URL` environment variable (optional, defaults to https://api.openai.com/v1)

## Architecture Overview

OpenCode Nano is a simplified AI programming assistant that demonstrates core AI agent concepts. It focuses on understanding the fundamental working principles of AI agent systems.

### Core Components
- **main.go**: Entry point that coordinates config, permissions, tools, and agent
- **config/**: Configuration management (loads OpenAI API key from environment)
- **agent/**: Core AI agent logic with OpenAI integration and streaming responses
- **tools/**: Tool system with unified interface for AI capabilities
- **permission/**: Interactive permission system for dangerous operations

### Key Design Patterns
- **Tool Interface**: Unified interface for all AI capabilities with `Name()`, `Description()`, `Parameters()`, and `Execute()` methods
- **Permission System**: Interactive approval for dangerous operations (file writes, bash commands)
- **Streaming Response**: Real-time display of AI responses using OpenAI streaming API
- **Single Conversation**: Stateless operation with no persistent session storage

### Available Tools
- **ReadTool**: Read file contents (no permission required)
- **WriteTool**: Write file contents (requires user permission)
- **BashTool**: Execute bash commands with safety checks (requires user permission)
- **TodoTool**: Manage session todo lists for task planning and tracking (no permission required)

## Security Features

### Permission System
- All file write operations require explicit user approval
- All bash command execution requires explicit user approval
- Interactive permission prompts with action description
- User must type 'y' or 'yes' to approve dangerous operations

### Command Safety
- Built-in dangerous command filtering (blocks `rm -rf`, `sudo`, `curl`, `wget`, etc.)
- Commands are executed through controlled `bash -c` wrapper
- Error handling with combined output capture

## Configuration

The application uses environment variables for configuration:
- `OPENAI_API_KEY`: Required OpenAI API key for GPT model access
- `OPENAI_BASE_URL`: Optional custom base URL for OpenAI API (defaults to https://api.openai.com/v1)
- No config files or persistent storage

### Example Configuration
```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_BASE_URL="https://api.rcouyi.com/v1"  # Optional custom endpoint
```

## Development Notes

### Adding New Tools
1. Implement the `Tool` interface in `tools/` package
2. Add tool to toolSet in `main.go`
3. Consider if the tool requires permission checks
4. Use `ToOpenAIFunction()` to convert to OpenAI function definition

### Session Management
- **Todo Lists**: Use the `todo` tool to manage task lists for complex multi-step operations
- **Persistent Storage**: Todo lists are persisted across sessions in `~/.opencode_nano/session_todos.json`
- **Operations**: Support add, update, delete, list, clear, and count operations
- **Priority Levels**: high, medium, low priority support
- **Status Tracking**: pending, in_progress, completed status management

### Key Differences from Full OpenCode
- No TUI interface (command-line only)
- Basic session management (todo lists only)
- Single AI provider (OpenAI only)
- Minimal tool set (4 core tools vs 10+ in full version)
- Environment variable configuration only
- No LSP or MCP integration

## Error Handling

- Configuration errors exit with descriptive messages
- Tool execution errors are returned to the AI for handling
- Permission denials are treated as tool execution failures
- Streaming response errors are handled gracefully

This is a learning-focused implementation that demonstrates AI agent fundamentals without the complexity of the full OpenCode system.

## Todo Tool Usage Examples

The todo tool helps manage complex multi-step tasks:

```bash
# Add a new todo
./opencode_nano "添加一个高优先级的 todo：实现用户认证功能"

# View todo list
./opencode_nano "显示我的 todo 列表"

# Update todo status
./opencode_nano "将 ID 为 123 的 todo 标记为进行中"

# Delete a todo
./opencode_nano "删除 ID 为 123 的 todo"

# Clear all todos
./opencode_nano "清空所有 todo"

# Get statistics
./opencode_nano "显示 todo 统计信息"
```

The AI will automatically use the todo tool when handling complex tasks that require multiple steps.