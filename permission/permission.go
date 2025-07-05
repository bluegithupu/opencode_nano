package permission

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Manager 权限管理器接口
type Manager interface {
	Request(action, description string) bool
}

// InteractiveManager 交互式权限管理器
type InteractiveManager struct{}

func New() Manager {
	return &InteractiveManager{}
}

// Request 请求执行权限，返回是否允许
func (m *InteractiveManager) Request(action, description string) bool {
	fmt.Printf("\n🔐 需要权限:\n")
	fmt.Printf("操作: %s\n", action)
	fmt.Printf("描述: %s\n", description)
	fmt.Printf("是否允许? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// AutoManager 自动批准权限管理器
type AutoManager struct{}

func NewAuto() Manager {
	return &AutoManager{}
}

// Request 自动批准所有请求
func (m *AutoManager) Request(action, description string) bool {
	fmt.Printf("✅ 自动批准: %s - %s\n", action, description)
	return true
}