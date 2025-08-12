package sharding

import (
	"context"
	"database/sql"
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/parser"
	"go-sharding/pkg/readwrite"
	"go-sharding/pkg/routing"
	"go-sharding/pkg/rewrite"
	"sync"
)

// EnhancedShardingDB 增强的分片数据库，支持读写分离
type EnhancedShardingDB struct {
	config           *config.ShardingConfig
	dataSources      map[string]*sql.DB
	readWriteSplitters map[string]*readwrite.ReadWriteSplitter
	router           *routing.ShardingRouter
	rewriter         *rewrite.SQLRewriter
	parserFactory    *parser.ParserFactory
	mutex            sync.RWMutex
}

// NewEnhancedShardingDB 创建增强的分片数据库实例
func NewEnhancedShardingDB(cfg *config.ShardingConfig) (*EnhancedShardingDB, error) {
	db := &EnhancedShardingDB{
		config:             cfg,
		dataSources:        make(map[string]*sql.DB),
		readWriteSplitters: make(map[string]*readwrite.ReadWriteSplitter),
		router:             routing.NewShardingRouter(cfg.DataSources, cfg.ShardingRule),
		rewriter:           rewrite.NewSQLRewriter(),
		parserFactory:      parser.DefaultParserFactory,
	}

	// 初始化数据源连接
	if err := db.initDataSources(); err != nil {
		return nil, fmt.Errorf("failed to initialize data sources: %w", err)
	}

	// 初始化读写分离器
	if err := db.initReadWriteSplitters(); err != nil {
		return nil, fmt.Errorf("failed to initialize read-write splitters: %w", err)
	}

	return db, nil
}

// initDataSources 初始化数据源连接
func (db *EnhancedShardingDB) initDataSources() error {
	for name, dsConfig := range db.config.DataSources {
		sqlDB, err := sql.Open(dsConfig.DriverName, dsConfig.URL)
		if err != nil {
			return fmt.Errorf("failed to open data source %s: %w", name, err)
		}

		// 设置连接池参数
		sqlDB.SetMaxIdleConns(dsConfig.MaxIdle)
		sqlDB.SetMaxOpenConns(dsConfig.MaxOpen)

		// 测试连接
		if err := sqlDB.Ping(); err != nil {
			sqlDB.Close()
			return fmt.Errorf("failed to ping data source %s: %w", name, err)
		}

		db.dataSources[name] = sqlDB
	}

	return nil
}

// initReadWriteSplitters 初始化读写分离器
func (db *EnhancedShardingDB) initReadWriteSplitters() error {
	for name, rwConfig := range db.config.ReadWriteSplits {
		splitter, err := readwrite.NewReadWriteSplitter(rwConfig, db.dataSources)
		if err != nil {
			return fmt.Errorf("failed to create read-write splitter %s: %w", name, err)
		}

		db.readWriteSplitters[name] = splitter
	}

	return nil
}

// Query 执行查询语句
func (db *EnhancedShardingDB) Query(query string, args ...interface{}) (*EnhancedShardingRows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行查询语句（带上下文）
func (db *EnhancedShardingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*EnhancedShardingRows, error) {
	// 解析 SQL 语句
	stmt, err := db.parserFactory.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	// 提取逻辑表名
	configuredTables := make(map[string]bool)
	if db.config.ShardingRule != nil {
		for tableName := range db.config.ShardingRule.Tables {
			configuredTables[tableName] = true
		}
	}

	logicTables := db.rewriter.ExtractLogicTables(query, configuredTables)
	if len(logicTables) == 0 {
		// 没有分片表，直接执行
		return db.executeNonShardedQuery(ctx, query, args...)
	}

	// 路由计算
	var routes []*routing.RouteResult
	if len(logicTables) > 0 {
		// 提取分片值（简化实现，实际应该从 SQL 和参数中解析）
		shardingValues := make(map[string]interface{})
		for i, arg := range args {
			shardingValues[fmt.Sprintf("param_%d", i)] = arg
		}
		
		// 对每个逻辑表进行路由
		for _, table := range logicTables {
			tableRoutes, err := db.router.Route(table, shardingValues)
			if err != nil {
				return nil, fmt.Errorf("failed to route query for table %s: %w", table, err)
			}
			routes = append(routes, tableRoutes...)
		}
	}

	// SQL 重写
	rewriteCtx := &rewrite.RewriteContext{
		OriginalSQL:  query,
		LogicTables:  logicTables,
		RouteResults: routes,
		Parameters:   args,
	}

	rewriteResults, err := db.rewriter.Rewrite(rewriteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite SQL: %w", err)
	}

	// 执行查询
	return db.executeShardedQuery(ctx, stmt, rewriteResults)
}

// Exec 执行非查询语句
func (db *EnhancedShardingDB) Exec(query string, args ...interface{}) (*EnhancedShardingResult, error) {
	return db.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行非查询语句（带上下文）
func (db *EnhancedShardingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (*EnhancedShardingResult, error) {
	// 解析 SQL 语句
	stmt, err := db.parserFactory.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	// 提取逻辑表名
	configuredTables := make(map[string]bool)
	if db.config.ShardingRule != nil {
		for tableName := range db.config.ShardingRule.Tables {
			configuredTables[tableName] = true
		}
	}

	logicTables := db.rewriter.ExtractLogicTables(query, configuredTables)
	if len(logicTables) == 0 {
		// 没有分片表，直接执行
		return db.executeNonShardedExec(ctx, query, args...)
	}

	// 路由计算
	var routes []*routing.RouteResult
	if len(logicTables) > 0 {
		// 提取分片值（简化实现，实际应该从 SQL 和参数中解析）
		shardingValues := make(map[string]interface{})
		for i, arg := range args {
			shardingValues[fmt.Sprintf("param_%d", i)] = arg
		}
		
		// 对每个逻辑表进行路由
		for _, table := range logicTables {
			tableRoutes, err := db.router.Route(table, shardingValues)
			if err != nil {
				return nil, fmt.Errorf("failed to route query for table %s: %w", table, err)
			}
			routes = append(routes, tableRoutes...)
		}
	}

	// SQL 重写
	rewriteCtx := &rewrite.RewriteContext{
		OriginalSQL:  query,
		LogicTables:  logicTables,
		RouteResults: routes,
		Parameters:   args,
	}

	rewriteResults, err := db.rewriter.Rewrite(rewriteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite SQL: %w", err)
	}

	// 执行语句
	return db.executeShardedExec(ctx, stmt, rewriteResults)
}

// executeNonShardedQuery 执行非分片查询
func (db *EnhancedShardingDB) executeNonShardedQuery(ctx context.Context, query string, args ...interface{}) (*EnhancedShardingRows, error) {
	// 选择第一个数据源或使用读写分离
	var targetDB *sql.DB
	
	if len(db.readWriteSplitters) > 0 {
		// 使用第一个读写分离器
		for _, splitter := range db.readWriteSplitters {
			targetDB = splitter.RouteContext(ctx, query)
			break
		}
	} else {
		// 使用第一个数据源
		for _, sqlDB := range db.dataSources {
			targetDB = sqlDB
			break
		}
	}

	if targetDB == nil {
		return nil, fmt.Errorf("no available data source")
	}

	rows, err := targetDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &EnhancedShardingRows{rows: rows}, nil
}

// executeNonShardedExec 执行非分片语句
func (db *EnhancedShardingDB) executeNonShardedExec(ctx context.Context, query string, args ...interface{}) (*EnhancedShardingResult, error) {
	// 选择第一个数据源或使用读写分离
	var targetDB *sql.DB
	
	if len(db.readWriteSplitters) > 0 {
		// 使用第一个读写分离器
		for _, splitter := range db.readWriteSplitters {
			targetDB = splitter.RouteContext(ctx, query)
			break
		}
	} else {
		// 使用第一个数据源
		for _, sqlDB := range db.dataSources {
			targetDB = sqlDB
			break
		}
	}

	if targetDB == nil {
		return nil, fmt.Errorf("no available data source")
	}

	result, err := targetDB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &EnhancedShardingResult{result: result}, nil
}

// executeShardedQuery 执行分片查询
func (db *EnhancedShardingDB) executeShardedQuery(ctx context.Context, stmt *parser.SQLStatement, rewriteResults []*rewrite.RewriteResult) (*EnhancedShardingRows, error) {
	var allRows []*sql.Rows

	for _, rewriteResult := range rewriteResults {
		var targetDB *sql.DB

		// 检查是否有对应的读写分离器
		if splitter, exists := db.readWriteSplitters[rewriteResult.DataSource]; exists {
			targetDB = splitter.RouteContext(ctx, rewriteResult.SQL)
		} else {
			// 直接使用数据源
			targetDB = db.dataSources[rewriteResult.DataSource]
		}

		if targetDB == nil {
			return nil, fmt.Errorf("data source %s not found", rewriteResult.DataSource)
		}

		rows, err := targetDB.QueryContext(ctx, rewriteResult.SQL, rewriteResult.Parameters...)
		if err != nil {
			// 关闭已打开的 rows
			for _, r := range allRows {
				r.Close()
			}
			return nil, fmt.Errorf("failed to execute query on %s: %w", rewriteResult.DataSource, err)
		}

		allRows = append(allRows, rows)
	}

	return &EnhancedShardingRows{
		rows:     allRows[0], // 简化实现，返回第一个结果
		allRows:  allRows,
		sqlType:  stmt.Type,
	}, nil
}

// executeShardedExec 执行分片语句
func (db *EnhancedShardingDB) executeShardedExec(ctx context.Context, stmt *parser.SQLStatement, rewriteResults []*rewrite.RewriteResult) (*EnhancedShardingResult, error) {
	var totalRowsAffected int64
	var lastInsertId int64

	for _, rewriteResult := range rewriteResults {
		var targetDB *sql.DB

		// 检查是否有对应的读写分离器
		if splitter, exists := db.readWriteSplitters[rewriteResult.DataSource]; exists {
			targetDB = splitter.RouteContext(ctx, rewriteResult.SQL)
		} else {
			// 直接使用数据源
			targetDB = db.dataSources[rewriteResult.DataSource]
		}

		if targetDB == nil {
			return nil, fmt.Errorf("data source %s not found", rewriteResult.DataSource)
		}

		result, err := targetDB.ExecContext(ctx, rewriteResult.SQL, rewriteResult.Parameters...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute statement on %s: %w", rewriteResult.DataSource, err)
		}

		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}

		if insertId, err := result.LastInsertId(); err == nil && insertId > 0 {
			lastInsertId = insertId
		}
	}

	return &EnhancedShardingResult{
		rowsAffected: totalRowsAffected,
		lastInsertId: lastInsertId,
	}, nil
}

// Close 关闭数据库连接
func (db *EnhancedShardingDB) Close() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var errors []string

	// 关闭读写分离器
	for name, splitter := range db.readWriteSplitters {
		if err := splitter.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close read-write splitter %s: %v", name, err))
		}
	}

	// 关闭数据源连接
	for name, sqlDB := range db.dataSources {
		if err := sqlDB.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close data source %s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing database: %s", errors)
	}

	return nil
}

// HealthCheck 健康检查
func (db *EnhancedShardingDB) HealthCheck() error {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// 检查数据源
	for name, sqlDB := range db.dataSources {
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("data source %s health check failed: %w", name, err)
		}
	}

	// 检查读写分离器
	for name, splitter := range db.readWriteSplitters {
		if err := splitter.HealthCheck(); err != nil {
			return fmt.Errorf("read-write splitter %s health check failed: %w", name, err)
		}
	}

	return nil
}

// EnhancedShardingRows 增强的分片查询结果
type EnhancedShardingRows struct {
	rows    *sql.Rows
	allRows []*sql.Rows
	sqlType parser.SQLType
}

// Next 移动到下一行
func (r *EnhancedShardingRows) Next() bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

// Scan 扫描当前行数据
func (r *EnhancedShardingRows) Scan(dest ...interface{}) error {
	if r.rows == nil {
		return fmt.Errorf("no rows available")
	}
	return r.rows.Scan(dest...)
}

// Columns 获取列名
func (r *EnhancedShardingRows) Columns() ([]string, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("no rows available")
	}
	return r.rows.Columns()
}

// Close 关闭结果集
func (r *EnhancedShardingRows) Close() error {
	var errors []string

	if r.rows != nil {
		if err := r.rows.Close(); err != nil {
			errors = append(errors, err.Error())
		}
	}

	for i, rows := range r.allRows {
		if rows != nil && rows != r.rows {
			if err := rows.Close(); err != nil {
				errors = append(errors, fmt.Sprintf("failed to close rows %d: %v", i, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing rows: %s", errors)
	}

	return nil
}

// EnhancedShardingResult 增强的分片执行结果
type EnhancedShardingResult struct {
	result       sql.Result
	rowsAffected int64
	lastInsertId int64
}

// LastInsertId 获取最后插入的 ID
func (r *EnhancedShardingResult) LastInsertId() (int64, error) {
	if r.result != nil {
		return r.result.LastInsertId()
	}
	return r.lastInsertId, nil
}

// RowsAffected 获取受影响的行数
func (r *EnhancedShardingResult) RowsAffected() (int64, error) {
	if r.result != nil {
		return r.result.RowsAffected()
	}
	return r.rowsAffected, nil
}