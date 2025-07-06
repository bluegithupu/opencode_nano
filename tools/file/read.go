package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"opencode_nano/tools/core"
)

// ReadTool 增强版文件读取工具
type ReadTool struct {
	*core.BaseTool
}

// NewReadTool 创建读取工具
func NewReadTool() *ReadTool {
	tool := &ReadTool{
		BaseTool: core.NewBaseTool("read", "file", "Read file contents with advanced options"),
	}
	
	tool.SetTags("file", "read", "content")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "File path to read",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8)",
				Default:     "utf-8",
			},
			"start_line": {
				Type:        "integer",
				Description: "Start line number (1-based, optional)",
				Default:     0,
			},
			"end_line": {
				Type:        "integer",
				Description: "End line number (1-based, optional)",
				Default:     0,
			},
			"max_size": {
				Type:        "integer",
				Description: "Maximum file size in bytes (default: 10MB)",
				Default:     10 * 1024 * 1024,
			},
		},
		Required: []string{"path"},
	})
	
	return tool
}

// Execute 执行读取操作
func (t *ReadTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	filePath, err := params.GetString("path")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid path parameter")
	}
	
	// 规范化路径
	filePath = filepath.Clean(filePath)
	
	// 获取可选参数
	startLine := 0
	if params.Has("start_line") {
		startLine, _ = params.GetInt("start_line")
	}
	
	endLine := 0
	if params.Has("end_line") {
		endLine, _ = params.GetInt("end_line")
	}
	
	maxSize := 10 * 1024 * 1024 // 默认 10MB
	if params.Has("max_size") {
		maxSize, _ = params.GetInt("max_size")
	}
	
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("file not found: %s", filePath))
		}
		return nil, core.ErrExecutionFailed(t.Info().Name, err.Error())
	}
	
	// 检查是否为目录
	if fileInfo.IsDir() {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("path is a directory: %s", filePath))
	}
	
	// 检查文件大小
	if fileInfo.Size() > int64(maxSize) {
		return nil, core.ErrExecutionFailed(t.Info().Name, 
			fmt.Sprintf("file too large: %d bytes (max: %d bytes)", fileInfo.Size(), maxSize))
	}
	
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to open file: %v", err))
	}
	defer file.Close()
	
	// 读取文件内容
	var content string
	var lineCount int
	
	if startLine > 0 || endLine > 0 {
		// 按行读取
		content, lineCount, err = t.readLines(file, startLine, endLine)
		if err != nil {
			return nil, core.ErrExecutionFailed(t.Info().Name, err.Error())
		}
	} else {
		// 读取全部内容
		bytes, err := io.ReadAll(file)
		if err != nil {
			return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to read file: %v", err))
		}
		content = string(bytes)
		lineCount = strings.Count(content, "\n") + 1
	}
	
	// 创建结果
	result := core.NewSimpleResult(content)
	result.WithMetadata("path", filePath)
	result.WithMetadata("size", fileInfo.Size())
	result.WithMetadata("lines", lineCount)
	result.WithMetadata("mode", fileInfo.Mode().String())
	
	if startLine > 0 || endLine > 0 {
		result.WithMetadata("start_line", startLine)
		result.WithMetadata("end_line", endLine)
	}
	
	return result, nil
}

// readLines 按行读取文件
func (t *ReadTool) readLines(file *os.File, startLine, endLine int) (string, int, error) {
	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 0
	totalLines := 0
	
	for scanner.Scan() {
		currentLine++
		totalLines++
		
		// 如果指定了起始行，跳过之前的行
		if startLine > 0 && currentLine < startLine {
			continue
		}
		
		// 如果指定了结束行，超过后停止
		if endLine > 0 && currentLine > endLine {
			break
		}
		
		// 在范围内，添加行
		if startLine == 0 || currentLine >= startLine {
			lines = append(lines, scanner.Text())
		}
	}
	
	if err := scanner.Err(); err != nil {
		return "", totalLines, fmt.Errorf("error reading file: %v", err)
	}
	
	return strings.Join(lines, "\n"), totalLines, nil
}

// ReadBinaryTool 二进制文件读取工具
type ReadBinaryTool struct {
	*core.BaseTool
}

// NewReadBinaryTool 创建二进制读取工具
func NewReadBinaryTool() *ReadBinaryTool {
	tool := &ReadBinaryTool{
		BaseTool: core.NewBaseTool("read_binary", "file", "Read binary file contents"),
	}
	
	tool.SetTags("file", "read", "binary")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "File path to read",
			},
			"offset": {
				Type:        "integer",
				Description: "Start offset in bytes",
				Default:     0,
			},
			"length": {
				Type:        "integer",
				Description: "Number of bytes to read (0 for all)",
				Default:     0,
			},
			"encoding": {
				Type:        "string",
				Description: "Output encoding: hex, base64, raw",
				Default:     "hex",
				Enum:        []string{"hex", "base64", "raw"},
			},
		},
		Required: []string{"path"},
	})
	
	return tool
}

// Execute 执行二进制读取
func (t *ReadBinaryTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 这里简化实现，实际应该实现完整的二进制读取逻辑
	return core.NewSimpleResult("binary read not implemented yet"), nil
}