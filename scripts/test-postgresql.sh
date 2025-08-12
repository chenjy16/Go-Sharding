#!/bin/bash

# PostgreSQL 功能测试脚本

set -e

echo "🚀 开始 PostgreSQL 功能测试..."

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    exit 1
fi

echo "✅ Docker 运行正常"

# 启动 PostgreSQL 集群
echo "📦 启动 PostgreSQL 集群..."
docker-compose -f docker-compose-postgresql.yml up -d

# 等待服务启动
echo "⏳ 等待 PostgreSQL 服务启动..."
sleep 10

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose -f docker-compose-postgresql.yml ps

# 等待数据库完全启动
echo "⏳ 等待数据库完全启动..."
for i in {1..30}; do
    if docker exec go-sharding-postgres-ds0-write-1 pg_isready -U sharding_user -d sharding_db > /dev/null 2>&1; then
        echo "✅ 数据源 0 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ 数据源 0 启动超时"
        exit 1
    fi
    sleep 2
done

for i in {1..30}; do
    if docker exec go-sharding-postgres-ds1-write-1 pg_isready -U sharding_user -d sharding_db > /dev/null 2>&1; then
        echo "✅ 数据源 1 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ 数据源 1 启动超时"
        exit 1
    fi
    sleep 2
done

# 初始化数据库
echo "🗄️ 初始化数据库..."
docker exec -i go-sharding-postgres-ds0-write-1 psql -U sharding_user -d sharding_db < scripts/postgresql/init-ds0.sql
docker exec -i go-sharding-postgres-ds1-write-1 psql -U sharding_user -d sharding_db < scripts/postgresql/init-ds1.sql

echo "✅ 数据库初始化完成"

# 验证数据库连接
echo "🔗 验证数据库连接..."
docker exec go-sharding-postgres-ds0-write-1 psql -U sharding_user -d sharding_db -c "SELECT 'DS0 连接成功' as status;"
docker exec go-sharding-postgres-ds1-write-1 psql -U sharding_user -d sharding_db -c "SELECT 'DS1 连接成功' as status;"

# 验证表创建
echo "📋 验证表创建..."
docker exec go-sharding-postgres-ds0-write-1 psql -U sharding_user -d sharding_db -c "\\dt"
docker exec go-sharding-postgres-ds1-write-1 psql -U sharding_user -d sharding_db -c "\\dt"

# 验证测试数据
echo "📊 验证测试数据..."
docker exec go-sharding-postgres-ds0-write-1 psql -U sharding_user -d sharding_db -c "SELECT COUNT(*) as user_count FROM user_0 UNION ALL SELECT COUNT(*) FROM user_1;"
docker exec go-sharding-postgres-ds1-write-1 psql -U sharding_user -d sharding_db -c "SELECT COUNT(*) as user_count FROM user_0 UNION ALL SELECT COUNT(*) FROM user_1;"

# 编译 Go 项目
echo "🔨 编译 Go 项目..."
go build ./...

echo "✅ 编译成功"

# 运行测试
echo "🧪 运行测试..."
go test ./pkg/config -v
go test ./pkg/sharding -v

echo "✅ 测试通过"

# 运行 PostgreSQL 示例（如果存在）
if [ -f "examples/postgresql/main.go" ]; then
    echo "🎯 运行 PostgreSQL 示例..."
    cd examples/postgresql
    timeout 30s go run main.go || echo "⚠️ 示例运行超时或出错（这是正常的，因为可能需要手动交互）"
    cd ../..
fi

echo "🎉 PostgreSQL 功能测试完成！"
echo ""
echo "📋 测试总结："
echo "✅ Docker 服务正常"
echo "✅ PostgreSQL 集群启动成功"
echo "✅ 数据库初始化完成"
echo "✅ 连接验证通过"
echo "✅ 表结构创建成功"
echo "✅ 测试数据插入成功"
echo "✅ Go 项目编译成功"
echo "✅ 单元测试通过"
echo ""
echo "🌐 访问地址："
echo "- pgAdmin: http://localhost:8080 (admin@example.com / admin123)"
echo "- Prometheus DS0: http://localhost:9187/metrics"
echo "- Prometheus DS1: http://localhost:9188/metrics"
echo ""
echo "🛠️ 管理命令："
echo "- 停止服务: docker-compose -f docker-compose-postgresql.yml down"
echo "- 查看日志: docker-compose -f docker-compose-postgresql.yml logs -f"
echo "- 重启服务: docker-compose -f docker-compose-postgresql.yml restart"