package tools

import (
	"opencode_nano/tools/core"
	"opencode_nano/tools/file"
	"opencode_nano/tools/system"
	"opencode_nano/tools/task"
)

// DefaultRegistry 默认工具注册表
var DefaultRegistry *core.ToolRegistry

// InitializeRegistry 初始化工具注册表
func InitializeRegistry() (*core.ToolRegistry, error) {
	registry := core.NewRegistry()
	
	// 注册文件操作工具
	if err := registerFileTools(registry); err != nil {
		return nil, err
	}
	
	// 注册系统工具
	if err := registerSystemTools(registry); err != nil {
		return nil, err
	}
	
	// 注册任务工具
	if err := registerTaskTools(registry); err != nil {
		return nil, err
	}
	
	DefaultRegistry = registry
	return registry, nil
}

// registerFileTools 注册文件操作工具
func registerFileTools(registry *core.ToolRegistry) error {
	// 读取工具
	if err := registry.Register(file.NewReadTool(), "r", "cat"); err != nil {
		return err
	}
	
	// 写入工具
	if err := registry.Register(file.NewWriteTool(), "w", "write"); err != nil {
		return err
	}
	
	// 编辑工具
	if err := registry.Register(file.NewEditTool(), "e", "ed"); err != nil {
		return err
	}
	
	// 多文件编辑工具
	if err := registry.Register(file.NewMultiEditTool()); err != nil {
		return err
	}
	
	// 补丁工具
	if err := registry.Register(file.NewPatchTool()); err != nil {
		return err
	}
	
	// 搜索工具
	if err := registry.Register(file.NewSearchTool(), "s", "grep", "find"); err != nil {
		return err
	}
	
	// 通配符工具
	if err := registry.Register(file.NewGlobTool(), "g", "glob"); err != nil {
		return err
	}
	
	// 列表工具
	if err := registry.Register(file.NewListTool(), "ls", "dir"); err != nil {
		return err
	}
	
	// 二进制读取工具
	if err := registry.Register(file.NewReadBinaryTool()); err != nil {
		return err
	}
	
	return nil
}

// registerSystemTools 注册系统工具
func registerSystemTools(registry *core.ToolRegistry) error {
	// Bash 工具
	if err := registry.Register(system.NewBashTool(), "sh", "shell", "cmd"); err != nil {
		return err
	}
	
	// 管道工具
	if err := registry.Register(system.NewPipelineTool(), "pipe"); err != nil {
		return err
	}
	
	// 环境变量工具
	if err := registry.Register(system.NewEnvTool(), "env"); err != nil {
		return err
	}
	
	// 进程工具
	if err := registry.Register(system.NewProcessTool(), "ps", "proc"); err != nil {
		return err
	}
	
	return nil
}

// registerTaskTools 注册任务工具
func registerTaskTools(registry *core.ToolRegistry) error {
	// 任务工具
	taskTool, err := task.NewTaskTool()
	if err != nil {
		return err
	}
	
	// 注册时使用 "todo" 作为主名称，保持向后兼容
	if err := registry.Register(taskTool, "todo", "todos", "task", "t"); err != nil {
		return err
	}
	
	return nil
}

// GetTool 获取工具
func GetTool(name string) (core.Tool, error) {
	if DefaultRegistry == nil {
		if _, err := InitializeRegistry(); err != nil {
			return nil, err
		}
	}
	
	return DefaultRegistry.Get(name)
}

// ListTools 列出所有工具
func ListTools() []core.Tool {
	if DefaultRegistry == nil {
		if _, err := InitializeRegistry(); err != nil {
			return []core.Tool{}
		}
	}
	
	return DefaultRegistry.All()
}

// SearchTools 搜索工具
func SearchTools(query string) []core.Tool {
	if DefaultRegistry == nil {
		if _, err := InitializeRegistry(); err != nil {
			return []core.Tool{}
		}
	}
	
	return DefaultRegistry.Find(query)
}