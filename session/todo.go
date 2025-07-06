package session

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TodoStatus 表示 todo 项的状态
type TodoStatus string

const (
	StatusPending    TodoStatus = "pending"
	StatusInProgress TodoStatus = "in_progress"
	StatusCompleted  TodoStatus = "completed"
)

// TodoPriority 表示 todo 项的优先级
type TodoPriority string

const (
	PriorityHigh   TodoPriority = "high"
	PriorityMedium TodoPriority = "medium"
	PriorityLow    TodoPriority = "low"
)

// TodoItem 表示单个 todo 项
type TodoItem struct {
	ID          string       `json:"id"`
	Content     string       `json:"content"`
	Status      TodoStatus   `json:"status"`
	Priority    TodoPriority `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TodoManager 管理 todo 列表
type TodoManager struct {
	items   map[string]*TodoItem
	storage Storage
}

// NewTodoManager 创建新的 TodoManager
func NewTodoManager(storage Storage) *TodoManager {
	return &TodoManager{
		items:   make(map[string]*TodoItem),
		storage: storage,
	}
}

// Load 从存储加载 todo 数据
func (tm *TodoManager) Load() error {
	items, err := tm.storage.Load()
	if err != nil {
		return fmt.Errorf("failed to load todos: %v", err)
	}
	tm.items = items
	return nil
}

// Save 保存 todo 数据到存储
func (tm *TodoManager) Save() error {
	if err := tm.storage.Save(tm.items); err != nil {
		return fmt.Errorf("failed to save todos: %v", err)
	}
	return nil
}

// Add 添加新的 todo 项
func (tm *TodoManager) Add(content string, priority TodoPriority) (*TodoItem, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("todo content cannot be empty")
	}

	id := generateID()
	now := time.Now()
	
	item := &TodoItem{
		ID:        id,
		Content:   strings.TrimSpace(content),
		Status:    StatusPending,
		Priority:  priority,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tm.items[id] = item
	return item, nil
}

// Update 更新 todo 项
func (tm *TodoManager) Update(id string, status TodoStatus, content string, priority TodoPriority) (*TodoItem, error) {
	item, exists := tm.items[id]
	if !exists {
		return nil, fmt.Errorf("todo item with id %s not found", id)
	}

	now := time.Now()
	
	if status != "" {
		item.Status = status
		item.UpdatedAt = now
	}
	
	if strings.TrimSpace(content) != "" {
		item.Content = strings.TrimSpace(content)
		item.UpdatedAt = now
	}
	
	if priority != "" {
		item.Priority = priority
		item.UpdatedAt = now
	}

	return item, nil
}

// Delete 删除 todo 项
func (tm *TodoManager) Delete(id string) error {
	if _, exists := tm.items[id]; !exists {
		return fmt.Errorf("todo item with id %s not found", id)
	}
	delete(tm.items, id)
	return nil
}

// Get 获取单个 todo 项
func (tm *TodoManager) Get(id string) (*TodoItem, error) {
	item, exists := tm.items[id]
	if !exists {
		return nil, fmt.Errorf("todo item with id %s not found", id)
	}
	return item, nil
}

// List 列出所有 todo 项，按优先级和创建时间排序
func (tm *TodoManager) List() []*TodoItem {
	items := make([]*TodoItem, 0, len(tm.items))
	for _, item := range tm.items {
		items = append(items, item)
	}

	// 按优先级和创建时间排序
	sort.Slice(items, func(i, j int) bool {
		// 先按状态排序：pending < in_progress < completed
		statusOrder := map[TodoStatus]int{
			StatusPending:    0,
			StatusInProgress: 1,
			StatusCompleted:  2,
		}
		
		if statusOrder[items[i].Status] != statusOrder[items[j].Status] {
			return statusOrder[items[i].Status] < statusOrder[items[j].Status]
		}
		
		// 再按优先级排序：high < medium < low
		priorityOrder := map[TodoPriority]int{
			PriorityHigh:   0,
			PriorityMedium: 1,
			PriorityLow:    2,
		}
		
		if priorityOrder[items[i].Priority] != priorityOrder[items[j].Priority] {
			return priorityOrder[items[i].Priority] < priorityOrder[items[j].Priority]
		}
		
		// 最后按创建时间排序
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items
}

// ListByStatus 按状态筛选 todo 项
func (tm *TodoManager) ListByStatus(status TodoStatus) []*TodoItem {
	items := tm.List()
	filtered := make([]*TodoItem, 0)
	
	for _, item := range items {
		if item.Status == status {
			filtered = append(filtered, item)
		}
	}
	
	return filtered
}

// Clear 清空所有 todo 项
func (tm *TodoManager) Clear() {
	tm.items = make(map[string]*TodoItem)
}

// Count 统计不同状态的 todo 数量
func (tm *TodoManager) Count() map[TodoStatus]int {
	counts := map[TodoStatus]int{
		StatusPending:    0,
		StatusInProgress: 0,
		StatusCompleted:  0,
	}
	
	for _, item := range tm.items {
		counts[item.Status]++
	}
	
	return counts
}

// String 返回 todo 项的字符串表示
func (item *TodoItem) String() string {
	statusSymbol := map[TodoStatus]string{
		StatusPending:    "⏳",
		StatusInProgress: "🔄",
		StatusCompleted:  "✅",
	}
	
	prioritySymbol := map[TodoPriority]string{
		PriorityHigh:   "🔴",
		PriorityMedium: "🟡",
		PriorityLow:    "🟢",
	}
	
	return fmt.Sprintf("%s %s [%s] %s",
		statusSymbol[item.Status],
		prioritySymbol[item.Priority],
		item.ID,
		item.Content,
	)
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}