package tools

import (
	"fmt"
	"os"

	"opencode_nano/permission"
)

type WriteTool struct {
	perm permission.Manager
}

func NewWriteTool(perm permission.Manager) *WriteTool {
	return &WriteTool{perm: perm}
}

func (t *WriteTool) Name() string {
	return "write_file"
}

func (t *WriteTool) Description() string {
	return "Write content to a file"
}

func (t *WriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"file_path": map[string]any{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"file_path", "content"},
	}
}

func (t *WriteTool) Execute(params map[string]any) (string, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path parameter is required and must be a string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required and must be a string")
	}

	// 请求权限
	if !t.perm.Request("write_file", fmt.Sprintf("Write to file: %s", filePath)) {
		return "", fmt.Errorf("permission denied for writing to file: %s", filePath)
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return fmt.Sprintf("Successfully wrote content to file: %s", filePath), nil
}