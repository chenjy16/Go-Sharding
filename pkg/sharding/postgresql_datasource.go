package sharding

import (
	"context"
	"database/sql"
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/database"
	"go-sharding/pkg/parser"
	"go-sharding/pkg/routing"
	"go-sharding/pkg/rewrite"
	"strings"
	
	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// PostgreSQLShardingDataSource PostgreSQL 分片数据源
type PostgreSQLShardingDataSource struct {
	*ShardingDataSource
	pgParser *parser.PostgreSQLParser
	dialect  database.DatabaseDialect
}

// NewPostgreSQLShardingDataSource 创建 PostgreSQL 分片数据源
func NewPostgreSQLShardingDataSource(cfg *config.ShardingConfig) (*PostgreSQLShardingDataSource, error) {
	// 验证所有数据源都是 PostgreSQL
	for name, dsConfig := range cfg.DataSources {
		dbType, err := database.GlobalDatabaseTypeRegistry.GetDatabaseType(dsConfig.DriverName)
		if err != nil {
			return nil, fmt.Errorf("unsupported driver for data source %s: %w", name, err)
		}
		if dbType != database.PostgreSQL {
			return nil, fmt.Errorf("data source %s is not PostgreSQL, got %s", name, dbType)
		}
	}
	
	// 创建基础分片数据源
	baseDS, err := NewShardingDataSource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create base sharding data source: %w", err)
	}
	
	// 获取 PostgreSQL 方言
	dialect, err := database.GlobalDialectRegistry.GetDialect(database.PostgreSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL dialect: %w", err)
	}
	
	pgDS := &PostgreSQLShardingDataSource{
		ShardingDataSource: baseDS,
		pgParser:           parser.NewPostgreSQLParser(),
		dialect:            dialect,
	}
	
	return pgDS, nil
}

// PostgreSQLDB PostgreSQL 分片数据库
type PostgreSQLDB struct {
	*ShardingDB
	pgDataSource *PostgreSQLShardingDataSource
}

// DB 获取 PostgreSQL 分片数据库连接
func (ds *PostgreSQLShardingDataSource) DB() *PostgreSQLDB {
	return &PostgreSQLDB{
		ShardingDB:   ds.ShardingDataSource.DB(),
		pgDataSource: ds,
	}
}

// QueryContext 执行 PostgreSQL 查询（带上下文）
func (db *PostgreSQLDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*ShardingRows, error) {
	// 验证 PostgreSQL SQL 语法
	if err := db.pgDataSource.pgParser.ValidatePostgreSQLSQL(query); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL SQL: %w", err)
	}
	
	// 解析 PostgreSQL 特定语法
	_, err := db.pgDataSource.pgParser.ParsePostgreSQLSpecific(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL SQL: %w", err)
	}
	
	// 转换参数占位符（PostgreSQL 使用 $1, $2, ... 而不是 ?）
	pgQuery, pgArgs := db.convertToPostgreSQLParams(query, args)
	
	// 调用基础查询方法
	return db.ShardingDB.QueryContext(ctx, pgQuery, pgArgs...)
}

// Query 执行 PostgreSQL 查询
func (db *PostgreSQLDB) Query(query string, args ...interface{}) (*ShardingRows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

// ExecContext 执行 PostgreSQL 命令（带上下文）
func (db *PostgreSQLDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 验证 PostgreSQL SQL 语法
	if err := db.pgDataSource.pgParser.ValidatePostgreSQLSQL(query); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL SQL: %w", err)
	}
	
	// 转换参数占位符
	pgQuery, pgArgs := db.convertToPostgreSQLParams(query, args)
	
	// 提取逻辑表名
	logicTables := db.extractLogicTables(pgQuery)
	if len(logicTables) == 0 {
		// 如果没有分片表，直接在第一个数据源执行
		return db.executeOnFirstDataSource(ctx, pgQuery, pgArgs...)
	}
	
	// 提取分片值
	shardingValues := db.extractShardingValues(pgQuery, pgArgs)
	
	// 路由计算
	var allRouteResults []*routing.RouteResult
	for _, logicTable := range logicTables {
		routeResults, err := db.pgDataSource.router.Route(logicTable, shardingValues)
		if err != nil {
			return nil, fmt.Errorf("routing failed for table %s: %w", logicTable, err)
		}
		allRouteResults = append(allRouteResults, routeResults...)
	}
	
	// SQL 重写
	rewriteCtx := &rewrite.RewriteContext{
		OriginalSQL:  pgQuery,
		LogicTables:  logicTables,
		RouteResults: allRouteResults,
		Parameters:   pgArgs,
	}
	rewriteResults, err := db.pgDataSource.rewriter.Rewrite(rewriteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite SQL: %w", err)
	}
	
	// 执行命令
	var lastResult sql.Result
	for _, rewriteResult := range rewriteResults {
		conn := db.pgDataSource.dataSources[rewriteResult.DataSource]
		result, err := conn.ExecContext(ctx, rewriteResult.SQL, rewriteResult.Parameters...)
		if err != nil {
			return nil, fmt.Errorf("exec failed on %s: %w", rewriteResult.DataSource, err)
		}
		lastResult = result
	}
	
	return lastResult, nil
}

// Exec 执行 PostgreSQL 命令
func (db *PostgreSQLDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

// convertToPostgreSQLParams 转换参数占位符为 PostgreSQL 格式
func (db *PostgreSQLDB) convertToPostgreSQLParams(query string, args []interface{}) (string, []interface{}) {
	// 如果查询已经使用 PostgreSQL 格式的参数占位符，直接返回
	if strings.Contains(query, "$1") || strings.Contains(query, "$2") {
		return query, args
	}
	
	// 将 ? 占位符转换为 PostgreSQL 的 $1, $2, ... 格式
	paramCount := 0
	result := ""
	
	for i, char := range query {
		if char == '?' {
			// 检查是否在字符串字面量中
			if !db.isInStringLiteral(query, i) {
				paramCount++
				result += fmt.Sprintf("$%d", paramCount)
			} else {
				result += string(char)
			}
		} else {
			result += string(char)
		}
	}
	
	return result, args
}

// isInStringLiteral 检查位置是否在字符串字面量中
func (db *PostgreSQLDB) isInStringLiteral(sql string, pos int) bool {
	inSingleQuote := false
	inDoubleQuote := false
	
	for i := 0; i < pos; i++ {
		char := sql[i]
		switch char {
		case '\'':
			if !inDoubleQuote {
				// 检查是否是转义的单引号
				if i > 0 && sql[i-1] == '\\' {
					continue
				}
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				// 检查是否是转义的双引号
				if i > 0 && sql[i-1] == '\\' {
					continue
				}
				inDoubleQuote = !inDoubleQuote
			}
		}
	}
	
	return inSingleQuote || inDoubleQuote
}

// executeOnFirstDataSource 在第一个数据源执行命令
func (db *PostgreSQLDB) executeOnFirstDataSource(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var firstDB *sql.DB
	for _, conn := range db.pgDataSource.dataSources {
		firstDB = conn
		break
	}
	
	if firstDB == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	
	return firstDB.ExecContext(ctx, query, args...)
}

// BeginTx 开始 PostgreSQL 事务
func (db *PostgreSQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*PostgreSQLTx, error) {
	// 对于分片环境，这里简化处理，只在第一个数据源开始事务
	// 实际生产环境中需要实现分布式事务
	var firstDB *sql.DB
	for _, conn := range db.pgDataSource.dataSources {
		firstDB = conn
		break
	}
	
	if firstDB == nil {
		return nil, fmt.Errorf("no database connection available")
	}
	
	tx, err := firstDB.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	return &PostgreSQLTx{
		Tx:     tx,
		pgDB:   db,
	}, nil
}

// Begin 开始 PostgreSQL 事务
func (db *PostgreSQLDB) Begin() (*PostgreSQLTx, error) {
	return db.BeginTx(context.Background(), nil)
}

// PostgreSQLTx PostgreSQL 事务
type PostgreSQLTx struct {
	*sql.Tx
	pgDB *PostgreSQLDB
}

// QueryContext 在事务中执行查询
func (tx *PostgreSQLTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// 验证 PostgreSQL SQL 语法
	if err := tx.pgDB.pgDataSource.pgParser.ValidatePostgreSQLSQL(query); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL SQL: %w", err)
	}
	
	// 转换参数占位符
	pgQuery, pgArgs := tx.pgDB.convertToPostgreSQLParams(query, args)
	
	return tx.Tx.QueryContext(ctx, pgQuery, pgArgs...)
}

// Query 在事务中执行查询
func (tx *PostgreSQLTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(context.Background(), query, args...)
}

// ExecContext 在事务中执行命令
func (tx *PostgreSQLTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 验证 PostgreSQL SQL 语法
	if err := tx.pgDB.pgDataSource.pgParser.ValidatePostgreSQLSQL(query); err != nil {
		return nil, fmt.Errorf("invalid PostgreSQL SQL: %w", err)
	}
	
	// 转换参数占位符
	pgQuery, pgArgs := tx.pgDB.convertToPostgreSQLParams(query, args)
	
	return tx.Tx.ExecContext(ctx, pgQuery, pgArgs...)
}

// Exec 在事务中执行命令
func (tx *PostgreSQLTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(context.Background(), query, args...)
}

// GetPostgreSQLDialect 获取 PostgreSQL 方言
func (ds *PostgreSQLShardingDataSource) GetPostgreSQLDialect() database.DatabaseDialect {
	return ds.dialect
}

// ValidateConnection 验证 PostgreSQL 连接
func (ds *PostgreSQLShardingDataSource) ValidateConnection() error {
	for name, db := range ds.dataSources {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping PostgreSQL database %s: %w", name, err)
		}
		
		// 验证是否确实是 PostgreSQL
		var version string
		err := db.QueryRow("SELECT version()").Scan(&version)
		if err != nil {
			return fmt.Errorf("failed to get version from database %s: %w", name, err)
		}
		
		if !strings.Contains(strings.ToLower(version), "postgresql") {
			return fmt.Errorf("database %s is not PostgreSQL: %s", name, version)
		}
	}
	
	return nil
}