package tools

import (
	"fmt"
	"os/exec"
	"strings"

	"opencode_nano/permission"
)

type BashTool struct {
	perm permission.Manager
}

func NewBashTool(perm permission.Manager) *BashTool {
	return &BashTool{perm: perm}
}

func (t *BashTool) Name() string {
	return "bash"
}

func (t *BashTool) Description() string {
	return "Execute bash commands. Use with caution."
}

func (t *BashTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The bash command to execute",
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Execute(params map[string]any) (string, error) {
	command, ok := params["command"].(string)
	if !ok {
		return "", fmt.Errorf("command parameter is required and must be a string")
	}

	// 简单的安全检查
	if t.isDangerous(command) {
		return "", fmt.Errorf("command contains dangerous operations: %s", command)
	}

	// 请求权限
	if !t.perm.Request("bash", fmt.Sprintf("Execute command: %s", command)) {
		return "", fmt.Errorf("permission denied for command: %s", command)
	}

	// 执行命令
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}

	return fmt.Sprintf("Command executed successfully:\n%s", string(output)), nil
}

func (t *BashTool) isDangerous(command string) bool {
	dangerous := []string{
		"rm -rf",
		"sudo",
		"curl",
		"wget",
		"dd if=",
		"mkfs",
		"fdisk",
		"> /dev/",
	}

	cmdLower := strings.ToLower(command)
	for _, danger := range dangerous {
		if strings.Contains(cmdLower, danger) {
			return true
		}
	}
	return false
}