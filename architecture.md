# OpenCode Nano 架构设计

## 项目结构

```
opencode_nano/
├── main.go                 # 程序入口点
├── go.mod                  # Go 模块定义
├── README.md               # 项目说明
├── examples.md             # 使用示例
├── architecture.md         # 架构文档
├── CLAUDE.md              # Claude Code 项目说明
├── .gitignore             # Git 忽略文件
├── config/
│   └── config.go          # 配置管理
├── permission/
│   └── permission.go      # 权限控制系统
├── session/               # 会话管理
│   ├── todo.go           # Todo 数据结构和管理
│   └── storage.go        # 持久化存储
├── tools/                 # 工具系统
│   ├── tool.go           # 旧版工具接口（兼容）
│   ├── converter.go      # 新旧接口转换器
│   ├── registry.go       # 工具注册表
│   ├── core/             # 核心接口和基础设施
│   │   ├── interfaces.go # 核心接口定义
│   │   ├── base.go       # 基础实现
│   │   ├── errors.go     # 错误处理
│   │   ├── registry.go   # 注册表实现
│   │   └── pipeline.go   # 管道功能
│   ├── file/             # 文件操作工具
│   │   ├── read.go       # 读取工具
│   │   ├── write.go      # 写入工具
│   │   ├── edit.go       # 编辑工具
│   │   ├── search.go     # 搜索工具
│   │   └── list.go       # 列表工具
│   ├── system/           # 系统工具
│   │   ├── bash.go       # 增强的命令执行
│   │   └── env.go        # 环境变量工具
│   └── task/             # 任务管理
│       └── task.go       # 简化的任务工具
└── agent/                 # AI 代理系统
    ├── agent.go          # 主代理逻辑
    └── provider.go       # OpenAI 提供商
```

## 核心组件

### 1. 配置系统 (`config/`)
**职责**: 管理应用配置，主要是 OpenAI API Key
**核心功能**:
- 从环境变量加载 API Key
- 配置验证

```go
type Config struct {
    OpenAIAPIKey string
}
```

### 2. 权限系统 (`permission/`)
**职责**: 控制危险操作的执行权限
**核心功能**:
- 交互式权限请求
- 用户确认机制

```go
type Manager struct{}

func (m *Manager) Request(action, description string) bool
```

### 3. 工具系统 (`tools/`)
**职责**: 提供 AI 可调用的功能工具
**核心功能**:
- 统一的工具接口
- 工具注册和发现
- 类型安全的参数处理
- 工具组合和管道

**新架构特点**:
- **分层设计**: core → 具体实现 → 转换器
- **模块化**: 按功能领域组织（file、system、task）
- **可扩展**: 易于添加新工具和功能
- **简洁转换**: 通过 converter.go 适配旧接口

**核心接口** (`tools/core/interfaces.go`):
```go
type Tool interface {
    Info() ToolInfo
    Execute(ctx context.Context, params Parameters) (Result, error)
    Schema() ParameterSchema
}

type Parameters interface {
    Get(key string) (any, error)
    GetString(key string) (string, error)
    Validate(schema ParameterSchema) error
}

type Result interface {
    String() string
    Data() any
    Success() bool
}
```

**工具类别**:
1. **文件工具** (`tools/file/`)
   - `ReadTool`: 读取文件内容
   - `WriteTool`: 写入文件内容（需要权限）
   - `EditTool`: 编辑文件内容
   - `SearchTool`: 搜索文件内容
   - `ListTool`: 列出目录内容

2. **系统工具** (`tools/system/`)
   - `BashTool`: 增强的命令执行（需要权限）
   - `EnvTool`: 环境变量管理

3. **任务工具** (`tools/task/`)
   - `TaskTool`: 简化的任务管理（list、add、update）

### 4. 代理系统 (`agent/`)
**职责**: 核心 AI 交互逻辑
**核心功能**:
- 与 OpenAI API 交互
- 流式响应处理
- 工具调用管理
- 多轮对话支持

**新增功能**:
- `StreamResponseWithTools`: 支持工具调用的流式响应
- `ExecuteToolCall`: 公开的工具执行方法
- 自动完成多步骤任务

### 5. 会话管理 (`session/`)
**职责**: 管理任务和会话状态
**核心功能**:
- Todo 项目的 CRUD 操作
- 持久化存储（JSON 文件）
- 状态和优先级管理

```go
type TodoItem struct {
    ID       string
    Content  string
    Status   TodoStatus    // pending, in_progress, completed
    Priority TodoPriority  // high, medium, low
}
```

## 核心流程

### 1. 启动流程
```
main() → config.Load() → permission.New() → tools.InitializeRegistry() → agent.New() → agent.Run()
```

### 2. 对话流程
```
用户输入 → 构建消息 → 发送到 OpenAI → 流式响应 → 工具调用 → 权限检查 → 执行工具 → 继续对话
```

### 3. 工具调用流程
```
LLM 决定调用工具 → 解析工具参数 → 查找对应工具 → 参数验证 → 权限检查 → 执行工具 → 返回结果
```

### 4. 多步骤任务流程
```
接收任务 → 创建 Todo → 执行第一步 → 更新状态 → 执行下一步 → ... → 完成任务
```

## 核心设计原则

### 1. 简单性
- 最小化的依赖（只有 go-openai）
- 清晰的模块分离
- 直观的接口设计

### 2. 安全性
- 所有危险操作都需要权限确认
- 命令过滤机制
- 明确的权限提示

### 3. 可扩展性
- 统一的工具接口
- 模块化的架构
- 易于添加新工具

### 4. 可学习性
- 清晰的代码结构
- 完整的注释
- 详细的示例

## 新增功能

### 1. 持续对话和多轮交互
- 支持持续的多轮对话
- 保持对话历史和上下文
- 内置命令：`clear`、`help`、`exit`
- 优雅的信号处理（Ctrl+C）

### 2. 多步骤任务自动执行
- AI 能够独立完成多步骤任务
- 不需要每轮用户介入
- 自动跟踪任务进度
- 最大执行轮数限制（防止无限循环）

### 3. 任务管理系统
- 持久化的任务列表
- 任务状态跟踪（pending → in_progress → completed）
- 优先级管理（high、medium、low）
- 简化的操作接口（list、add、update）

### 4. 新工具系统架构
- 类型安全的参数和结果
- 工具注册和发现机制
- 支持工具组合（Pipeline）
- 更好的错误处理

## 与原版 OpenCode 的对比

| 特性 | OpenCode | OpenCode Nano |
|------|----------|---------------|
| 界面 | 复杂 TUI | 交互式命令行 |
| 数据库 | SQLite | 内存存储 + JSON 文件 |
| 提供商 | 多个 | 仅 OpenAI |
| 工具数量 | 10+ | 8 个核心工具 |
| 配置 | 复杂配置文件 | 环境变量 |
| 会话管理 | 持久化会话 | 内存对话 + 任务持久化 |
| 多轮对话 | ✅ | ✅ |
| 多步骤任务 | ✅ | ✅ |
| 任务管理 | 复杂 | 简化（3个操作） |
| 工具架构 | 简单接口 | 分层架构 |
| LSP 集成 | ✅ | ❌ |
| MCP 支持 | ✅ | ❌ |
| 代码行数 | ~10000+ | ~2000 |

## 学习价值

通过 OpenCode Nano，你可以学习到：

1. **AI Agent 基本架构**: 理解代理、工具、权限的关系
2. **LLM 工具调用**: 如何让 AI 执行具体操作
3. **流式响应处理**: 实时显示 AI 的思考过程
4. **权限控制设计**: 如何安全地给 AI 执行权限
5. **模块化设计**: 如何构建可扩展的系统架构
6. **多步骤任务处理**: AI 如何自主完成复杂任务
7. **工具系统设计**: 类型安全、可组合的工具架构
8. **持久化策略**: 简单有效的数据存储方案

## 技术亮点

1. **新工具系统**
   - 接口驱动的设计
   - 类型安全的参数处理
   - 统一的错误处理
   - 支持工具组合

2. **任务自动化**
   - AI 自主完成多步骤任务
   - 智能的任务状态管理
   - 防止无限循环的保护机制

3. **简化但不简单**
   - 保留核心功能
   - 优化的用户体验
   - 清晰的代码结构
   - 易于理解和扩展

这是理解现代 AI 编程助手工作原理的绝佳入门项目！