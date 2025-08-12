package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLType SQL 语句类型
type SQLType string

const (
	SQLTypeSelect SQLType = "SELECT"
	SQLTypeInsert SQLType = "INSERT"
	SQLTypeUpdate SQLType = "UPDATE"
	SQLTypeDelete SQLType = "DELETE"
	SQLTypeCreate SQLType = "CREATE"
	SQLTypeDrop   SQLType = "DROP"
	SQLTypeAlter  SQLType = "ALTER"
	SQLTypeShow   SQLType = "SHOW"
	SQLTypeOther  SQLType = "OTHER"
)

// SQLStatement SQL 语句结构
type SQLStatement struct {
	Type         SQLType
	Tables       []string
	Columns      []string
	Conditions   []Condition
	JoinTables   []JoinTable
	OrderBy      []OrderByClause
	GroupBy      []string
	Having       []Condition
	Limit        *LimitClause
	OriginalSQL  string
}

// Condition 条件结构
type Condition struct {
	Column   string
	Operator string
	Value    interface{}
	Logic    string // AND, OR
}

// JoinTable 连接表结构
type JoinTable struct {
	Type      string // INNER, LEFT, RIGHT, FULL
	Table     string
	Condition string
}

// OrderByClause 排序子句
type OrderByClause struct {
	Column    string
	Direction string // ASC, DESC
}

// LimitClause 限制子句
type LimitClause struct {
	Offset int
	Count  int
}

// SQLParser SQL 解析器
type SQLParser struct {
	keywords map[string]bool
}

// NewSQLParser 创建 SQL 解析器
func NewSQLParser() *SQLParser {
	parser := &SQLParser{
		keywords: make(map[string]bool),
	}
	
	// 初始化 SQL 关键字
	keywords := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "INTO", "VALUES", "UPDATE", "SET",
		"DELETE", "CREATE", "TABLE", "DROP", "ALTER", "INDEX", "VIEW", "DATABASE",
		"SCHEMA", "AND", "OR", "NOT", "NULL", "TRUE", "FALSE", "JOIN", "INNER",
		"LEFT", "RIGHT", "FULL", "OUTER", "ON", "USING", "GROUP", "BY", "ORDER",
		"HAVING", "LIMIT", "OFFSET", "UNION", "ALL", "DISTINCT", "AS", "ASC",
		"DESC", "COUNT", "SUM", "AVG", "MIN", "MAX", "CASE", "WHEN", "THEN",
		"ELSE", "END", "IF", "EXISTS", "IN", "BETWEEN", "LIKE", "IS", "PRIMARY",
		"KEY", "FOREIGN", "REFERENCES", "UNIQUE", "CHECK", "DEFAULT", "AUTO_INCREMENT",
		"TIMESTAMP", "DATETIME", "DATE", "TIME", "VARCHAR", "CHAR", "TEXT", "INT",
		"INTEGER", "BIGINT", "SMALLINT", "TINYINT", "DECIMAL", "FLOAT", "DOUBLE",
		"BOOLEAN", "BLOB", "CLOB", "BINARY", "VARBINARY",
	}
	
	for _, keyword := range keywords {
		parser.keywords[keyword] = true
	}
	
	return parser
}

// Parse 解析 SQL 语句
func (p *SQLParser) Parse(sql string) (*SQLStatement, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, fmt.Errorf("empty SQL statement")
	}

	stmt := &SQLStatement{
		OriginalSQL: sql,
	}

	// 确定 SQL 类型
	stmt.Type = p.determineSQLType(sql)

	// 根据类型解析
	switch stmt.Type {
	case SQLTypeSelect:
		return p.parseSelect(sql, stmt)
	case SQLTypeInsert:
		return p.parseInsert(sql, stmt)
	case SQLTypeUpdate:
		return p.parseUpdate(sql, stmt)
	case SQLTypeDelete:
		return p.parseDelete(sql, stmt)
	case SQLTypeCreate:
		return p.parseCreate(sql, stmt)
	case SQLTypeDrop:
		return p.parseDrop(sql, stmt)
	case SQLTypeAlter:
		return p.parseAlter(sql, stmt)
	default:
		// 对于其他类型，尝试提取表名
		stmt.Tables = p.extractTables(sql)
		return stmt, nil
	}
}

// determineSQLType 确定 SQL 类型
func (p *SQLParser) determineSQLType(sql string) SQLType {
	sql = strings.TrimSpace(strings.ToUpper(sql))
	
	if strings.HasPrefix(sql, "SELECT") {
		return SQLTypeSelect
	} else if strings.HasPrefix(sql, "INSERT") {
		return SQLTypeInsert
	} else if strings.HasPrefix(sql, "UPDATE") {
		return SQLTypeUpdate
	} else if strings.HasPrefix(sql, "DELETE") {
		return SQLTypeDelete
	} else if strings.HasPrefix(sql, "CREATE") {
		return SQLTypeCreate
	} else if strings.HasPrefix(sql, "DROP") {
		return SQLTypeDrop
	} else if strings.HasPrefix(sql, "ALTER") {
		return SQLTypeAlter
	} else if strings.HasPrefix(sql, "SHOW") {
		return SQLTypeShow
	}
	
	return SQLTypeOther
}

// parseSelect 解析 SELECT 语句
func (p *SQLParser) parseSelect(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	// 移除注释和多余空格
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromSelect(sql)
	
	// 提取列名
	stmt.Columns = p.extractColumnsFromSelect(sql)
	
	// 提取 WHERE 条件
	stmt.Conditions = p.extractConditions(sql)
	
	// 提取 JOIN 表
	stmt.JoinTables = p.extractJoinTables(sql)
	
	// 提取 ORDER BY
	stmt.OrderBy = p.extractOrderBy(sql)
	
	// 提取 GROUP BY
	stmt.GroupBy = p.extractGroupBy(sql)
	
	// 提取 HAVING
	stmt.Having = p.extractHaving(sql)
	
	// 提取 LIMIT
	stmt.Limit = p.extractLimit(sql)
	
	return stmt, nil
}

// parseInsert 解析 INSERT 语句
func (p *SQLParser) parseInsert(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromInsert(sql)
	
	// 提取列名
	stmt.Columns = p.extractColumnsFromInsert(sql)
	
	return stmt, nil
}

// parseUpdate 解析 UPDATE 语句
func (p *SQLParser) parseUpdate(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromUpdate(sql)
	
	// 提取 SET 子句中的列
	stmt.Columns = p.extractColumnsFromUpdate(sql)
	
	// 提取 WHERE 条件
	stmt.Conditions = p.extractConditions(sql)
	
	return stmt, nil
}

// parseDelete 解析 DELETE 语句
func (p *SQLParser) parseDelete(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromDelete(sql)
	
	// 提取 WHERE 条件
	stmt.Conditions = p.extractConditions(sql)
	
	return stmt, nil
}

// parseCreate 解析 CREATE 语句
func (p *SQLParser) parseCreate(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromCreate(sql)
	
	return stmt, nil
}

// parseDrop 解析 DROP 语句
func (p *SQLParser) parseDrop(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromDrop(sql)
	
	return stmt, nil
}

// parseAlter 解析 ALTER 语句
func (p *SQLParser) parseAlter(sql string, stmt *SQLStatement) (*SQLStatement, error) {
	sql = p.cleanSQL(sql)
	
	// 提取表名
	stmt.Tables = p.extractTablesFromAlter(sql)
	
	return stmt, nil
}

// cleanSQL 清理 SQL 语句
func (p *SQLParser) cleanSQL(sql string) string {
	// 移除行注释
	sql = regexp.MustCompile(`--.*`).ReplaceAllString(sql, "")
	
	// 移除块注释
	sql = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(sql, "")
	
	// 标准化空格
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")
	
	return strings.TrimSpace(sql)
}

// extractTablesFromSelect 从 SELECT 语句中提取表名
func (p *SQLParser) extractTablesFromSelect(sql string) []string {
	var tables []string
	
	// 匹配 FROM 子句
	fromRegex := regexp.MustCompile(`(?i)\bFROM\s+([^\s]+(?:\s+[^\s]+)*?)(?:\s+WHERE|\s+GROUP\s+BY|\s+ORDER\s+BY|\s+HAVING|\s+LIMIT|\s+UNION|\s+JOIN|$)`)
	matches := fromRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		fromClause := strings.TrimSpace(matches[1])
		tables = append(tables, p.parseTableNames(fromClause)...)
	}
	
	return p.removeDuplicates(tables)
}

// extractTablesFromInsert 从 INSERT 语句中提取表名
func (p *SQLParser) extractTablesFromInsert(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bINTO\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// extractTablesFromUpdate 从 UPDATE 语句中提取表名
func (p *SQLParser) extractTablesFromUpdate(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bUPDATE\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// extractTablesFromDelete 从 DELETE 语句中提取表名
func (p *SQLParser) extractTablesFromDelete(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// extractTablesFromCreate 从 CREATE 语句中提取表名
func (p *SQLParser) extractTablesFromCreate(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bCREATE\s+TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// extractTablesFromDrop 从 DROP 语句中提取表名
func (p *SQLParser) extractTablesFromDrop(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bDROP\s+TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// extractTablesFromAlter 从 ALTER 语句中提取表名
func (p *SQLParser) extractTablesFromAlter(sql string) []string {
	regex := regexp.MustCompile(`(?i)\bALTER\s+TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z_][a-zA-Z0-9_]*)?)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		return []string{p.cleanTableName(matches[1])}
	}
	return []string{}
}

// ExtractTables 公共接口：提取表名
func (p *SQLParser) ExtractTables(sql string) []string {
	return p.extractTables(sql)
}

// extractTables 通用表名提取（兼容旧接口）
func (p *SQLParser) extractTables(sql string) []string {
	stmt, err := p.Parse(sql)
	if err != nil {
		// 如果解析失败，使用简单的正则表达式
		return p.extractTablesSimple(sql)
	}
	return stmt.Tables
}

// extractTablesSimple 简单的表名提取
func (p *SQLParser) extractTablesSimple(sql string) []string {
	var tables []string
	
	// 使用正则表达式提取可能的表名
	regex := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
	matches := regex.FindAllString(sql, -1)
	
	for _, match := range matches {
		if !p.keywords[strings.ToUpper(match)] {
			tables = append(tables, match)
		}
	}
	
	return p.removeDuplicates(tables)
}

// parseTableNames 解析表名列表
func (p *SQLParser) parseTableNames(tableClause string) []string {
	var tables []string
	
	// 分割多个表（逗号分隔）
	parts := strings.Split(tableClause, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// 移除别名
		words := strings.Fields(part)
		if len(words) > 0 {
			tableName := p.cleanTableName(words[0])
			if tableName != "" {
				tables = append(tables, tableName)
			}
		}
	}
	
	return tables
}

// cleanTableName 清理表名
func (p *SQLParser) cleanTableName(tableName string) string {
	// 移除引号
	tableName = strings.Trim(tableName, "`\"'")
	
	// 移除数据库前缀（如果存在）
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1]
	}
	
	return tableName
}

// extractColumnsFromSelect 从 SELECT 语句中提取列名
func (p *SQLParser) extractColumnsFromSelect(sql string) []string {
	regex := regexp.MustCompile(`(?i)SELECT\s+(.*?)\s+FROM`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		columnClause := strings.TrimSpace(matches[1])
		if columnClause == "*" {
			return []string{"*"}
		}
		return p.parseColumnNames(columnClause)
	}
	return []string{}
}

// extractColumnsFromInsert 从 INSERT 语句中提取列名
func (p *SQLParser) extractColumnsFromInsert(sql string) []string {
	regex := regexp.MustCompile(`(?i)\([^)]*\)\s*VALUES`)
	match := regex.FindString(sql)
	if match != "" {
		// 提取括号内的列名
		start := strings.Index(match, "(")
		end := strings.LastIndex(match, ")")
		if start >= 0 && end > start {
			columnClause := match[start+1 : end]
			return p.parseColumnNames(columnClause)
		}
	}
	return []string{}
}

// extractColumnsFromUpdate 从 UPDATE 语句中提取列名
func (p *SQLParser) extractColumnsFromUpdate(sql string) []string {
	regex := regexp.MustCompile(`(?i)SET\s+(.*?)(?:\s+WHERE|$)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		setClause := strings.TrimSpace(matches[1])
		return p.parseSetColumns(setClause)
	}
	return []string{}
}

// parseColumnNames 解析列名列表
func (p *SQLParser) parseColumnNames(columnClause string) []string {
	var columns []string
	
	// 分割多个列（逗号分隔）
	parts := strings.Split(columnClause, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// 移除别名和函数
		words := strings.Fields(part)
		if len(words) > 0 {
			columnName := p.cleanColumnName(words[0])
			if columnName != "" && columnName != "*" {
				columns = append(columns, columnName)
			}
		}
	}
	
	return p.removeDuplicates(columns)
}

// parseSetColumns 解析 SET 子句中的列名
func (p *SQLParser) parseSetColumns(setClause string) []string {
	var columns []string
	
	// 分割多个赋值（逗号分隔）
	parts := strings.Split(setClause, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// 提取等号前的列名
		if equalIndex := strings.Index(part, "="); equalIndex > 0 {
			columnName := strings.TrimSpace(part[:equalIndex])
			columnName = p.cleanColumnName(columnName)
			if columnName != "" {
				columns = append(columns, columnName)
			}
		}
	}
	
	return columns
}

// cleanColumnName 清理列名
func (p *SQLParser) cleanColumnName(columnName string) string {
	// 移除引号
	columnName = strings.Trim(columnName, "`\"'")
	
	// 移除表前缀（如果存在）
	if strings.Contains(columnName, ".") {
		parts := strings.Split(columnName, ".")
		columnName = parts[len(parts)-1]
	}
	
	return columnName
}

// extractConditions 提取 WHERE 条件
func (p *SQLParser) extractConditions(sql string) []Condition {
	// 简化实现，仅提取基本条件
	var conditions []Condition
	
	regex := regexp.MustCompile(`(?i)WHERE\s+(.*?)(?:\s+GROUP\s+BY|\s+ORDER\s+BY|\s+HAVING|\s+LIMIT|$)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		whereClause := strings.TrimSpace(matches[1])
		// 这里可以进一步解析具体的条件
		// 简化实现，仅记录存在 WHERE 子句
		if whereClause != "" {
			conditions = append(conditions, Condition{
				Column:   "unknown",
				Operator: "unknown",
				Value:    whereClause,
				Logic:    "AND",
			})
		}
	}
	
	return conditions
}

// extractJoinTables 提取 JOIN 表
func (p *SQLParser) extractJoinTables(sql string) []JoinTable {
	var joinTables []JoinTable
	
	// 分步处理多个 JOIN
	// 首先找到所有 JOIN 的位置
	joinRegex := regexp.MustCompile(`(?i)(INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|FULL\s+JOIN|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+(?:[a-zA-Z_][a-zA-Z0-9_]*\s+)?ON\s+`)
	joinPositions := joinRegex.FindAllStringIndex(sql, -1)
	joinMatches := joinRegex.FindAllStringSubmatch(sql, -1)
	
	for i, match := range joinMatches {
		if len(match) >= 3 {
			joinType := strings.ToUpper(strings.TrimSpace(match[1]))
			tableName := p.cleanTableName(match[2])
			
			// 找到 ON 子句的开始位置
			onStart := joinPositions[i][1]
			
			// 找到下一个 JOIN 或 SQL 子句的位置作为结束位置
			var onEnd int
			if i+1 < len(joinPositions) {
				onEnd = joinPositions[i+1][0]
			} else {
				// 查找其他 SQL 子句
				endRegex := regexp.MustCompile(`(?i)\s+(WHERE|GROUP\s+BY|ORDER\s+BY|HAVING|LIMIT)`)
				endMatch := endRegex.FindStringIndex(sql[onStart:])
				if endMatch != nil {
					onEnd = onStart + endMatch[0]
				} else {
					onEnd = len(sql)
				}
			}
			
			// 提取 ON 条件
			condition := strings.TrimSpace(sql[onStart:onEnd])
			
			joinTables = append(joinTables, JoinTable{
				Type:      joinType,
				Table:     tableName,
				Condition: condition,
			})
		}
	}
	
	return joinTables
}

// extractOrderBy 提取 ORDER BY 子句
func (p *SQLParser) extractOrderBy(sql string) []OrderByClause {
	var orderBy []OrderByClause
	
	regex := regexp.MustCompile(`(?i)ORDER\s+BY\s+(.*?)(?:\s+LIMIT|$)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		orderClause := strings.TrimSpace(matches[1])
		parts := strings.Split(orderClause, ",")
		
		for _, part := range parts {
			part = strings.TrimSpace(part)
			words := strings.Fields(part)
			if len(words) > 0 {
				column := p.cleanColumnName(words[0])
				direction := "ASC"
				if len(words) > 1 && strings.ToUpper(words[1]) == "DESC" {
					direction = "DESC"
				}
				
				orderBy = append(orderBy, OrderByClause{
					Column:    column,
					Direction: direction,
				})
			}
		}
	}
	
	return orderBy
}

// extractGroupBy 提取 GROUP BY 子句
func (p *SQLParser) extractGroupBy(sql string) []string {
	regex := regexp.MustCompile(`(?i)GROUP\s+BY\s+(.*?)(?:\s+HAVING|\s+ORDER\s+BY|\s+LIMIT|$)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		groupClause := strings.TrimSpace(matches[1])
		return p.parseColumnNames(groupClause)
	}
	return []string{}
}

// extractHaving 提取 HAVING 子句
func (p *SQLParser) extractHaving(sql string) []Condition {
	// 简化实现，类似 WHERE 条件
	var conditions []Condition
	
	regex := regexp.MustCompile(`(?i)HAVING\s+(.*?)(?:\s+ORDER\s+BY|\s+LIMIT|$)`)
	matches := regex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		havingClause := strings.TrimSpace(matches[1])
		if havingClause != "" {
			conditions = append(conditions, Condition{
				Column:   "unknown",
				Operator: "unknown",
				Value:    havingClause,
				Logic:    "AND",
			})
		}
	}
	
	return conditions
}

// extractLimit 提取 LIMIT 子句
func (p *SQLParser) extractLimit(sql string) *LimitClause {
	// 首先检查 MySQL 风格的 LIMIT offset, count
	mysqlRegex := regexp.MustCompile(`(?i)LIMIT\s+(\d+)\s*,\s*(\d+)`)
	matches := mysqlRegex.FindStringSubmatch(sql)
	if len(matches) > 2 {
		limit := &LimitClause{}
		
		if offset, err := parseInt(matches[1]); err == nil {
			limit.Offset = offset
		}
		
		if count, err := parseInt(matches[2]); err == nil {
			limit.Count = count
		}
		
		return limit
	}
	
	// 然后检查标准的 LIMIT count [OFFSET offset]
	standardRegex := regexp.MustCompile(`(?i)LIMIT\s+(\d+)(?:\s+OFFSET\s+(\d+))?`)
	matches = standardRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		limit := &LimitClause{}
		
		// 解析 count
		if count := matches[1]; count != "" {
			if c, err := parseInt(count); err == nil {
				limit.Count = c
			}
		}
		
		// 解析 offset
		if len(matches) > 2 && matches[2] != "" {
			if offset, err := parseInt(matches[2]); err == nil {
				limit.Offset = offset
			}
		}
		
		return limit
	}
	
	return nil
}

// removeDuplicates 移除重复项
func (p *SQLParser) removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range items {
		if !seen[item] && item != "" {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid integer: %s", s)
		}
		result = result*10 + int(r-'0')
	}
	return result, nil
}

// IsKeyword 检查是否为 SQL 关键字
func (p *SQLParser) IsKeyword(word string) bool {
	return p.keywords[strings.ToUpper(word)]
}

// GetTables 获取 SQL 中的表名（兼容旧接口）
func (p *SQLParser) GetTables(sql string) []string {
	return p.extractTables(sql)
}