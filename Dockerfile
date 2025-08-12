# Go-Sharding Dockerfile
# 基于 Apache ShardingSphere 设计的 Go 语言分片数据库中间件

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-sharding-demo cmd/demo/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-sharding-basic examples/basic/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-sharding-yaml examples/yaml_config/main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/go-sharding-demo .
COPY --from=builder /app/go-sharding-basic .
COPY --from=builder /app/go-sharding-yaml .

# 复制配置文件
COPY --from=builder /app/examples/yaml_config/config.yaml ./config/
COPY --from=builder /app/scripts/init_database.sql ./scripts/

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 更改文件所有者
RUN chown -R appuser:appgroup /root/

# 切换到非 root 用户
USER appuser

# 暴露端口（如果需要）
EXPOSE 8080

# 设置默认命令
CMD ["./go-sharding-demo"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD echo "Go-Sharding is running" || exit 1

# 标签
LABEL maintainer="Go-Sharding Team" \
      version="1.0.0" \
      description="Go-Sharding: 基于 Apache ShardingSphere 设计的 Go 语言分片数据库中间件" \
      org.opencontainers.image.source="https://github.com/your-org/go-sharding" \
      org.opencontainers.image.documentation="https://github.com/your-org/go-sharding/blob/main/README.md" \
      org.opencontainers.image.licenses="Apache-2.0"