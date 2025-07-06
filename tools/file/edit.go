package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"opencode_nano/tools/core"
)

// EditTool 文件编辑工具
type EditTool struct {
	*core.BaseTool
}

// NewEditTool 创建编辑工具
func NewEditTool() *EditTool {
	tool := &EditTool{
		BaseTool: core.NewBaseTool("edit", "file", "Edit file contents with find and replace"),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("file", "edit", "modify", "replace")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "File path to edit",
			},
			"operations": {
				Type:        "array",
				Description: "List of edit operations to perform",
			},
		},
		Required: []string{"path", "operations"},
	})
	
	return tool
}

// EditOperation 编辑操作
type EditOperation struct {
	Type        string `json:"type"`        // replace, regex_replace, insert, delete
	Find        string `json:"find"`        // 查找内容
	Replace     string `json:"replace"`     // 替换内容
	Line        int    `json:"line"`        // 行号（用于 insert/delete）
	All         bool   `json:"all"`         // 是否替换所有匹配
	CaseSensitive bool `json:"case_sensitive"` // 是否区分大小写
}

// Execute 执行编辑操作
func (t *EditTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
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
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("file not found: %s", filePath))
		}
		return nil, core.ErrExecutionFailed(t.Info().Name, err.Error())
	}
	
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to read file: %v", err))
	}
	
	// 将内容转换为行
	lines := strings.Split(string(content), "\n")
	originalLineCount := len(lines)
	
	// 获取操作列表
	operationsRaw, err := params.Get("operations")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid operations parameter")
	}

	// 解析操作列表
	operations, err := t.parseOperations(operationsRaw)
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("invalid operations: %v", err))
	}

	// 执行编辑操作
	editCount := 0
	for _, op := range operations {
		switch op.Type {
		case "replace", "regex_replace":
			var count int
			newContent := strings.Join(lines, "\n")
			if op.Type == "regex_replace" {
				newContent, count = regexReplace(newContent, op.Find, op.Replace, op.All, op.CaseSensitive)
			} else {
				newContent, count = findAndReplace(newContent, op.Find, op.Replace, op.All, op.CaseSensitive)
			}
			lines = strings.Split(newContent, "\n")
			editCount += count
		
		case "insert":
			if op.Line > 0 && op.Line <= len(lines)+1 {
				lines = insertLine(lines, op.Line, op.Replace)
				editCount++
			}
		
		case "delete":
			if op.Line > 0 && op.Line <= len(lines) {
				lines = deleteLine(lines, op.Line)
				editCount++
			}
		
		default:
			return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("unknown operation type: %s", op.Type))
		}
	}
	
	// 写回文件
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to write file: %v", err))
	}
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Successfully edited %s", filePath))
	result.WithMetadata("path", filePath)
	result.WithMetadata("edits", editCount)
	result.WithMetadata("original_lines", originalLineCount)
	result.WithMetadata("new_lines", len(lines))
	result.WithMetadata("operations", operationsRaw)
	
	return result, nil
}

// MultiEditTool 多文件编辑工具
type MultiEditTool struct {
	*core.BaseTool
	editTool *EditTool
}

// NewMultiEditTool 创建多文件编辑工具
func NewMultiEditTool() *MultiEditTool {
	tool := &MultiEditTool{
		BaseTool: core.NewBaseTool("multi_edit", "file", "Edit multiple files in one operation"),
		editTool: NewEditTool(),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("file", "edit", "batch", "multiple")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"edits": {
				Type:        "array",
				Description: "List of file edits to perform",
			},
		},
		Required: []string{"edits"},
	})
	
	return tool
}

// Execute 执行多文件编辑
func (t *MultiEditTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取编辑列表
	editsRaw, err := params.Get("edits")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid edits parameter")
	}
	
	// 解析编辑列表
	edits, err := t.parseEdits(editsRaw)
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, fmt.Sprintf("invalid edits: %v", err))
	}
	
	// 执行所有编辑
	results := make([]map[string]interface{}, 0, len(edits))
	successCount := 0
	failCount := 0
	
	for _, edit := range edits {
		// 为每个文件创建参数
		editParams := core.NewMapParameters(map[string]any{
			"path":       edit.Path,
			"operations": edit.Operations,
		})
		
		// 执行编辑
		result, err := t.editTool.Execute(ctx, editParams)
		if err != nil {
			failCount++
			results = append(results, map[string]interface{}{
				"path":  edit.Path,
				"error": err.Error(),
			})
		} else {
			successCount++
			results = append(results, map[string]interface{}{
				"path":     edit.Path,
				"success":  true,
				"metadata": result.Metadata(),
			})
		}
	}
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Edited %d files successfully, %d failed", successCount, failCount))
	result.WithMetadata("success_count", successCount)
	result.WithMetadata("fail_count", failCount)
	result.WithMetadata("results", results)
	
	return result, nil
}

// FileEdit 文件编辑信息
type FileEdit struct {
	Path       string        `json:"path"`
	Operations []interface{} `json:"operations"`
}

// parseEdits 解析编辑列表
func (t *MultiEditTool) parseEdits(raw interface{}) ([]FileEdit, error) {
	var edits []FileEdit
	
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			editMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid edit format")
			}
			
			path, ok := editMap["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("edit must have a path")
			}
			
			operations, ok := editMap["operations"].([]interface{})
			if !ok {
				return nil, fmt.Errorf("edit must have operations")
			}
			
			edits = append(edits, FileEdit{
				Path:       path,
				Operations: operations,
			})
		}
	default:
		return nil, fmt.Errorf("edits must be an array")
	}
	
	return edits, nil
}

// PatchTool 补丁应用工具
type PatchTool struct {
	*core.BaseTool
}

// NewPatchTool 创建补丁工具
func NewPatchTool() *PatchTool {
	tool := &PatchTool{
		BaseTool: core.NewBaseTool("patch", "file", "Apply unified diff patches to files"),
	}
	
	tool.SetRequiresPerm(true)
	tool.SetTags("file", "edit", "patch", "diff")
	tool.SetSchema(core.ParameterSchema{
		Type: "object",
		Properties: map[string]core.PropertySchema{
			"path": {
				Type:        "string",
				Description: "File path to patch",
			},
			"patch": {
				Type:        "string",
				Description: "Unified diff patch content",
			},
			"reverse": {
				Type:        "boolean",
				Description: "Apply patch in reverse",
				Default:     false,
			},
		},
		Required: []string{"path", "patch"},
	})
	
	return tool
}

// Execute 应用补丁
func (t *PatchTool) Execute(ctx context.Context, params core.Parameters) (core.Result, error) {
	// 参数验证
	if err := params.Validate(t.Schema()); err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, err.Error())
	}
	
	// 获取参数
	filePath, err := params.GetString("path")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid path parameter")
	}
	
	patchContent, err := params.GetString("patch")
	if err != nil {
		return nil, core.ErrInvalidParams(t.Info().Name, "invalid patch parameter")
	}
	
	reverse := false
	if params.Has("reverse") {
		reverse, _ = params.GetBool("reverse")
	}
	
	// 规范化路径
	filePath = filepath.Clean(filePath)
	
	// 读取原文件
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to read file: %v", err))
	}
	
	// 应用补丁（简化实现）
	// 在实际实现中，应该使用专门的 diff/patch 库
	newContent, applied, err := t.applySimplePatch(string(originalContent), patchContent, reverse)
	if err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to apply patch: %v", err))
	}
	
	// 写回文件
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, core.ErrExecutionFailed(t.Info().Name, fmt.Sprintf("failed to write file: %v", err))
	}
	
	// 创建结果
	result := core.NewSimpleResult(fmt.Sprintf("Successfully applied patch to %s", filePath))
	result.WithMetadata("path", filePath)
	result.WithMetadata("hunks_applied", applied)
	result.WithMetadata("reverse", reverse)
	
	return result, nil
}

// applySimplePatch 简单的补丁应用（仅用于演示）
func (t *PatchTool) applySimplePatch(content, patch string, reverse bool) (string, int, error) {
	// 这是一个极简的实现，仅支持简单的行替换
	// 实际应该使用 github.com/sourcegraph/go-diff 或类似库
	
	lines := strings.Split(content, "\n")
	patchLines := strings.Split(patch, "\n")
	applied := 0
	
	for i := 0; i < len(patchLines); i++ {
		line := patchLines[i]
		
		// 简单查找以 - 开头的行并替换为 + 开头的行
		if strings.HasPrefix(line, "-") && i+1 < len(patchLines) && strings.HasPrefix(patchLines[i+1], "+") {
			oldLine := strings.TrimPrefix(line, "-")
			newLine := strings.TrimPrefix(patchLines[i+1], "+")
			
			if reverse {
				oldLine, newLine = newLine, oldLine
			}
			
			// 在内容中查找并替换
			for j, contentLine := range lines {
				if strings.TrimSpace(contentLine) == strings.TrimSpace(oldLine) {
					lines[j] = newLine
					applied++
					break
				}
			}
			
			i++ // 跳过下一行
		}
	}
	
	return strings.Join(lines, "\n"), applied, nil
}

// findAndReplace 执行查找替换
func findAndReplace(content, find, replace string, all, caseSensitive bool) (string, int) {
	count := 0
	
	if !caseSensitive {
		// 不区分大小写的替换
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(find))
		if all {
			content = re.ReplaceAllStringFunc(content, func(match string) string {
				count++
				return replace
			})
		} else {
			content = re.ReplaceAllStringFunc(content, func(match string) string {
				if count == 0 {
					count++
					return replace
				}
				return match
			})
		}
	} else {
		// 区分大小写的替换
		if all {
			newContent := strings.ReplaceAll(content, find, replace)
			count = strings.Count(content, find)
			content = newContent
		} else {
			index := strings.Index(content, find)
			if index != -1 {
				content = content[:index] + replace + content[index+len(find):]
				count = 1
			}
		}
	}
	
	return content, count
}

// regexReplace 执行正则表达式替换
func regexReplace(content, pattern, replace string, all, caseSensitive bool) (string, int) {
	count := 0
	
	// 构建正则表达式
	flags := ""
	if !caseSensitive {
		flags = "(?i)"
	}
	
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		// 如果正则表达式无效，返回原内容
		return content, 0
	}
	
	if all {
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			count++
			return replace
		})
	} else {
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			if count == 0 {
				count++
				return replace
			}
			return match
		})
	}
	
	return content, count
}

// parseOperations 解析操作列表
func (t *EditTool) parseOperations(raw interface{}) ([]EditOperation, error) {
	// 尝试转换为切片
	var operations []EditOperation
	
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			opMap, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid operation format")
			}
			
			op := EditOperation{
				Type:          getStringValue(opMap, "type", "replace"),
				Find:          getStringValue(opMap, "find", ""),
				Replace:       getStringValue(opMap, "replace", ""),
				Line:          getIntValue(opMap, "line", 0),
				All:           getBoolValue(opMap, "all", true),
				CaseSensitive: getBoolValue(opMap, "case_sensitive", true),
			}
			
			// 验证操作
			if err := t.validateOperation(op); err != nil {
				return nil, err
			}
			
			operations = append(operations, op)
		}
	default:
		return nil, fmt.Errorf("operations must be an array")
	}
	
	return operations, nil
}

// validateOperation 验证操作
func (t *EditTool) validateOperation(op EditOperation) error {
	switch op.Type {
	case "replace", "regex_replace":
		if op.Find == "" {
			return fmt.Errorf("%s operation requires 'find' field", op.Type)
		}
	case "insert":
		if op.Line <= 0 {
			return fmt.Errorf("insert operation requires positive 'line' field")
		}
	case "delete":
		if op.Line <= 0 {
			return fmt.Errorf("delete operation requires positive 'line' field")
		}
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
	return nil
}

// Helper functions for parsing
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getIntValue(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return defaultValue
}

func getBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// insertLine 在指定行插入内容
func insertLine(lines []string, lineNum int, content string) []string {
	if lineNum <= 0 || lineNum > len(lines)+1 {
		return lines
	}
	
	// 转换为 0 基索引
	index := lineNum - 1
	
	// 创建新切片
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:index]...)
	result = append(result, content)
	result = append(result, lines[index:]...)
	
	return result
}

// deleteLine 删除指定行
func deleteLine(lines []string, lineNum int) []string {
	if lineNum <= 0 || lineNum > len(lines) {
		return lines
	}
	
	// 转换为 0 基索引
	index := lineNum - 1
	
	// 创建新切片
	result := make([]string, 0, len(lines)-1)
	result = append(result, lines[:index]...)
	result = append(result, lines[index+1:]...)
	
	return result
}