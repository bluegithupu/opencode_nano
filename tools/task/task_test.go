package task

import (
	"context"
	"opencode_nano/session"
	"opencode_nano/tools/core"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskTool(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "task_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create task tool with custom storage path
	storagePath := filepath.Join(tmpDir, "test_todos.json")
	tool, err := NewTaskTool()
	if err != nil {
		t.Fatal(err)
	}

	// Override the storage path by recreating with custom storage
	storage := session.NewFileStorage(storagePath)
	tool.manager = session.NewTodoManager(storage)

	// Test 1: List empty todos
	t.Run("ListEmpty", func(t *testing.T) {
		params := core.NewMapParameters(map[string]any{
			"action": "list",
		})
		
		result, err := tool.Execute(context.Background(), params)
		if err != nil {
			t.Errorf("Failed to list empty todos: %v", err)
		}
		
		if !strings.Contains(result.String(), "No todos found") {
			t.Errorf("Expected 'No todos found', got: %s", result.String())
		}
	})

	// Test 2: Add todo
	t.Run("AddTodo", func(t *testing.T) {
		params := core.NewMapParameters(map[string]any{
			"action":   "add",
			"content":  "Test task 1",
			"priority": "high",
		})
		
		result, err := tool.Execute(context.Background(), params)
		if err != nil {
			t.Errorf("Failed to add todo: %v", err)
		}
		
		if !result.Success() {
			t.Errorf("Add todo failed: %v", result.Error())
		}
		
		if !strings.Contains(result.String(), "Todo added successfully") {
			t.Errorf("Unexpected result: %s", result.String())
		}
		
		// Check if ID is in metadata
		if result.Metadata()["id"] == nil {
			t.Error("Todo ID not found in metadata")
		}
	})

	// Test 3: Add todo with default priority
	t.Run("AddTodoDefaultPriority", func(t *testing.T) {
		params := core.NewMapParameters(map[string]any{
			"action":  "add",
			"content": "Test task 2",
		})
		
		result, err := tool.Execute(context.Background(), params)
		if err != nil {
			t.Errorf("Failed to add todo: %v", err)
		}
		
		if !result.Success() {
			t.Errorf("Add todo failed: %v", result.Error())
		}
	})

	// Test 4: List todos
	t.Run("ListTodos", func(t *testing.T) {
		params := core.NewMapParameters(map[string]any{
			"action": "list",
		})
		
		result, err := tool.Execute(context.Background(), params)
		if err != nil {
			t.Errorf("Failed to list todos: %v", err)
		}
		
		output := result.String()
		if !strings.Contains(output, "Todo List") {
			t.Errorf("Missing todo list header")
		}
		
		if !strings.Contains(output, "Test task 1") {
			t.Errorf("Missing first todo")
		}
		
		if !strings.Contains(output, "Test task 2") {
			t.Errorf("Missing second todo")
		}
		
		if !strings.Contains(output, "Summary") {
			t.Errorf("Missing summary section")
		}
		
		if !strings.Contains(output, "Pending: 2") {
			t.Errorf("Wrong pending count")
		}
	})

	// Test 5: Update todo status
	t.Run("UpdateTodoStatus", func(t *testing.T) {
		// First get the list to find an ID
		listParams := core.NewMapParameters(map[string]any{
			"action": "list",
		})
		listResult, _ := tool.Execute(context.Background(), listParams)
		
		// Extract first ID from the output (hacky but works for test)
		lines := strings.Split(listResult.String(), "\n")
		var todoID string
		for _, line := range lines {
			if strings.Contains(line, "[") && strings.Contains(line, "]") {
				start := strings.Index(line, "[") + 1
				end := strings.Index(line, "]")
				todoID = line[start:end]
				break
			}
		}
		
		if todoID == "" {
			t.Fatal("Could not find todo ID")
		}
		
		// Update the todo
		updateParams := core.NewMapParameters(map[string]any{
			"action": "update",
			"id":     todoID,
			"status": "in_progress",
		})
		
		result, err := tool.Execute(context.Background(), updateParams)
		if err != nil {
			t.Errorf("Failed to update todo: %v", err)
		}
		
		if !strings.Contains(result.String(), "Todo updated successfully") {
			t.Errorf("Unexpected result: %s", result.String())
		}
		
		// Verify the update
		listResult2, _ := tool.Execute(context.Background(), listParams)
		if !strings.Contains(listResult2.String(), "In Progress: 1") {
			t.Errorf("Todo status not updated correctly")
		}
	})

	// Test 6: Update todo priority and content
	t.Run("UpdateTodoPriorityContent", func(t *testing.T) {
		// Get list to find ID
		listParams := core.NewMapParameters(map[string]any{
			"action": "list",
		})
		listResult, _ := tool.Execute(context.Background(), listParams)
		
		// Extract second ID
		lines := strings.Split(listResult.String(), "\n")
		var todoID string
		count := 0
		for _, line := range lines {
			if strings.Contains(line, "[") && strings.Contains(line, "]") {
				if count == 1 {
					start := strings.Index(line, "[") + 1
					end := strings.Index(line, "]")
					todoID = line[start:end]
					break
				}
				count++
			}
		}
		
		if todoID == "" {
			t.Fatal("Could not find second todo ID")
		}
		
		// Update priority and content
		updateParams := core.NewMapParameters(map[string]any{
			"action":   "update",
			"id":       todoID,
			"priority": "low",
			"content":  "Updated task 2",
		})
		
		_, err := tool.Execute(context.Background(), updateParams)
		if err != nil {
			t.Errorf("Failed to update todo: %v", err)
		}
		
		// Verify updates
		listResult2, _ := tool.Execute(context.Background(), listParams)
		output := listResult2.String()
		
		if !strings.Contains(output, "Updated task 2") {
			t.Errorf("Content not updated")
		}
		
		// Check for low priority symbol (🟢)
		if !strings.Contains(output, "🟢") {
			t.Errorf("Priority not updated to low")
		}
	})

	// Test 7: Error cases
	t.Run("ErrorCases", func(t *testing.T) {
		// Missing action
		params := core.NewMapParameters(map[string]any{})
		_, err := tool.Execute(context.Background(), params)
		if err == nil {
			t.Error("Expected error for missing action")
		}
		
		// Invalid action
		params = core.NewMapParameters(map[string]any{
			"action": "delete", // Not supported anymore
		})
		_, err = tool.Execute(context.Background(), params)
		if err == nil {
			t.Error("Expected error for invalid action")
		}
		
		// Add without content
		params = core.NewMapParameters(map[string]any{
			"action": "add",
		})
		_, err = tool.Execute(context.Background(), params)
		if err == nil {
			t.Error("Expected error for add without content")
		}
		
		// Update without ID
		params = core.NewMapParameters(map[string]any{
			"action": "update",
			"status": "completed",
		})
		_, err = tool.Execute(context.Background(), params)
		if err == nil {
			t.Error("Expected error for update without ID")
		}
		
		// Update with non-existent ID
		params = core.NewMapParameters(map[string]any{
			"action": "update",
			"id":     "non-existent-id",
			"status": "completed",
		})
		_, err = tool.Execute(context.Background(), params)
		if err == nil {
			t.Error("Expected error for non-existent ID")
		}
	})

	// Test 8: Complete todo workflow
	t.Run("CompleteTodoWorkflow", func(t *testing.T) {
		// Add a new todo
		addParams := core.NewMapParameters(map[string]any{
			"action":   "add",
			"content":  "Complete workflow test",
			"priority": "high",
		})
		
		addResult, err := tool.Execute(context.Background(), addParams)
		if err != nil {
			t.Fatal(err)
		}
		
		todoID := addResult.Metadata()["id"].(string)
		
		// Update to in_progress
		updateParams := core.NewMapParameters(map[string]any{
			"action": "update",
			"id":     todoID,
			"status": "in_progress",
		})
		
		_, err = tool.Execute(context.Background(), updateParams)
		if err != nil {
			t.Error(err)
		}
		
		// Update to completed
		updateParams = core.NewMapParameters(map[string]any{
			"action": "update",
			"id":     todoID,
			"status": "completed",
		})
		
		_, err = tool.Execute(context.Background(), updateParams)
		if err != nil {
			t.Error(err)
		}
		
		// Verify final state
		listParams := core.NewMapParameters(map[string]any{
			"action": "list",
		})
		listResult, _ := tool.Execute(context.Background(), listParams)
		
		if !strings.Contains(listResult.String(), "Completed: 1") {
			t.Error("Todo not marked as completed")
		}
	})

	// Test 9: Schema validation
	t.Run("SchemaValidation", func(t *testing.T) {
		schema := tool.Schema()
		
		// Check required fields
		if len(schema.Required) != 1 || schema.Required[0] != "action" {
			t.Error("Schema should require only 'action'")
		}
		
		// Check action enum
		actionProp := schema.Properties["action"]
		if len(actionProp.Enum) != 3 {
			t.Error("Action should have exactly 3 options")
		}
		
		expectedActions := map[string]bool{
			"list":   true,
			"add":    true,
			"update": true,
		}
		
		for _, action := range actionProp.Enum {
			if !expectedActions[action] {
				t.Errorf("Unexpected action in enum: %s", action)
			}
		}
		
		// Check status enum
		statusProp := schema.Properties["status"]
		if len(statusProp.Enum) != 3 {
			t.Error("Status should have 3 options")
		}
		
		// Check priority enum
		priorityProp := schema.Properties["priority"]
		if len(priorityProp.Enum) != 3 {
			t.Error("Priority should have 3 options")
		}
	})

	// Test 10: Tool info
	t.Run("ToolInfo", func(t *testing.T) {
		info := tool.Info()
		
		if info.Name != "todo" {
			t.Errorf("Expected name 'todo', got '%s'", info.Name)
		}
		
		if info.Category != "development" {
			t.Errorf("Expected category 'development', got '%s'", info.Category)
		}
		
		if !strings.Contains(info.Description, "list, add, update") {
			t.Error("Description should mention the three supported operations")
		}
		
		// Check tags
		expectedTags := map[string]bool{
			"task":     true,
			"todo":     true,
			"project":  true,
			"planning": true,
		}
		
		for _, tag := range info.Tags {
			if !expectedTags[tag] {
				t.Errorf("Unexpected tag: %s", tag)
			}
		}
	})
}