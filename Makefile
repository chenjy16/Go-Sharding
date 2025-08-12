# Go-Sharding Makefile
# 基于 Apache ShardingSphere 设计的 Go 语言分片数据库中间件

.PHONY: help build test clean run demo deps fmt lint vet check init-db

# 默认目标
help:
	@echo "Go-Sharding 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  build     - 构建项目"
	@echo "  test      - 运行测试"
	@echo "  clean     - 清理构建文件"
	@echo "  run       - 运行基础示例"
	@echo "  demo      - 运行演示程序"
	@echo "  deps      - 安装依赖"
	@echo "  fmt       - 格式化代码"
	@echo "  lint      - 代码检查"
	@echo "  vet       - 静态分析"
	@echo "  check     - 运行所有检查"
	@echo "  init-db   - 初始化数据库"

# 构建项目
build:
	@echo "构建 Go-Sharding..."
	go build -o bin/go-sharding-demo cmd/demo/main.go
	go build -o bin/go-sharding-basic examples/basic/main.go
	go build -o bin/go-sharding-yaml examples/yaml_config/main.go
	@echo "构建完成！"

# 运行测试
test:
	@echo "运行测试..."
	go test -v ./...
	@echo "测试完成！"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean
	@echo "清理完成！"

# 运行基础示例
run:
	@echo "运行基础示例..."
	go run examples/basic/main.go

# 运行演示程序
demo:
	@echo "运行演示程序..."
	go run cmd/demo/main.go

# 运行 YAML 配置示例
yaml:
	@echo "运行 YAML 配置示例..."
	go run examples/yaml_config/main.go

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod tidy
	go mod download
	@echo "依赖安装完成！"

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	@echo "代码格式化完成！"

# 代码检查 (需要安装 golangci-lint)
lint:
	@echo "运行代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过 lint 检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 静态分析
vet:
	@echo "运行静态分析..."
	go vet ./...
	@echo "静态分析完成！"

# 运行所有检查
check: fmt vet lint test
	@echo "所有检查完成！"

# 初始化数据库
init-db:
	@echo "初始化数据库..."
	@if command -v mysql >/dev/null 2>&1; then \
		echo "正在执行数据库初始化脚本..."; \
		mysql -u root -p < scripts/init_database.sql; \
		echo "数据库初始化完成！"; \
	else \
		echo "MySQL 客户端未安装，请手动执行 scripts/init_database.sql"; \
	fi

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "开发工具安装完成！"

# 生成文档
docs:
	@echo "生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "启动文档服务器: http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc 未安装，安装命令: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 性能测试
bench:
	@echo "运行性能测试..."
	go test -bench=. -benchmem ./...

# 覆盖率测试
coverage:
	@echo "运行覆盖率测试..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

# 交叉编译
build-all:
	@echo "交叉编译..."
	GOOS=linux GOARCH=amd64 go build -o bin/go-sharding-demo-linux-amd64 cmd/demo/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/go-sharding-demo-windows-amd64.exe cmd/demo/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/go-sharding-demo-darwin-amd64 cmd/demo/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/go-sharding-demo-darwin-arm64 cmd/demo/main.go
	@echo "交叉编译完成！"

# Docker 构建
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t go-sharding:latest .
	@echo "Docker 镜像构建完成！"

# Docker 运行
docker-run:
	@echo "运行 Docker 容器..."
	docker run --rm -it go-sharding:latest

# 版本信息
version:
	@echo "Go-Sharding 版本信息:"
	@echo "Go 版本: $(shell go version)"
	@echo "项目版本: v1.0.0"
	@echo "构建时间: $(shell date)"