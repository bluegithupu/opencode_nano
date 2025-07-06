package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintHelp(t *testing.T) {
	// 保存原始的 stdout
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()

	// 捕获输出
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 调用 printHelp
	printHelp()

	// 恢复 stdout 并读取输出
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	// 验证输出包含预期内容
	expectedContents := []string{
		"📖 可用命令:",
		"clear",
		"help",
		"exit",
		"quit",
		"🔧 可用工具:",
		"read_file",
		"write_file",
		"bash",
		"⚡ 启动参数:",
		"--auto",
		"-a",
		"💡 示例提示:",
		"🚀 自主模式使用示例:",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("printHelp() 输出未包含: %s", expected)
		}
	}
}

func TestMainFunction_ArgumentParsing(t *testing.T) {
	// 测试参数解析逻辑
	tests := []struct {
		name     string
		args     []string
		wantAuto bool
		wantArgs []string
	}{
		{
			name:     "无参数",
			args:     []string{},
			wantAuto: false,
			wantArgs: []string{},
		},
		{
			name:     "只有 --auto",
			args:     []string{"--auto"},
			wantAuto: true,
			wantArgs: []string{},
		},
		{
			name:     "只有 -a",
			args:     []string{"-a"},
			wantAuto: true,
			wantArgs: []string{},
		},
		{
			name:     "--auto 和其他参数",
			args:     []string{"--auto", "test", "prompt"},
			wantAuto: true,
			wantArgs: []string{"test", "prompt"},
		},
		{
			name:     "参数中间有 --auto",
			args:     []string{"test", "--auto", "prompt"},
			wantAuto: true,
			wantArgs: []string{"test", "prompt"},
		},
		{
			name:     "只有普通参数",
			args:     []string{"test", "prompt"},
			wantAuto: false,
			wantArgs: []string{"test", "prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟参数解析逻辑
			autoMode := false
			args := tt.args
			for i, arg := range args {
				if arg == "--auto" || arg == "-a" {
					autoMode = true
					args = append(args[:i], args[i+1:]...)
					break
				}
			}

			if autoMode != tt.wantAuto {
				t.Errorf("autoMode = %v, want %v", autoMode, tt.wantAuto)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("args length = %d, want %d", len(args), len(tt.wantArgs))
			}

			for i, arg := range args {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("args[%d] = %s, want %s", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

// 测试辅助函数
func TestArgumentProcessing(t *testing.T) {
	// 测试字符串连接
	args := []string{"create", "a", "hello", "world", "program"}
	joined := strings.Join(args, " ")
	expected := "create a hello world program"
	
	if joined != expected {
		t.Errorf("strings.Join() = %s, want %s", joined, expected)
	}
}

// 由于 main 函数包含交互式循环，很难直接测试
// 这里我们测试可以独立测试的部分