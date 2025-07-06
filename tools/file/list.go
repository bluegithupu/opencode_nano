package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"opencode_nano/tools/core"
)

// ListTool 列出目录内容工具
type ListTool struct {
	*core.BaseTool
}

// NewListTool 创建列表工具
func NewListTool() *ListTool {
	tool := &ListTool{
		BaseTool: core.NewBaseTool("list", "file", "List directory contents with detailed information"),
	}
	
	tool.SetTags("file", "list", "ls", "dir")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "Directory path to list",
				Default:     ".",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List recursively",
				Default:     false,
			},
			"show_hidden": {
				Type:        "boolean",
				Description: "Show hidden files (starting with .)",
				Default:     false,
			},
			"sort_by": {
				Type:        "string",
				Description: "Sort by: name, size, time",
				Default:     "name",
				Enum:        []string{"name", "size", "time"},
			},
			"reverse": {
				Type:        "boolean",
				Description: "Reverse sort order",
				Default:     false,
			},
			"max_depth": {
				Type:        "integer",
				Description: "Maximum depth for recursive listing",
				Default:     10,
			},
			"include_details": {
				Type:        "boolean",
				Description: "Include file details (size, permissions, etc)",
				Default:     true,
			},
		},
		Required: []string{},
	})
	
	return tool
}

// FileInfo 文件信息
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"mod_time"`
	IsDir       bool      `json:"is_dir"`
	IsSymlink   bool      `json:"is_symlink"`
	Target      string    `json:"target,omitempty"`      // 符号链接目标
	Children    []FileInfo `json:"children,omitempty"`   // 子目录内容（递归时）
}

// Execute 执行列表操作
func (t *ListTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	path := "."
	if params.Has("path") {
		path, _ = params.GetString("path")
	}
	
	recursive := false
	if params.Has("recursive") {
		recursive, _ = params.GetBool("recursive")
	}
	
	showHidden := false
	if params.Has("show_hidden") {
		showHidden, _ = params.GetBool("show_hidden")
	}
	
	sortBy := "name"
	if params.Has("sort_by") {
		sortBy, _ = params.GetString("sort_by")
	}
	
	reverse := false
	if params.Has("reverse") {
		reverse, _ = params.GetBool("reverse")
	}
	
	maxDepth := 10
	if params.Has("max_depth") {
		maxDepth, _ = params.GetInt("max_depth")
	}
	
	includeDetails := true
	if params.Has("include_details") {
		includeDetails, _ = params.GetBool("include_details")
	}
	
	// 规范化路径
	path = filepath.Clean(path)
	
	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to stat path: %v", err))
	}
	
	var files []FileInfo
	var totalSize int64
	var fileCount, dirCount int
	
	if info.IsDir() {
		// 列出目录内容
		if recursive {
			rootInfo, err := t.listRecursive(ctx, path, showHidden, includeDetails, 0, maxDepth)
			if err != nil {
				return nil, core.ErrExecutionFailed(t.Info().Name, err.Error())
			}
			files = rootInfo.Children
			t.countStats(rootInfo, &totalSize, &fileCount, &dirCount)
		} else {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to read directory: %v", err))
			}
			
			for _, entry := range entries {
				// 检查是否显示隐藏文件
				if !showHidden && strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				
				fileInfo, err := t.getFileInfo(filepath.Join(path, entry.Name()), includeDetails)
				if err == nil {
					files = append(files, fileInfo)
					totalSize += fileInfo.Size
					if fileInfo.IsDir {
						dirCount++
					} else {
						fileCount++
					}
				}
			}
		}
		
		// 排序
		t.sortFiles(files, sortBy, reverse)
	} else {
		// 单个文件
		fileInfo, err := t.getFileInfo(path, includeDetails)
		if err != nil {
			return nil, core.ErrExecutionFailed(t.Info().Name, err.Error())
		}
		files = []FileInfo{fileInfo}
		totalSize = fileInfo.Size
		fileCount = 1
	}
	
	// 创建结果
	var summary string
	if info.IsDir() {
		summary = fmt.Sprintf("Listed %d files and %d directories (total size: %s)", 
			fileCount, dirCount, formatSize(totalSize))
	} else {
		summary = fmt.Sprintf("File info: %s (size: %s)", path, formatSize(totalSize))
	}
	
	result := core.NewSimpleResult(summary)
	result.WithMetadata("files", files)
	result.WithMetadata("total_files", fileCount)
	result.WithMetadata("total_dirs", dirCount)
	result.WithMetadata("total_size", totalSize)
	result.WithMetadata("path", path)
	
	return result, nil
}

// getFileInfo 获取文件信息
func (t *ListTool) getFileInfo(path string, includeDetails bool) (FileInfo, error) {
	info, err := os.Lstat(path) // 使用 Lstat 以获取符号链接信息
	if err != nil {
		return FileInfo{}, err
	}
	
	fileInfo := FileInfo{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
	}
	
	if includeDetails {
		fileInfo.Size = info.Size()
		fileInfo.Mode = info.Mode().String()
		fileInfo.ModTime = info.ModTime()
		
		// 检查是否为符号链接
		if info.Mode()&os.ModeSymlink != 0 {
			fileInfo.IsSymlink = true
			if target, err := os.Readlink(path); err == nil {
				fileInfo.Target = target
			}
		}
	}
	
	return fileInfo, nil
}

// listRecursive 递归列出目录
func (t *ListTool) listRecursive(ctx context.Context, path string, showHidden, includeDetails bool, depth, maxDepth int) (FileInfo, error) {
	if depth > maxDepth {
		return FileInfo{}, fmt.Errorf("max depth exceeded")
	}
	
	// 检查上下文取消
	select {
	case <-ctx.Done():
		return FileInfo{}, ctx.Err()
	default:
	}
	
	fileInfo, err := t.getFileInfo(path, includeDetails)
	if err != nil {
		return FileInfo{}, err
	}
	
	if fileInfo.IsDir && !fileInfo.IsSymlink {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fileInfo, nil // 返回目录信息但不包含内容
		}
		
		fileInfo.Children = make([]FileInfo, 0)
		
		for _, entry := range entries {
			// 检查是否显示隐藏文件
			if !showHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			
			childPath := filepath.Join(path, entry.Name())
			if entry.IsDir() {
				// 递归处理子目录
				childInfo, err := t.listRecursive(ctx, childPath, showHidden, includeDetails, depth+1, maxDepth)
				if err == nil {
					fileInfo.Children = append(fileInfo.Children, childInfo)
				}
			} else {
				// 添加文件
				childInfo, err := t.getFileInfo(childPath, includeDetails)
				if err == nil {
					fileInfo.Children = append(fileInfo.Children, childInfo)
				}
			}
		}
	}
	
	return fileInfo, nil
}

// sortFiles 排序文件列表
func (t *ListTool) sortFiles(files []FileInfo, sortBy string, reverse bool) {
	sort.Slice(files, func(i, j int) bool {
		var result bool
		
		switch sortBy {
		case "size":
			result = files[i].Size < files[j].Size
		case "time":
			result = files[i].ModTime.Before(files[j].ModTime)
		default: // name
			result = files[i].Name < files[j].Name
		}
		
		if reverse {
			result = !result
		}
		
		return result
	})
	
	// 递归排序子目录
	for i := range files {
		if len(files[i].Children) > 0 {
			t.sortFiles(files[i].Children, sortBy, reverse)
		}
	}
}

// countStats 统计文件信息
func (t *ListTool) countStats(info FileInfo, totalSize *int64, fileCount, dirCount *int) {
	if info.IsDir {
		*dirCount++
		for _, child := range info.Children {
			t.countStats(child, totalSize, fileCount, dirCount)
		}
	} else {
		*fileCount++
		*totalSize += info.Size
	}
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	
	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}