package core

import (
	"fmt"
	"strings"
	"sync"
)

// ToolRegistry 工具注册表实现
type ToolRegistry struct {
	mu         sync.RWMutex
	tools      map[string]Tool
	aliases    map[string]string
	categories map[string][]Tool
	tagIndex   map[string][]Tool
}

// NewRegistry 创建新的注册表
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:      make(map[string]Tool),
		aliases:    make(map[string]string),
		categories: make(map[string][]Tool),
		tagIndex:   make(map[string][]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool, aliases ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := tool.Info()
	
	// 检查名称是否已存在
	if _, exists := r.tools[info.Name]; exists {
		return fmt.Errorf("tool %s already registered", info.Name)
	}
	
	// 注册工具
	r.tools[info.Name] = tool
	
	// 注册别名
	for _, alias := range aliases {
		if _, exists := r.aliases[alias]; exists {
			return fmt.Errorf("alias %s already in use", alias)
		}
		r.aliases[alias] = info.Name
	}
	
	// 更新分类索引
	if info.Category != "" {
		r.categories[info.Category] = append(r.categories[info.Category], tool)
	}
	
	// 更新标签索引
	for _, tag := range info.Tags {
		r.tagIndex[tag] = append(r.tagIndex[tag], tool)
	}
	
	return nil
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// 先尝试直接查找
	if tool, exists := r.tools[name]; exists {
		return tool, nil
	}
	
	// 尝试通过别名查找
	if realName, exists := r.aliases[name]; exists {
		if tool, exists := r.tools[realName]; exists {
			return tool, nil
		}
	}
	
	return nil, ErrToolNotFound(name)
}

// Find 查找工具（支持模糊搜索）
func (r *ToolRegistry) Find(query string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	query = strings.ToLower(query)
	var results []Tool
	seen := make(map[string]bool)
	
	// 搜索工具名称
	for name, tool := range r.tools {
		if strings.Contains(strings.ToLower(name), query) {
			if !seen[name] {
				results = append(results, tool)
				seen[name] = true
			}
		}
	}
	
	// 搜索别名
	for alias, realName := range r.aliases {
		if strings.Contains(strings.ToLower(alias), query) {
			if tool, exists := r.tools[realName]; exists && !seen[realName] {
				results = append(results, tool)
				seen[realName] = true
			}
		}
	}
	
	// 搜索描述
	for name, tool := range r.tools {
		info := tool.Info()
		if strings.Contains(strings.ToLower(info.Description), query) && !seen[name] {
			results = append(results, tool)
			seen[name] = true
		}
	}
	
	// 搜索标签
	for tag, tools := range r.tagIndex {
		if strings.Contains(strings.ToLower(tag), query) {
			for _, tool := range tools {
				info := tool.Info()
				if !seen[info.Name] {
					results = append(results, tool)
					seen[info.Name] = true
				}
			}
		}
	}
	
	return results
}

// GetByCategory 按分类获取工具
func (r *ToolRegistry) GetByCategory(category string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools, exists := r.categories[category]
	if !exists {
		return []Tool{}
	}
	
	// 返回副本以避免并发问题
	result := make([]Tool, len(tools))
	copy(result, tools)
	return result
}

// GetByTags 按标签获取工具
func (r *ToolRegistry) GetByTags(tags ...string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if len(tags) == 0 {
		return []Tool{}
	}
	
	// 使用第一个标签作为基准
	toolMap := make(map[string]Tool)
	for _, tool := range r.tagIndex[tags[0]] {
		info := tool.Info()
		toolMap[info.Name] = tool
	}
	
	// 对于后续标签，保留交集
	for i := 1; i < len(tags); i++ {
		nextMap := make(map[string]Tool)
		for _, tool := range r.tagIndex[tags[i]] {
			info := tool.Info()
			if _, exists := toolMap[info.Name]; exists {
				nextMap[info.Name] = tool
			}
		}
		toolMap = nextMap
	}
	
	// 转换为切片
	results := make([]Tool, 0, len(toolMap))
	for _, tool := range toolMap {
		results = append(results, tool)
	}
	
	return results
}

// All 获取所有工具
func (r *ToolRegistry) All() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	results := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		results = append(results, tool)
	}
	
	return results
}

// Categories 获取所有分类
func (r *ToolRegistry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	categories := make([]string, 0, len(r.categories))
	for category := range r.categories {
		categories = append(categories, category)
	}
	
	return categories
}

// Unregister 注销工具（用于测试或动态管理）
func (r *ToolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	tool, exists := r.tools[name]
	if !exists {
		return ErrToolNotFound(name)
	}
	
	info := tool.Info()
	
	// 删除主注册
	delete(r.tools, name)
	
	// 删除别名
	for alias, toolName := range r.aliases {
		if toolName == name {
			delete(r.aliases, alias)
		}
	}
	
	// 从分类中删除
	if info.Category != "" {
		newTools := []Tool{}
		for _, t := range r.categories[info.Category] {
			if t.Info().Name != name {
				newTools = append(newTools, t)
			}
		}
		if len(newTools) == 0 {
			delete(r.categories, info.Category)
		} else {
			r.categories[info.Category] = newTools
		}
	}
	
	// 从标签索引中删除
	for _, tag := range info.Tags {
		newTools := []Tool{}
		for _, t := range r.tagIndex[tag] {
			if t.Info().Name != name {
				newTools = append(newTools, t)
			}
		}
		if len(newTools) == 0 {
			delete(r.tagIndex, tag)
		} else {
			r.tagIndex[tag] = newTools
		}
	}
	
	return nil
}

// Has 检查工具是否存在
func (r *ToolRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if _, exists := r.tools[name]; exists {
		return true
	}
	
	if realName, exists := r.aliases[name]; exists {
		_, exists = r.tools[realName]
		return exists
	}
	
	return false
}

// GetAlias 获取别名对应的真实名称
func (r *ToolRegistry) GetAlias(alias string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	realName, exists := r.aliases[alias]
	return realName, exists
}

// GetAliases 获取工具的所有别名
func (r *ToolRegistry) GetAliases(toolName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var aliases []string
	for alias, name := range r.aliases {
		if name == toolName {
			aliases = append(aliases, alias)
		}
	}
	
	return aliases
}

// Stats 获取注册表统计信息
func (r *ToolRegistry) Stats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return map[string]int{
		"tools":      len(r.tools),
		"aliases":    len(r.aliases),
		"categories": len(r.categories),
		"tags":       len(r.tagIndex),
	}
}