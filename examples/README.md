# Go-Sharding 示例集合

本目录包含了 go-sharding 框架的各种功能示例，帮助开发者快速了解和使用框架的核心特性。

## 示例列表

### 1. 负载均衡示例 (`load_balancing/`)

展示了 go-sharding 中读写分离的负载均衡功能：

- **轮询负载均衡**: 按顺序依次分配请求到各个从库
- **随机负载均衡**: 随机选择从库处理请求
- **加权负载均衡**: 根据服务器性能分配不同权重
- **读写分离配置**: 演示如何配置主从数据库的读写分离

```bash
cd examples/load_balancing
go run main.go
```

### 2. SQL 优化器示例 (`sql_optimizer/`)

展示了 SQL 查询优化功能：

- **谓词下推**: 将过滤条件尽早应用以减少数据传输
- **列裁剪**: 只选择需要的列以减少I/O开销
- **索引提示**: 智能建议使用合适的索引
- **JOIN重排序**: 优化多表连接的执行顺序
- **成本估算**: 基于统计信息估算查询执行成本

```bash
cd examples/sql_optimizer
go run main.go
```

### 3. ID 生成器示例 (`id_generator/`)

展示了分布式ID生成功能：

- **雪花算法**: 分布式唯一ID生成，支持多节点部署
- **UUID生成**: 基于随机数的UUID生成
- **自增ID**: 简单的自增ID生成器
- **并发测试**: 验证多协程环境下的ID唯一性
- **性能测试**: 对比不同算法的生成性能
- **唯一性验证**: 确保生成的ID不重复

```bash
cd examples/id_generator
go run main.go
```

## 其他示例

### 4. 数据库解析器配置示例 (`database_parsers_config/`)

展示了 MySQL 和 PostgreSQL 数据库 SQL 解析器的各种配置使用方式：

- **MySQL TiDB 解析器配置**: 启用和配置 TiDB 解析器用于 MySQL
- **PostgreSQL 基础解析器**: PostgreSQL 特有语法解析
- **PostgreSQL 增强解析器**: 深度 AST 分析和 SQL 优化建议
- **配置文件初始化**: 从 YAML 配置文件初始化解析器
- **环境变量配置**: 通过环境变量配置解析器
- **动态解析器切换**: 运行时切换不同解析器
- **性能对比测试**: 解析器性能基准测试

```bash
cd examples/database_parsers_config
go run main.go
```

## 其他示例

### 基础功能示例

- `basic/`: 基本的分库分表功能演示
- `yaml_config/`: 基于YAML配置文件的使用方式
- `config_file_parser/`: 配置文件解析功能
- `read_write_splitting/`: 读写分离的详细实现

### 数据库适配器示例

- `postgresql/`: PostgreSQL 数据库适配器使用
- `postgresql_config/`: PostgreSQL 配置示例
- `postgresql_parser/`: PostgreSQL SQL解析器
- `postgresql_enhanced_parser/`: 增强的PostgreSQL解析器
- `cockroachdb_adapter/`: CockroachDB 适配器

### 高级功能示例

- `monitoring/`: 监控和指标收集
- `base_transaction/`: 基础事务处理
- `enable_tidb_parser/`: TiDB解析器集成

## 运行要求

- Go 1.16 或更高版本
- 根据具体示例可能需要相应的数据库环境（MySQL、PostgreSQL等）

## 快速开始

1. 克隆项目并进入目录：
```bash
git clone <repository-url>
cd go-sharding
```

2. 安装依赖：
```bash
go mod tidy
```

3. 运行任意示例：
```bash
cd examples/<example-name>
go run main.go
```

## 注意事项

- 某些示例可能需要配置数据库连接信息
- 建议按照示例中的注释说明进行配置
- 如遇到问题，请查看项目根目录的文档或提交Issue

## 贡献

欢迎提交新的示例或改进现有示例。请确保：

1. 代码风格一致
2. 包含充分的注释说明
3. 提供运行说明
4. 测试通过

更多详细信息请参考项目主README文件。