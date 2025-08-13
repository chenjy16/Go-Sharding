package parser

import (
	"fmt"
	"go-sharding/pkg/database"
	"strings"

	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/parser"
	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
)

// CockroachDBAdapter CockroachDB Parser 适配器
// 实现与现有 PostgreSQL 解析器接口的兼容
type CockroachDBAdapter struct {
	*EnhancedSQLParser
	dialect database.DatabaseDialect
	parser  *parser.Parser
}

// NewCockroachDBAdapter 创建 CockroachDB 适配器
func NewCockroachDBAdapter() *CockroachDBAdapter {
	dialect, _ := database.GlobalDialectRegistry.GetDialect(database.PostgreSQL)
	return &CockroachDBAdapter{
		EnhancedSQLParser: NewEnhancedSQLParser(),
		dialect:           dialect,
		parser:            &parser.Parser{},
	}
}

// ParsePostgreSQLSpecific 使用 CockroachDB Parser 解析 PostgreSQL 特定语法
func (c *CockroachDBAdapter) ParsePostgreSQLSpecific(sql string) (*EnhancedSQLStatement, error) {
	// 使用 CockroachDB Parser 解析 SQL
	stmts, err := c.parser.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("CockroachDB parser error: %w", err)
	}

	if len(stmts) == 0 {
		return nil, fmt.Errorf("no statements found in SQL")
	}

	// 转换为 EnhancedSQLStatement
	result := &EnhancedSQLStatement{
		OriginalSQL:        sql,
		Type:               c.determineStatementType(stmts[0].AST),
		Tables:             c.extractTablesFromAST(stmts[0].AST),
		Columns:            c.extractColumnsFromAST(stmts[0].AST),
		Conditions:         make(map[string]interface{}),
		PostgreSQLFeatures: make(map[string]interface{}),
	}

	// 增强 PostgreSQL 特定功能
	c.enhancePostgreSQLFeatures(result, stmts[0].AST, sql)

	return result, nil
}

// determineStatementType 确定语句类型
func (c *CockroachDBAdapter) determineStatementType(stmt tree.Statement) SQLType {
	switch stmt.(type) {
	case *tree.Select:
		return SQLTypeSelect
	case *tree.Insert:
		return SQLTypeInsert
	case *tree.Update:
		return SQLTypeUpdate
	case *tree.Delete:
		return SQLTypeDelete
	case *tree.CreateTable:
		return SQLTypeCreate
	case *tree.DropTable:
		return SQLTypeDrop
	case *tree.AlterTable:
		return SQLTypeAlter
	default:
		return SQLTypeOther
	}
}

// extractTablesFromAST 从 AST 中提取表名
func (c *CockroachDBAdapter) extractTablesFromAST(stmt tree.Statement) []string {
	tables := make([]string, 0)
	
	// 根据语句类型提取表名
	switch s := stmt.(type) {
	case *tree.Select:
		tables = append(tables, c.extractTablesFromSelect(s)...)
	case *tree.Insert:
		if s.Table != nil {
			tables = append(tables, c.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.Update:
		if s.Table != nil {
			tables = append(tables, c.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.Delete:
		if s.Table != nil {
			tables = append(tables, c.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.CreateTable:
		if tableName := c.extractTableNameFromTableName(&s.Table); tableName != "" {
			tables = append(tables, tableName)
		}
	case *tree.DropTable:
		for _, name := range s.Names {
			if tableName := c.extractTableNameFromTableName(&name); tableName != "" {
				tables = append(tables, tableName)
			}
		}
	case *tree.AlterTable:
		if tableName := c.extractTableNameFromUnresolvedObjectName(s.Table); tableName != "" {
			tables = append(tables, tableName)
		}
	}
	
	return tables
}

// extractTablesFromSelect 从 SELECT 语句中提取表名
func (c *CockroachDBAdapter) extractTablesFromSelect(sel *tree.Select) []string {
	tables := make([]string, 0)
	
	if sel.Select != nil {
		switch s := sel.Select.(type) {
		case *tree.SelectClause:
				if s.From.Tables != nil {
				tables = append(tables, c.extractTablesFromTableExprs(s.From.Tables)...)
			}
		}
	}
	
	return tables
}

// extractTablesFromTableExprs 从表表达式列表中提取表名
func (c *CockroachDBAdapter) extractTablesFromTableExprs(exprs tree.TableExprs) []string {
	tables := make([]string, 0)
	
	for _, expr := range exprs {
		tables = append(tables, c.extractTableNameFromTableExpr(expr)...)
	}
	
	return tables
}

// extractTableNameFromTableExpr 从表表达式中提取表名
func (c *CockroachDBAdapter) extractTableNameFromTableExpr(expr tree.TableExpr) []string {
	tables := make([]string, 0)
	
	switch e := expr.(type) {
	case *tree.AliasedTableExpr:
		tables = append(tables, c.extractTableNameFromTableExpr(e.Expr)...)
	case *tree.TableName:
		if tableName := c.extractTableNameFromTableName(e); tableName != "" {
			tables = append(tables, tableName)
		}
	case *tree.JoinTableExpr:
		tables = append(tables, c.extractTableNameFromTableExpr(e.Left)...)
		tables = append(tables, c.extractTableNameFromTableExpr(e.Right)...)
	}
	
	return tables
}

// extractTableNameFromTableName 从 TableName 中提取表名
func (c *CockroachDBAdapter) extractTableNameFromTableName(tableName *tree.TableName) string {
	if tableName != nil {
		return tableName.Table()
	}
	return ""
}

// extractTableNameFromUnresolvedObjectName 从 UnresolvedObjectName 中提取表名
func (c *CockroachDBAdapter) extractTableNameFromUnresolvedObjectName(objName *tree.UnresolvedObjectName) string {
	if objName != nil {
		return objName.Object()
	}
	return ""
}

// extractColumnsFromAST 从 AST 中提取列名
func (c *CockroachDBAdapter) extractColumnsFromAST(stmt tree.Statement) []string {
	columns := make([]string, 0)
	columnVisitor := &columnExtractor{columns: &columns}
	
	// 使用 visitor 模式遍历 AST
	tree.WalkStmt(columnVisitor, stmt)
	
	return columns
}

// extractConditionsFromAST 从 AST 中提取条件
func (c *CockroachDBAdapter) extractConditionsFromAST(stmt tree.Statement) []string {
	conditions := make([]string, 0)
	
	// 根据语句类型提取 WHERE 条件
	switch s := stmt.(type) {
	case *tree.Select:
		if s.Select != nil {
			if selectClause, ok := s.Select.(*tree.SelectClause); ok && selectClause.Where != nil {
				conditions = append(conditions, selectClause.Where.Expr.String())
			}
		}
	case *tree.Update:
		if s.Where != nil {
			conditions = append(conditions, s.Where.Expr.String())
		}
	case *tree.Delete:
		if s.Where != nil {
			conditions = append(conditions, s.Where.Expr.String())
		}
	}
	
	return conditions
}

// enhancePostgreSQLFeatures 增强 PostgreSQL 特定功能
func (c *CockroachDBAdapter) enhancePostgreSQLFeatures(result *EnhancedSQLStatement, stmt tree.Statement, sql string) {
	// 处理 LIMIT/OFFSET
	c.parsePostgreSQLLimit(result, stmt)
	
	// 处理 RETURNING 子句
	c.parseReturningClause(result, stmt)
	
	// 处理 PostgreSQL 特定函数
	c.parsePostgreSQLFunctions(result, sql)
	
	// 处理 PostgreSQL 特定操作符
	c.parsePostgreSQLOperators(result, sql)
}

// parsePostgreSQLLimit 解析 LIMIT/OFFSET
func (c *CockroachDBAdapter) parsePostgreSQLLimit(result *EnhancedSQLStatement, stmt tree.Statement) {
	if selectStmt, ok := stmt.(*tree.Select); ok {
		if selectStmt.Limit != nil {
			if selectStmt.Limit.Count != nil {
				result.PostgreSQLFeatures["limit"] = selectStmt.Limit.Count.String()
			}
			if selectStmt.Limit.Offset != nil {
				result.PostgreSQLFeatures["offset"] = selectStmt.Limit.Offset.String()
			}
		}
	}
}

// parseReturningClause 解析 RETURNING 子句
func (c *CockroachDBAdapter) parseReturningClause(result *EnhancedSQLStatement, stmt tree.Statement) {
	switch s := stmt.(type) {
	case *tree.Insert:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				result.PostgreSQLFeatures["returning"] = returningCols
			}
		}
	case *tree.Update:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				result.PostgreSQLFeatures["returning"] = returningCols
			}
		}
	case *tree.Delete:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				result.PostgreSQLFeatures["returning"] = returningCols
			}
		}
	}
}

// parsePostgreSQLFunctions 解析 PostgreSQL 特定函数
func (c *CockroachDBAdapter) parsePostgreSQLFunctions(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL 特定函数列表
	pgFunctions := []string{
		"COALESCE", "NULLIF", "GREATEST", "LEAST",
		"ARRAY_AGG", "STRING_AGG", "ARRAY_TO_STRING",
		"ROW_NUMBER", "RANK", "DENSE_RANK", "LAG", "LEAD",
		"GENERATE_SERIES", "UNNEST",
	}
	
	functions := make([]string, 0)
	sqlUpper := strings.ToUpper(sql)
	
	for _, fn := range pgFunctions {
		if strings.Contains(sqlUpper, fn+"(") {
			functions = append(functions, fn)
		}
	}
	
	if len(functions) > 0 {
		result.PostgreSQLFeatures["functions"] = functions
	}
}

// parsePostgreSQLOperators 解析 PostgreSQL 特定操作符
func (c *CockroachDBAdapter) parsePostgreSQLOperators(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL 特定操作符
	pgOperators := map[string]string{
		"ILIKE":    "case-insensitive LIKE",
		"~":        "regex match",
		"~*":       "case-insensitive regex match",
		"!~":       "regex not match",
		"!~*":      "case-insensitive regex not match",
		"@@":       "text search match",
		"||":       "string concatenation",
		"#>":       "JSON path",
		"#>>":      "JSON path as text",
		"?":        "JSON key exists",
		"?&":       "JSON all keys exist",
		"?|":       "JSON any key exists",
	}
	
	operators := make([]string, 0)
	
	for op := range pgOperators {
		if strings.Contains(sql, op) {
			operators = append(operators, op)
		}
	}
	
	if len(operators) > 0 {
		result.PostgreSQLFeatures["operators"] = operators
	}
}

// RewriteForPostgreSQL SQL 重写功能
func (c *CockroachDBAdapter) RewriteForPostgreSQL(sql string, tableName string, actualTableName string) (string, error) {
	// 替换表名
	rewrittenSQL := strings.ReplaceAll(sql, tableName, actualTableName)
	
	// 处理引用字符
	rewrittenSQL = c.handleQuoting(rewrittenSQL)
	
	// 转换参数占位符
	rewrittenSQL = c.convertParameterPlaceholders(rewrittenSQL)
	
	return rewrittenSQL, nil
}

// handleQuoting 处理引用字符
func (c *CockroachDBAdapter) handleQuoting(sql string) string {
	// PostgreSQL 使用双引号进行标识符引用
	// 这里可以根据需要进行转换
	return sql
}

// convertParameterPlaceholders 转换参数占位符
func (c *CockroachDBAdapter) convertParameterPlaceholders(sql string) string {
	// 将 ? 占位符转换为 PostgreSQL 的 $1, $2, ... 格式
	paramCount := 1
	result := ""
	inStringLiteral := false
	
	for i, char := range sql {
		if char == '\'' && (i == 0 || sql[i-1] != '\\') {
			inStringLiteral = !inStringLiteral
			result += string(char)
		} else if char == '?' && !inStringLiteral {
			result += fmt.Sprintf("$%d", paramCount)
			paramCount++
		} else {
			result += string(char)
		}
	}
	
	return result
}

// ValidatePostgreSQLSQL 验证 PostgreSQL SQL 语法
func (c *CockroachDBAdapter) ValidatePostgreSQLSQL(sql string) error {
	// 使用 CockroachDB Parser 进行语法验证
	_, err := c.parser.Parse(sql)
	if err != nil {
		return fmt.Errorf("PostgreSQL SQL validation failed: %w", err)
	}
	return nil
}

// ExtractTables 提取表名
func (c *CockroachDBAdapter) ExtractTables(sql string) []string {
	stmts, err := c.parser.Parse(sql)
	if err != nil {
		return []string{}
	}
	
	if len(stmts) == 0 {
		return []string{}
	}
	
	return c.extractTablesFromAST(stmts[0].AST)
}

// GetPostgreSQLDialect 获取 PostgreSQL 方言
func (c *CockroachDBAdapter) GetPostgreSQLDialect() database.DatabaseDialect {
	return c.dialect
}



// columnExtractor AST 列名提取器
type columnExtractor struct {
	columns *[]string
}

func (ce *columnExtractor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	return true, expr
}

func (ce *columnExtractor) VisitPost(expr tree.Expr) tree.Expr {
	if columnItem, ok := expr.(*tree.ColumnItem); ok {
		*ce.columns = append(*ce.columns, columnItem.String())
	}
	return expr
}