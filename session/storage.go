package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage 定义存储接口
type Storage interface {
	Load() (map[string]*TodoItem, error)
	Save(items map[string]*TodoItem) error
}

// FileStorage 实现基于文件的存储
type FileStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewFileStorage 创建新的文件存储
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

// NewDefaultFileStorage 创建默认的文件存储（存储在用户目录）
func NewDefaultFileStorage() (*FileStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".opencode_nano")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	filePath := filepath.Join(configDir, "session_todos.json")
	return NewFileStorage(filePath), nil
}

// Load 从文件加载 todo 数据
func (fs *FileStorage) Load() (map[string]*TodoItem, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	items := make(map[string]*TodoItem)

	// 如果文件不存在，返回空的 map
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return items, nil
	}

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 如果文件为空，返回空的 map
	if len(data) == 0 {
		return items, nil
	}

	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return items, nil
}

// Save 保存 todo 数据到文件
func (fs *FileStorage) Save(items map[string]*TodoItem) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 写入临时文件后重命名，确保原子性
	tempFile := fs.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := os.Rename(tempFile, fs.filePath); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// MemoryStorage 实现基于内存的存储（主要用于测试）
type MemoryStorage struct {
	items map[string]*TodoItem
	mu    sync.RWMutex
}

// NewMemoryStorage 创建新的内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		items: make(map[string]*TodoItem),
	}
}

// Load 从内存加载 todo 数据
func (ms *MemoryStorage) Load() (map[string]*TodoItem, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// 深拷贝避免并发问题
	items := make(map[string]*TodoItem)
	for k, v := range ms.items {
		itemCopy := *v
		items[k] = &itemCopy
	}

	return items, nil
}

// Save 保存 todo 数据到内存
func (ms *MemoryStorage) Save(items map[string]*TodoItem) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 深拷贝避免并发问题
	ms.items = make(map[string]*TodoItem)
	for k, v := range items {
		itemCopy := *v
		ms.items[k] = &itemCopy
	}

	return nil
}