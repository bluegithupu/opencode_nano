package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"opencode_nano/tools/core"
)

// EnvTool 环境变量管理工具
type EnvTool struct {
	*core.BaseTool
}

// NewEnvTool 创建环境变量工具
func NewEnvTool() *EnvTool {
	tool := &EnvTool{
		BaseTool: core.NewBaseTool("env", "system", "Manage environment variables"),
	}
	
	tool.SetTags("system", "environment", "config")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"action": {
				Type:        "string",
				Description: "Action to perform: get, set, list, delete",
				Enum:        []string{"get", "set", "list", "delete"},
			},
			"name": {
				Type:        "string",
				Description: "Environment variable name",
			},
			"value": {
				Type:        "string",
				Description: "Environment variable value (for set action)",
			},
			"pattern": {
				Type:        "string",
				Description: "Pattern to filter variables (for list action)",
				Default:     "*",
			},
		},
		Required: []string{"action"},
	})
	
	return tool
}

// Execute 执行环境变量操作
func (t *EnvTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取操作类型
	action, err := params.GetString("action")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid action parameter")
	}
	
	switch action {
	case "get":
		return t.getEnv(params)
	case "set":
		return t.setEnv(params)
	case "list":
		return t.listEnv(params)
	case "delete":
		return t.deleteEnv(params)
	default:
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("unknown action: %s", action))
	}
}

// getEnv 获取环境变量
func (t *EnvTool) getEnv(params core.Parameters) (core.Result, error) {
	name, err := params.GetString("name")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "name parameter required for get action")
	}
	
	value := os.Getenv(name)
	
	result := core.NewSimpleResult(value)
	result.WithMetadata("name", name)
	result.WithMetadata("exists", value != "")
	
	return result, nil
}

// setEnv 设置环境变量
func (t *EnvTool) setEnv(params core.Parameters) (core.Result, error) {
	name, err := params.GetString("name")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "name parameter required for set action")
	}
	
	value, err := params.GetString("value")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "value parameter required for set action")
	}
	
	// 设置环境变量
	if err := os.Setenv(name, value); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to set environment variable: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("Set %s=%s", name, value))
	result.WithMetadata("name", name)
	result.WithMetadata("value", value)
	
	return result, nil
}

// listEnv 列出环境变量
func (t *EnvTool) listEnv(params core.Parameters) (core.Result, error) {
	pattern := "*"
	if params.Has("pattern") {
		pattern, _ = params.GetString("pattern")
	}
	
	envVars := make(map[string]string)
	count := 0
	
	// 获取所有环境变量
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			name := parts[0]
			value := parts[1]
			
			// 检查是否匹配模式
			if pattern == "*" || strings.Contains(strings.ToLower(name), strings.ToLower(pattern)) {
				envVars[name] = value
				count++
			}
		}
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("Found %d environment variables", count))
	result.WithMetadata("variables", envVars)
	result.WithMetadata("count", count)
	result.WithMetadata("pattern", pattern)
	
	return result, nil
}

// deleteEnv 删除环境变量
func (t *EnvTool) deleteEnv(params core.Parameters) (core.Result, error) {
	name, err := params.GetString("name")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "name parameter required for delete action")
	}
	
	// 检查变量是否存在
	oldValue := os.Getenv(name)
	exists := oldValue != ""
	
	// 删除环境变量
	if err := os.Unsetenv(name); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to delete environment variable: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("Deleted environment variable: %s", name))
	result.WithMetadata("name", name)
	result.WithMetadata("existed", exists)
	if exists {
		result.WithMetadata("old_value", oldValue)
	}
	
	return result, nil
}

// ProcessTool 进程管理工具
type ProcessTool struct {
	*core.BaseTool
}

// NewProcessTool 创建进程工具
func NewProcessTool() *ProcessTool {
	tool := &ProcessTool{
		BaseTool: core.NewBaseTool("process", "system", "Manage system processes"),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("system", "process", "pid")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"action": {
				Type:        "string",
				Description: "Action to perform: list, info, kill",
				Enum:        []string{"list", "info", "kill"},
			},
			"pid": {
				Type:        "integer",
				Description: "Process ID (for info and kill actions)",
			},
			"signal": {
				Type:        "string",
				Description: "Signal to send (for kill action)",
				Default:     "TERM",
			},
			"pattern": {
				Type:        "string",
				Description: "Pattern to filter processes (for list action)",
				Default:     "",
			},
		},
		Required: []string{"action"},
	})
	
	return tool
}

// Execute 执行进程操作
func (t *ProcessTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取操作类型
	action, err := params.GetString("action")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid action parameter")
	}
	
	switch action {
	case "list":
		return t.listProcesses(params)
	case "info":
		return t.getProcessInfo(params)
	case "kill":
		return t.killProcess(params)
	default:
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("unknown action: %s", action))
	}
}

// listProcesses 列出进程（简化实现）
func (t *ProcessTool) listProcesses(params core.Parameters) (core.Result, error) {
	// 这是一个简化的实现
	// 在实际应用中，应该使用更复杂的进程列表获取方法
	
	result := core.NewSimpleResult("Process listing not fully implemented")
	result.WithMetadata("os", runtime.GOOS)
	result.WithMetadata("arch", runtime.GOARCH)
	result.WithMetadata("pid", os.Getpid())
	result.WithMetadata("ppid", os.Getppid())
	
	return result, nil
}

// getProcessInfo 获取进程信息
func (t *ProcessTool) getProcessInfo(params core.Parameters) (core.Result, error) {
	pid, err := params.GetInt("pid")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "pid parameter required for info action")
	}
	
	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("process not found: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("Found process with PID: %d", pid))
	result.WithMetadata("pid", pid)
	result.WithMetadata("process", process)
	
	return result, nil
}

// killProcess 终止进程
func (t *ProcessTool) killProcess(params core.Parameters) (core.Result, error) {
	pid, err := params.GetInt("pid")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "pid parameter required for kill action")
	}
	
	// 安全检查：不允许终止自己
	if pid == os.Getpid() {
		return nil, core.ErrExecutionFailed(t.Info().Name, "cannot kill self")
	}
	
	// 查找进程
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("process not found: %v", err))
	}
	
	// 发送信号（在 Windows 上总是发送 Kill 信号）
	var signal os.Signal
	if runtime.GOOS == "windows" {
		err = process.Kill()
	} else {
		// Unix 系统可以发送不同的信号
		signalName := "TERM"
		if params.Has("signal") {
			signalName, _ = params.GetString("signal")
		}
		
		// 这里简化处理，只支持 TERM 和 KILL
		switch signalName {
		case "KILL":
			err = process.Kill()
		default:
			err = process.Signal(os.Interrupt)
		}
		signal = os.Interrupt
	}
	
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to kill process: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("Successfully sent signal to process %d", pid))
	result.WithMetadata("pid", pid)
	if signal != nil {
		result.WithMetadata("signal", signal.String())
	}
	
	return result, nil
}