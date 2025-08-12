# Go-Sharding

Go 语言分片数据库中间件 - 基于 Apache ShardingSphere 设计理念的高性能分片解决方案

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [架构设计](#架构设计)
- [核心组件](#核心组件)
- [数据库支持](#数据库支持)
  - [MySQL 支持](#mysql-支持)
  - [PostgreSQL 支持](#postgresql-支持)
- [SQL 解析器](#sql-解析器)
  - [解析器配置和启用](#解析器配置和启用)
- [分片策略](#分片策略)
- [读写分离](#读写分离)
- [事务管理](#事务管理)
- [配置说明](#配置说明)
- [示例代码](#示例代码)
- [性能优化](#性能优化)
- [测试覆盖](#测试覆盖)
- [部署运维](#部署运维)
- [开发指南](#开发指南)
- [贡献指南](#贡献指南)

## 🚀 功能特性

### 核心功能
- ✅ **数据库分片和表分片**：支持水平分片，提高数据处理能力
- ✅ **多种分片算法**：取模、范围、哈希、自定义算法
- ✅ **跨分片查询和聚合**：智能路由和结果合并
- ✅ **分布式主键生成**：Snowflake 算法保证全局唯一性
- ✅ **读写分离**：主从数据库自动路由，提升性能
- ✅ **分布式事务**：支持本地事务、XA事务、BASE事务
- ✅ **SQL 路由和重写**：智能 SQL 解析和重写
- ✅ **结果合并**：支持排序、分组、聚合、分页
- ✅ **监控和指标收集**：完整的性能监控体系

### 数据库支持
- ✅ **MySQL**：完整支持，包括复杂查询和事务
- ✅ **PostgreSQL**：全面支持，包括特有功能
  - JSONB 数据类型支持
  - 数组类型支持
  - 全文搜索（tsvector/tsquery）
  - 窗口函数
  - CTE（公共表表达式）
  - RETURNING 子句
  - 参数占位符转换（? → $1, $2, ...）

### 高级功能
- ✅ **多解析器架构**：支持原生、TiDB、PostgreSQL、增强解析器
- ✅ **智能路由**：基于分片键的自动路由
- ✅ **连接池管理**：优化的数据库连接池
- ✅ **健康检查**：实时监控数据源状态
- ✅ **配置热更新**：支持运行时配置更新

## 🏃‍♂️ 快速开始

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

### 运行演示

```bash
# 构建演示程序
go build -o bin/go-sharding-demo ./cmd/demo

# 运行演示
./bin/go-sharding-demo
```

## 🏗️ 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    应用程序层                                │
├─────────────────────────────────────────────────────────────┤
│                Go-Sharding 中间件                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   路由引擎   │ │  SQL重写器  │ │  结果合并器  │ │ID生成器 │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                配置管理器                                │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                  数据库驱动层                                │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   数据库1    │ │   数据库2    │ │   数据库3    │ │   ...   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 核心设计原则

1. **高性能**：优化的 SQL 解析和路由算法
2. **高可用**：支持故障转移和负载均衡
3. **易扩展**：模块化设计，支持自定义扩展
4. **透明性**：对应用程序透明，无需修改业务代码

## 🔧 核心组件

### 1. 配置管理器 (Config Manager)

负责管理分片规则、数据源配置等。

**主要功能：**
- 数据源配置管理
- 分片规则配置
- 读写分离配置
- YAML/JSON 配置文件支持
- 配置验证

### 2. 路由引擎 (Routing Engine)

根据分片规则和 SQL 参数计算目标数据源和表。

**主要功能：**
- 分片键提取
- 分片算法执行
- 路由结果计算
- 支持多种分片策略

### 3. SQL 重写器 (SQL Rewriter)

将逻辑 SQL 重写为针对实际数据源的物理 SQL。

**主要功能：**
- 逻辑表名替换为实际表名
- 多表 UNION 查询生成
- SQL 语法解析和重构
- 参数绑定处理

### 4. 结果合并器 (Result Merger)

将多个分片的查询结果合并为统一的结果集。

**主要功能：**
- 流式结果合并
- 排序合并（ORDER BY）
- 分组聚合（GROUP BY）
- 分页处理（LIMIT/OFFSET）
- 聚合函数计算

### 5. ID 生成器 (ID Generator)

为分片表生成全局唯一的主键。

**支持算法：**
- 雪花算法（Snowflake）
- UUID 生成
- 自增序列
- 自定义生成器

## 🗄️ 数据库支持

### MySQL 支持

完整支持 MySQL 数据库，包括：
- 标准 SQL 语法
- MySQL 特有函数
- 事务支持
- 连接池管理

### PostgreSQL 支持

全面支持 PostgreSQL 数据库及其特有功能：

#### 特有功能支持
- **JSONB 数据类型**：完整的 JSON 操作支持
- **数组类型**：数组操作和函数
- **全文搜索**：tsvector/tsquery 支持
- **窗口函数**：完整的窗口函数支持
- **CTE**：公共表表达式
- **RETURNING 子句**：INSERT/UPDATE/DELETE 返回值
- **自定义数据类型**：用户定义类型支持
- **参数占位符转换**：自动转换 ? 为 $1, $2, ...

#### 快速开始 PostgreSQL

```bash
# 启动 PostgreSQL 集群
docker-compose -f docker-compose-postgresql.yml up -d

# 运行测试脚本
./scripts/test-postgresql.sh

# 运行 PostgreSQL 示例
cd examples/postgresql && go run main.go
```

#### PostgreSQL 代码示例

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

// RETURNING 子句
var newOrderID int64
err = ds.QueryRowContext(ctx, `
    INSERT INTO order_table (user_id, product_name, total_amount) 
    VALUES (?, ?, ?) 
    RETURNING order_id`, userID, "Product", 99.99).Scan(&newOrderID)
```

## 🔍 SQL 解析器

### 多解析器架构

项目采用多层解析器架构，支持不同的解析策略：

#### 1. 原始解析器 (Original Parser)
- **技术实现**：基于正则表达式
- **性能特点**：轻量级，启动快
- **适用场景**：简单 SQL 语句
- **兼容性**：MySQL 85%，PostgreSQL 75%

#### 2. TiDB 解析器 (TiDB Parser)
- **技术实现**：集成 `pingcap/tidb/pkg/parser`
- **性能特点**：高性能，低内存使用
- **适用场景**：复杂 MySQL 查询
- **兼容性**：MySQL 98%+

**性能对比：**
| 测试场景 | 原始解析器 | TiDB Parser | 性能提升 |
|---------|-----------|-------------|----------|
| 简单查询 | 70μs | 5μs | **14x** |
| 复杂 JOIN | 150μs | 25μs | **6x** |
| INSERT 语句 | 80μs | 8μs | **10x** |
| 内存使用 | 101,300 B/op | 3,993 B/op | **96% 减少** |

#### 3. PostgreSQL 解析器
- **技术实现**：专门针对 PostgreSQL 语法
- **功能特点**：支持 PostgreSQL 特有语法
- **适用场景**：PostgreSQL 数据库

#### 4. 增强解析器 (Enhanced Parser)
- **技术实现**：集成多种解析器
- **功能特点**：智能选择最适合的解析器
- **适用场景**：混合数据库环境

### 解析器配置和启用

#### 配置文件方式（推荐）

创建 `config.yaml` 配置文件：

```yaml
parser:
  # 启用 TiDB 解析器作为默认解析器
  enable_tidb_parser: true
  # 启用 PostgreSQL 解析器
  enable_postgresql_parser: false
  # 当解析失败时是否回退到原始解析器
  fallback_to_original: true
  # 启用性能基准测试
  enable_benchmarking: true
  # 记录解析错误
  log_parsing_errors: true
```

在代码中只需一行初始化：

```go
import "go-sharding/pkg/parser"

// 从配置文件初始化解析器（最简单的方式）
err := parser.InitializeParserFromConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// 现在解析器已根据配置文件设置好了
stmt, err := parser.DefaultParserFactory.Parse("SELECT * FROM users")
```

#### 程序化配置方式

```go
// 方法 1: 直接启用 TiDB 解析器
err := parser.EnableTiDBParserAsDefault()
if err != nil {
    log.Fatal(err)
}

// 方法 2: 使用配置结构体
config := &parser.InitConfig{
    EnableTiDBParser:       true,
    EnablePostgreSQLParser: false,
    FallbackToOriginal:     true,
    EnableBenchmarking:     true,
    LogParsingErrors:       true,
    AutoEnableTiDB:         true,
}

err := parser.InitializeParser(config)
if err != nil {
    log.Fatal(err)
}

// 方法 3: 环境变量配置
// 设置环境变量: ENABLE_TIDB_PARSER=true
err := parser.InitializeParserFromEnv()
if err != nil {
    log.Fatal(err)
}
```

#### 验证配置是否生效

```go
// 检查当前默认解析器
parserType := parser.GetDefaultParserType()
fmt.Printf("当前默认解析器: %s\n", parserType) // 应该输出: tidb

// 打印详细信息
parser.PrintParserInfo()

// 获取统计信息
stats := parser.GetParserFactoryStats()
fmt.Printf("解析器统计: %+v\n", stats)
```

#### 配置优先级

解析器配置的优先级顺序（从高到低）：

1. **代码中直接调用** - `parser.EnableTiDBParserAsDefault()`
2. **环境变量** - `ENABLE_TIDB_PARSER=true`
3. **配置文件** - `config.yaml` 中的 `parser` 配置
4. **默认配置** - 系统默认设置

### 解析器工厂模式

```go
// 创建解析器
parser := parser.NewParserFactory().CreateParser("tidb")

// 解析 SQL
stmt, err := parser.Parse("SELECT * FROM users WHERE id = ?")

// 提取表名
tables := parser.ExtractTables(sql)
```

## 📊 分片策略

### 1. 数据库分片

根据分片键将数据分散到不同的数据库实例。

```yaml
databaseStrategy:
  type: inline
  shardingColumn: user_id
  algorithm: "ds_${user_id % 2}"
```

### 2. 表分片

在同一数据库内将数据分散到不同的表。

```yaml
tableStrategy:
  type: inline
  shardingColumn: order_id
  algorithm: "t_order_${order_id % 4}"
```

### 3. 复合分片

同时进行数据库分片和表分片。

```yaml
actualDataNodes: "ds_${0..1}.t_order_${0..3}"
databaseStrategy:
  shardingColumn: user_id
  algorithm: "ds_${user_id % 2}"
tableStrategy:
  shardingColumn: order_id
  algorithm: "t_order_${order_id % 4}"
```

### 支持的分片算法

- **取模分片**：`ds_${user_id % 2}`
- **范围分片**：`ds_${user_id / 1000}`
- **哈希分片**：`ds_${hash(user_id) % 4}`
- **自定义算法**：实现 `ShardingAlgorithm` 接口

## 🔄 读写分离

支持主从数据库的读写分离，提高系统性能。

### 配置示例

```yaml
readWriteSplits:
  rw_ds_0:
    masterDataSource: ds_0_master
    slaveDataSources:
      - ds_0_slave_0
      - ds_0_slave_1
    loadBalanceAlgorithm: round_robin
```

### 负载均衡算法

- **轮询（Round Robin）**：依次访问从库
- **随机（Random）**：随机选择从库
- **权重轮询（Weighted Round Robin）**：基于权重的轮询

### 使用示例

```go
// 创建读写分离器
splitter, err := readwrite.NewReadWriteSplitter(rwConfig, dataSources)

// 自动路由查询（读操作 -> 从库）
db := splitter.Route("SELECT * FROM users WHERE id = ?")

// 自动路由写操作（写操作 -> 主库）
db := splitter.Route("INSERT INTO users (name) VALUES (?)")

// 强制使用主库
ctx := context.WithValue(context.Background(), "force_master", true)
db := splitter.RouteContext(ctx, "SELECT * FROM users WHERE id = ?")
```

## 💾 事务管理

### 1. 本地事务

单分片内的事务，直接使用数据库的本地事务。

```go
tx, err := db.Begin()
if err != nil {
    return err
}

// 执行操作
_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
if err != nil {
    tx.Rollback()
    return err
}

// 提交事务
return tx.Commit()
```

### 2. XA 分布式事务

跨分片的强一致性事务，使用两阶段提交协议。

```go
// 开始 XA 事务
tx, err := tm.Begin(ctx, transaction.XATransaction)
if err != nil {
    return err
}

// 执行跨分片操作
err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
if err != nil {
    tx.Rollback()
    return err
}

// 提交事务
return tx.Commit()
```

### 3. BASE 事务

最终一致性的分布式事务，适用于对一致性要求不严格的场景。

#### BASE 事务特性

- **Basically Available（基本可用）**：系统在出现故障时仍能保证核心功能可用
- **Soft state（软状态）**：允许系统存在中间状态，不要求实时一致性
- **Eventually consistent（最终一致性）**：系统最终会达到一致状态

#### 使用示例

```go
// 创建事务管理器
tm := transaction.NewTransactionManager()
defer tm.Close()

// 开始 BASE 事务
ctx := context.Background()
tx, err := tm.Begin(ctx, transaction.BaseTransaction)
if err != nil {
    log.Fatalf("Failed to begin BASE transaction: %v", err)
}

baseTx := tx.(*transaction.BASETransactionImpl)

// 添加操作
op := transaction.BASEOperation{
    Type:       "INSERT",
    SQL:        "INSERT INTO orders (user_id, amount) VALUES (?, ?)",
    DataSource: "order_db",
    Parameters: []interface{}{123, 99.99},
}

err := baseTx.AddOperation(op)
if err != nil {
    log.Fatalf("Failed to add operation: %v", err)
}

// 添加补偿操作
comp := transaction.BASECompensation{
    OperationID: "op1",
    SQL:         "DELETE FROM orders WHERE user_id = ? AND amount = ?",
    DataSource:  "order_db",
    Parameters:  []interface{}{123, 99.99},
}

err := baseTx.AddCompensation(comp)
if err != nil {
    log.Fatalf("Failed to add compensation: %v", err)
}

// 提交事务
err := baseTx.Commit(ctx)
if err != nil {
    log.Fatalf("Failed to commit transaction: %v", err)
}
```

#### 事务状态管理

- **StatusActive (0)**：事务活跃状态，可以添加操作
- **StatusPrepared (1)**：事务正在执行中
- **StatusCommitted (2)**：事务成功提交
- **StatusRolledBack (3)**：事务已回滚
- **StatusFailed (4)**：事务执行失败

### 事务类型对比

| 特性 | LOCAL事务 | XA事务 | BASE事务 |
|------|-----------|--------|----------|
| 一致性 | 强一致性 | 强一致性 | 最终一致性 |
| 性能 | 高 | 中 | 高 |
| 可用性 | 中 | 低 | 高 |
| 复杂度 | 低 | 高 | 中 |
| 适用场景 | 单数据源 | 多数据源强一致性 | 多数据源最终一致性 |

## ⚙️ 配置说明

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

## 📝 示例代码

查看 `examples/` 目录下的示例代码：

### 基础示例
- `examples/basic/` - 基本使用示例
- `examples/yaml_config/` - YAML 配置示例

### 解析器示例
- `examples/enable_tidb_parser/` - TiDB 解析器启用示例
- `examples/config_file_parser/` - 配置文件解析器设置示例

### 数据库示例
- `examples/postgresql/` - PostgreSQL 使用示例

### 事务示例
- `examples/base_transaction/` - BASE事务使用示例

### 快速开始示例

#### 1. 基本分片使用

```bash
cd examples/basic
go run main.go
```

#### 2. 启用 TiDB 解析器

```bash
cd examples/enable_tidb_parser
go run main.go
```

#### 3. 配置文件解析器设置

```bash
cd examples/config_file_parser
go run main.go
```

#### 4. PostgreSQL 支持

```bash
# 启动 PostgreSQL 集群
docker-compose -f docker-compose-postgresql.yml up -d

# 运行示例
cd examples/postgresql
go run main.go
```

#### 5. BASE 事务示例

```bash
cd examples/base_transaction
go run main.go
```

### 增强功能示例

```go
// 创建增强的分片数据库
db, err := sharding.NewEnhancedShardingDB(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 健康检查
if err := db.HealthCheck(); err != nil {
    log.Printf("Health check failed: %v", err)
}

// 执行查询（自动分片 + 读写分离）
rows, err := db.QueryContext(ctx, 
    "SELECT * FROM t_order WHERE user_id = ?", userID)

// 执行写操作（自动分片 + 主库路由）
result, err := db.ExecContext(ctx,
    "INSERT INTO t_order (user_id, amount) VALUES (?, ?)", 
    userID, amount)
```

## 🚀 性能优化

### 1. 连接池管理

- 每个数据源独立的连接池
- 可配置的最大连接数和空闲连接数
- 连接复用和自动回收

### 2. 查询优化

- SQL 解析缓存
- 路由结果缓存
- 预编译语句支持

### 3. 结果流式处理

- 大结果集的流式合并
- 内存使用优化
- 分页查询优化

### 4. 解析器性能

TiDB Parser 相比原始解析器的性能提升：

- **解析速度**：提升 5-20 倍
- **内存使用**：减少 90%+
- **CPU 使用**：减少 80-90%

## 🧪 测试覆盖

### 测试覆盖率统计

- **总体语句覆盖率**: 58.3%
- **transaction 包覆盖率**: 75.8%

### 各包测试状态

- ✅ `algorithm` - 完整测试套件
- ✅ `config` - 已有测试
- ✅ `database` - 已有测试
- ✅ `executor` - 完整测试套件
- ✅ `id` - 已有测试
- ✅ `merge` - 已有测试
- ✅ `monitoring` - 已有测试
- ✅ `optimizer` - 完整测试套件
- ✅ `parser` - 已有测试
- ✅ `readwrite` - 已有测试
- ✅ `rewrite` - 已有测试
- ✅ `routing` - 已有测试
- ✅ `sharding` - 已有测试
- ✅ `transaction` - 已有测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行核心包测试
go test ./pkg/...

# 生成覆盖率报告
go test -v -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

## 🚢 部署运维

### Docker 部署

#### MySQL 环境

```bash
# 启动 MySQL 集群
docker-compose up -d

# 查看服务状态
docker-compose ps
```

#### PostgreSQL 环境

```bash
# 启动 PostgreSQL 集群
docker-compose -f docker-compose-postgresql.yml up -d

# 运行测试脚本
./scripts/test-postgresql.sh
```

### 监控指标

- SQL 执行时间统计
- 连接池状态监控
- 分片路由统计
- 错误率监控

### 管理界面

- **pgAdmin** (PostgreSQL): http://localhost:8080
- **Prometheus 监控**: 
  - DS0: http://localhost:9187/metrics
  - DS1: http://localhost:9188/metrics

## 👨‍💻 开发指南

### 项目结构

```
go-sharding/
├── cmd/                    # 命令行工具
├── pkg/                    # 核心包
│   ├── algorithm/          # 分片算法
│   ├── config/            # 配置管理
│   ├── database/          # 数据库管理
│   ├── executor/          # 执行器
│   ├── id/                # ID 生成器
│   ├── merge/             # 结果合并
│   ├── monitoring/        # 监控指标
│   ├── optimizer/         # 查询优化器
│   ├── parser/            # SQL 解析器
│   ├── readwrite/         # 读写分离
│   ├── rewrite/           # SQL 重写
│   ├── routing/           # 路由引擎
│   ├── sharding/          # 分片管理
│   └── transaction/       # 事务管理
├── examples/              # 示例代码
├── scripts/               # 脚本文件
├── docs/                  # 文档
└── docker-compose*.yml    # Docker 配置
```

### 核心接口

```go
// 解析器接口
type ParserInterface interface {
    Parse(sql string) (*SQLStatement, error)
    ExtractTables(sql string) []string
}

// 路由器接口
type Router interface {
    Route(logicTable string, shardingValues map[string]interface{}) ([]*RouteResult, error)
}

// 事务管理器接口
type TransactionManager interface {
    Begin(ctx context.Context, txType TransactionType) (Transaction, error)
    Commit(ctx context.Context, tx Transaction) error
    Rollback(ctx context.Context, tx Transaction) error
}
```

### 扩展开发

1. **自定义分片算法**

```go
type CustomShardingAlgorithm struct{}

func (a *CustomShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) []string {
    // 实现自定义分片逻辑
    return []string{"target_table"}
}
```

2. **自定义解析器**

```go
type CustomParser struct{}

func (p *CustomParser) Parse(sql string) (*SQLStatement, error) {
    // 实现自定义解析逻辑
    return &SQLStatement{}, nil
}
```

## 🤝 贡献指南

### 贡献流程

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 代码规范
- 添加必要的注释和文档
- 编写单元测试
- 确保测试通过

### 问题报告

如果发现 bug 或有功能建议，请创建 Issue 并提供：

- 详细的问题描述
- 复现步骤
- 期望行为
- 实际行为
- 环境信息

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [Apache ShardingSphere](https://shardingsphere.apache.org/) - 设计理念参考
- [TiDB Parser](https://github.com/pingcap/parser) - SQL 解析器
- [PostgreSQL](https://www.postgresql.org/) - 数据库支持

## 📞 联系我们

- 项目主页：https://github.com/your-username/go-sharding
- 问题反馈：https://github.com/your-username/go-sharding/issues
