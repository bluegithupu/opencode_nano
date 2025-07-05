# OpenCode Nano 架构设计

## 项目结构

```
opencode_nano/
├── main.go                 # 程序入口点
├── go.mod                  # Go 模块定义
├── README.md               # 项目说明
├── examples.md             # 使用示例
├── architecture.md         # 架构文档
├── .gitignore             # Git 忽略文件
├── config/
│   └── config.go          # 配置管理
├── permission/
│   └── permission.go      # 权限控制系统
├── tools/                 # 工具系统
│   ├── tool.go           # 工具接口定义
│   ├── read.go           # 文件读取工具
│   ├── write.go          # 文件写入工具
│   └── bash.go           # 命令执行工具
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
- 文件读写操作
- 安全的命令执行

**工具接口**:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(params map[string]interface{}) (string, error)
}
```

**内置工具**:
- `ReadTool`: 读取文件内容
- `WriteTool`: 写入文件内容（需要权限）
- `BashTool`: 执行 bash 命令（需要权限）

### 4. 代理系统 (`agent/`)
**职责**: 核心 AI 交互逻辑
**核心功能**:
- 与 OpenAI API 交互
- 流式响应处理
- 工具调用管理

## 核心流程

### 1. 启动流程
```
main() → config.Load() → permission.New() → tools.New*() → agent.New() → agent.Run()
```

### 2. 对话流程
```
用户输入 → 构建消息 → 发送到 OpenAI → 流式响应 → 工具调用 → 权限检查 → 执行工具 → 返回结果
```

### 3. 工具调用流程
```
LLM 决定调用工具 → 解析工具参数 → 查找对应工具 → 权限检查 → 执行工具 → 返回结果
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

## 新增功能：持续对话

### 1. 交互式模式
- 支持持续的多轮对话
- 保持对话历史和上下文
- 内置命令：`clear`、`help`、`exit`
- 优雅的信号处理（Ctrl+C）

### 2. 对话管理
- 内存中的对话历史保存
- 系统提示词持久化
- 工具执行结果集成到对话
- `clear` 命令重置对话状态

## 与原版 OpenCode 的对比

| 特性 | OpenCode | OpenCode Nano |
|------|----------|---------------|
| 界面 | 复杂 TUI | 交互式命令行 |
| 数据库 | SQLite | 内存存储 |
| 提供商 | 多个 | 仅 OpenAI |
| 工具数量 | 10+ | 3 个核心工具 |
| 配置 | 复杂配置文件 | 环境变量 |
| 会话管理 | 持久化会话 | 内存中对话历史 |
| 多轮对话 | ✅ | ✅ |
| LSP 集成 | ✅ | ❌ |
| MCP 支持 | ✅ | ❌ |
| 代码行数 | ~10000+ | ~600 |

## 学习价值

通过 OpenCode Nano，你可以学习到：

1. **AI Agent 基本架构**: 理解代理、工具、权限的关系
2. **LLM 工具调用**: 如何让 AI 执行具体操作
3. **流式响应处理**: 实时显示 AI 的思考过程
4. **权限控制设计**: 如何安全地给 AI 执行权限
5. **模块化设计**: 如何构建可扩展的系统架构

这是理解现代 AI 编程助手工作原理的绝佳入门项目！