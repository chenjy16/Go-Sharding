# PostgreSQL 支持

本项目已增加对 PostgreSQL 数据库的完整支持，包括分片、读写分离、事务管理等功能。

## 功能特性

### 核心功能
- ✅ PostgreSQL 数据库连接和管理
- ✅ 分片支持（水平分片）
- ✅ 读写分离
- ✅ 事务管理（提交/回滚）
- ✅ 连接池管理
- ✅ SQL 解析和重写

### PostgreSQL 特有功能
- ✅ JSONB 数据类型支持
- ✅ 数组类型支持
- ✅ 全文搜索（tsvector/tsquery）
- ✅ 窗口函数
- ✅ CTE（公共表表达式）
- ✅ RETURNING 子句
- ✅ 自定义数据类型
- ✅ 扩展函数和操作符
- ✅ 参数占位符转换（? → $1, $2, ...）

## 快速开始

### 1. 环境准备

使用 Docker Compose 启动 PostgreSQL 集群：

```bash
# 启动 PostgreSQL 集群
docker-compose -f docker-compose-postgresql.yml up -d

# 查看服务状态
docker-compose -f docker-compose-postgresql.yml ps
```

### 2. 初始化数据库

```bash
# 初始化数据源 0
docker exec -i go-sharding-postgres-ds0-write-1 psql -U sharding_user -d sharding_db < scripts/postgresql/init-ds0.sql

# 初始化数据源 1
docker exec -i go-sharding-postgres-ds1-write-1 psql -U sharding_user -d sharding_db < scripts/postgresql/init-ds1.sql
```

### 3. 运行示例

```bash
# 运行 PostgreSQL 示例
cd examples/postgresql
go run main.go
```

## 配置说明

### 基本配置

```yaml
# examples/postgresql_config/config.yaml
dataSources:
  ds0:
    driverName: "postgres"
    dataSourceName: "postgres://sharding_user:sharding_pass@localhost:5432/sharding_db?sslmode=disable"
    maxOpenConns: 20
    maxIdleConns: 10
    connMaxLifetime: "1h"
    
  ds1:
    driverName: "postgres"
    dataSourceName: "postgres://sharding_user:sharding_pass@localhost:5433/sharding_db?sslmode=disable"
    maxOpenConns: 20
    maxIdleConns: 10
    connMaxLifetime: "1h"

shardingRules:
  tables:
    user:
      actualDataNodes: "ds${0..1}.user_${0..1}"
      databaseShardingStrategy:
        shardingColumn: "user_id"
        algorithmExpression: "ds${user_id % 2}"
      tableShardingStrategy:
        shardingColumn: "user_id"
        algorithmExpression: "user_${user_id % 2}"
```

### PostgreSQL 特有配置

```yaml
postgresql:
  features:
    jsonb: true
    arrays: true
    fullTextSearch: true
    windowFunctions: true
    cte: true
    returning: true
    customTypes: true
    extensions: true
  
  extensions:
    - "uuid-ossp"
    - "pg_stat_statements"
    - "pg_trgm"
    - "btree_gin"
    - "btree_gist"
```

## 代码示例

### 基本 CRUD 操作

```go
package main

import (
    "context"
    "log"
    
    "go-sharding/pkg/config"
    "go-sharding/pkg/sharding"
)

func main() {
    // 加载配置
    cfg, err := config.LoadFromYAML("../postgresql_config/config.yaml")
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }

    // 创建 PostgreSQL 分片数据源
    ds, err := sharding.NewPostgreSQLShardingDataSource(cfg)
    if err != nil {
        log.Fatal("创建数据源失败:", err)
    }
    defer ds.Close()

    ctx := context.Background()

    // 插入用户
    _, err = ds.ExecContext(ctx, `
        INSERT INTO user (username, email, password_hash, address, tags) 
        VALUES (?, ?, ?, ?, ?)`,
        "john_doe", "john@example.com", "$2a$10$hash",
        `{"street": "123 Main St", "city": "Beijing"}`,
        `{"premium", "web"}`)
    if err != nil {
        log.Fatal("插入用户失败:", err)
    }

    // 查询用户
    rows, err := ds.QueryContext(ctx, `
        SELECT user_id, username, email, address, tags 
        FROM user WHERE username = ?`, "john_doe")
    if err != nil {
        log.Fatal("查询用户失败:", err)
    }
    defer rows.Close()
}
```

### PostgreSQL 特有功能

```go
// JSONB 查询
rows, err := ds.QueryContext(ctx, `
    SELECT username, address->>'city' as city 
    FROM user 
    WHERE address @> '{"city": "Beijing"}'`)

// 数组操作
_, err = ds.ExecContext(ctx, `
    UPDATE user 
    SET tags = array_append(tags, ?) 
    WHERE user_id = ?`, "new_tag", userID)

// 全文搜索
rows, err := ds.QueryContext(ctx, `
    SELECT username, email 
    FROM user 
    WHERE search_vector @@ to_tsquery('english', ?)`, "john")

// 窗口函数
rows, err := ds.QueryContext(ctx, `
    SELECT 
        username,
        total_amount,
        ROW_NUMBER() OVER (ORDER BY total_amount DESC) as rank
    FROM user_order_summary`)

// CTE（公共表表达式）
rows, err := ds.QueryContext(ctx, `
    WITH monthly_sales AS (
        SELECT 
            DATE_TRUNC('month', created_at) as month,
            SUM(total_amount) as total
        FROM order_table
        GROUP BY DATE_TRUNC('month', created_at)
    )
    SELECT month, total FROM monthly_sales ORDER BY month`)

// RETURNING 子句
var newOrderID int64
err = ds.QueryRowContext(ctx, `
    INSERT INTO order_table (user_id, product_name, total_amount) 
    VALUES (?, ?, ?) 
    RETURNING order_id`, userID, "Product", 99.99).Scan(&newOrderID)
```

### 事务操作

```go
// 开始事务
tx, err := ds.BeginTx(ctx, nil)
if err != nil {
    log.Fatal("开始事务失败:", err)
}

// 执行操作
_, err = tx.ExecContext(ctx, `
    INSERT INTO user (username, email) VALUES (?, ?)`,
    "user1", "user1@example.com")
if err != nil {
    tx.Rollback()
    log.Fatal("插入失败:", err)
}

_, err = tx.ExecContext(ctx, `
    INSERT INTO order_table (user_id, total_amount) VALUES (?, ?)`,
    1, 100.00)
if err != nil {
    tx.Rollback()
    log.Fatal("插入订单失败:", err)
}

// 提交事务
err = tx.Commit()
if err != nil {
    log.Fatal("提交事务失败:", err)
}
```

## 架构说明

### 文件结构

```
pkg/
├── parser/
│   ├── postgresql_parser.go      # PostgreSQL SQL 解析器
│   └── enhanced_parser.go        # 增强的 SQL 解析器
├── sharding/
│   └── postgresql_datasource.go  # PostgreSQL 分片数据源
└── config/
    └── config.go                 # 配置管理

examples/
├── postgresql/
│   └── main.go                   # PostgreSQL 示例
└── postgresql_config/
    └── config.yaml               # PostgreSQL 配置

scripts/
└── postgresql/
    ├── init-ds0.sql              # 数据源 0 初始化脚本
    └── init-ds1.sql              # 数据源 1 初始化脚本

docker-compose-postgresql.yml     # PostgreSQL Docker Compose 配置
```

### 核心组件

1. **PostgreSQL 解析器** (`postgresql_parser.go`)
   - 解析 PostgreSQL 特有的 SQL 语法
   - 支持 JSONB、数组、全文搜索等特性
   - 参数占位符转换

2. **PostgreSQL 数据源** (`postgresql_datasource.go`)
   - 管理 PostgreSQL 连接池
   - 实现分片路由逻辑
   - 处理事务和连接管理

3. **配置管理** (`config.go`)
   - 支持 PostgreSQL 特有配置
   - 数据源配置和分片规则

## 监控和管理

### pgAdmin 管理界面

访问 http://localhost:8080 使用 pgAdmin 管理 PostgreSQL：

- 邮箱: admin@example.com
- 密码: admin123

### Prometheus 监控

PostgreSQL 指标通过 postgres_exporter 暴露：

- ds0 指标: http://localhost:9187/metrics
- ds1 指标: http://localhost:9188/metrics

### 常用监控指标

- `pg_up`: 数据库连接状态
- `pg_stat_database_*`: 数据库统计信息
- `pg_stat_user_tables_*`: 表统计信息
- `pg_locks_*`: 锁信息

## 性能优化

### 连接池配置

```yaml
dataSources:
  ds0:
    maxOpenConns: 20        # 最大连接数
    maxIdleConns: 10        # 最大空闲连接数
    connMaxLifetime: "1h"   # 连接最大生存时间
```

### 索引优化

```sql
-- 为分片键创建索引
CREATE INDEX idx_user_id ON user_table(user_id);

-- 为 JSONB 字段创建 GIN 索引
CREATE INDEX idx_address_gin ON user_table USING gin(address);

-- 为数组字段创建 GIN 索引
CREATE INDEX idx_tags_gin ON user_table USING gin(tags);

-- 为全文搜索创建 GIN 索引
CREATE INDEX idx_search_vector ON user_table USING gin(search_vector);
```

### 查询优化

1. 使用合适的分片键进行查询
2. 避免跨分片的复杂 JOIN 操作
3. 合理使用 JSONB 操作符
4. 利用 PostgreSQL 的查询计划器

## 故障排除

### 常见问题

1. **连接失败**
   ```
   检查 PostgreSQL 服务是否启动
   验证连接字符串是否正确
   确认防火墙设置
   ```

2. **分片路由错误**
   ```
   检查分片键配置
   验证分片算法
   查看日志输出
   ```

3. **性能问题**
   ```
   检查索引使用情况
   分析查询计划
   监控连接池状态
   ```

### 日志配置

```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  postgresql:
    slowQuery: "1s"
    logQueries: true
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 许可证

本项目采用 Apache 2.0 许可证。