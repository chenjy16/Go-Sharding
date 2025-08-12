package sharding

import (
	"context"
	"database/sql"
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/id"
	"go-sharding/pkg/merge"
	"go-sharding/pkg/rewrite"
	"go-sharding/pkg/routing"
	"regexp"
	"strings"
)

// ShardingDataSource 分片数据源
type ShardingDataSource struct {
	dataSources      map[string]*sql.DB
	shardingRule     *config.ShardingRuleConfig
	configuredTables map[string]*config.TableRuleConfig
	router           routing.Router
	rewriter         *rewrite.SQLRewriter
	merger           *merge.ResultMerger
	idGenerator      id.Generator
}

// NewShardingDataSource 创建分片数据源
func NewShardingDataSource(cfg *config.ShardingConfig) (*ShardingDataSource, error) {
	ds := &ShardingDataSource{
		dataSources:      make(map[string]*sql.DB),
		shardingRule:     cfg.ShardingRule,
		configuredTables: cfg.ShardingRule.Tables,
	}

	// 初始化数据源连接
	for name, dsConfig := range cfg.DataSources {
		db, err := sql.Open(dsConfig.DriverName, dsConfig.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database %s: %w", name, err)
		}

		// 设置连接池参数
		if dsConfig.MaxOpen > 0 {
			db.SetMaxOpenConns(dsConfig.MaxOpen)
		}
		if dsConfig.MaxIdle > 0 {
			db.SetMaxIdleConns(dsConfig.MaxIdle)
		}

		ds.dataSources[name] = db
	}

	// 创建路由器
	router := routing.NewShardingRouter(cfg.DataSources, ds.shardingRule)

	// 创建 SQL 重写器
	rewriter := rewrite.NewSQLRewriter()

	// 创建结果合并器
	merger := merge.NewResultMerger()

	// 创建 ID 生成器工厂
	factory := id.NewGeneratorFactory()
	idGenerator, err := factory.CreateGenerator("snowflake", map[string]interface{}{
		"workerId":     1,
		"datacenterId": 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ID generator: %w", err)
	}

	ds.router = router
	ds.rewriter = rewriter
	ds.merger = merger
	ds.idGenerator = idGenerator

	return ds, nil
}

// GetConnection 获取数据源连接
func (ds *ShardingDataSource) GetConnection(name string) *sql.DB {
	return ds.dataSources[name]
}

// GetShardingRule 获取分片规则
func (ds *ShardingDataSource) GetShardingRule() *config.ShardingRuleConfig {
	return ds.shardingRule
}

// GetConfiguredTables 获取配置的表
func (ds *ShardingDataSource) GetConfiguredTables() map[string]*config.TableRuleConfig {
	return ds.configuredTables
}

// DB 获取分片数据库连接
func (ds *ShardingDataSource) DB() *ShardingDB {
	return &ShardingDB{
		dataSource: ds,
	}
}

// Close 关闭所有连接
func (ds *ShardingDataSource) Close() error {
	for name, db := range ds.dataSources {
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close database %s: %w", name, err)
		}
	}
	return nil
}

// ShardingDB 分片数据库
type ShardingDB struct {
	dataSource *ShardingDataSource
}

// Query 执行查询
func (db *ShardingDB) Query(query string, args ...interface{}) (*ShardingRows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行查询（带上下文）
func (db *ShardingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*ShardingRows, error) {
	// 提取逻辑表名
	logicTables := db.extractLogicTables(query)
	if len(logicTables) == 0 {
		// 如果没有分片表，直接在第一个数据源执行
		return db.executeQueryOnFirstDataSource(ctx, query, args...)
	}

	// 提取分片值
	shardingValues := db.extractShardingValues(query, args)

	// 路由计算
	var allRouteResults []*routing.RouteResult
	for _, logicTable := range logicTables {
		routeResults, err := db.dataSource.router.Route(logicTable, shardingValues)
		if err != nil {
			return nil, fmt.Errorf("routing failed for table %s: %w", logicTable, err)
		}
		allRouteResults = append(allRouteResults, routeResults...)
	}

	// SQL 重写
	rewriteCtx := &rewrite.RewriteContext{
		OriginalSQL:  query,
		LogicTables:  logicTables,
		RouteResults: allRouteResults,
		Parameters:   args,
	}
	rewriteResults, err := db.dataSource.rewriter.Rewrite(rewriteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite SQL: %w", err)
	}
	// 执行查询
	var allRows []*sql.Rows
	for _, rewriteResult := range rewriteResults {
		conn := db.dataSource.dataSources[rewriteResult.DataSource]
		rows, err := conn.QueryContext(ctx, rewriteResult.SQL, rewriteResult.Parameters...)
		if err != nil {
			// 关闭已打开的结果集
			for _, r := range allRows {
				r.Close()
			}
			return nil, fmt.Errorf("query failed on %s: %w", rewriteResult.DataSource, err)
		}
		allRows = append(allRows, rows)
	}

	// 简化处理：如果只有一个结果集，直接返回
	if len(allRows) == 1 {
		return &ShardingRows{
			rows: allRows[0],
		}, nil
	}

	// 多个结果集需要合并，这里简化处理，返回第一个
	if len(allRows) > 0 {
		// 关闭其他结果集
		for i := 1; i < len(allRows); i++ {
			allRows[i].Close()
		}
		return &ShardingRows{
			rows: allRows[0],
		}, nil
	}
	
	return &ShardingRows{}, nil
}

// executeQueryOnFirstDataSource 在第一个数据源执行查询
func (db *ShardingDB) executeQueryOnFirstDataSource(ctx context.Context, query string, args ...interface{}) (*ShardingRows, error) {
	var firstDB *sql.DB
	for _, conn := range db.dataSource.dataSources {
		firstDB = conn
		break
	}

	if firstDB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	rows, err := firstDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &ShardingRows{
		rows:    rows,
		columns: nil,
	}, nil
}

// Exec 执行非查询语句
func (db *ShardingDB) Exec(query string, args ...interface{}) (*ShardingResult, error) {
	return db.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行非查询语句（带上下文）
func (db *ShardingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (*ShardingResult, error) {
	// 提取逻辑表名
	logicTables := db.extractLogicTables(query)
	if len(logicTables) == 0 {
		// 如果没有分片表，直接在第一个数据源执行
		return db.executeExecOnFirstDataSource(ctx, query, args...)
	}

	// 对于 INSERT 语句，可能需要生成 ID
	if strings.ToUpper(strings.TrimSpace(query)[:6]) == "INSERT" {
		query, args = db.handleInsertWithGeneratedID(query, args, logicTables[0])
	}

	// 提取分片值
	shardingValues := db.extractShardingValues(query, args)

	// 路由计算
	var allRouteResults []*routing.RouteResult
	for _, logicTable := range logicTables {
		routeResults, err := db.dataSource.router.Route(logicTable, shardingValues)
		if err != nil {
			return nil, fmt.Errorf("routing failed for table %s: %w", logicTable, err)
		}
		allRouteResults = append(allRouteResults, routeResults...)
	}

	// SQL 重写
	rewriteCtx := &rewrite.RewriteContext{
		OriginalSQL:  query,
		LogicTables:  logicTables,
		RouteResults: allRouteResults,
		Parameters:   args,
	}
	rewriteResults, err := db.dataSource.rewriter.Rewrite(rewriteCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite SQL: %w", err)
	}
	// 执行语句
	var totalAffected int64
	var lastInsertID int64

	for _, rewriteResult := range rewriteResults {
		conn := db.dataSource.dataSources[rewriteResult.DataSource]
		result, err := conn.ExecContext(ctx, rewriteResult.SQL, rewriteResult.Parameters...)
		if err != nil {
			return nil, fmt.Errorf("exec failed on %s: %w", rewriteResult.DataSource, err)
		}

		if affected, err := result.RowsAffected(); err == nil {
			totalAffected += affected
		}

		if insertID, err := result.LastInsertId(); err == nil && insertID > 0 {
			lastInsertID = insertID
		}
	}

	return &ShardingResult{
		affectedRows: totalAffected,
		lastInsertID: lastInsertID,
	}, nil
}

// executeExecOnFirstDataSource 在第一个数据源执行非查询语句
func (db *ShardingDB) executeExecOnFirstDataSource(ctx context.Context, query string, args ...interface{}) (*ShardingResult, error) {
	var firstDB *sql.DB
	for _, conn := range db.dataSource.dataSources {
		firstDB = conn
		break
	}

	if firstDB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	result, err := firstDB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	affected, _ := result.RowsAffected()
	lastID, _ := result.LastInsertId()

	return &ShardingResult{
		affectedRows: affected,
		lastInsertID: lastID,
	}, nil
}

// handleInsertWithGeneratedID 处理带生成 ID 的插入语句
func (db *ShardingDB) handleInsertWithGeneratedID(query string, args []interface{}, logicTable string) (string, []interface{}) {
	tableConfig, exists := db.dataSource.configuredTables[logicTable]
	if !exists || tableConfig.KeyGenerator == nil {
		return query, args
	}

	// 简化实现：如果是 INSERT 语句且配置了主键生成器，生成 ID
	if tableConfig.KeyGenerator.Type == "snowflake" {
		generatedID, err := db.dataSource.idGenerator.NextID()
		if err != nil {
			// 如果生成 ID 失败，返回原始参数
			return query, args
		}
		// 这里应该解析 SQL 并插入生成的 ID
		// 简化实现，假设 ID 是第一个参数
		newArgs := make([]interface{}, len(args)+1)
		newArgs[0] = generatedID
		copy(newArgs[1:], args)
		return query, newArgs
	}

	return query, args
}

// extractLogicTables 提取逻辑表名
func (db *ShardingDB) extractLogicTables(query string) []string {
	var tables []string
	for tableName := range db.dataSource.configuredTables {
		if strings.Contains(strings.ToLower(query), strings.ToLower(tableName)) {
			tables = append(tables, tableName)
		}
	}
	return tables
}

// extractShardingValues 提取分片值
func (db *ShardingDB) extractShardingValues(query string, args []interface{}) map[string]interface{} {
	shardingValues := make(map[string]interface{})
	
	// 使用正则表达式提取 WHERE 条件中的分片值
	whereRegex := regexp.MustCompile(`(?i)where\s+(.+?)(?:\s+order\s+by|\s+group\s+by|\s+limit|$)`)
	whereMatch := whereRegex.FindStringSubmatch(query)
	
	if len(whereMatch) > 1 {
		whereClause := whereMatch[1]
		
		// 提取 user_id 条件
		userIDRegex := regexp.MustCompile(`(?i)user_id\s*=\s*\?`)
		if userIDRegex.MatchString(whereClause) && len(args) > 0 {
			shardingValues["user_id"] = args[0]
		}
		
		// 提取 order_id 条件
		orderIDRegex := regexp.MustCompile(`(?i)order_id\s*=\s*\?`)
		if orderIDRegex.MatchString(whereClause) {
			// 查找 order_id 对应的参数位置
			for i, arg := range args {
				if i == 0 && userIDRegex.MatchString(whereClause) {
					continue // 跳过 user_id 参数
				}
				shardingValues["order_id"] = arg
				break
			}
		}
	}
	
	// 如果没有找到具体的分片值，使用默认逻辑
	if len(shardingValues) == 0 && len(args) > 0 {
		shardingValues["user_id"] = args[0]
		if len(args) > 1 {
			shardingValues["order_id"] = args[1]
		}
	}

	return shardingValues
}

// ShardingRows 分片查询结果
type ShardingRows struct {
	rows    *sql.Rows
	columns []string
}

// Next 移动到下一行
func (sr *ShardingRows) Next() bool {
	return sr.rows.Next()
}

// Scan 扫描当前行
func (sr *ShardingRows) Scan(dest ...interface{}) error {
	return sr.rows.Scan(dest...)
}

// Columns 获取列名
func (sr *ShardingRows) Columns() ([]string, error) {
	if sr.columns == nil {
		cols, err := sr.rows.Columns()
		if err != nil {
			return nil, err
		}
		sr.columns = cols
	}

	return sr.columns, nil
}

// Close 关闭结果集
func (sr *ShardingRows) Close() error {
	return sr.rows.Close()
}

// Err 获取错误
func (sr *ShardingRows) Err() error {
	return sr.rows.Err()
}

// ShardingResult 分片执行结果
type ShardingResult struct {
	affectedRows int64
	lastInsertID int64
}

// RowsAffected 获取影响的行数
func (sr *ShardingResult) RowsAffected() (int64, error) {
	return sr.affectedRows, nil
}

// LastInsertId 获取最后插入的 ID
func (sr *ShardingResult) LastInsertId() (int64, error) {
	return sr.lastInsertID, nil
}