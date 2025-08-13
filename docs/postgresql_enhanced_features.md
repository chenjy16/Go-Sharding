# PostgreSQL 增强解析器功能

本文档描述了 go-sharding 项目中 PostgreSQL 增强解析器的功能和使用方法。

## 概述

PostgreSQL 增强解析器基于 CockroachDB Parser 构建，提供了比传统正则表达式解析更精确和强大的 SQL 分析能力。它支持复杂的 SQL 语句分析，包括 CTE、窗口函数、子查询等高级特性。

## 主要功能

### 1. 精确的 AST 解析

- **基于 CockroachDB Parser**: 使用成熟的 SQL 解析器，支持完整的 PostgreSQL 语法
- **AST 分析**: 深度分析抽象语法树，提供准确的语法结构信息
- **错误处理**: 优雅的错误处理和回退机制

### 2. 增强的表名提取

```go
// 基本用法
parser := parser.NewPostgreSQLEnhancedParser()
tables, err := parser.ExtractTablesEnhanced(sql)

// 返回分类的表名
// map[string][]string{
//     "main": ["users", "posts"],
//     "cte": ["user_stats"],
//     "subqueries": ["comments"]
// }
```

**支持的表名提取场景**:
- 主查询表名
- JOIN 表名
- CTE (Common Table Expression) 表名
- 子查询中的表名
- 递归 CTE 表名

### 3. 复杂 SQL 重写

```go
// 分片重写
shardingRules := map[string]string{
    "users": "users_shard_1",
    "posts": "posts_shard_2",
}

rewrittenSQL, err := parser.RewriteForSharding(originalSQL, shardingRules)
```

**重写功能**:
- 表名替换（支持别名）
- CTE 表名重写
- 子查询表名重写
- 保持 SQL 语法正确性

### 4. 深度 SQL 分析

```go
analysis, err := parser.AnalyzeSQL(sql)

// 返回详细的分析结果
type EnhancedSQLAnalysis struct {
    Type              SQLType
    Tables            []string
    Columns           []string
    Subqueries        []SubqueryInfo
    CTEs              []CTEInfo
    Joins             []EnhancedJoinInfo
    WindowFunctions   []WindowFunctionInfo
    Complexity        ComplexityMetrics
    Optimizations     []OptimizationSuggestion
    PostgreSQLFeatures map[string]interface{}
}
```

**分析维度**:
- **语句类型**: SELECT, INSERT, UPDATE, DELETE, CREATE, DROP 等
- **表和列**: 涉及的所有表名和列名
- **子查询**: 标量子查询、WHERE 子查询、FROM 子查询等
- **CTE**: 普通 CTE 和递归 CTE
- **JOIN**: 内连接、外连接、交叉连接等
- **窗口函数**: ROW_NUMBER, RANK, LAG, LEAD 等
- **复杂度指标**: 表数量、JOIN 数量、嵌套级别等

### 5. PostgreSQL 特性支持

**数据类型**:
- JSONB 操作符 (`@>`, `?&`, `?|`)
- 数组类型和操作
- 几何类型 (point, polygon 等)
- 范围类型

**函数**:
- 聚合函数 (ARRAY_AGG, STRING_AGG)
- 窗口函数 (ROW_NUMBER, RANK, LAG, LEAD)
- JSON 函数
- 全文搜索函数

**语法特性**:
- RETURNING 子句
- UPSERT (ON CONFLICT)
- 递归 CTE
- 窗口函数

### 6. SQL 优化建议

```go
suggestions, err := parser.GetOptimizationSuggestions(sql)

// 返回优化建议
type OptimizationSuggestion struct {
    Type       string // "performance", "readability", "best_practice"
    Severity   string // "info", "warning", "error"
    Message    string
    Suggestion string
}
```

**优化建议类型**:
- **性能优化**: 索引建议、查询重写建议
- **可读性改进**: 代码格式、命名规范
- **最佳实践**: 安全性、维护性建议

### 7. 复杂度分析

```go
type ComplexityMetrics struct {
    Score           int // 总体复杂度分数
    TableCount      int // 涉及的表数量
    JoinCount       int // JOIN 数量
    SubqueryCount   int // 子查询数量
    CTECount        int // CTE 数量
    WindowFuncCount int // 窗口函数数量
    NestingLevel    int // 最大嵌套级别
}
```

### 8. 表依赖关系分析

```go
dependencies, err := parser.AnalyzeTableDependencies(sql)

// 返回表之间的依赖关系
// map[string][]string{
//     "users": ["posts", "comments"],
//     "posts": ["categories"]
// }
```

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "go-sharding/pkg/parser"
)

func main() {
    // 创建增强解析器
    enhancedParser := parser.NewPostgreSQLEnhancedParser()
    
    sql := `
        WITH user_stats AS (
            SELECT user_id, COUNT(*) as post_count 
            FROM posts 
            GROUP BY user_id
        )
        SELECT u.name, us.post_count,
               ROW_NUMBER() OVER (ORDER BY us.post_count DESC) as rank
        FROM users u
        JOIN user_stats us ON u.id = us.user_id
        WHERE u.active = true
    `
    
    // 分析 SQL
    analysis, err := enhancedParser.AnalyzeSQL(sql)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("语句类型: %s\n", analysis.Type)
    fmt.Printf("涉及表: %v\n", analysis.Tables)
    fmt.Printf("CTE 数量: %d\n", len(analysis.CTEs))
    fmt.Printf("窗口函数数量: %d\n", len(analysis.WindowFunctions))
    fmt.Printf("复杂度分数: %d\n", analysis.Complexity.Score)
}
```

### 分片重写示例

```go
// 原始 SQL
sql := `
    SELECT u.name, p.title 
    FROM users u 
    JOIN posts p ON u.id = p.user_id 
    WHERE u.id = $1
`

// 分片规则
shardingRules := map[string]string{
    "users": "users_shard_1",
    "posts": "posts_shard_1",
}

// 重写 SQL
rewrittenSQL, err := enhancedParser.RewriteForSharding(sql, shardingRules)
// 结果: SELECT u.name, p.title FROM users_shard_1 u JOIN posts_shard_1 p ON u.id = p.user_id WHERE u.id = $1
```

### 复杂查询分析示例

```go
sql := `
    WITH RECURSIVE employee_hierarchy AS (
        SELECT id, name, manager_id, 1 as level
        FROM employees 
        WHERE manager_id IS NULL
        UNION ALL
        SELECT e.id, e.name, e.manager_id, eh.level + 1
        FROM employees e
        JOIN employee_hierarchy eh ON e.manager_id = eh.id
    )
    SELECT eh.name, eh.level,
           COUNT(*) OVER (PARTITION BY eh.level) as peers_count
    FROM employee_hierarchy eh
    ORDER BY eh.level, eh.name
`

analysis, err := enhancedParser.AnalyzeSQL(sql)
if err == nil {
    fmt.Printf("递归 CTE: %v\n", analysis.CTEs[0].Recursive)
    fmt.Printf("窗口函数: %v\n", analysis.WindowFunctions)
    fmt.Printf("嵌套级别: %d\n", analysis.Complexity.NestingLevel)
}
```

## 性能特点

- **高效解析**: 基于成熟的 CockroachDB Parser，解析性能优异
- **内存优化**: 合理的内存使用，支持大型 SQL 语句
- **缓存机制**: 内置解析结果缓存，提高重复查询性能
- **并发安全**: 线程安全的设计，支持并发使用

## 错误处理

增强解析器提供了完善的错误处理机制：

```go
analysis, err := parser.AnalyzeSQL(invalidSQL)
if err != nil {
    // 处理解析错误
    fmt.Printf("解析失败: %v\n", err)
    return
}

// 检查验证错误
validationErrors, err := parser.ValidateComplexSQL(sql)
if err == nil && len(validationErrors) > 0 {
    for _, verr := range validationErrors {
        fmt.Printf("验证错误: [%s] %s\n", verr.Type, verr.Message)
    }
}
```

## 扩展性

增强解析器设计为可扩展的架构：

- **插件化优化器**: 可以添加自定义的优化规则
- **自定义分析器**: 支持添加特定的分析逻辑
- **方言支持**: 可以扩展支持其他 PostgreSQL 方言

## 最佳实践

1. **合理使用缓存**: 对于重复的 SQL 语句，利用解析结果缓存
2. **分批处理**: 对于大量 SQL 语句，考虑分批处理以控制内存使用
3. **错误处理**: 始终检查解析错误和验证错误
4. **性能监控**: 监控解析性能，特别是对于复杂的 SQL 语句

## 限制和注意事项

- **PostgreSQL 兼容性**: 主要针对 PostgreSQL 语法，其他数据库可能有兼容性问题
- **复杂度限制**: 极其复杂的 SQL 语句可能影响解析性能
- **内存使用**: 大型 SQL 语句会消耗更多内存

## 未来计划

- **更多优化规则**: 添加更多的 SQL 优化建议
- **可视化支持**: 提供 SQL 结构的可视化展示
- **性能分析**: 集成查询性能分析功能
- **多数据库支持**: 扩展支持 MySQL、Oracle 等数据库