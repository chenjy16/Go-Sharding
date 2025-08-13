# Go-Sharding

Go Language Database Sharding Middleware - High-Performance Sharding Solution Based on Apache ShardingSphere Design Concepts

## 📋 Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Architecture Design](#architecture-design)
- [Core Components](#core-components)
- [Database Support](#database-support)
  - [MySQL Support](#mysql-support)
  - [PostgreSQL Support](#postgresql-support)
- [SQL Parser](#sql-parser)
  - [Parser Configuration and Enabling](#parser-configuration-and-enabling)
- [Sharding Strategies](#sharding-strategies)
- [Read-Write Splitting](#read-write-splitting)
- [Transaction Management](#transaction-management)
- [Configuration](#configuration)
- [Example Code](#example-code)
- [Performance Optimization](#performance-optimization)
- [Test Coverage](#test-coverage)
- [Deployment & Operations](#deployment--operations)
- [Development Guide](#development-guide)
- [Contributing](#contributing)

## 🚀 Features

### Core Features
- ✅ **Database and Table Sharding**: Support horizontal sharding to improve data processing capabilities
- ✅ **Multiple Sharding Algorithms**: Modulo, range, hash, and custom algorithms
- ✅ **Cross-Shard Queries and Aggregation**: Intelligent routing and result merging
- ✅ **Distributed Primary Key Generation**: Snowflake algorithm ensures global uniqueness
- ✅ **Read-Write Splitting**: Automatic routing for master-slave databases to improve performance
- ✅ **Distributed Transactions**: Support local transactions, XA transactions, and BASE transactions
- ✅ **SQL Routing and Rewriting**: Intelligent SQL parsing and rewriting
- ✅ **Result Merging**: Support sorting, grouping, aggregation, and pagination
- ✅ **Monitoring and Metrics Collection**: Complete performance monitoring system

### Database Support
- ✅ **MySQL**: Complete support including complex queries and transactions
- ✅ **PostgreSQL**: Comprehensive support including unique features
  - JSONB data type support
  - Array type support
  - Full-text search (tsvector/tsquery)
  - Window functions
  - CTE (Common Table Expressions)
  - RETURNING clause
  - Parameter placeholder conversion (? → $1, $2, ...)

### Advanced Features
- ✅ **Multi-Parser Architecture**: Support native, TiDB, PostgreSQL, and enhanced parsers
- ✅ **Intelligent Routing**: Automatic routing based on sharding keys
- ✅ **Connection Pool Management**: Optimized database connection pools
- ✅ **Health Checks**: Real-time monitoring of data source status
- ✅ **Hot Configuration Updates**: Support runtime configuration updates

## 🏃‍♂️ Quick Start

### Installation

```bash
go get github.com/your-username/go-sharding
```

### Basic Usage

```go
package main

import (
    "go-sharding/pkg/config"
    "go-sharding/pkg/sharding"
    "log"
)

func main() {
    // Create data source configuration
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

    // Create sharding rule configuration
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

    // Create sharding configuration
    shardingConfig := &config.ShardingConfig{
        DataSources:  dataSources,
        ShardingRule: shardingRule,
    }

    // Create sharding data source
    dataSource, err := sharding.NewShardingDataSource(shardingConfig)
    if err != nil {
        log.Fatalf("Failed to create sharding data source: %v", err)
    }
    defer dataSource.Close()

    // Get database connection
    db := dataSource.DB()

    // Execute SQL
    result, err := db.Exec("INSERT INTO t_user (user_name, user_email) VALUES (?, ?)", "John Doe", "john@example.com")
    if err != nil {
        log.Printf("Insert failed: %v", err)
    }
}
```

### Run Demo

```bash
# Build demo program
go build -o bin/go-sharding-demo ./cmd/demo

# Run demo
./bin/go-sharding-demo
```

## 🏗️ Architecture Design

### Overall Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│                Go-Sharding Middleware                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │ Routing     │ │ SQL         │ │ Result      │ │ ID      │ │
│  │ Engine      │ │ Rewriter    │ │ Merger      │ │Generator│ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                Configuration Manager                     │ │
│  └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                  Database Driver Layer                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │ Database 1  │ │ Database 2  │ │ Database 3  │ │   ...   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Core Design Principles

1. **High Performance**: Optimized SQL parsing and routing algorithms
2. **High Availability**: Support failover and load balancing
3. **Easy Extension**: Modular design supporting custom extensions
4. **Transparency**: Transparent to applications, no business code modification required

## 🔧 Core Components

### 1. Configuration Manager

Manages sharding rules, data source configurations, etc.

**Main Functions:**
- Data source configuration management
- Sharding rule configuration
- Read-write splitting configuration
- YAML/JSON configuration file support
- Configuration validation

### 2. Routing Engine

Calculates target data sources and tables based on sharding rules and SQL parameters.

**Main Functions:**
- Sharding key extraction
- Sharding algorithm execution
- Route result calculation
- Support for multiple sharding strategies

### 3. SQL Rewriter

Rewrites logical SQL into physical SQL for actual data sources.

**Main Functions:**
- Replace logical table names with actual table names
- Generate multi-table UNION queries
- SQL syntax parsing and reconstruction
- Parameter binding handling

### 4. Result Merger

Merges query results from multiple shards into a unified result set.

**Main Functions:**
- Streaming result merging
- Sort merging (ORDER BY)
- Group aggregation (GROUP BY)
- Pagination handling (LIMIT/OFFSET)
- Aggregate function calculation

### 5. ID Generator

Generates globally unique primary keys for sharded tables.

**Supported Algorithms:**
- Snowflake algorithm
- UUID generation
- Auto-increment sequences
- Custom generators

## 🗄️ Database Support

### MySQL Support

Complete support for MySQL database, including:
- Standard SQL syntax
- MySQL-specific functions
- Transaction support
- Connection pool management

### PostgreSQL Support

Comprehensive support for PostgreSQL database and its unique features:

#### Unique Feature Support
- **JSONB Data Type**: Complete JSON operation support
- **Array Types**: Array operations and functions
- **Full-Text Search**: tsvector/tsquery support
- **Window Functions**: Complete window function support
- **CTE**: Common Table Expressions
- **RETURNING Clause**: INSERT/UPDATE/DELETE return values
- **Custom Data Types**: User-defined type support
- **Parameter Placeholder Conversion**: Automatic conversion of ? to $1, $2, ...

#### PostgreSQL Quick Start

```bash
# Start PostgreSQL cluster
docker-compose -f docker-compose-postgresql.yml up -d

# Run test script
./scripts/test-postgresql.sh

# Run PostgreSQL example
cd examples/postgresql && go run main.go
```

#### PostgreSQL Code Examples

```go
// JSONB queries
rows, err := ds.QueryContext(ctx, `
    SELECT username, address->>'city' as city 
    FROM user 
    WHERE address @> '{"city": "Beijing"}'`)

// Array operations
_, err = ds.ExecContext(ctx, `
    UPDATE user 
    SET tags = array_append(tags, ?) 
    WHERE user_id = ?`, "new_tag", userID)

// Full-text search
rows, err := ds.QueryContext(ctx, `
    SELECT username, email 
    FROM user 
    WHERE search_vector @@ to_tsquery('english', ?)`, "john")

// Window functions
rows, err := ds.QueryContext(ctx, `
    SELECT 
        username,
        total_amount,
        ROW_NUMBER() OVER (ORDER BY total_amount DESC) as rank
    FROM user_order_summary`)

// RETURNING clause
var newOrderID int64
err = ds.QueryRowContext(ctx, `
    INSERT INTO order_table (user_id, product_name, total_amount) 
    VALUES (?, ?, ?) 
    RETURNING order_id`, userID, "Product", 99.99).Scan(&newOrderID)
```

#### PostgreSQL Enhanced Parser Features

The PostgreSQL Enhanced Parser provides advanced SQL analysis capabilities:

```go
// Create enhanced parser
enhancedParser := parser.NewPostgreSQLEnhancedParser()

// Deep SQL analysis
analysis, err := enhancedParser.AnalyzeSQL(`
    WITH RECURSIVE employee_hierarchy AS (
        SELECT id, name, manager_id, 1 as level
        FROM employees WHERE manager_id IS NULL
        UNION ALL
        SELECT e.id, e.name, e.manager_id, eh.level + 1
        FROM employees e
        JOIN employee_hierarchy eh ON e.manager_id = eh.id
    )
    SELECT eh.name, eh.level,
           COUNT(*) OVER (PARTITION BY eh.level) as peers_count
    FROM employee_hierarchy eh
    ORDER BY eh.level, eh.name
`)

if err == nil {
    fmt.Printf("Query Type: %s\n", analysis.Type)
    fmt.Printf("Tables: %v\n", analysis.Tables)
    fmt.Printf("CTEs: %d (Recursive: %v)\n", len(analysis.CTEs), analysis.CTEs[0].Recursive)
    fmt.Printf("Window Functions: %d\n", len(analysis.WindowFunctions))
    fmt.Printf("Complexity Score: %d\n", analysis.Complexity.Score)
    fmt.Printf("Optimization Suggestions: %d\n", len(analysis.Optimizations))
}

// Table dependency analysis
dependencies, err := enhancedParser.AnalyzeTableDependencies(sql)
if err == nil {
    fmt.Printf("Table Dependencies: %v\n", dependencies)
}

// SQL optimization suggestions
suggestions, err := enhancedParser.GetOptimizationSuggestions(sql)
if err == nil {
    for _, suggestion := range suggestions {
        fmt.Printf("[%s] %s: %s\n", 
            suggestion.Severity, suggestion.Type, suggestion.Message)
    }
}
```

**Enhanced Parser Capabilities:**
- **Deep AST Analysis**: Complete syntax tree analysis
- **CTE Support**: Regular and recursive Common Table Expressions
- **Window Function Analysis**: ROW_NUMBER, RANK, LAG, LEAD, etc.
- **Subquery Detection**: Scalar, WHERE, FROM subqueries
- **Join Analysis**: INNER, LEFT, RIGHT, FULL OUTER joins
- **Complexity Metrics**: Query complexity scoring
- **Optimization Suggestions**: Performance and best practice recommendations
- **Table Dependencies**: Automatic relationship detection

## 🔍 SQL Parser

### Multi-Parser Architecture

The project adopts a multi-layer parser architecture supporting different parsing strategies:

#### 1. Original Parser
- **Technical Implementation**: Based on regular expressions
- **Performance Characteristics**: Lightweight, fast startup
- **Use Cases**: Simple SQL statements
- **Compatibility**: MySQL 85%, PostgreSQL 75%

#### 2. TiDB Parser
- **Technical Implementation**: Integrates `pingcap/tidb/pkg/parser`
- **Performance Characteristics**: High performance, low memory usage
- **Use Cases**: Complex MySQL queries
- **Compatibility**: MySQL 98%+

**Performance Comparison:**
| Test Scenario | Original Parser | TiDB Parser | PostgreSQL Enhanced | Performance Improvement |
|---------------|----------------|-------------|--------------------|-----------------------|
| Simple Query | 70μs | 5μs | 73μs | **TiDB: 14x faster** |
| Complex JOIN | 150μs | 25μs | 163μs | **TiDB: 6x faster** |
| INSERT Statement | 80μs | 8μs | - | **TiDB: 10x faster** |
| Window Functions | - | - | 94μs | **Specialized support** |
| CTE Queries | - | - | 163μs | **Advanced analysis** |
| Memory Usage | 101,300 B/op | 3,993 B/op | 82,287 B/op | **TiDB: 96% reduction** |

#### 3. PostgreSQL Parser
- **Technical Implementation**: Based on CockroachDB Parser
- **Features**: Supports PostgreSQL-specific syntax and advanced features
- **Use Cases**: PostgreSQL databases with complex queries
- **Compatibility**: PostgreSQL 95%+

#### 4. PostgreSQL Enhanced Parser
- **Technical Implementation**: Advanced PostgreSQL parser with deep AST analysis
- **Features**: 
  - **Deep SQL Analysis**: CTE, window functions, subqueries
  - **Table Dependency Analysis**: Automatic dependency detection
  - **SQL Optimization Suggestions**: Performance and best practice recommendations
  - **Complex Query Support**: Recursive CTE, advanced joins, JSONB operations
- **Performance**: Optimized for complex PostgreSQL queries
- **Use Cases**: Enterprise PostgreSQL applications

#### 5. Enhanced Parser
- **Technical Implementation**: Integrates multiple parsers
- **Features**: Intelligently selects the most suitable parser
- **Use Cases**: Mixed database environments

### Parser Configuration and Enabling

#### Configuration File Method (Recommended)

Create a `config.yaml` configuration file:

```yaml
parser:
  # Enable TiDB parser as default parser
  enable_tidb_parser: true
  # Enable PostgreSQL parser
  enable_postgresql_parser: false
  # Whether to fallback to original parser when parsing fails
  fallback_to_original: true
  # Enable performance benchmarking
  enable_benchmarking: true
  # Log parsing errors
  log_parsing_errors: true
```

Initialize with just one line in code:

```go
import "go-sharding/pkg/parser"

// Initialize parser from config file (simplest way)
err := parser.InitializeParserFromConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Now the parser is set up according to the config file
stmt, err := parser.DefaultParserFactory.Parse("SELECT * FROM users")
```

#### Programmatic Configuration

```go
// Method 1: Directly enable TiDB parser
err := parser.EnableTiDBParserAsDefault()
if err != nil {
    log.Fatal(err)
}

// Method 2: Use configuration struct
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

// Method 3: Environment variable configuration
// Set environment variable: ENABLE_TIDB_PARSER=true
err := parser.InitializeParserFromEnv()
if err != nil {
    log.Fatal(err)
}
```

#### Verify Configuration

```go
// Check current default parser
parserType := parser.GetDefaultParserType()
fmt.Printf("Current default parser: %s\n", parserType) // Should output: tidb

// Print detailed information
parser.PrintParserInfo()

// Get statistics
stats := parser.GetParserFactoryStats()
fmt.Printf("Parser statistics: %+v\n", stats)
```

#### Configuration Priority

Parser configuration priority order (from high to low):

1. **Direct code calls** - `parser.EnableTiDBParserAsDefault()`
2. **Environment variables** - `ENABLE_TIDB_PARSER=true`
3. **Configuration file** - `parser` configuration in `config.yaml`
4. **Default configuration** - System default settings

### Parser Factory Pattern

```go
// Create parser
parser := parser.NewParserFactory().CreateParser("tidb")

// Parse SQL
stmt, err := parser.Parse("SELECT * FROM users WHERE id = ?")

// Extract table names
tables := parser.ExtractTables(sql)
```

## 📊 Sharding Strategies

### 1. Database Sharding

Distribute data to different database instances based on sharding keys.

```yaml
databaseStrategy:
  type: inline
  shardingColumn: user_id
  algorithm: "ds_${user_id % 2}"
```

### 2. Table Sharding

Distribute data to different tables within the same database.

```yaml
tableStrategy:
  type: inline
  shardingColumn: order_id
  algorithm: "t_order_${order_id % 4}"
```

### 3. Compound Sharding

Perform both database sharding and table sharding simultaneously.

```yaml
actualDataNodes: "ds_${0..1}.t_order_${0..3}"
databaseStrategy:
  shardingColumn: user_id
  algorithm: "ds_${user_id % 2}"
tableStrategy:
  shardingColumn: order_id
  algorithm: "t_order_${order_id % 4}"
```

### Supported Sharding Algorithms

- **Modulo Sharding**: `ds_${user_id % 2}`
- **Range Sharding**: `ds_${user_id / 1000}`
- **Hash Sharding**: `ds_${hash(user_id) % 4}`
- **Custom Algorithm**: Implement `ShardingAlgorithm` interface

## 🔄 Read-Write Splitting

Support read-write splitting for master-slave databases to improve system performance.

### Configuration Example

```yaml
readWriteSplits:
  rw_ds_0:
    masterDataSource: ds_0_master
    slaveDataSources:
      - ds_0_slave_0
      - ds_0_slave_1
    loadBalanceAlgorithm: round_robin
```

### Load Balancing Algorithms

- **Round Robin**: Access slave databases in turn
- **Random**: Randomly select slave databases
- **Weighted Round Robin**: Weight-based round robin

### Usage Example

```go
// Create read-write splitter
splitter, err := readwrite.NewReadWriteSplitter(rwConfig, dataSources)

// Auto-route queries (read operations -> slave)
db := splitter.Route("SELECT * FROM users WHERE id = ?")

// Auto-route write operations (write operations -> master)
db := splitter.Route("INSERT INTO users (name) VALUES (?)")

// Force use master
ctx := context.WithValue(context.Background(), "force_master", true)
db := splitter.RouteContext(ctx, "SELECT * FROM users WHERE id = ?")
```

## 💾 Transaction Management

### 1. Local Transactions

Transactions within a single shard, directly using database local transactions.

```go
tx, err := db.Begin()
if err != nil {
    return err
}

// Execute operations
_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
if err != nil {
    tx.Rollback()
    return err
}

// Commit transaction
return tx.Commit()
```

### 2. XA Distributed Transactions

Strong consistency transactions across shards using two-phase commit protocol.

```go
// Begin XA transaction
tx, err := tm.Begin(ctx, transaction.XATransaction)
if err != nil {
    return err
}

// Execute cross-shard operations
err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
if err != nil {
    tx.Rollback()
    return err
}

// Commit transaction
return tx.Commit()
```

### 3. BASE Transactions

Eventually consistent distributed transactions suitable for scenarios with relaxed consistency requirements.

#### BASE Transaction Characteristics

- **Basically Available**: System maintains core functionality even during failures
- **Soft state**: Allows intermediate states without requiring real-time consistency
- **Eventually consistent**: System eventually reaches a consistent state

#### Usage Example

```go
// Create transaction manager
tm := transaction.NewTransactionManager()
defer tm.Close()

// Begin BASE transaction
ctx := context.Background()
tx, err := tm.Begin(ctx, transaction.BaseTransaction)
if err != nil {
    log.Fatalf("Failed to begin BASE transaction: %v", err)
}

baseTx := tx.(*transaction.BASETransactionImpl)

// Add operation
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

// Add compensation operation
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

// Commit transaction
err := baseTx.Commit(ctx)
if err != nil {
    log.Fatalf("Failed to commit transaction: %v", err)
}
```

#### Transaction State Management

- **StatusActive (0)**: Transaction active state, can add operations
- **StatusPrepared (1)**: Transaction executing
- **StatusCommitted (2)**: Transaction successfully committed
- **StatusRolledBack (3)**: Transaction rolled back
- **StatusFailed (4)**: Transaction execution failed

### Transaction Type Comparison

| Feature | LOCAL Transaction | XA Transaction | BASE Transaction |
|---------|------------------|----------------|------------------|
| Consistency | Strong | Strong | Eventual |
| Performance | High | Medium | High |
| Availability | Medium | Low | High |
| Complexity | Low | High | Medium |
| Use Cases | Single data source | Multi-source strong consistency | Multi-source eventual consistency |

## ⚙️ Configuration

### Data Source Configuration

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

### Sharding Rule Configuration

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

### PostgreSQL-Specific Configuration

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

## 📝 Example Code

Check example code in the `examples/` directory:

### Basic Examples
- `examples/basic/` - Basic usage example
- `examples/yaml_config/` - YAML configuration example

### Parser Examples
- `examples/enable_tidb_parser/` - TiDB parser enabling example
- `examples/config_file_parser/` - Configuration file parser setup example

### Database Examples
- `examples/postgresql/` - PostgreSQL usage example
- `examples/postgresql_config/` - PostgreSQL configuration example
- `examples/postgresql_parser/` - PostgreSQL parser example
- `examples/postgresql_enhanced_parser/` - **PostgreSQL Enhanced Parser example**
- `examples/cockroachdb_adapter/` - CockroachDB adapter example

### Transaction Examples
- `examples/base_transaction/` - BASE transaction usage example

### Quick Start Examples

#### 1. Basic Sharding Usage

```bash
cd examples/basic
go run main.go
```

#### 2. Enable TiDB Parser

```bash
cd examples/enable_tidb_parser
go run main.go
```

#### 3. Configuration File Parser Setup

```bash
cd examples/config_file_parser
go run main.go
```

#### 4. PostgreSQL Support

```bash
# Start PostgreSQL cluster
docker-compose -f docker-compose-postgresql.yml up -d

# Run example
cd examples/postgresql
go run main.go
```

#### 5. PostgreSQL Enhanced Parser Example

```bash
cd examples/postgresql_enhanced_parser
go run main.go
```

#### 6. BASE Transaction Example

```bash
cd examples/base_transaction
go run main.go
```

### Enhanced Features Example

```go
// Create enhanced sharding database
db, err := sharding.NewEnhancedShardingDB(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Health check
if err := db.HealthCheck(); err != nil {
    log.Printf("Health check failed: %v", err)
}

// Execute query (auto sharding + read-write splitting)
rows, err := db.QueryContext(ctx, 
    "SELECT * FROM t_order WHERE user_id = ?", userID)

// Execute write operation (auto sharding + master routing)
result, err := db.ExecContext(ctx,
    "INSERT INTO t_order (user_id, amount) VALUES (?, ?)", 
    userID, amount)
```

## 🚀 Performance Optimization

### 1. Connection Pool Management

- Independent connection pools for each data source
- Configurable maximum connections and idle connections
- Connection reuse and automatic recycling

### 2. Query Optimization

- SQL parsing cache
- Route result cache
- Prepared statement support

### 3. Streaming Result Processing

- Streaming merge for large result sets
- Memory usage optimization
- Pagination query optimization

### 4. Parser Performance

**TiDB Parser** performance improvements over original parser:
- **Parsing Speed**: 5-20x improvement
- **Memory Usage**: 90%+ reduction
- **CPU Usage**: 80-90% reduction

**PostgreSQL Enhanced Parser** performance characteristics:
- **Simple Queries**: ~73μs per operation
- **Complex Queries**: ~163μs per operation
- **Window Functions**: ~94μs per operation
- **Memory Efficiency**: ~82KB per operation
- **Advanced Features**: Deep AST analysis with minimal overhead

## 🧪 Test Coverage

### Test Coverage Statistics

- **Overall Statement Coverage**: 58.9%
- **Transaction Package Coverage**: 75.8%

### Package Test Status

- ✅ `algorithm` - Complete test suite with comprehensive coverage
- ✅ `config` - Configuration validation and parser tests
- ✅ `database` - Database type and dialect tests
- ✅ `executor` - Complete execution plan test suite
- ✅ `id` - ID generator tests with performance benchmarks
- ✅ `merge` - Result merger tests with complex scenarios
- ✅ `monitoring` - Metrics collection and monitoring tests
- ✅ `optimizer` - SQL optimizer comprehensive test suite
- ✅ `parser` - **Enhanced parser test suite** including:
  - PostgreSQL Enhanced Parser comprehensive tests
  - CockroachDB Adapter tests
  - TiDB Parser performance tests
  - Multi-parser factory tests
- ✅ `readwrite` - Read-write splitting tests
- ✅ `rewrite` - SQL rewriting tests
- ✅ `routing` - Routing engine tests
- ✅ `sharding` - Enhanced sharding tests including PostgreSQL support
- ✅ `transaction` - Complete transaction management tests

### Running Tests

```bash
# Run all tests
go test ./...

# Run core package tests
go test ./pkg/...

# Generate coverage report
go test -v -coverprofile=coverage.out ./pkg/...
go tool cover -html=coverage.out -o coverage.html
```

## 🚢 Deployment & Operations

### Docker Deployment

#### MySQL Environment

```bash
# Start MySQL cluster
docker-compose up -d

# Check service status
docker-compose ps
```

#### PostgreSQL Environment

```bash
# Start PostgreSQL cluster
docker-compose -f docker-compose-postgresql.yml up -d

# Run test script
./scripts/test-postgresql.sh
```

### Monitoring Metrics

- SQL execution time statistics
- Connection pool status monitoring
- Sharding routing statistics
- Error rate monitoring

### Management Interface

- **pgAdmin** (PostgreSQL): http://localhost:8080
- **Prometheus Monitoring**: 
  - DS0: http://localhost:9187/metrics
  - DS1: http://localhost:9188/metrics

## 👨‍💻 Development Guide

### Project Structure

```
go-sharding/
├── cmd/                    # Command line tools
├── pkg/                    # Core packages
│   ├── algorithm/          # Sharding algorithms
│   ├── config/            # Configuration management
│   ├── database/          # Database management
│   ├── executor/          # Executors
│   ├── id/                # ID generators
│   ├── merge/             # Result merging
│   ├── monitoring/        # Monitoring metrics
│   ├── optimizer/         # Query optimizer
│   ├── parser/            # **Enhanced SQL parsers**
│   │   ├── postgresql_enhanced_parser.go  # PostgreSQL Enhanced Parser
│   │   ├── cockroachdb_adapter.go         # CockroachDB Adapter
│   │   ├── tidb_parser.go                 # TiDB Parser
│   │   └── parser_factory.go              # Multi-parser factory
│   ├── readwrite/         # Read-write splitting
│   ├── rewrite/           # SQL rewriting
│   ├── routing/           # Routing engine
│   ├── sharding/          # **Enhanced sharding management**
│   │   ├── postgresql_datasource.go       # PostgreSQL data source
│   │   └── enhanced_datasource.go         # Enhanced data source
│   └── transaction/       # Transaction management
├── examples/              # **Comprehensive example code**
│   ├── postgresql_enhanced_parser/        # PostgreSQL Enhanced Parser demo
│   ├── cockroachdb_adapter/               # CockroachDB Adapter demo
│   ├── postgresql_config/                 # PostgreSQL configuration demo
│   └── base_transaction/                  # BASE transaction demo
├── scripts/               # Script files
├── docs/                  # **Enhanced documentation**
│   └── postgresql_enhanced_features.md    # PostgreSQL Enhanced Features
├── benchmarks/            # Performance benchmarks
└── docker-compose*.yml    # Docker configuration
```

### Core Interfaces

```go
// Parser interface
type ParserInterface interface {
    Parse(sql string) (*SQLStatement, error)
    ExtractTables(sql string) []string
}

// Router interface
type Router interface {
    Route(logicTable string, shardingValues map[string]interface{}) ([]*RouteResult, error)
}

// Transaction manager interface
type TransactionManager interface {
    Begin(ctx context.Context, txType TransactionType) (Transaction, error)
    Commit(ctx context.Context, tx Transaction) error
    Rollback(ctx context.Context, tx Transaction) error
}
```

### Extension Development

1. **Custom Sharding Algorithm**

```go
type CustomShardingAlgorithm struct{}

func (a *CustomShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) []string {
    // Implement custom sharding logic
    return []string{"target_table"}
}
```

2. **Custom Parser**

```go
type CustomParser struct{}

func (p *CustomParser) Parse(sql string) (*SQLStatement, error) {
    // Implement custom parsing logic
    return &SQLStatement{}, nil
}
```

## 🤝 Contributing

### Contribution Process

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Create a Pull Request

### Code Standards

- Follow Go code conventions
- Add necessary comments and documentation
- Write unit tests
- Ensure tests pass

### Issue Reporting

If you find bugs or have feature suggestions, please create an Issue with:

- Detailed problem description
- Reproduction steps
- Expected behavior
- Actual behavior
- Environment information

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Apache ShardingSphere](https://shardingsphere.apache.org/) - Design concept reference
- [TiDB Parser](https://github.com/pingcap/parser) - SQL parser
- [PostgreSQL](https://www.postgresql.org/) - Database support

## 📞 Contact Us

- Project Homepage: https://github.com/your-username/go-sharding
- Issue Reporting: https://github.com/your-username/go-sharding/issues