# Go-Sharding 增强功能

本项目基于 Apache ShardingSphere 的设计理念，实现了 Go 语言版本的分片中间件，并增强了以下功能：

## 🚀 新增功能

### 1. 读写分离功能

#### 特性
- **主从数据源配置**：支持一主多从的数据源配置
- **自动读写路由**：根据 SQL 类型自动路由到主库或从库
- **负载均衡算法**：支持轮询（round_robin）和随机（random）两种负载均衡策略
- **强制主库访问**：支持通过上下文强制访问主库
- **事务支持**：事务中的所有操作自动路由到主库

#### 使用示例

```go
// 配置读写分离
readWriteSplits := map[string]*config.ReadWriteSplitConfig{
    "rw_ds_0": {
        MasterDataSource: "master_ds_0",
        SlaveDataSources: []string{"slave_ds_0_1", "slave_ds_0_2"},
        LoadBalanceAlgorithm: "round_robin",
    },
}

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

### 2. 增强的 SQL 解析能力

#### 特性
- **多种 SQL 类型支持**：SELECT、INSERT、UPDATE、DELETE、CREATE、DROP、ALTER
- **复杂查询解析**：支持 JOIN、子查询、聚合函数等
- **语法元素提取**：
  - 表名提取
  - 列名提取
  - WHERE 条件解析
  - JOIN 表解析
  - ORDER BY 子句
  - GROUP BY 子句
  - HAVING 子句
  - LIMIT 子句

#### 使用示例

```go
// 创建 SQL 解析器
parser := parser.NewSQLParser()

// 解析复杂 SQL
sql := `
    SELECT o.order_id, o.amount, u.username 
    FROM t_order o 
    JOIN t_user u ON o.user_id = u.user_id 
    WHERE o.user_id = ? AND o.status = ?
    ORDER BY o.order_id DESC 
    LIMIT 10
`

stmt, err := parser.Parse(sql)
if err != nil {
    log.Fatal(err)
}

// 获取解析结果
fmt.Printf("SQL Type: %s\n", stmt.Type)
fmt.Printf("Tables: %v\n", stmt.Tables)
fmt.Printf("Columns: %v\n", stmt.Columns)
fmt.Printf("JOIN Tables: %v\n", stmt.JoinTables)
fmt.Printf("ORDER BY: %v\n", stmt.OrderBy)
fmt.Printf("LIMIT: %v\n", stmt.Limit)
```

### 3. 增强的分片数据源管理

#### 特性
- **集成读写分离**：分片和读写分离功能无缝集成
- **智能路由**：根据 SQL 类型和分片规则智能路由
- **健康检查**：支持数据源和读写分离器的健康检查
- **连接池管理**：优化的数据库连接池管理

#### 使用示例

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

## 📁 项目结构

```
pkg/
├── config/          # 配置管理
├── readwrite/       # 读写分离功能
├── parser/          # 增强的 SQL 解析器
├── sharding/        # 增强的分片管理
├── routing/         # 路由功能
├── rewrite/         # SQL 重写功能
└── ...

examples/
└── enhanced_sharding_example.go  # 完整使用示例
```

## 🔧 核心组件

### ReadWriteSplitter
- 负责读写分离的核心逻辑
- 支持多种负载均衡算法
- 提供健康检查和连接管理

### SQLParser
- 增强的 SQL 解析器
- 支持复杂 SQL 语句解析
- 提供详细的语法元素提取

### EnhancedShardingDB
- 集成分片和读写分离的数据库实例
- 提供统一的数据库操作接口
- 支持事务和上下文管理

## 🚦 使用流程

1. **配置数据源**：配置主从数据源和分片规则
2. **创建实例**：创建 EnhancedShardingDB 实例
3. **执行操作**：使用标准的数据库操作接口
4. **自动路由**：系统自动处理分片和读写分离

## 📊 性能特性

- **连接池优化**：智能的数据库连接池管理
- **解析缓存**：SQL 解析结果缓存
- **负载均衡**：多种负载均衡策略
- **健康检查**：实时的数据源健康监控

## 🔍 监控和调试

- 支持详细的执行日志
- 提供性能指标监控
- 支持 SQL 执行追踪
- 健康检查接口

## 📝 配置示例

完整的配置示例请参考 `examples/enhanced_sharding_example.go` 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 📄 许可证

本项目采用 MIT 许可证。