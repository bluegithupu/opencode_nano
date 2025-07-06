package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		wantErr bool
		check   func(*Config) error
	}{
		{
			name: "成功加载 - 使用环境变量",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "test-api-key")
				os.Setenv("OPENAI_BASE_URL", "https://test.api.com/v1")
			},
			cleanup: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("OPENAI_BASE_URL")
			},
			wantErr: false,
			check: func(cfg *Config) error {
				if cfg.OpenAIAPIKey != "test-api-key" {
					t.Errorf("OpenAIAPIKey = %v, want %v", cfg.OpenAIAPIKey, "test-api-key")
				}
				if cfg.OpenAIBaseURL != "https://test.api.com/v1" {
					t.Errorf("OpenAIBaseURL = %v, want %v", cfg.OpenAIBaseURL, "https://test.api.com/v1")
				}
				return nil
			},
		},
		{
			name: "成功加载 - 使用默认 BaseURL",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "test-api-key-2")
				os.Unsetenv("OPENAI_BASE_URL")
			},
			cleanup: func() {
				os.Unsetenv("OPENAI_API_KEY")
			},
			wantErr: false,
			check: func(cfg *Config) error {
				if cfg.OpenAIAPIKey != "test-api-key-2" {
					t.Errorf("OpenAIAPIKey = %v, want %v", cfg.OpenAIAPIKey, "test-api-key-2")
				}
				if cfg.OpenAIBaseURL != "https://api.openai.com/v1" {
					t.Errorf("OpenAIBaseURL = %v, want %v", cfg.OpenAIBaseURL, "https://api.openai.com/v1")
				}
				return nil
			},
		},
		{
			name: "失败 - 缺少 API Key",
			setup: func() {
				os.Unsetenv("OPENAI_API_KEY")
			},
			cleanup: func() {},
			wantErr: true,
		},
		{
			name: "成功加载 - 空格被去除",
			setup: func() {
				os.Setenv("OPENAI_API_KEY", "  test-api-key-spaces  ")
				os.Setenv("OPENAI_BASE_URL", "  https://test.api.com/v1  ")
			},
			cleanup: func() {
				os.Unsetenv("OPENAI_API_KEY")
				os.Unsetenv("OPENAI_BASE_URL")
			},
			wantErr: false,
			check: func(cfg *Config) error {
				if cfg.OpenAIAPIKey != "test-api-key-spaces" {
					t.Errorf("OpenAIAPIKey = %v, want %v", cfg.OpenAIAPIKey, "test-api-key-spaces")
				}
				if cfg.OpenAIBaseURL != "https://test.api.com/v1" {
					t.Errorf("OpenAIBaseURL = %v, want %v", cfg.OpenAIBaseURL, "https://test.api.com/v1")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置测试环境
			if tt.setup != nil {
				tt.setup()
			}
			
			// 清理测试环境
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			// 执行测试
			got, err := Load()
			
			// 检查错误
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// 如果不期望错误，检查返回值
			if !tt.wantErr {
				if got == nil {
					t.Fatal("Load() 返回 nil 配置")
				}
				
				// 执行自定义检查
				if tt.check != nil {
					if err := tt.check(got); err != nil {
						t.Error(err)
					}
				}
			}
		})
	}
}

func TestConfig_Structure(t *testing.T) {
	// 测试 Config 结构体的字段
	cfg := &Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "test-url",
	}
	
	if cfg.OpenAIAPIKey != "test-key" {
		t.Errorf("OpenAIAPIKey 字段设置失败")
	}
	
	if cfg.OpenAIBaseURL != "test-url" {
		t.Errorf("OpenAIBaseURL 字段设置失败")
	}
}

// 基准测试
func BenchmarkLoad(b *testing.B) {
	// 设置环境变量
	os.Setenv("OPENAI_API_KEY", "benchmark-key")
	defer os.Unsetenv("OPENAI_API_KEY")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Load()
	}
}