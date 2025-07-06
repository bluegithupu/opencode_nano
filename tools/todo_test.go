package tools

import (
	"strings"
	"testing"

	"opencode_nano/session"
)

func TestTodoTool_Name(t *testing.T) {
	tool := createTestTodoTool(t)
	if tool.Name() != "todo" {
		t.Errorf("Name() = %v, want %v", tool.Name(), "todo")
	}
}

func TestTodoTool_Description(t *testing.T) {
	tool := createTestTodoTool(t)
	desc := tool.Description()
	if desc == "" {
		t.Error("Description() should not return empty string")
	}
	if !strings.Contains(desc, "todo") {
		t.Error("Description() should contain 'todo'")
	}
}

func TestTodoTool_Parameters(t *testing.T) {
	tool := createTestTodoTool(t)
	params := tool.Parameters()
	
	if params == nil {
		t.Fatal("Parameters() should not return nil")
	}
	
	// 检查必需的字段
	if params["type"] != "object" {
		t.Error("Parameters() should have type 'object'")
	}
	
	properties, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters() should have properties field")
	}
	
	// 检查 action 参数
	action, ok := properties["action"].(map[string]any)
	if !ok {
		t.Fatal("Parameters() should have action property")
	}
	
	if action["type"] != "string" {
		t.Error("Action parameter should be string type")
	}
	
	// 检查必需字段
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Parameters() should have required field")
	}
	
	if len(required) == 0 || required[0] != "action" {
		t.Error("Parameters() should require action parameter")
	}
}

func TestTodoTool_Execute_List(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 测试列出空的 todo 列表
	result, err := tool.Execute(map[string]any{"action": "list"})
	if err != nil {
		t.Fatalf("Execute(list) failed: %v", err)
	}
	if !strings.Contains(result, "No todos found") {
		t.Error("Execute(list) should return 'No todos found' for empty list")
	}
	
	// 添加一些 todo 后再测试
	_, err = tool.Execute(map[string]any{
		"action":   "add",
		"content":  "Test todo",
		"priority": "high",
	})
	if err != nil {
		t.Fatalf("Execute(add) failed: %v", err)
	}
	
	result, err = tool.Execute(map[string]any{"action": "list"})
	if err != nil {
		t.Fatalf("Execute(list) failed: %v", err)
	}
	if !strings.Contains(result, "Test todo") {
		t.Error("Execute(list) should contain added todo")
	}
	if !strings.Contains(result, "📋 Todo List") {
		t.Error("Execute(list) should contain list header")
	}
}

func TestTodoTool_Execute_Add(t *testing.T) {
	tool := createTestTodoTool(t)
	
	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{
			name: "成功添加 todo",
			params: map[string]any{
				"action":   "add",
				"content":  "Test todo",
				"priority": "medium",
			},
			wantErr: false,
		},
		{
			name: "添加高优先级 todo",
			params: map[string]any{
				"action":   "add",
				"content":  "High priority todo",
				"priority": "high",
			},
			wantErr: false,
		},
		{
			name: "添加默认优先级 todo",
			params: map[string]any{
				"action":  "add",
				"content": "Default priority todo",
			},
			wantErr: false,
		},
		{
			name: "缺少 content 参数",
			params: map[string]any{
				"action": "add",
			},
			wantErr: true,
		},
		{
			name: "空 content",
			params: map[string]any{
				"action":  "add",
				"content": "",
			},
			wantErr: true,
		},
		{
			name: "只有空格的 content",
			params: map[string]any{
				"action":  "add",
				"content": "   ",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if !strings.Contains(result, "✅ Todo added successfully") {
					t.Error("Execute(add) should return success message")
				}
				if content, ok := tt.params["content"].(string); ok {
					if !strings.Contains(result, content) {
						t.Error("Execute(add) should contain todo content")
					}
				}
			}
		})
	}
}

func TestTodoTool_Execute_Update(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 先添加一个 todo
	addResult, err := tool.Execute(map[string]any{
		"action":  "add",
		"content": "Test todo",
	})
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}
	
	// 从结果中提取 ID
	todoID := extractTodoID(addResult)
	if todoID == "" {
		t.Fatal("Failed to extract todo ID from add result")
	}
	
	tests := []struct {
		name    string
		params  map[string]any
		wantErr bool
	}{
		{
			name: "更新状态",
			params: map[string]any{
				"action": "update",
				"id":     todoID,
				"status": "in_progress",
			},
			wantErr: false,
		},
		{
			name: "更新内容",
			params: map[string]any{
				"action":  "update",
				"id":      todoID,
				"content": "Updated content",
			},
			wantErr: false,
		},
		{
			name: "更新优先级",
			params: map[string]any{
				"action":   "update",
				"id":       todoID,
				"priority": "low",
			},
			wantErr: false,
		},
		{
			name: "缺少 ID 参数",
			params: map[string]any{
				"action": "update",
				"status": "completed",
			},
			wantErr: true,
		},
		{
			name: "不存在的 ID",
			params: map[string]any{
				"action": "update",
				"id":     "nonexistent",
				"status": "completed",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if !strings.Contains(result, "✅ Todo updated successfully") {
					t.Error("Execute(update) should return success message")
				}
			}
		})
	}
}

func TestTodoTool_Execute_Delete(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 先添加一个 todo
	addResult, err := tool.Execute(map[string]any{
		"action":  "add",
		"content": "Test todo",
	})
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}
	
	todoID := extractTodoID(addResult)
	if todoID == "" {
		t.Fatal("Failed to extract todo ID from add result")
	}
	
	// 测试删除存在的 todo
	result, err := tool.Execute(map[string]any{
		"action": "delete",
		"id":     todoID,
	})
	if err != nil {
		t.Fatalf("Execute(delete) failed: %v", err)
	}
	if !strings.Contains(result, "✅ Todo deleted successfully") {
		t.Error("Execute(delete) should return success message")
	}
	
	// 测试删除不存在的 todo
	_, err = tool.Execute(map[string]any{
		"action": "delete",
		"id":     "nonexistent",
	})
	if err == nil {
		t.Error("Execute(delete) should fail for nonexistent todo")
	}
	
	// 测试缺少 ID 参数
	_, err = tool.Execute(map[string]any{
		"action": "delete",
	})
	if err == nil {
		t.Error("Execute(delete) should fail without ID parameter")
	}
}

func TestTodoTool_Execute_Clear(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 测试清空空列表
	result, err := tool.Execute(map[string]any{"action": "clear"})
	if err != nil {
		t.Fatalf("Execute(clear) failed: %v", err)
	}
	if !strings.Contains(result, "No todos to clear") {
		t.Error("Execute(clear) should return 'No todos to clear' for empty list")
	}
	
	// 添加一些 todo
	tool.Execute(map[string]any{"action": "add", "content": "Todo 1"})
	tool.Execute(map[string]any{"action": "add", "content": "Todo 2"})
	
	// 测试清空
	result, err = tool.Execute(map[string]any{"action": "clear"})
	if err != nil {
		t.Fatalf("Execute(clear) failed: %v", err)
	}
	if !strings.Contains(result, "✅ Cleared") {
		t.Error("Execute(clear) should return success message")
	}
	
	// 验证已清空
	listResult, err := tool.Execute(map[string]any{"action": "list"})
	if err != nil {
		t.Fatalf("Execute(list) failed: %v", err)
	}
	if !strings.Contains(listResult, "No todos found") {
		t.Error("List should be empty after clear")
	}
}

func TestTodoTool_Execute_Count(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 测试计数空列表
	result, err := tool.Execute(map[string]any{"action": "count"})
	if err != nil {
		t.Fatalf("Execute(count) failed: %v", err)
	}
	if !strings.Contains(result, "📊 Todo Statistics") {
		t.Error("Execute(count) should contain statistics header")
	}
	if !strings.Contains(result, "Total: 0") {
		t.Error("Execute(count) should show total 0 for empty list")
	}
	
	// 添加一些 todo
	addResult, _ := tool.Execute(map[string]any{"action": "add", "content": "Todo 1"})
	todoID := extractTodoID(addResult)
	tool.Execute(map[string]any{"action": "add", "content": "Todo 2"})
	tool.Execute(map[string]any{"action": "update", "id": todoID, "status": "in_progress"})
	
	// 测试计数
	result, err = tool.Execute(map[string]any{"action": "count"})
	if err != nil {
		t.Fatalf("Execute(count) failed: %v", err)
	}
	if !strings.Contains(result, "Total: 2") {
		t.Error("Execute(count) should show correct total")
	}
	if !strings.Contains(result, "Pending: 1") {
		t.Error("Execute(count) should show correct pending count")
	}
	if !strings.Contains(result, "In Progress: 1") {
		t.Error("Execute(count) should show correct in progress count")
	}
}

func TestTodoTool_Execute_FilterByStatus(t *testing.T) {
	tool := createTestTodoTool(t)
	
	// 添加不同状态的 todo
	_, _ = tool.Execute(map[string]any{"action": "add", "content": "Pending todo"})
	addResult2, _ := tool.Execute(map[string]any{"action": "add", "content": "Progress todo"})
	
	todoID2 := extractTodoID(addResult2)
	tool.Execute(map[string]any{"action": "update", "id": todoID2, "status": "in_progress"})
	
	// 测试按状态筛选
	result, err := tool.Execute(map[string]any{
		"action":        "list",
		"filter_status": "pending",
	})
	if err != nil {
		t.Fatalf("Execute(list with filter) failed: %v", err)
	}
	if !strings.Contains(result, "Pending todo") {
		t.Error("Filtered list should contain pending todo")
	}
	if strings.Contains(result, "Progress todo") {
		t.Error("Filtered list should not contain in-progress todo")
	}
	
	// 测试筛选进行中的 todo
	result, err = tool.Execute(map[string]any{
		"action":        "list",
		"filter_status": "in_progress",
	})
	if err != nil {
		t.Fatalf("Execute(list with filter) failed: %v", err)
	}
	if strings.Contains(result, "Pending todo") {
		t.Error("Filtered list should not contain pending todo")
	}
	if !strings.Contains(result, "Progress todo") {
		t.Error("Filtered list should contain in-progress todo")
	}
}

func TestTodoTool_Execute_InvalidAction(t *testing.T) {
	tool := createTestTodoTool(t)
	
	_, err := tool.Execute(map[string]any{"action": "invalid"})
	if err == nil {
		t.Error("Execute() should fail with invalid action")
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Error("Error should mention unknown action")
	}
}

func TestTodoTool_Execute_MissingAction(t *testing.T) {
	tool := createTestTodoTool(t)
	
	_, err := tool.Execute(map[string]any{})
	if err == nil {
		t.Error("Execute() should fail without action parameter")
	}
	if !strings.Contains(err.Error(), "action parameter is required") {
		t.Error("Error should mention missing action parameter")
	}
}

// 辅助函数
func createTestTodoTool(_ *testing.T) *TodoTool {
	storage := session.NewMemoryStorage()
	manager := session.NewTodoManager(storage)
	
	return &TodoTool{
		manager: manager,
	}
}

func extractTodoID(result string) string {
	// 从结果中提取 ID，格式如：[ID] content
	start := strings.Index(result, "[")
	if start == -1 {
		return ""
	}
	end := strings.Index(result[start:], "]")
	if end == -1 {
		return ""
	}
	return result[start+1 : start+end]
}