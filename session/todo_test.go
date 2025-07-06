package session

import (
	"os"
	"testing"
	"time"
)

func TestTodoItem_String(t *testing.T) {
	item := &TodoItem{
		ID:       "123",
		Content:  "Test todo",
		Status:   StatusPending,
		Priority: PriorityHigh,
	}

	str := item.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}
	
	// 应该包含状态和优先级符号
	if !contains(str, "⏳") {
		t.Error("String() should contain pending status symbol")
	}
	if !contains(str, "🔴") {
		t.Error("String() should contain high priority symbol")
	}
	if !contains(str, "Test todo") {
		t.Error("String() should contain content")
	}
}

func TestTodoManager_Add(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 测试添加正常 todo
	item, err := manager.Add("Test todo", PriorityHigh)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if item.Content != "Test todo" {
		t.Errorf("Add() content = %v, want %v", item.Content, "Test todo")
	}
	if item.Status != StatusPending {
		t.Errorf("Add() status = %v, want %v", item.Status, StatusPending)
	}
	if item.Priority != PriorityHigh {
		t.Errorf("Add() priority = %v, want %v", item.Priority, PriorityHigh)
	}

	// 测试添加空内容
	_, err = manager.Add("", PriorityMedium)
	if err == nil {
		t.Error("Add() should fail with empty content")
	}

	// 测试添加只有空格的内容
	_, err = manager.Add("   ", PriorityMedium)
	if err == nil {
		t.Error("Add() should fail with whitespace-only content")
	}
}

func TestTodoManager_Update(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加一个 todo
	item, err := manager.Add("Test todo", PriorityMedium)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 测试更新状态
	updatedItem, err := manager.Update(item.ID, StatusInProgress, "", TodoPriority(""))
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updatedItem.Status != StatusInProgress {
		t.Errorf("Update() status = %v, want %v", updatedItem.Status, StatusInProgress)
	}

	// 测试更新内容
	_, err = manager.Update(item.ID, TodoStatus(""), "Updated content", TodoPriority(""))
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updatedItem.Content != "Updated content" {
		t.Errorf("Update() content = %v, want %v", updatedItem.Content, "Updated content")
	}

	// 测试更新优先级
	_, err = manager.Update(item.ID, TodoStatus(""), "", PriorityLow)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if updatedItem.Priority != PriorityLow {
		t.Errorf("Update() priority = %v, want %v", updatedItem.Priority, PriorityLow)
	}

	// 测试更新不存在的 todo
	_, err = manager.Update("nonexistent", StatusCompleted, "", TodoPriority(""))
	if err == nil {
		t.Error("Update() should fail with nonexistent ID")
	}
}

func TestTodoManager_Delete(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加一个 todo
	item, err := manager.Add("Test todo", PriorityMedium)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 测试删除存在的 todo
	err = manager.Delete(item.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// 验证 todo 已被删除
	_, err = manager.Get(item.ID)
	if err == nil {
		t.Error("Get() should fail after deletion")
	}

	// 测试删除不存在的 todo
	err = manager.Delete("nonexistent")
	if err == nil {
		t.Error("Delete() should fail with nonexistent ID")
	}
}

func TestTodoManager_List(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加多个 todo
	_, err := manager.Add("High priority pending", PriorityHigh)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	item2, err := manager.Add("Medium priority", PriorityMedium)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	_, err = manager.Add("Low priority", PriorityLow)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// 更新一个 todo 的状态
	_, err = manager.Update(item2.ID, StatusInProgress, "", TodoPriority(""))
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// 测试列表排序
	items := manager.List()
	if len(items) != 3 {
		t.Fatalf("List() returned %d items, want 3", len(items))
	}

	// 验证排序：pending > in_progress，在同一状态内按优先级排序
	if items[0].Status != StatusPending || items[0].Priority != PriorityHigh {
		t.Error("First item should be pending high priority")
	}
	if items[1].Status != StatusPending || items[1].Priority != PriorityLow {
		t.Error("Second item should be pending low priority")
	}
	if items[2].Status != StatusInProgress {
		t.Error("Third item should be in progress")
	}
}

func TestTodoManager_ListByStatus(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加不同状态的 todo
	item1, _ := manager.Add("Pending todo", PriorityMedium)
	item2, _ := manager.Add("Another todo", PriorityMedium)
	manager.Update(item2.ID, StatusInProgress, "", TodoPriority(""))

	// 测试按状态筛选
	pendingItems := manager.ListByStatus(StatusPending)
	if len(pendingItems) != 1 {
		t.Errorf("ListByStatus(pending) returned %d items, want 1", len(pendingItems))
	}
	if pendingItems[0].ID != item1.ID {
		t.Error("ListByStatus(pending) returned wrong item")
	}

	inProgressItems := manager.ListByStatus(StatusInProgress)
	if len(inProgressItems) != 1 {
		t.Errorf("ListByStatus(in_progress) returned %d items, want 1", len(inProgressItems))
	}

	completedItems := manager.ListByStatus(StatusCompleted)
	if len(completedItems) != 0 {
		t.Errorf("ListByStatus(completed) returned %d items, want 0", len(completedItems))
	}
}

func TestTodoManager_Count(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 初始计数应该为 0
	counts := manager.Count()
	if counts[StatusPending] != 0 || counts[StatusInProgress] != 0 || counts[StatusCompleted] != 0 {
		t.Error("Initial counts should be zero")
	}

	// 添加不同状态的 todo
	_, _ = manager.Add("Todo 1", PriorityMedium)
	item2, _ := manager.Add("Todo 2", PriorityMedium)
	item3, _ := manager.Add("Todo 3", PriorityMedium)

	manager.Update(item2.ID, StatusInProgress, "", TodoPriority(""))
	manager.Update(item3.ID, StatusCompleted, "", TodoPriority(""))

	// 验证计数
	counts = manager.Count()
	if counts[StatusPending] != 1 {
		t.Errorf("Pending count = %d, want 1", counts[StatusPending])
	}
	if counts[StatusInProgress] != 1 {
		t.Errorf("InProgress count = %d, want 1", counts[StatusInProgress])
	}
	if counts[StatusCompleted] != 1 {
		t.Errorf("Completed count = %d, want 1", counts[StatusCompleted])
	}
}

func TestTodoManager_Clear(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加一些 todo
	manager.Add("Todo 1", PriorityMedium)
	manager.Add("Todo 2", PriorityMedium)

	// 清空
	manager.Clear()

	// 验证已清空
	items := manager.List()
	if len(items) != 0 {
		t.Errorf("After Clear(), List() returned %d items, want 0", len(items))
	}

	counts := manager.Count()
	total := counts[StatusPending] + counts[StatusInProgress] + counts[StatusCompleted]
	if total != 0 {
		t.Errorf("After Clear(), total count = %d, want 0", total)
	}
}

func TestTodoManager_SaveLoad(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTodoManager(storage)

	// 添加一些 todo
	item1, _ := manager.Add("Todo 1", PriorityHigh)
	item2, _ := manager.Add("Todo 2", PriorityMedium)
	manager.Update(item2.ID, StatusInProgress, "", TodoPriority(""))

	// 保存
	err := manager.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// 创建新的 manager 并加载
	newManager := NewTodoManager(storage)
	err = newManager.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 验证数据一致性
	items := newManager.List()
	if len(items) != 2 {
		t.Fatalf("After Load(), List() returned %d items, want 2", len(items))
	}

	// 验证具体数据
	loadedItem1, err := newManager.Get(item1.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if loadedItem1.Content != "Todo 1" || loadedItem1.Priority != PriorityHigh {
		t.Error("Loaded item1 data mismatch")
	}

	loadedItem2, err := newManager.Get(item2.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if loadedItem2.Status != StatusInProgress {
		t.Error("Loaded item2 status mismatch")
	}
}

func TestFileStorage(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "todo_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	storage := NewFileStorage(tmpFile.Name())

	// 测试保存和加载空数据
	err = storage.Save(make(map[string]*TodoItem))
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	items, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if len(items) != 0 {
		t.Error("Load() should return empty map for empty file")
	}

	// 测试保存和加载实际数据
	testItems := map[string]*TodoItem{
		"1": {
			ID:        "1",
			Content:   "Test todo",
			Status:    StatusPending,
			Priority:  PriorityHigh,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err = storage.Save(testItems)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	loadedItems, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if len(loadedItems) != 1 {
		t.Fatalf("Load() returned %d items, want 1", len(loadedItems))
	}

	loadedItem := loadedItems["1"]
	if loadedItem.Content != "Test todo" {
		t.Errorf("Loaded content = %v, want %v", loadedItem.Content, "Test todo")
	}
}

func TestFileStorage_NonexistentFile(t *testing.T) {
	storage := NewFileStorage("nonexistent.json")

	// 测试加载不存在的文件
	items, err := storage.Load()
	if err != nil {
		t.Fatalf("Load() should not fail for nonexistent file: %v", err)
	}
	if len(items) != 0 {
		t.Error("Load() should return empty map for nonexistent file")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) >= len(substr) && s[:len(substr)] == substr ||
		   len(s) > len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}