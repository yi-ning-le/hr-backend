# Makefile for HR Backend

.PHONY: setup lint test generate run build

# 核心设置：安装并初始化 Lefthook
setup:
	@echo "Installing tools and setting up git hooks..."
	go install github.com/evilmartians/lefthook@latest
	go install github.com/air-verse/air@latest
	lefthook install

# 运行静态检查
lint:
	golangci-lint run ./...

# 运行测试
test:
	go test -v ./...

# 生成 SQLC 代码
generate:
	sqlc generate

# 运行开发服务器
run:
	go run main.go

# 使用 Air 运行支持热重载的开发服务器
dev:
	air -c .air.toml

# 编译项目
build:
	go build -o server main.go
