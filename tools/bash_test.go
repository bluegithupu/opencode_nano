package tools

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestBashTool_Name(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewBashTool(perm)
	if got := tool.Name(); got != "bash" {
		t.Errorf("Name() = %v, want %v", got, "bash")
	}
}

func TestBashTool_Description(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewBashTool(perm)
	if got := tool.Description(); got != "Execute bash commands. Use with caution." {
		t.Errorf("Description() = %v, want %v", got, "Execute bash commands. Use with caution.")
	}
}

func TestBashTool_Parameters(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewBashTool(perm)
	params := tool.Parameters()
	
	// 检查参数结构
	if params["type"] != "object" {
		t.Errorf("Parameters type = %v, want object", params["type"])
	}
	
	// 检查 required 字段
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("Parameters required 字段类型错误")
	}
	if len(required) != 1 || required[0] != "command" {
		t.Errorf("Parameters required = %v, want [command]", required)
	}
	
	// 检查 properties
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters properties 字段类型错误")
	}
	
	// 检查 command 属性
	command, ok := props["command"].(map[string]any)
	if !ok {
		t.Fatal("command 属性类型错误")
	}
	if command["type"] != "string" {
		t.Errorf("command type = %v, want string", command["type"])
	}
}

func TestBashTool_Execute(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]any
		allowPerm   bool
		wantErr     bool
		wantPerm    bool
		checkOutput bool
	}{
		{
			name: "成功执行命令 - echo",
			params: map[string]any{
				"command": "echo 'Hello, World!'",
			},
			allowPerm:   true,
			wantErr:     false,
			wantPerm:    true,
			checkOutput: true,
		},
		{
			name: "成功执行命令 - pwd",
			params: map[string]any{
				"command": "pwd",
			},
			allowPerm:   true,
			wantErr:     false,
			wantPerm:    true,
			checkOutput: true,
		},
		{
			name: "权限被拒绝",
			params: map[string]any{
				"command": "ls",
			},
			allowPerm:   false,
			wantErr:     true,
			wantPerm:    true,
			checkOutput: false,
		},
		{
			name: "危险命令 - rm -rf",
			params: map[string]any{
				"command": "rm -rf /tmp/test",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "危险命令 - sudo",
			params: map[string]any{
				"command": "sudo ls",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "危险命令 - curl",
			params: map[string]any{
				"command": "curl http://example.com",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "危险命令 - wget",
			params: map[string]any{
				"command": "wget http://example.com",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "缺少 command 参数",
			params:      map[string]any{},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "command 参数类型错误",
			params: map[string]any{
				"command": 123,
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    false,
			checkOutput: false,
		},
		{
			name: "命令执行失败",
			params: map[string]any{
				"command": "false",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    true,
			checkOutput: false,
		},
		{
			name: "命令不存在",
			params: map[string]any{
				"command": "nonexistentcommand123456",
			},
			allowPerm:   true,
			wantErr:     true,
			wantPerm:    true,
			checkOutput: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 跳过在 Windows 上不兼容的测试
			if runtime.GOOS == "windows" && strings.Contains(tt.name, "pwd") {
				t.Skip("跳过 Windows 上的 pwd 测试")
			}
			
			perm := &MockPermissionManager{shouldAllow: tt.allowPerm}
			tool := NewBashTool(perm)
			
			got, err := tool.Execute(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			// 检查是否请求了权限
			if tt.wantPerm && len(perm.requests) == 0 {
				t.Errorf("Execute() 未请求权限")
			}
			if tt.wantPerm && len(perm.requests) > 0 {
				req := perm.requests[0]
				if req.action != "bash" {
					t.Errorf("权限请求 action = %v, want bash", req.action)
				}
			}
			
			// 检查输出
			if tt.checkOutput && !tt.wantErr {
				if tt.params["command"] == "echo 'Hello, World!'" {
					if !strings.Contains(got, "Hello, World!") {
						t.Errorf("Execute() 输出未包含预期内容")
					}
				}
				if !strings.Contains(got, "Command executed successfully") {
					t.Errorf("Execute() 未返回成功消息")
				}
			}
			
			// 检查危险命令错误消息
			if tt.wantErr && !tt.wantPerm && err != nil {
				// 只有当确实是危险命令时才检查错误消息
				if strings.Contains(tt.name, "危险命令") && !strings.Contains(err.Error(), "dangerous operations") {
					t.Errorf("Execute() 错误消息未包含危险操作提示: %v", err)
				}
			}
		})
	}
}

func TestBashTool_IsDangerous(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"echo hello", false},
		{"ls -la", false},
		{"pwd", false},
		{"rm -rf /", true},
		{"rm -rf .", true},
		{"sudo apt-get update", true},
		{"curl http://example.com", true},
		{"wget http://example.com", true},
		{"ssh user@host", true},
		{"scp file user@host:", true},
		{"chmod 777 /", true},
		{"chown -R user /", true},
		{"> /dev/null", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{":(){ :|: & };:", true},
		{"mkfs.ext4 /dev/sda", true},
		{"mv /* /tmp", true},
		{"find / -delete", true},
		{"echo safe > file.txt", false},
		{"cat file.txt", false},
		{"grep pattern file.txt", false},
	}
	
	perm := &MockPermissionManager{}
	tool := &BashTool{perm: perm}
	
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := tool.isDangerous(tt.command); got != tt.want {
				t.Errorf("isDangerous(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

func TestBashTool_MultilineCommand(t *testing.T) {
	perm := &MockPermissionManager{shouldAllow: true}
	tool := NewBashTool(perm)
	
	// 测试多行命令
	multilineCmd := `echo "Line 1"
echo "Line 2"
echo "Line 3"`
	
	got, err := tool.Execute(map[string]any{
		"command": multilineCmd,
	})
	
	if err != nil {
		t.Errorf("Execute() 多行命令失败: %v", err)
	}
	
	// 检查输出包含所有行
	for i := 1; i <= 3; i++ {
		expected := fmt.Sprintf("Line %d", i)
		if !strings.Contains(got, expected) {
			t.Errorf("Execute() 输出未包含 %q", expected)
		}
	}
}

func TestNewBashTool(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewBashTool(perm)
	if tool == nil {
		t.Fatal("NewBashTool() 返回 nil")
	}
	
	// 验证是否实现了 Tool 接口
	var _ Tool = tool
}