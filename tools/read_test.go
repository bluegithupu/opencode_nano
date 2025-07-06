package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTool_Name(t *testing.T) {
	tool := NewReadTool()
	if got := tool.Name(); got != "read_file" {
		t.Errorf("Name() = %v, want %v", got, "read_file")
	}
}

func TestReadTool_Description(t *testing.T) {
	tool := NewReadTool()
	if got := tool.Description(); got != "Read the contents of a file" {
		t.Errorf("Description() = %v, want %v", got, "Read the contents of a file")
	}
}

func TestReadTool_Parameters(t *testing.T) {
	tool := NewReadTool()
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
	if len(required) != 1 || required[0] != "file_path" {
		t.Errorf("Parameters required = %v, want [file_path]", required)
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
}

func TestReadTool_Execute(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!\nThis is a test file."
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name    string
		params  map[string]any
		want    string
		wantErr bool
	}{
		{
			name: "成功读取文件",
			params: map[string]any{
				"file_path": testFile,
			},
			want:    testContent,
			wantErr: false,
		},
		{
			name: "文件不存在",
			params: map[string]any{
				"file_path": filepath.Join(tmpDir, "nonexistent.txt"),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "缺少 file_path 参数",
			params: map[string]any{},
			want:   "",
			wantErr: true,
		},
		{
			name: "file_path 参数类型错误",
			params: map[string]any{
				"file_path": 123,
			},
			want:    "",
			wantErr: true,
		},
	}
	
	tool := NewReadTool()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tool.Execute(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// 检查返回的内容是否包含预期的文件内容
				if !strings.Contains(got, tt.want) {
					t.Errorf("Execute() 未包含预期内容 = %v", tt.want)
				}
				// 检查是否包含文件路径
				if !strings.Contains(got, "File content of") {
					t.Errorf("Execute() 未包含文件路径信息")
				}
			}
		})
	}
}

func TestReadTool_ExecuteEmptyFile(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "read_test_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// 创建空文件
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	
	tool := NewReadTool()
	got, err := tool.Execute(map[string]any{"file_path": emptyFile})
	
	if err != nil {
		t.Errorf("Execute() 读取空文件失败: %v", err)
	}
	
	if !strings.Contains(got, "File content of") {
		t.Errorf("Execute() 未包含文件路径信息")
	}
}

func TestNewReadTool(t *testing.T) {
	tool := NewReadTool()
	if tool == nil {
		t.Fatal("NewReadTool() 返回 nil")
	}
	
	// 验证是否实现了 Tool 接口
	var _ Tool = tool
}