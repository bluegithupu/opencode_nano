package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"opencode_nano/session"
)

type TodoTool struct {
	manager *session.TodoManager
}

// NewTodoTool 创建新的 TodoTool
func NewTodoTool() (*TodoTool, error) {
	storage, err := session.NewDefaultFileStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %v", err)
	}

	manager := session.NewTodoManager(storage)
	if err := manager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load todos: %v", err)
	}

	return &TodoTool{
		manager: manager,
	}, nil
}

func (t *TodoTool) Name() string {
	return "todo"
}

func (t *TodoTool) Description() string {
	return "Manage session todo list. Support operations: list, add, update, delete, clear, count."
}

func (t *TodoTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action to perform: list, add, update, delete, clear, count",
				"enum":        []string{"list", "add", "update", "delete", "clear", "count"},
			},
			"id": map[string]any{
				"type":        "string",
				"description": "Todo item ID (required for update, delete, get)",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Todo item content (required for add, optional for update)",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "Todo item status: pending, in_progress, completed (optional for update)",
				"enum":        []string{"pending", "in_progress", "completed"},
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "Todo item priority: high, medium, low (optional for add and update, default: medium)",
				"enum":        []string{"high", "medium", "low"},
			},
			"filter_status": map[string]any{
				"type":        "string",
				"description": "Filter todos by status (optional for list)",
				"enum":        []string{"pending", "in_progress", "completed"},
			},
		},
		"required": []string{"action"},
	}
}

func (t *TodoTool) Execute(params map[string]any) (string, error) {
	action, ok := params["action"].(string)
	if !ok {
		return "", fmt.Errorf("action parameter is required and must be a string")
	}

	switch action {
	case "list":
		return t.listTodos(params)
	case "add":
		return t.addTodo(params)
	case "update":
		return t.updateTodo(params)
	case "delete":
		return t.deleteTodo(params)
	case "clear":
		return t.clearTodos(params)
	case "count":
		return t.countTodos(params)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *TodoTool) listTodos(params map[string]any) (string, error) {
	var items []*session.TodoItem

	if filterStatus, ok := params["filter_status"].(string); ok {
		status := session.TodoStatus(filterStatus)
		items = t.manager.ListByStatus(status)
	} else {
		items = t.manager.List()
	}

	if len(items) == 0 {
		return "No todos found.", nil
	}

	var result strings.Builder
	result.WriteString("📋 Todo List:\n")
	result.WriteString("================\n")

	for i, item := range items {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.String()))
	}

	// 添加统计信息
	counts := t.manager.Count()
	result.WriteString("\n📊 Summary:\n")
	result.WriteString(fmt.Sprintf("• Pending: %d\n", counts[session.StatusPending]))
	result.WriteString(fmt.Sprintf("• In Progress: %d\n", counts[session.StatusInProgress]))
	result.WriteString(fmt.Sprintf("• Completed: %d\n", counts[session.StatusCompleted]))
	result.WriteString(fmt.Sprintf("• Total: %d\n", counts[session.StatusPending]+counts[session.StatusInProgress]+counts[session.StatusCompleted]))

	return result.String(), nil
}

func (t *TodoTool) addTodo(params map[string]any) (string, error) {
	content, ok := params["content"].(string)
	if !ok || strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("content parameter is required and must be a non-empty string")
	}

	priority := session.PriorityMedium
	if p, ok := params["priority"].(string); ok {
		priority = session.TodoPriority(p)
	}

	item, err := t.manager.Add(content, priority)
	if err != nil {
		return "", fmt.Errorf("failed to add todo: %v", err)
	}

	if err := t.manager.Save(); err != nil {
		return "", fmt.Errorf("failed to save todos: %v", err)
	}

	return fmt.Sprintf("✅ Todo added successfully:\n%s", item.String()), nil
}

func (t *TodoTool) updateTodo(params map[string]any) (string, error) {
	id, ok := params["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id parameter is required and must be a non-empty string")
	}

	status := session.TodoStatus("")
	if s, ok := params["status"].(string); ok {
		status = session.TodoStatus(s)
	}

	content := ""
	if c, ok := params["content"].(string); ok {
		content = c
	}

	priority := session.TodoPriority("")
	if p, ok := params["priority"].(string); ok {
		priority = session.TodoPriority(p)
	}

	item, err := t.manager.Update(id, status, content, priority)
	if err != nil {
		return "", fmt.Errorf("failed to update todo: %v", err)
	}

	if err := t.manager.Save(); err != nil {
		return "", fmt.Errorf("failed to save todos: %v", err)
	}

	return fmt.Sprintf("✅ Todo updated successfully:\n%s", item.String()), nil
}

func (t *TodoTool) deleteTodo(params map[string]any) (string, error) {
	id, ok := params["id"].(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("id parameter is required and must be a non-empty string")
	}

	// 获取要删除的项目信息
	item, err := t.manager.Get(id)
	if err != nil {
		return "", fmt.Errorf("failed to get todo: %v", err)
	}

	if err := t.manager.Delete(id); err != nil {
		return "", fmt.Errorf("failed to delete todo: %v", err)
	}

	if err := t.manager.Save(); err != nil {
		return "", fmt.Errorf("failed to save todos: %v", err)
	}

	return fmt.Sprintf("✅ Todo deleted successfully:\n%s", item.String()), nil
}

func (t *TodoTool) clearTodos(_ map[string]any) (string, error) {
	counts := t.manager.Count()
	total := counts[session.StatusPending] + counts[session.StatusInProgress] + counts[session.StatusCompleted]

	if total == 0 {
		return "No todos to clear.", nil
	}

	t.manager.Clear()
	if err := t.manager.Save(); err != nil {
		return "", fmt.Errorf("failed to save todos: %v", err)
	}

	return fmt.Sprintf("✅ Cleared %d todos successfully.", total), nil
}

func (t *TodoTool) countTodos(_ map[string]any) (string, error) {
	counts := t.manager.Count()
	total := counts[session.StatusPending] + counts[session.StatusInProgress] + counts[session.StatusCompleted]

	result := fmt.Sprintf("📊 Todo Statistics:\n" +
		"• Pending: %d\n" +
		"• In Progress: %d\n" +
		"• Completed: %d\n" +
		"• Total: %d",
		counts[session.StatusPending],
		counts[session.StatusInProgress],
		counts[session.StatusCompleted],
		total)

	return result, nil
}

// ToJSONString 将参数转换为 JSON 字符串（用于调试）
func (t *TodoTool) ToJSONString(params map[string]any) string {
	data, _ := json.MarshalIndent(params, "", "  ")
	return string(data)
}