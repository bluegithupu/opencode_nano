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

### Key Differences from Full OpenCode
- No TUI interface (command-line only)
- No persistent storage or session management
- Single AI provider (OpenAI only)
- Minimal tool set (3 core tools vs 10+ in full version)
- Environment variable configuration only
- No LSP or MCP integration

## Error Handling

- Configuration errors exit with descriptive messages
- Tool execution errors are returned to the AI for handling
- Permission denials are treated as tool execution failures
- Streaming response errors are handled gracefully

This is a learning-focused implementation that demonstrates AI agent fundamentals without the complexity of the full OpenCode system.