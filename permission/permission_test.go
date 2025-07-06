package permission

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestInteractiveManager_Request(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		action      string
		description string
		want        bool
	}{
		{
			name:        "允许操作 - 输入 y",
			input:       "y\n",
			action:      "write_file",
			description: "写入文件: test.txt",
			want:        true,
		},
		{
			name:        "允许操作 - 输入 yes",
			input:       "yes\n",
			action:      "bash",
			description: "执行命令: ls",
			want:        true,
		},
		{
			name:        "拒绝操作 - 输入 n",
			input:       "n\n",
			action:      "write_file",
			description: "写入文件: dangerous.txt",
			want:        false,
		},
		{
			name:        "拒绝操作 - 输入空",
			input:       "\n",
			action:      "bash",
			description: "执行命令: rm -rf",
			want:        false,
		},
		{
			name:        "拒绝操作 - 输入其他",
			input:       "maybe\n",
			action:      "write_file",
			description: "写入文件: test.txt",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始的 stdin 和 stdout
			oldStdin := os.Stdin
			oldStdout := os.Stdout
			defer func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
			}()

			// 创建模拟的输入
			r, w, _ := os.Pipe()
			os.Stdin = r
			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			// 捕获输出
			outR, outW, _ := os.Pipe()
			os.Stdout = outW

			// 创建管理器并测试
			m := &InteractiveManager{}
			got := m.Request(tt.action, tt.description)

			// 恢复 stdout 并读取输出
			outW.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, outR)

			// 验证结果
			if got != tt.want {
				t.Errorf("Request() = %v, want %v", got, tt.want)
			}

			// 验证输出包含预期内容
			output := buf.String()
			if !bytes.Contains([]byte(output), []byte("🔐 需要权限:")) {
				t.Errorf("输出未包含权限提示")
			}
			if !bytes.Contains([]byte(output), []byte(tt.action)) {
				t.Errorf("输出未包含操作: %s", tt.action)
			}
			if !bytes.Contains([]byte(output), []byte(tt.description)) {
				t.Errorf("输出未包含描述: %s", tt.description)
			}
		})
	}
}

func TestAutoManager_Request(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		description string
		want        bool
	}{
		{
			name:        "自动批准写文件",
			action:      "write_file",
			description: "写入文件: test.txt",
			want:        true,
		},
		{
			name:        "自动批准执行命令",
			action:      "bash",
			description: "执行命令: ls",
			want:        true,
		},
		{
			name:        "自动批准危险操作",
			action:      "bash",
			description: "执行命令: rm -rf /",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始的 stdout
			oldStdout := os.Stdout
			defer func() {
				os.Stdout = oldStdout
			}()

			// 捕获输出
			r, w, _ := os.Pipe()
			os.Stdout = w

			// 创建自动管理器并测试
			m := &AutoManager{}
			got := m.Request(tt.action, tt.description)

			// 恢复 stdout 并读取输出
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// 验证结果
			if got != tt.want {
				t.Errorf("Request() = %v, want %v", got, tt.want)
			}

			// 验证输出包含预期内容
			output := buf.String()
			if !bytes.Contains([]byte(output), []byte("✅ 自动批准:")) {
				t.Errorf("输出未包含自动批准提示")
			}
			if !bytes.Contains([]byte(output), []byte(tt.action)) {
				t.Errorf("输出未包含操作: %s", tt.action)
			}
			if !bytes.Contains([]byte(output), []byte(tt.description)) {
				t.Errorf("输出未包含描述: %s", tt.description)
			}
		})
	}
}

func TestNew(t *testing.T) {
	manager := New()
	if manager == nil {
		t.Fatal("New() 返回 nil")
	}
	
	// 验证返回的是 InteractiveManager 类型
	if _, ok := manager.(*InteractiveManager); !ok {
		t.Errorf("New() 返回的不是 *InteractiveManager 类型")
	}
}

func TestNewAuto(t *testing.T) {
	manager := NewAuto()
	if manager == nil {
		t.Fatal("NewAuto() 返回 nil")
	}
	
	// 验证返回的是 AutoManager 类型
	if _, ok := manager.(*AutoManager); !ok {
		t.Errorf("NewAuto() 返回的不是 *AutoManager 类型")
	}
}