package parser

import (
	"fmt"
	"go-sharding/pkg/database"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/parser"
	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
)

// PostgreSQLParser PostgreSQL 解析器
type PostgreSQLParser struct {
	*EnhancedSQLParser
	dialect database.DatabaseDialect
	parser  *parser.Parser
}

// NewPostgreSQLParser 创建 PostgreSQL 解析器
func NewPostgreSQLParser() *PostgreSQLParser {
	dialect, _ := database.GlobalDialectRegistry.GetDialect(database.PostgreSQL)
	return &PostgreSQLParser{
		EnhancedSQLParser: NewEnhancedSQLParser(),
		dialect:           dialect,
		parser:            &parser.Parser{},
	}
}

// ParsePostgreSQLSpecific 解析 PostgreSQL 特定语法
func (p *PostgreSQLParser) ParsePostgreSQLSpecific(sql string) (*EnhancedSQLStatement, error) {
	// 使用 CockroachDB Parser 解析 SQL
	stmts, err := p.parser.Parse(sql)
	if err != nil {
		// 如果 CockroachDB Parser 失败，回退到基础解析器
		result, fallbackErr := p.EnhancedSQLParser.Parse(sql)
		if fallbackErr != nil {
			return nil, fmt.Errorf("both CockroachDB parser and fallback parser failed: %w, %w", err, fallbackErr)
		}
		// 使用正则表达式增强 PostgreSQL 特性（保持向后兼容）
		p.enhancePostgreSQLFeaturesRegex(result, sql)
		return result, nil
	}

	if len(stmts) == 0 {
		return nil, fmt.Errorf("no statements found in SQL")
	}

	// 转换为 EnhancedSQLStatement
	result := &EnhancedSQLStatement{
		OriginalSQL:        sql,
		Type:               p.determineStatementType(stmts[0].AST),
		Tables:             p.extractTablesFromAST(stmts[0].AST),
		Columns:            p.extractColumnsFromAST(stmts[0].AST),
		Conditions:         make(map[string]interface{}),
		PostgreSQLFeatures: make(map[string]interface{}),
	}

	// 使用 AST 增强 PostgreSQL 特定功能
	p.enhancePostgreSQLFeaturesAST(result, stmts[0].AST, sql)

	return result, nil
}

// determineStatementType 确定语句类型
func (p *PostgreSQLParser) determineStatementType(stmt tree.Statement) SQLType {
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
func (p *PostgreSQLParser) extractTablesFromAST(stmt tree.Statement) []string {
	tables := make([]string, 0)
	
	// 根据语句类型提取表名
	switch s := stmt.(type) {
	case *tree.Select:
		tables = append(tables, p.extractTablesFromSelect(s)...)
	case *tree.Insert:
		if s.Table != nil {
			tables = append(tables, p.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.Update:
		if s.Table != nil {
			tables = append(tables, p.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.Delete:
		if s.Table != nil {
			tables = append(tables, p.extractTableNameFromTableExpr(s.Table)...)
		}
	case *tree.CreateTable:
		if s.Table.Table() != "" {
			tables = append(tables, p.extractTableNameFromTableName(&s.Table))
		}
	case *tree.DropTable:
		for _, name := range s.Names {
			tables = append(tables, p.extractTableNameFromTableName(&name))
		}
	case *tree.AlterTable:
		if s.Table != nil {
			tables = append(tables, p.extractTableNameFromUnresolvedObjectName(s.Table))
		}
	}
	
	return tables
}

// extractTablesFromSelect 从 SELECT 语句中提取表名
func (p *PostgreSQLParser) extractTablesFromSelect(sel *tree.Select) []string {
	tables := make([]string, 0)
	
	if selectClause, ok := sel.Select.(*tree.SelectClause); ok {
		if selectClause.From.Tables != nil {
			tables = append(tables, p.extractTablesFromTableExprs(selectClause.From.Tables)...)
		}
	}
	
	return tables
}

// extractTablesFromTableExprs 从表表达式列表中提取表名
func (p *PostgreSQLParser) extractTablesFromTableExprs(exprs tree.TableExprs) []string {
	tables := make([]string, 0)
	for _, expr := range exprs {
		tables = append(tables, p.extractTableNameFromTableExpr(expr)...)
	}
	return tables
}

// extractTableNameFromTableExpr 从表表达式中提取表名
func (p *PostgreSQLParser) extractTableNameFromTableExpr(expr tree.TableExpr) []string {
	tables := make([]string, 0)
	
	switch e := expr.(type) {
	case *tree.AliasedTableExpr:
		if tableName, ok := e.Expr.(*tree.TableName); ok {
			tables = append(tables, p.extractTableNameFromTableName(tableName))
		}
	case *tree.TableName:
		tables = append(tables, p.extractTableNameFromTableName(e))
	}
	
	return tables
}

// extractTableNameFromTableName 从 TableName 中提取表名
func (p *PostgreSQLParser) extractTableNameFromTableName(tableName *tree.TableName) string {
	return tableName.Table()
}

// extractTableNameFromUnresolvedObjectName 从 UnresolvedObjectName 中提取表名
func (p *PostgreSQLParser) extractTableNameFromUnresolvedObjectName(objName *tree.UnresolvedObjectName) string {
	return objName.Object()
}

// extractColumnsFromAST 从 AST 中提取列名
func (p *PostgreSQLParser) extractColumnsFromAST(stmt tree.Statement) []string {
	// 简单实现，可以根据需要扩展
	columns := make([]string, 0)
	// TODO: 实现列名提取逻辑
	return columns
}

// enhancePostgreSQLFeaturesAST 使用 AST 增强 PostgreSQL 特定功能
func (p *PostgreSQLParser) enhancePostgreSQLFeaturesAST(result *EnhancedSQLStatement, stmt tree.Statement, sql string) {
	// 处理 LIMIT/OFFSET
	p.parsePostgreSQLLimitAST(result, stmt)
	
	// 处理 RETURNING 子句
	p.parseReturningClauseAST(result, stmt)
	
	// 处理 PostgreSQL 特定函数（仍使用正则表达式作为补充）
	p.parsePostgreSQLFunctions(result, sql)
	
	// 处理 PostgreSQL 特定操作符（仍使用正则表达式作为补充）
	p.parsePostgreSQLOperators(result, sql)
	
	// 处理 PostgreSQL 特定数据类型（仍使用正则表达式作为补充）
	p.parsePostgreSQLDataTypes(result, sql)
}

// parsePostgreSQLLimitAST 使用 AST 解析 LIMIT/OFFSET
func (p *PostgreSQLParser) parsePostgreSQLLimitAST(result *EnhancedSQLStatement, stmt tree.Statement) {
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

// parseReturningClauseAST 使用 AST 解析 RETURNING 子句
func (p *PostgreSQLParser) parseReturningClauseAST(result *EnhancedSQLStatement, stmt tree.Statement) {
	switch s := stmt.(type) {
	case *tree.Insert:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				// 为了保持与现有测试的兼容性，如果只有一个列，返回字符串
				if len(returningCols) == 1 {
					result.PostgreSQLFeatures["returning"] = returningCols[0]
				} else {
					result.PostgreSQLFeatures["returning"] = returningCols
				}
			}
		}
	case *tree.Update:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				// 为了保持与现有测试的兼容性，如果只有一个列，返回字符串
				if len(returningCols) == 1 {
					result.PostgreSQLFeatures["returning"] = returningCols[0]
				} else {
					result.PostgreSQLFeatures["returning"] = returningCols
				}
			}
		}
	case *tree.Delete:
		if s.Returning != nil {
			if returningExprs, ok := s.Returning.(*tree.ReturningExprs); ok {
				returningCols := make([]string, len(*returningExprs))
				for i, selectExpr := range *returningExprs {
					returningCols[i] = tree.AsString(selectExpr.Expr)
				}
				// 为了保持与现有测试的兼容性，如果只有一个列，返回字符串
				if len(returningCols) == 1 {
					result.PostgreSQLFeatures["returning"] = returningCols[0]
				} else {
					result.PostgreSQLFeatures["returning"] = returningCols
				}
			}
		}
	}
}

// enhancePostgreSQLFeaturesRegex 使用正则表达式增强 PostgreSQL 特定功能（向后兼容）
func (p *PostgreSQLParser) enhancePostgreSQLFeaturesRegex(result *EnhancedSQLStatement, sql string) {
	// 初始化 PostgreSQL 特性字段
	if result.PostgreSQLFeatures == nil {
		result.PostgreSQLFeatures = make(map[string]interface{})
	}
	
	// 处理 PostgreSQL 特定的 LIMIT/OFFSET 语法
	p.parsePostgreSQLLimit(result, sql)
	
	// 处理 PostgreSQL 特定的数据类型
	p.parsePostgreSQLDataTypes(result, sql)
	
	// 处理 PostgreSQL 特定的函数
	p.parsePostgreSQLFunctions(result, sql)
	
	// 处理 PostgreSQL 特定的操作符
	p.parsePostgreSQLOperators(result, sql)
	
	// 处理 PostgreSQL 特定的 RETURNING 子句
	p.parseReturningClause(result, sql)
}

// parsePostgreSQLLimit 解析 PostgreSQL 的 LIMIT/OFFSET 语法
func (p *PostgreSQLParser) parsePostgreSQLLimit(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL: LIMIT count OFFSET start
	limitOffsetRegex := regexp.MustCompile(`(?i)\s+LIMIT\s+(\d+|\$\d+|\?)\s+OFFSET\s+(\d+|\$\d+|\?)`)
	matches := limitOffsetRegex.FindStringSubmatch(sql)
	
	if len(matches) == 3 {
		if result.PostgreSQLFeatures == nil {
			result.PostgreSQLFeatures = make(map[string]interface{})
		}
		result.PostgreSQLFeatures["limit"] = matches[1]
		result.PostgreSQLFeatures["offset"] = matches[2]
		return
	}
	
	// 只有 LIMIT 的情况
	limitRegex := regexp.MustCompile(`(?i)\s+LIMIT\s+(\d+|\$\d+|\?)`)
	matches = limitRegex.FindStringSubmatch(sql)
	if len(matches) == 2 {
		if result.PostgreSQLFeatures == nil {
			result.PostgreSQLFeatures = make(map[string]interface{})
		}
		result.PostgreSQLFeatures["limit"] = matches[1]
	}
}

// parsePostgreSQLDataTypes 解析 PostgreSQL 特定数据类型
func (p *PostgreSQLParser) parsePostgreSQLDataTypes(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL 特有数据类型
	pgTypes := []string{
		"SERIAL", "BIGSERIAL", "SMALLSERIAL",
		"UUID", "JSON", "JSONB", "XML",
		"INET", "CIDR", "MACADDR", "MACADDR8",
		"POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE",
		"TSVECTOR", "TSQUERY",
		"ARRAY", "HSTORE",
	}
	
	for _, pgType := range pgTypes {
		if strings.Contains(strings.ToUpper(sql), pgType) {
			if result.PostgreSQLFeatures == nil {
				result.PostgreSQLFeatures = make(map[string]interface{})
			}
			if result.PostgreSQLFeatures["dataTypes"] == nil {
				result.PostgreSQLFeatures["dataTypes"] = make([]string, 0)
			}
			dataTypes := result.PostgreSQLFeatures["dataTypes"].([]string)
			result.PostgreSQLFeatures["dataTypes"] = append(dataTypes, pgType)
		}
	}
	
	// 检查数组语法 (例如 TEXT[], INTEGER[])
	arrayRegex := regexp.MustCompile(`(?i)\w+\[\s*\]`)
	if arrayRegex.MatchString(sql) {
		if result.PostgreSQLFeatures == nil {
			result.PostgreSQLFeatures = make(map[string]interface{})
		}
		if result.PostgreSQLFeatures["dataTypes"] == nil {
			result.PostgreSQLFeatures["dataTypes"] = make([]string, 0)
		}
		dataTypes := result.PostgreSQLFeatures["dataTypes"].([]string)
		// 检查是否已经包含 ARRAY
		found := false
		for _, dt := range dataTypes {
			if dt == "ARRAY" {
				found = true
				break
			}
		}
		if !found {
			result.PostgreSQLFeatures["dataTypes"] = append(dataTypes, "ARRAY")
		}
	}
}

// parsePostgreSQLFunctions 解析 PostgreSQL 特定函数
func (p *PostgreSQLParser) parsePostgreSQLFunctions(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL 特有函数
	pgFunctions := map[string]*regexp.Regexp{
		"COALESCE":     regexp.MustCompile(`(?i)\bCOALESCE\s*\(`),
		"NULLIF":       regexp.MustCompile(`(?i)\bNULLIF\s*\(`),
		"GREATEST":     regexp.MustCompile(`(?i)\bGREATEST\s*\(`),
		"LEAST":        regexp.MustCompile(`(?i)\bLEAST\s*\(`),
		"GENERATE_SERIES": regexp.MustCompile(`(?i)\bGENERATE_SERIES\s*\(`),
		"ARRAY_AGG":    regexp.MustCompile(`(?i)\bARRAY_AGG\s*\(`),
		"STRING_AGG":   regexp.MustCompile(`(?i)\bSTRING_AGG\s*\(`),
		"JSONB_AGG":    regexp.MustCompile(`(?i)\bJSONB_AGG\s*\(`),
		"ROW_NUMBER":   regexp.MustCompile(`(?i)\bROW_NUMBER\s*\(\s*\)\s+OVER`),
		"RANK":         regexp.MustCompile(`(?i)\bRANK\s*\(\s*\)\s+OVER`),
		"DENSE_RANK":   regexp.MustCompile(`(?i)\bDENSE_RANK\s*\(\s*\)\s+OVER`),
		"LAG":          regexp.MustCompile(`(?i)\bLAG\s*\(`),
		"LEAD":         regexp.MustCompile(`(?i)\bLEAD\s*\(`),
	}
	
	for funcName, regex := range pgFunctions {
		if regex.MatchString(sql) {
			if result.PostgreSQLFeatures == nil {
				result.PostgreSQLFeatures = make(map[string]interface{})
			}
			if result.PostgreSQLFeatures["functions"] == nil {
				result.PostgreSQLFeatures["functions"] = make([]string, 0)
			}
			functions := result.PostgreSQLFeatures["functions"].([]string)
			result.PostgreSQLFeatures["functions"] = append(functions, funcName)
		}
	}
}

// parsePostgreSQLOperators 解析 PostgreSQL 特定操作符
func (p *PostgreSQLParser) parsePostgreSQLOperators(result *EnhancedSQLStatement, sql string) {
	// PostgreSQL 特有操作符
	pgOperators := map[string]*regexp.Regexp{
		"ILIKE":        regexp.MustCompile(`(?i)\s+ILIKE\s+`),
		"SIMILAR TO":   regexp.MustCompile(`(?i)\s+SIMILAR\s+TO\s+`),
		"~":            regexp.MustCompile(`\s+~\s+`),
		"~*":           regexp.MustCompile(`\s+~\*\s+`),
		"!~":           regexp.MustCompile(`\s+!~\s+`),
		"!~*":          regexp.MustCompile(`\s+!~\*\s+`),
		"@@":           regexp.MustCompile(`\s+@@\s+`),
		"@>":           regexp.MustCompile(`\s+@>\s+`),
		"<@":           regexp.MustCompile(`\s+<@\s+`),
		"?":            regexp.MustCompile(`\s+\?\s+`),
		"?&":           regexp.MustCompile(`\s+\?\&\s+`),
		"?|":           regexp.MustCompile(`\s+\?\|\s+`),
		"&&":           regexp.MustCompile(`\s+&&\s+`), // 数组重叠操作符
		"||":           regexp.MustCompile(`\s+\|\|\s+`), // 字符串连接
	}
	
	for opName, regex := range pgOperators {
		if regex.MatchString(sql) {
			if result.PostgreSQLFeatures == nil {
				result.PostgreSQLFeatures = make(map[string]interface{})
			}
			if result.PostgreSQLFeatures["operators"] == nil {
				result.PostgreSQLFeatures["operators"] = make([]string, 0)
			}
			operators := result.PostgreSQLFeatures["operators"].([]string)
			result.PostgreSQLFeatures["operators"] = append(operators, opName)
		}
	}
}

// parseReturningClause 解析 RETURNING 子句
func (p *PostgreSQLParser) parseReturningClause(result *EnhancedSQLStatement, sql string) {
	returningRegex := regexp.MustCompile(`(?i)\s+RETURNING\s+(.+?)(?:\s*$|\s+(?:LIMIT|ORDER|GROUP|HAVING))`)
	matches := returningRegex.FindStringSubmatch(sql)
	
	if len(matches) >= 2 {
		if result.PostgreSQLFeatures == nil {
			result.PostgreSQLFeatures = make(map[string]interface{})
		}
		result.PostgreSQLFeatures["returning"] = strings.TrimSpace(matches[1])
	}
}

// RewriteForPostgreSQL 为 PostgreSQL 重写 SQL
func (p *PostgreSQLParser) RewriteForPostgreSQL(sql string, tableName string, actualTableName string) (string, error) {
	// 替换表名
	rewrittenSQL := strings.ReplaceAll(sql, tableName, actualTableName)
	
	// 处理 PostgreSQL 特定的引用字符
	quoteChar := p.dialect.GetQuoteCharacter()
	if quoteChar != "`" {
		// 将 MySQL 风格的引用字符替换为 PostgreSQL 风格
		rewrittenSQL = strings.ReplaceAll(rewrittenSQL, "`", quoteChar)
	}
	
	// 处理参数占位符（PostgreSQL 使用 $1, $2, ... 而不是 ?）
	rewrittenSQL = p.convertParameterPlaceholders(rewrittenSQL)
	
	return rewrittenSQL, nil
}

// convertParameterPlaceholders 转换参数占位符
func (p *PostgreSQLParser) convertParameterPlaceholders(sql string) string {
	// 将 ? 占位符转换为 PostgreSQL 的 $1, $2, ... 格式
	paramCount := 0
	result := ""
	
	for i, char := range sql {
		if char == '?' {
			// 检查是否在字符串字面量中
			if !p.isInStringLiteral(sql, i) {
				paramCount++
				result += fmt.Sprintf("$%d", paramCount)
			} else {
				result += string(char)
			}
		} else {
			result += string(char)
		}
	}
	
	return result
}

// isInStringLiteral 检查位置是否在字符串字面量中
func (p *PostgreSQLParser) isInStringLiteral(sql string, pos int) bool {
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

// ValidatePostgreSQLSQL 验证 PostgreSQL SQL 语法
func (p *PostgreSQLParser) ValidatePostgreSQLSQL(sql string) error {
	// 基本的 PostgreSQL SQL 验证
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return fmt.Errorf("empty SQL statement")
	}
	
	// 检查是否包含不支持的 MySQL 特定语法
	mysqlSpecific := []string{
		"LIMIT \\d+,\\s*\\d+", // MySQL 风格的 LIMIT offset, count
		"`[^`]*`",             // MySQL 风格的反引号
		"AUTO_INCREMENT",      // MySQL 自增关键字
	}
	
	for _, pattern := range mysqlSpecific {
		matched, _ := regexp.MatchString("(?i)"+pattern, sql)
		if matched {
			return fmt.Errorf("MySQL-specific syntax detected, please use PostgreSQL syntax")
		}
	}
	
	return nil
}

// ExtractTables 提取表名（实现 ParserInterface 接口）
func (p *PostgreSQLParser) ExtractTables(sql string) []string {
	// 首先尝试使用 CockroachDB Parser
	stmts, err := p.parser.Parse(sql)
	if err == nil && len(stmts) > 0 {
		return p.extractTablesFromAST(stmts[0].AST)
	}
	
	// 如果 CockroachDB Parser 失败，回退到增强解析器
	tables, err := p.EnhancedSQLParser.extractTables(sql)
	if err != nil {
		// 如果解析失败，返回空切片
		return []string{}
	}
	return tables
}

// GetPostgreSQLDialect 获取 PostgreSQL 方言
func (p *PostgreSQLParser) GetPostgreSQLDialect() database.DatabaseDialect {
	return p.dialect
}