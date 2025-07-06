# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `go build -o opencode_nano` - Build the main binary
- `./opencode_nano` - Run in interactive mode
- `./opencode_nano "your prompt here"` - Run with a single command
- `./opencode_nano --auto "prompt"` or `./opencode_nano -a "prompt"` - Run in auto mode (auto-approves all operations)
- `go run main.go` - Run without building binary

### Testing
- `go test ./...` - Run all tests
- `go test -v ./...` - Run tests with verbose output
- `go test ./tools/...` - Run tests for specific package
- `make test` - Run tests using Makefile
- `make test-coverage` - Generate coverage report
- `make test-coverage-html` - Generate HTML coverage report
- `make test-race` - Run tests with race detection
- `make bench` - Run benchmark tests

### Development Utilities
- `make fmt` - Format code
- `make lint` - Run linter (requires golangci-lint)
- `make check` - Run fmt, test, and lint
- `make clean` - Clean build artifacts and test cache

### Requirements
- Go 1.21 or higher
- `OPENAI_API_KEY` environment variable must be set
- `OPENAI_BASE_URL` environment variable (optional, defaults to https://api.openai.com/v1)

## Architecture Overview

OpenCode Nano is a simplified AI programming assistant that demonstrates core AI agent concepts. The codebase is undergoing a major refactoring to introduce a modular tool system.

### Core Components

**Entry Points:**
- **main.go**: Entry point supporting interactive mode, single command mode, and auto mode
- **config/**: Configuration management (loads OpenAI API key from environment)
- **agent/**: Core AI agent logic with OpenAI integration and streaming responses
  - `RunOnce()`: Execute single task with multi-round conversation support
  - `RunInteractive()`: Continuous conversation mode
  - `StreamResponseWithTools()`: Multi-round tool execution

**Tool System (Dual Architecture):**
- **tools/** (Legacy): Original simple tool interface
- **tools/core/** (New): Modular, interface-based tool system
  - Type-safe parameters and results
  - Tool registry with categories and tags
  - Pipeline support for tool composition
  - Comprehensive error handling

**Permission System:**
- **permission/**: Interactive permission system for dangerous operations
- Supports interactive mode and auto-approval mode
- Integrated with both old and new tool systems

**Session Management:**
- **session/**: Todo list management with persistent storage
- Stores todos in `~/.opencode_nano/session_todos.json`

### New Tool System Architecture

The new modular tool system provides significant improvements:

**Core Interfaces** (`tools/core/interfaces.go`):
```go
type Tool interface {
    Info() ToolInfo
    Execute(ctx context.Context, params Parameters) (Result, error)
    Schema() ParameterSchema
}
```

**Tool Registry** (`tools/core/registry.go`):
- Register tools with aliases
- Search by name, category, or tags
- Thread-safe operations
- Tool discovery and categorization

**Tool Categories:**
- **file/**: Read, Write, Edit, Search, Glob, List operations
- **system/**: Bash (enhanced), Pipeline, Env, Process management
- **task/**: Todo/task management with import/export

**Migration Layer** (`tools/migration.go`):
- `CreateLegacyToolSet()`: Creates backward-compatible tool set
- `PermissionWrappedTool`: Integrates permission checks
- Ensures smooth transition from old to new system

### Key Design Patterns

**Multi-Round Conversation:**
- Agent can complete multi-step tasks autonomously
- Configurable maximum rounds (10 for RunOnce, 5 for RunInteractive)
- Continues execution as long as tools are called

**Tool Execution Flow:**
1. Agent analyzes user request
2. Selects appropriate tool(s)
3. Checks permissions if required
4. Executes tool with parameters
5. Processes results and continues if needed

**Error Handling:**
- Typed errors with codes and context
- Retryable error support
- Rich error information for debugging

### Available Tools

**File Operations:**
- **read**: Read file contents with line ranges
- **write**: Write files with atomic operations
- **edit**: Find/replace with regex support
- **search**: Content search with regex
- **glob**: File pattern matching
- **list**: Directory listing

**System Operations:**
- **bash**: Execute commands with safety checks
- **pipeline**: Sequential/parallel command execution
- **env**: Environment variable management
- **process**: Process management

**Development Tools:**
- **todo**: Todo/task management with priorities and statuses (formerly task tool)

### Security Features

**Permission System:**
- Interactive approval for dangerous operations
- Auto mode for CI/CD scenarios
- Tool-specific permission requirements

**Command Safety:**
- Dangerous command filtering
- Timeout support
- Environment isolation
- Working directory control

### Configuration

Environment variables only:
- `OPENAI_API_KEY`: Required for OpenAI API access
- `OPENAI_BASE_URL`: Optional custom API endpoint

No configuration files - designed for simplicity.

### Development Notes

**Adding New Tools:**
1. Implement `core.Tool` interface in appropriate category
2. Register in `tools/registry.go`
3. Add permission wrapper if needed
4. Tool automatically available to agent

**Testing New Tools:**
```bash
# Run specific tool tests
go test -v ./tools/file/...

# Test with coverage
go test -coverprofile=coverage.out ./tools/...
```

**Tool Development Best Practices:**
- Use `core.BaseTool` for common functionality
- Implement proper parameter validation
- Return structured `Result` with metadata
- Handle context cancellation
- Add comprehensive error messages

This is a learning-focused implementation that demonstrates AI agent fundamentals while introducing professional software engineering patterns.