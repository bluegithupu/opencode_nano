package permission

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Manager struct{}

func New() *Manager {
	return &Manager{}
}

// Request 请求执行权限，返回是否允许
func (m *Manager) Request(action, description string) bool {
	fmt.Printf("\n🔐 Permission required:\n")
	fmt.Printf("Action: %s\n", action)
	fmt.Printf("Description: %s\n", description)
	fmt.Printf("Allow? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}