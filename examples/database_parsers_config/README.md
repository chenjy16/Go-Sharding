# 数据库解析器配置使用方式示例

本示例演示了 `go-sharding` 框架中 MySQL 和 PostgreSQL 数据库 SQL 解析器的各种配置使用方式。

## 📁 文件结构

```
database_parsers_config/
├── main.go                          # 主演示程序
├── mysql_parser_config.yaml         # MySQL 解析器配置文件
├── postgresql_parser_config.yaml    # PostgreSQL 解析器配置文件
├── mixed_parser_config.yaml         # 混合解析器配置文件
└── README.md                        # 说明文档
```

## 🚀 快速开始

### 运行示例

```bash
cd examples/database_parsers_config
go run main.go
```

## 📊 演示内容

### 1. MySQL TiDB 解析器配置

演示如何配置和使用 MySQL TiDB 解析器：

- **默认配置启用**：使用内置默认配置
- **自定义配置**：通过代码设置自定义配置
- **MySQL 特有语法支持**：
  - 反引号标识符 (\`table\`)
  - AUTO_INCREMENT
  - MySQL 风格的 LIMIT offset, count
  - MySQL 特有函数和存储引擎语法

### 2. PostgreSQL 基础解析器配置

演示 PostgreSQL 基础解析器的使用：

- **创建解析器实例**
- **PostgreSQL 特有语法支持**：
  - 参数占位符 ($1, $2, ...)
  - RETURNING 子句
  - JSONB 操作符 (@>, ->>, 等)
  - 数组操作
  - 类型转换 (::)
  - 窗口函数

### 3. PostgreSQL 增强解析器配置

演示 PostgreSQL 增强解析器的高级功能：

- **深度 AST 分析**
- **复杂 SQL 支持**：
  - CTE (Common Table Expression)
  - 复杂 JOIN 查询
  - 子查询分析
- **SQL 优化建议**
- **性能分析和复杂度计算**

### 4. 从配置文件初始化解析器

演示如何从 YAML 配置文件初始化解析器：

- `mysql_parser_config.yaml` - MySQL 解析器配置
- `postgresql_parser_config.yaml` - PostgreSQL 解析器配置
- `mixed_parser_config.yaml` - 混合解析器配置

### 5. 从环境变量初始化解析器

演示如何通过环境变量配置解析器：

```bash
export ENABLE_TIDB_PARSER=true
export ENABLE_POSTGRESQL_PARSER=false
export AUTO_ENABLE_TIDB=true
export FALLBACK_TO_ORIGINAL=true
export ENABLE_BENCHMARKING=true
export LOG_PARSING_ERRORS=true
```

### 6. 动态切换解析器

演示运行时动态切换不同解析器：

- TiDB 解析器 ↔ PostgreSQL 解析器
- 配置更新和状态验证

### 7. 解析器性能对比

演示解析器性能基准测试：

- 启用性能监控
- 解析统计信息
- 性能指标对比

## ⚙️ 配置文件详解

### MySQL 解析器配置 (`mysql_parser_config.yaml`)

```yaml
parser:
  enable_tidb_parser: true
  enable_postgresql_parser: false
  fallback_to_original: true
  enable_benchmarking: true
  log_parsing_errors: true

mysql:
  enableAdvancedFeatures: true
  dialect: "mysql"
  charset: "utf8mb4"
  version: "8.0"
  
  features:
    backtickIdentifiers: true
    autoIncrement: true
    mysqlLimitSyntax: true
    mysqlFunctions: true
    storageEngines: true
```

### PostgreSQL 解析器配置 (`postgresql_parser_config.yaml`)

```yaml
parser:
  enable_tidb_parser: false
  enable_postgresql_parser: true
  fallback_to_original: true
  enable_benchmarking: true
  log_parsing_errors: true

postgresql:
  enableAdvancedFeatures: true
  enableEnhancedParser: true
  dialect: "postgresql"
  defaultSchema: "public"
  version: "14.0"
  
  features:
    jsonb: true
    arrays: true
    fullTextSearch: true
    windowFunctions: true
    cte: true
    returning: true
    customTypes: true
    extensions: true
    parameterPlaceholders: true
    upsert: true
```

### 混合解析器配置 (`mixed_parser_config.yaml`)

支持同时使用 MySQL 和 PostgreSQL 解析器，包括：

- **自动方言检测**
- **基于表名的路由规则**
- **多数据源配置**
- **分离的分片规则**
- **性能优化配置**

## 🔧 环境变量配置

支持的环境变量：

| 环境变量 | 描述 | 默认值 |
|---------|------|--------|
| `ENABLE_TIDB_PARSER` | 启用 TiDB 解析器 | `true` |
| `ENABLE_POSTGRESQL_PARSER` | 启用 PostgreSQL 解析器 | `false` |
| `AUTO_ENABLE_TIDB` | 自动启用 TiDB 作为默认解析器 | `true` |
| `FALLBACK_TO_ORIGINAL` | 解析失败时回退到原始解析器 | `true` |
| `ENABLE_BENCHMARKING` | 启用性能基准测试 | `true` |
| `LOG_PARSING_ERRORS` | 记录解析错误 | `true` |

## 📈 性能监控

示例包含性能监控功能：

- **解析时间统计**
- **成功/失败率统计**
- **内存使用监控**
- **缓存命中率**

## 🧪 测试用例

### MySQL 测试 SQL

```sql
-- 基本查询
SELECT * FROM users WHERE id = 1

-- MySQL 风格的 LIMIT
SELECT * FROM users LIMIT 10, 20

-- 反引号标识符
SELECT * FROM `users` WHERE `name` = 'John'

-- AUTO_INCREMENT
CREATE TABLE test (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))
```

### PostgreSQL 测试 SQL

```sql
-- 参数占位符
SELECT * FROM users WHERE id = $1

-- RETURNING 子句
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id

-- JSONB 操作
SELECT username, profile->>'age' as age FROM users WHERE profile @> '{"city": "Beijing"}'

-- 数组操作
UPDATE users SET tags = array_append(tags, $1) WHERE user_id = $2

-- 窗口函数
SELECT username, ROW_NUMBER() OVER (ORDER BY created_at DESC) as rank FROM users

-- CTE 查询
WITH active_users AS (
    SELECT id, name FROM users WHERE active = true
)
SELECT * FROM active_users
```

## 🚨 注意事项

1. **配置文件路径**：确保配置文件在正确的路径下
2. **数据库连接**：配置文件中的数据库连接信息需要根据实际环境调整
3. **解析器兼容性**：不同解析器对 SQL 方言的支持程度不同
4. **性能影响**：启用详细日志和性能监控可能影响性能
5. **内存使用**：复杂 SQL 的 AST 分析会消耗更多内存

## 🔗 相关示例

- `examples/enable_tidb_parser/` - TiDB 解析器启用示例
- `examples/config_file_parser/` - 配置文件解析器示例
- `examples/postgresql_parser/` - PostgreSQL 解析器示例
- `examples/postgresql_enhanced_parser/` - PostgreSQL 增强解析器示例
- `examples/postgresql_config/` - PostgreSQL 配置示例

## 📚 更多信息

- [Go-Sharding 文档](../../README.md)
- [解析器架构设计](../../docs/parser_architecture.md)
- [PostgreSQL 增强功能](../../docs/postgresql_enhanced_features.md)