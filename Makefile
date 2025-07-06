.PHONY: test test-verbose test-coverage test-coverage-html test-race clean build run help

# 默认目标
default: help

# 帮助信息
help:
	@echo "可用的命令："
	@echo "  make test              - 运行所有测试"
	@echo "  make test-verbose      - 运行测试并显示详细输出"
	@echo "  make test-coverage     - 运行测试并生成覆盖率报告"
	@echo "  make test-coverage-html- 生成 HTML 格式的覆盖率报告并打开"
	@echo "  make test-race         - 运行测试并检测数据竞争"
	@echo "  make build             - 构建项目"
	@echo "  make run               - 运行项目"
	@echo "  make clean             - 清理构建文件和测试缓存"

# 运行测试
test:
	@echo "运行测试..."
	@go test ./... -short

# 运行测试（详细输出）
test-verbose:
	@echo "运行测试（详细输出）..."
	@go test ./... -v

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo ""
	@echo "总体覆盖率："
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'

# 生成 HTML 格式的覆盖率报告
test-coverage-html: test-coverage
	@echo "生成 HTML 覆盖率报告..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "打开覆盖率报告..."
	@open coverage.html || xdg-open coverage.html || echo "请手动打开 coverage.html"

# 运行数据竞争检测
test-race:
	@echo "运行测试并检测数据竞争..."
	@go test ./... -race

# 构建项目
build:
	@echo "构建项目..."
	@go build -o opencode_nano

# 运行项目
run:
	@echo "运行项目..."
	@go run main.go

# 清理
clean:
	@echo "清理构建文件和测试缓存..."
	@rm -f opencode_nano
	@rm -f coverage.out coverage.html
	@go clean -testcache

# 检查代码格式
fmt:
	@echo "检查代码格式..."
	@gofmt -s -w .

# 运行 linter
lint:
	@echo "运行 linter..."
	@golangci-lint run || echo "提示：使用 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest' 安装 golangci-lint"

# 运行所有检查
check: fmt test lint

# 基准测试
bench:
	@echo "运行基准测试..."
	@go test -bench=. -benchmem ./...