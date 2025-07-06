package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"opencode_nano/tools/core"
)

// WriteTool 增强版文件写入工具
type WriteTool struct {
	*core.BaseTool
}

// NewWriteTool 创建写入工具
func NewWriteTool() *WriteTool {
	tool := &WriteTool{
		BaseTool: core.NewBaseTool("write", "file", "Write content to file with advanced options"),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("file", "write", "content")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "File path to write",
			},
			"content": {
				Type:        "string",
				Description: "Content to write",
			},
			"mode": {
				Type:        "string",
				Description: "Write mode: overwrite, append, create",
				Default:     "overwrite",
				Enum:        []string{"overwrite", "append", "create"},
			},
			"create_dirs": {
				Type:        "boolean",
				Description: "Create parent directories if they don't exist",
				Default:     true,
			},
			"backup": {
				Type:        "boolean",
				Description: "Create backup of existing file",
				Default:     false,
			},
			"file_mode": {
				Type:        "string",
				Description: "File permissions (e.g., '0644')",
				Default:     "0644",
			},
		},
		Required: []string{"path", "content"},
	})
	
	return tool
}

// Execute 执行写入操作
func (t *WriteTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	filePath, err := params.GetString("path")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid path parameter")
	}
	
	content, err := params.GetString("content")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid content parameter")
	}
	
	// 规范化路径
	filePath = filepath.Clean(filePath)
	
	// 获取可选参数
	mode := "overwrite"
	if params.Has("mode") {
		mode, _ = params.GetString("mode")
	}
	
	createDirs := true
	if params.Has("create_dirs") {
		createDirs, _ = params.GetBool("create_dirs")
	}
	
	backup := false
	if params.Has("backup") {
		backup, _ = params.GetBool("backup")
	}
	
	// 检查文件是否存在
	fileExists := false
	if fileInfo, err := os.Stat(filePath); err == nil {
		if fileInfo.IsDir() {
			return nil, core.ErrExecutionFailed(t.Info().Name, "path is a directory")
		}
		fileExists = true
	}
	
	// 处理写入模式
	if mode == "create" && fileExists {
		return nil, core.ErrExecutionFailed(t.Info().Name, "file already exists")
	}
	
	// 创建父目录
	if createDirs {
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, core.ErrExecutionFailed(t.Info().Name, 
				fmt.Sprintf("failed to create directories: %v", err))
		}
	}
	
	// 备份现有文件
	if backup && fileExists {
		backupPath := filePath + ".backup"
		if err := t.copyFile(filePath, backupPath); err != nil {
			return nil, core.ErrExecutionFailed(t.Info().Name, 
				fmt.Sprintf("failed to create backup: %v", err))
		}
	}
	
	// 写入文件
	var writeErr error
	switch mode {
	case "append":
		writeErr = t.appendToFile(filePath, content)
	default: // overwrite or create
		writeErr = t.writeFile(filePath, content)
	}
	
	if writeErr != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, writeErr.Error())
	}
	
	// 获取文件信息
	fileInfo, _ := os.Stat(filePath)
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filePath))
	result.WithMetadata("path", filePath)
	result.WithMetadata("size", len(content))
	result.WithMetadata("mode", mode)
	if fileInfo != nil {
		result.WithMetadata("file_size", fileInfo.Size())
	}
	if backup && fileExists {
		result.WithMetadata("backup_path", filePath+".backup")
	}
	
	return result, nil
}

// writeFile 写入文件（覆盖模式）
func (t *WriteTool) writeFile(path, content string) error {
	// 使用原子写入：先写入临时文件，然后重命名
	tempPath := path + ".tmp"
	
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	
	_, err = file.WriteString(content)
	if err != nil {
		file.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write content: %v", err)
	}
	
	if err := file.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close file: %v", err)
	}
	
	// 原子重命名
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename file: %v", err)
	}
	
	return nil
}

// appendToFile 追加到文件
func (t *WriteTool) appendToFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to append content: %v", err)
	}
	
	return nil
}

// copyFile 复制文件
func (t *WriteTool) copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	
	_, err = io.Copy(destination, source)
	return err
}