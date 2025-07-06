package task

import (
	"context"
	"fmt"
	"strings"

	"opencode_nano/session"
	"opencode_nano/tools/core"
)

// TaskTool 通用任务管理工具
type TaskTool struct {
	*core.BaseTool
	manager *session.TodoManager
}

// NewTaskTool 创建任务工具
func NewTaskTool() (*TaskTool, error) {
	// 创建默认存储
	storage, err := session.NewDefaultFileStorage()
	if err != nil {
		return nil, err
	}
	
	// 创建管理器
	manager := session.NewTodoManager(storage)
	
	tool := &TaskTool{
		BaseTool: core.NewBaseTool("todo", "development", "Manage session todo list. Support operations: list, add, update."),
		manager:  manager,
	}
	
	tool.SetTags("task", "todo", "project", "planning")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"action": {
				Type:        "string",
				Description: "Action to perform",
				Enum:        []string{"list", "add", "update"},
			},
			"id": {
				Type:        "string",
				Description: "Task ID (required for update)",
			},
			"content": {
				Type:        "string",
				Description: "Task content/description",
			},
			"status": {
				Type:        "string",
				Description: "Task status",
				Enum:        []string{"pending", "in_progress", "completed"},
				Default:     "pending",
			},
			"priority": {
				Type:        "string",
				Description: "Task priority",
				Enum:        []string{"low", "medium", "high"},
				Default:     "medium",
			},
		},
		Required: []string{"action"},
	})
	
	return tool, nil
}

// Execute 执行任务操作
func (t *TaskTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
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
		return t.listTasks(params)
	case "add":
		return t.addTask(params)
	case "update":
		return t.updateTask(params)
	default:
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("unknown action: %s", action))
	}
}

// listTasks 列出任务
func (t *TaskTool) listTasks(params core.Parameters) (core.Result, error) {
	todos := t.manager.List()
	
	if len(todos) == 0 {
		return core.NewSimpleResult("No todos found."), nil
	}
	
	// 构建输出
	var output strings.Builder
	output.WriteString("📋 Todo List:\n")
	output.WriteString("================\n")
	
	for i, todo := range todos {
		statusSymbol := map[session.TodoStatus]string{
			session.StatusPending:    "⏳",
			session.StatusInProgress: "🔄",
			session.StatusCompleted:  "✅",
		}[todo.Status]
		
		prioritySymbol := map[session.TodoPriority]string{
			session.PriorityHigh:   "🔴",
			session.PriorityMedium: "🟡",
			session.PriorityLow:    "🟢",
		}[todo.Priority]
		
		output.WriteString(fmt.Sprintf("%d. %s %s [%s] %s\n", 
			i+1, statusSymbol, prioritySymbol, todo.ID, todo.Content))
	}
	
	// 统计信息
	counts := t.manager.Count()
	output.WriteString("\n📊 Summary:\n")
	output.WriteString(fmt.Sprintf("• Pending: %d\n", counts[session.StatusPending]))
	output.WriteString(fmt.Sprintf("• In Progress: %d\n", counts[session.StatusInProgress]))
	output.WriteString(fmt.Sprintf("• Completed: %d\n", counts[session.StatusCompleted]))
	
	return core.NewSimpleResult(output.String()), nil
}

// addTask 添加任务
func (t *TaskTool) addTask(params core.Parameters) (core.Result, error) {
	content, err := params.GetString("content")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "content parameter required")
	}
	
	// 获取优先级（可选）
	priority := session.TodoPriority("medium")
	if params.Has("priority") {
		if p, _ := params.GetString("priority"); p != "" {
			priority = session.TodoPriority(p)
		}
	}
	
	// 创建任务
	todo, err := t.manager.Add(content, priority)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to add task: %v", err))
	}
	
	// 保存
	if err := t.manager.Save(); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to save: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("✅ Todo added successfully:\n%s", todo.String()))
	result.WithMetadata("id", todo.ID)
	
	return result, nil
}

// updateTask 更新任务
func (t *TaskTool) updateTask(params core.Parameters) (core.Result, error) {
	id, err := params.GetString("id")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "id parameter required")
	}
	
	// 验证任务存在
	_, err = t.manager.Get(id)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("task not found: %s", id))
	}
	
	// 准备更新参数
	var status session.TodoStatus = ""
	var priority session.TodoPriority = ""
	var content string = ""
	
	// 获取要更新的字段
	if params.Has("status") {
		if s, _ := params.GetString("status"); s != "" {
			status = session.TodoStatus(s)
		}
	}
	
	if params.Has("priority") {
		if p, _ := params.GetString("priority"); p != "" {
			priority = session.TodoPriority(p)
		}
	}
	
	if params.Has("content") {
		content, _ = params.GetString("content")
	}
	
	// 执行更新
	updatedTodo, err := t.manager.Update(id, status, content, priority)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to update task: %v", err))
	}
	
	// 保存
	if err := t.manager.Save(); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to save: %v", err))
	}
	
	result := core.NewSimpleResult(fmt.Sprintf("✅ Todo updated successfully:\n%s", updatedTodo.String()))
	result.WithMetadata("id", id)
	
	return result, nil
}