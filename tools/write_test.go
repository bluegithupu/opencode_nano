package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockPermissionManager 用于测试的模拟权限管理器
type MockPermissionManager struct {
	shouldAllow bool
	requests    []struct {
		action      string
		description string
	}
}

func (m *MockPermissionManager) Request(action, description string) bool {
	m.requests = append(m.requests, struct {
		action      string
		description string
	}{action, description})
	return m.shouldAllow
}

func TestWriteTool_Name(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewWriteTool(perm)
	if got := tool.Name(); got != "write_file" {
		t.Errorf("Name() = %v, want %v", got, "write_file")
	}
}

func TestWriteTool_Description(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewWriteTool(perm)
	if got := tool.Description(); got != "Write content to a file" {
		t.Errorf("Description() = %v, want %v", got, "Write content to a file")
	}
}

func TestWriteTool_Parameters(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewWriteTool(perm)
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
	if len(required) != 2 || required[0] != "file_path" || required[1] != "content" {
		t.Errorf("Parameters required = %v, want [file_path content]", required)
	}
	
	// 检查 properties
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("Parameters properties 字段类型错误")
	}
	
	// 检查 file_path 属性
	filePath, ok := props["file_path"].(map[string]any)
	if !ok {
		t.Fatal("file_path 属性类型错误")
	}
	if filePath["type"] != "string" {
		t.Errorf("file_path type = %v, want string", filePath["type"])
	}
	
	// 检查 content 属性
	content, ok := props["content"].(map[string]any)
	if !ok {
		t.Fatal("content 属性类型错误")
	}
	if content["type"] != "string" {
		t.Errorf("content type = %v, want string", content["type"])
	}
}

func TestWriteTool_Execute(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	
	tests := []struct {
		name        string
		params      map[string]any
		allowPerm   bool
		wantErr     bool
		wantPerm    bool
		checkFile   bool
	}{
		{
			name: "成功写入文件",
			params: map[string]any{
				"file_path": testFile,
				"content":   testContent,
			},
			allowPerm: true,
			wantErr:   false,
			wantPerm:  true,
			checkFile: true,
		},
		{
			name: "权限被拒绝",
			params: map[string]any{
				"file_path": testFile,
				"content":   testContent,
			},
			allowPerm: false,
			wantErr:   true,
			wantPerm:  true,
			checkFile: false,
		},
		{
			name: "缺少 file_path 参数",
			params: map[string]any{
				"content": testContent,
			},
			allowPerm: true,
			wantErr:   true,
			wantPerm:  false,
			checkFile: false,
		},
		{
			name: "缺少 content 参数",
			params: map[string]any{
				"file_path": testFile,
			},
			allowPerm: true,
			wantErr:   true,
			wantPerm:  false,
			checkFile: false,
		},
		{
			name: "file_path 参数类型错误",
			params: map[string]any{
				"file_path": 123,
				"content":   testContent,
			},
			allowPerm: true,
			wantErr:   true,
			wantPerm:  false,
			checkFile: false,
		},
		{
			name: "content 参数类型错误",
			params: map[string]any{
				"file_path": testFile,
				"content":   123,
			},
			allowPerm: true,
			wantErr:   true,
			wantPerm:  false,
			checkFile: false,
		},
		{
			name: "写入空内容",
			params: map[string]any{
				"file_path": filepath.Join(tmpDir, "empty.txt"),
				"content":   "",
			},
			allowPerm: true,
			wantErr:   false,
			wantPerm:  true,
			checkFile: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理之前的测试文件
			os.Remove(testFile)
			
			perm := &MockPermissionManager{shouldAllow: tt.allowPerm}
			tool := NewWriteTool(perm)
			
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
				if req.action != "write_file" {
					t.Errorf("权限请求 action = %v, want write_file", req.action)
				}
			}
			
			// 检查文件是否被创建
			if tt.checkFile {
				filePath := tt.params["file_path"].(string)
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("无法读取创建的文件: %v", err)
				} else {
					expectedContent := tt.params["content"].(string)
					if string(content) != expectedContent {
						t.Errorf("文件内容 = %v, want %v", string(content), expectedContent)
					}
				}
			}
			
			// 检查成功消息
			if !tt.wantErr && !strings.Contains(got, "Successfully wrote content to file") {
				t.Errorf("Execute() 未返回成功消息")
			}
		})
	}
}

func TestWriteTool_WriteToSubdirectory(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "write_test_subdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// 创建子目录
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	testFile := filepath.Join(subDir, "test.txt")
	testContent := "Subdirectory test"
	
	perm := &MockPermissionManager{shouldAllow: true}
	tool := NewWriteTool(perm)
	
	_, err = tool.Execute(map[string]any{
		"file_path": testFile,
		"content":   testContent,
	})
	
	if err != nil {
		t.Errorf("Execute() 写入子目录失败: %v", err)
	}
	
	// 验证文件内容
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("无法读取文件: %v", err)
	} else if string(content) != testContent {
		t.Errorf("文件内容 = %v, want %v", string(content), testContent)
	}
}

func TestNewWriteTool(t *testing.T) {
	perm := &MockPermissionManager{}
	tool := NewWriteTool(perm)
	if tool == nil {
		t.Fatal("NewWriteTool() 返回 nil")
	}
	
	// 验证是否实现了 Tool 接口
	var _ Tool = tool
}