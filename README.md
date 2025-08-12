# Go-Sharding

Go 语言分片数据库中间件

## 功能特性

- 支持数据库分片和表分片
- 支持多种分片算法（取模、范围、哈希等）
- 支持跨分片查询和聚合
- 支持分布式主键生成（Snowflake）
- 支持读写分离
- 支持分布式事务
- 支持 SQL 路由和重写
- 支持结果合并
- 支持监控和指标收集
- **🆕 支持 PostgreSQL 数据库**
  - JSONB 数据类型支持
  - 数组类型支持
  - 全文搜索（tsvector/tsquery）
  - 窗口函数
  - CTE（公共表表达式）
  - RETURNING 子句
  - 参数占位符转换（? → $1, $2, ...）

## 快速开始

### 安装

```bash
go get github.com/your-username/go-sharding
```

### 基本使用

```go
package main

import (
    "go-sharding/pkg/config"
    "go-sharding/pkg/sharding"
    "log"
)

func main() {
    // 创建数据源配置
    dataSources := map[string]*config.DataSourceConfig{
        "ds_0": {
            DriverName: "mysql",
            URL:        "root:password@tcp(localhost:3306)/ds_0",
            MaxIdle:    10,
            MaxOpen:    100,
        },
        "ds_1": {
            DriverName: "mysql", 
            URL:        "root:password@tcp(localhost:3306)/ds_1",
            MaxIdle:    10,
            MaxOpen:    100,
        },
    }

    // 创建分片规则配置
    shardingRule := &config.ShardingRuleConfig{
        Tables: map[string]*config.TableRuleConfig{
            "t_user": {
                LogicTable:      "t_user",
                ActualDataNodes: "ds_${0..1}.t_user",
                DatabaseStrategy: &config.ShardingStrategyConfig{
                    ShardingColumn: "user_id",
                    Algorithm:      "ds_${user_id % 2}",
                    Type:           "inline",
                },
                KeyGenerator: &config.KeyGeneratorConfig{
                    Column: "user_id",
                    Type:   "snowflake",
                },
            },
        },
    }

    // 创建分片配置
    shardingConfig := &config.ShardingConfig{
        DataSources:  dataSources,
        ShardingRule: shardingRule,
    }

    // 创建分片数据源
    dataSource, err := sharding.NewShardingDataSource(shardingConfig)
    if err != nil {
        log.Fatalf("创建分片数据源失败: %v", err)
    }
    defer dataSource.Close()

    // 获取数据库连接
    db := dataSource.DB()

    // 执行 SQL
    result, err := db.Exec("INSERT INTO t_user (user_name, user_email) VALUES (?, ?)", "张三", "zhangsan@example.com")
    if err != nil {
        log.Printf("插入失败: %v", err)
    }
}
```

## 配置

### 数据源配置

```yaml
dataSources:
  ds_0:
    driverName: mysql
    url: "root:password@tcp(localhost:3306)/ds_0"
    maxIdle: 10
    maxOpen: 100
  ds_1:
    driverName: mysql
    url: "root:password@tcp(localhost:3306)/ds_1"
    maxIdle: 10
    maxOpen: 100
```

### 分片规则配置

```yaml
shardingRule:
  tables:
    t_user:
      logicTable: t_user
      actualDataNodes: "ds_${0..1}.t_user"
      databaseStrategy:
        shardingColumn: user_id
        algorithm: "ds_${user_id % 2}"
        type: inline
      keyGenerator:
        column: user_id
        type: snowflake
    t_order:
      logicTable: t_order
      actualDataNodes: "ds_${0..1}.t_order_${0..1}"
      databaseStrategy:
        shardingColumn: user_id
        algorithm: "ds_${user_id % 2}"
        type: inline
      tableStrategy:
        shardingColumn: order_id
        algorithm: "t_order_${order_id % 2}"
        type: inline
      keyGenerator:
        column: order_id
        type: snowflake
```

## 运行演示

```bash
# 构建演示程序
go build -o bin/go-sharding-demo ./cmd/demo

# 运行演示
./bin/go-sharding-demo
```

演示程序将展示：
- 分片配置信息
- 分片表配置
- SQL 路由逻辑演示
- 不同类型查询的路由结果

## 示例

查看 `examples/` 目录下的示例代码：

- `examples/basic/` - 基本使用示例
- `examples/yaml_config/` - YAML 配置示例
- `examples/base_transaction/` - BASE事务使用示例
- `examples/postgresql/` - **🆕 PostgreSQL 使用示例**

### PostgreSQL 快速开始

```bash
# 启动 PostgreSQL 集群
docker-compose -f docker-compose-postgresql.yml up -d

# 运行测试脚本
./scripts/test-postgresql.sh

# 运行 PostgreSQL 示例
cd examples/postgresql && go run main.go
```

详细的 PostgreSQL 支持文档请参考 [README-PostgreSQL.md](README-PostgreSQL.md)

## 核心组件

### 1. 配置管理 (pkg/config)
- 数据源配置
- 分片规则配置
- YAML 配置文件支持

### 2. 路由引擎 (pkg/routing)
- SQL 解析和路由
- 分片算法实现
- 数据节点计算

### 3. SQL 重写 (pkg/rewrite)
- 逻辑表到物理表的转换
- SQL 语句重写
- 参数绑定

### 4. 结果合并 (pkg/merge)
- 跨分片结果合并
- 聚合函数处理
- 排序和分页

### 5. ID 生成器 (pkg/id)
- Snowflake 算法
- 分布式唯一 ID 生成

### 6. 事务管理 (pkg/transaction)
- 本地事务支持
- XA分布式事务（两阶段提交）
- BASE事务（最终一致性）
- 事务状态管理和监控

### 7. 监控指标 (pkg/monitoring)
- 性能指标收集
- 监控数据导出

## 架构设计

详细的架构设计请参考 [docs/architecture.md](docs/architecture.md)

## 测试

```bash
# 运行所有测试
go test ./...

# 运行核心包测试
go test ./pkg/...
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License