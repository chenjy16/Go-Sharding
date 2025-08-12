package rewrite

import (
	"fmt"
	"go-sharding/pkg/parser"
	"go-sharding/pkg/routing"
	"regexp"
	"strings"
)

// SQLRewriter SQL 重写器
type SQLRewriter struct {
	parser *parser.SQLParser
}

// NewSQLRewriter 创建 SQL 重写器
func NewSQLRewriter() *SQLRewriter {
	return &SQLRewriter{
		parser: parser.NewSQLParser(),
	}
}

// RewriteContext 重写上下文
type RewriteContext struct {
	OriginalSQL    string
	LogicTables    []string
	RouteResults   []*routing.RouteResult
	Parameters     []interface{}
}

// RewriteResult 重写结果
type RewriteResult struct {
	SQL        string
	Parameters []interface{}
	DataSource string
}

// Rewrite 重写 SQL
func (r *SQLRewriter) Rewrite(ctx *RewriteContext) ([]*RewriteResult, error) {
	var results []*RewriteResult

	// 按数据源分组路由结果
	dataSourceGroups := r.groupByDataSource(ctx.RouteResults)

	for dataSource, routes := range dataSourceGroups {
		// 为每个数据源重写 SQL
		rewrittenSQL, err := r.rewriteSQLForDataSource(ctx.OriginalSQL, ctx.LogicTables, routes)
		if err != nil {
			return nil, fmt.Errorf("failed to rewrite SQL for data source %s: %w", dataSource, err)
		}

		results = append(results, &RewriteResult{
			SQL:        rewrittenSQL,
			Parameters: ctx.Parameters,
			DataSource: dataSource,
		})
	}

	return results, nil
}

// groupByDataSource 按数据源分组路由结果
func (r *SQLRewriter) groupByDataSource(routes []*routing.RouteResult) map[string][]*routing.RouteResult {
	groups := make(map[string][]*routing.RouteResult)
	
	for _, route := range routes {
		groups[route.DataSource] = append(groups[route.DataSource], route)
	}
	
	return groups
}

// rewriteSQLForDataSource 为指定数据源重写 SQL
func (r *SQLRewriter) rewriteSQLForDataSource(originalSQL string, logicTables []string, routes []*routing.RouteResult) (string, error) {
	sql := originalSQL

	// 替换逻辑表名为实际表名
	for _, logicTable := range logicTables {
		actualTables := r.getActualTablesForLogicTable(logicTable, routes)
		if len(actualTables) == 0 {
			continue
		}

		if len(actualTables) == 1 {
			// 单表替换
			sql = r.replaceTableName(sql, logicTable, actualTables[0])
		} else {
			// 多表需要生成 UNION 查询
			sql = r.generateUnionSQL(sql, logicTable, actualTables)
		}
	}

	return sql, nil
}

// getActualTablesForLogicTable 获取逻辑表对应的实际表
func (r *SQLRewriter) getActualTablesForLogicTable(logicTable string, routes []*routing.RouteResult) []string {
	var actualTables []string
	
	for _, route := range routes {
		actualTables = append(actualTables, route.Table)
	}
	
	return actualTables
}

// replaceTableName 替换表名
func (r *SQLRewriter) replaceTableName(sql, logicTable, actualTable string) string {
	// 使用正则表达式精确匹配表名，避免部分匹配
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(logicTable))
	regex := regexp.MustCompile(pattern)
	return regex.ReplaceAllString(sql, actualTable)
}

// generateUnionSQL 生成 UNION 查询
func (r *SQLRewriter) generateUnionSQL(originalSQL, logicTable string, actualTables []string) string {
	var unionParts []string
	
	for _, actualTable := range actualTables {
		rewrittenSQL := r.replaceTableName(originalSQL, logicTable, actualTable)
		unionParts = append(unionParts, fmt.Sprintf("(%s)", rewrittenSQL))
	}
	
	return strings.Join(unionParts, " UNION ALL ")
}

// ExtractLogicTables 从 SQL 中提取逻辑表名
func (r *SQLRewriter) ExtractLogicTables(sql string, configuredTables map[string]bool) []string {
	var logicTables []string
	
	// 使用增强的 SQL 解析器提取表名
	stmt, err := r.parser.Parse(sql)
	if err != nil {
		// 如果解析失败，回退到简单的单词提取方法
		words := r.extractWords(sql)
		for _, word := range words {
			if configuredTables[word] {
				if !r.contains(logicTables, word) {
					logicTables = append(logicTables, word)
				}
			}
		}
		return logicTables
	}
	
	// 从解析结果中提取配置的逻辑表
	for _, table := range stmt.Tables {
		if configuredTables[table] {
			if !r.contains(logicTables, table) {
				logicTables = append(logicTables, table)
			}
		}
	}
	
	return logicTables
}

// extractWords 提取 SQL 中的单词
func (r *SQLRewriter) extractWords(sql string) []string {
	// 移除 SQL 注释和字符串字面量
	sql = r.removeComments(sql)
	sql = r.removeStringLiterals(sql)
	
	// 使用正则表达式提取标识符
	regex := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
	matches := regex.FindAllString(sql, -1)
	
	var words []string
	for _, match := range matches {
		// 过滤 SQL 关键字
		if !r.isSQLKeyword(strings.ToUpper(match)) {
			words = append(words, match)
		}
	}
	
	return words
}

// removeComments 移除 SQL 注释
func (r *SQLRewriter) removeComments(sql string) string {
	// 移除单行注释 --
	lineCommentRegex := regexp.MustCompile(`--.*`)
	sql = lineCommentRegex.ReplaceAllString(sql, "")
	
	// 移除多行注释 /* */
	blockCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	sql = blockCommentRegex.ReplaceAllString(sql, "")
	
	return sql
}

// removeStringLiterals 移除字符串字面量
func (r *SQLRewriter) removeStringLiterals(sql string) string {
	// 移除单引号字符串
	singleQuoteRegex := regexp.MustCompile(`'([^'\\]|\\.)*'`)
	sql = singleQuoteRegex.ReplaceAllString(sql, "''")
	
	// 移除双引号字符串
	doubleQuoteRegex := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	sql = doubleQuoteRegex.ReplaceAllString(sql, `""`)
	
	return sql
}

// isSQLKeyword 检查是否为 SQL 关键字
func (r *SQLRewriter) isSQLKeyword(word string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true, "UPDATE": true,
		"DELETE": true, "CREATE": true, "DROP": true, "ALTER": true, "TABLE": true,
		"INDEX": true, "VIEW": true, "DATABASE": true, "SCHEMA": true, "AND": true,
		"OR": true, "NOT": true, "NULL": true, "TRUE": true, "FALSE": true,
		"JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true, "FULL": true,
		"OUTER": true, "ON": true, "USING": true, "GROUP": true, "BY": true,
		"ORDER": true, "HAVING": true, "LIMIT": true, "OFFSET": true, "UNION": true,
		"ALL": true, "DISTINCT": true, "AS": true, "ASC": true, "DESC": true,
		"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
		"CASE": true, "WHEN": true, "THEN": true, "ELSE": true, "END": true,
		"IF": true, "EXISTS": true, "IN": true, "BETWEEN": true, "LIKE": true,
		"IS": true, "PRIMARY": true, "KEY": true, "FOREIGN": true, "REFERENCES": true,
		"UNIQUE": true, "CHECK": true, "DEFAULT": true, "AUTO_INCREMENT": true,
		"TIMESTAMP": true, "DATETIME": true, "DATE": true, "TIME": true,
		"VARCHAR": true, "CHAR": true, "TEXT": true, "INT": true, "INTEGER": true,
		"BIGINT": true, "SMALLINT": true, "TINYINT": true, "DECIMAL": true,
		"FLOAT": true, "DOUBLE": true, "BOOLEAN": true, "BOOL": true,
	}
	
	return keywords[word]
}

// contains 检查字符串数组是否包含指定元素
func (r *SQLRewriter) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}