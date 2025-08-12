# Go-Sharding 架构设计

## 整体架构

Go-Sharding 是一个基于 Go 语言实现的分片数据库中间件，提供透明的数据库分片功能。

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

## 核心组件

### 1. 配置管理器 (Config Manager)

负责管理分片规则、数据源配置等。

**主要功能：**
- 数据源配置管理
- 分片规则配置
- 读写分离配置
- YAML/JSON 配置文件支持
- 配置验证

**核心接口：**
```go
type ShardingConfig struct {
    DataSources     map[string]*DataSourceConfig
    ShardingRule    *ShardingRuleConfig
    ReadWriteSplits map[string]*ReadWriteSplitConfig
}
```

### 2. 路由引擎 (Routing Engine)

根据分片规则和 SQL 参数计算目标数据源和表。

**主要功能：**
- 分片键提取
- 分片算法执行
- 路由结果计算
- 支持多种分片策略（内联表达式、标准分片、复合分片、Hint）

**核心接口：**
```go
type Router interface {
    Route(logicTable string, shardingValues map[string]interface{}) ([]*RouteResult, error)
}
```

**支持的分片算法：**
- 取模分片：`ds_${user_id % 2}`
- 范围分片：`ds_${user_id / 1000}`
- 哈希分片：`ds_${hash(user_id) % 4}`
- 自定义算法

### 3. SQL 重写器 (SQL Rewriter)

将逻辑 SQL 重写为针对实际数据源的物理 SQL。

**主要功能：**
- 逻辑表名替换为实际表名
- 多表 UNION 查询生成
- SQL 语法解析和重构
- 参数绑定处理

**核心接口：**
```go
type SQLRewriter interface {
    Rewrite(ctx *RewriteContext) ([]*RewriteResult, error)
}
```

**重写示例：**
```sql
-- 逻辑 SQL
SELECT * FROM t_order WHERE user_id = 10

-- 重写后的物理 SQL
SELECT * FROM t_order_0 WHERE user_id = 10  -- 在 ds_0 上执行
```

### 4. 结果合并器 (Result Merger)

将多个分片的查询结果合并为统一的结果集。

**主要功能：**
- 流式结果合并
- 排序合并（ORDER BY）
- 分组聚合（GROUP BY）
- 分页处理（LIMIT/OFFSET）
- 聚合函数计算（COUNT、SUM、AVG、MIN、MAX）

**核心接口：**
```go
type ResultMerger interface {
    Merge(results []*sql.Rows, ctx *MergeContext) (*MergedRows, error)
}
```

**合并策略：**
- **流式合并**：适用于无需排序的查询
- **内存合并**：适用于需要排序、分组的查询
- **装饰器合并**：适用于聚合函数查询

### 5. ID 生成器 (ID Generator)

为分片表生成全局唯一的主键。

**主要功能：**
- 雪花算法（Snowflake）
- UUID 生成
- 自增序列
- 自定义生成器

**核心接口：**
```go
type Generator interface {
    NextID() (int64, error)
}
```

**雪花算法结构：**
```
0 - 0000000000 0000000000 0000000000 0000000000 0 - 00000 - 00000 - 000000000000
|   |                                             |   |       |       |
|   |<-------------- 41位时间戳 ---------------->|   |       |       |
|                                                     |       |       |
|<- 1位符号位                                    5位数据中心  |       |
                                                         5位机器ID   |
                                                                 12位序列号
```

## 分片策略

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

## 读写分离

支持主从数据库的读写分离，提高系统性能。

```yaml
readWriteSplits:
  master_slave_ds_0:
    masterDataSource: ds_0_master
    slaveDataSources:
      - ds_0_slave_0
      - ds_0_slave_1
    loadBalanceAlgorithm: round_robin
```

**负载均衡算法：**
- 轮询（Round Robin）
- 随机（Random）
- 权重轮询（Weighted Round Robin）

## 事务管理

### 1. 本地事务

单分片内的事务，直接使用数据库的本地事务。

### 2. 分布式事务

跨分片的事务，支持以下模式：

**最大努力交付型事务：**
- 适用于对一致性要求不高的场景
- 通过重试机制保证最终一致性
- 性能较好，实现简单

**两阶段提交（2PC）：**
- 强一致性保证
- 性能相对较差
- 实现复杂度高

## 性能优化

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

## 监控和运维

### 1. 性能监控

- SQL 执行时间统计
- 分片路由性能监控
- 连接池状态监控

### 2. 日志记录

- SQL 执行日志
- 路由决策日志
- 错误和异常日志

### 3. 健康检查

- 数据源连接状态检查
- 分片规则验证
- 系统资源监控

## 扩展性设计

### 1. 插件化架构

- 自定义分片算法
- 自定义 ID 生成器
- 自定义负载均衡算法

### 2. 配置热更新

- 动态添加/删除数据源
- 分片规则在线调整
- 无需重启应用

### 3. 多数据库支持

- MySQL
- PostgreSQL
- SQLite
- 其他兼容 database/sql 的数据库