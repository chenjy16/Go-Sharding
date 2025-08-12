package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// EnhancedSQLParser 增强的SQL解析器
type EnhancedSQLParser struct {
	keywords map[string]bool
}

// NewEnhancedSQLParser 创建增强的SQL解析器
func NewEnhancedSQLParser() *EnhancedSQLParser {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true, "UPDATE": true, "DELETE": true,
		"JOIN": true, "INNER": true, "LEFT": true, "RIGHT": true, "FULL": true, "OUTER": true,
		"ON": true, "AND": true, "OR": true, "NOT": true, "IN": true, "EXISTS": true,
		"GROUP": true, "BY": true, "HAVING": true, "ORDER": true, "LIMIT": true, "OFFSET": true,
		"UNION": true, "ALL": true, "DISTINCT": true, "AS": true, "CASE": true, "WHEN": true,
		"THEN": true, "ELSE": true, "END": true, "NULL": true, "IS": true, "LIKE": true,
		"BETWEEN": true, "ASC": true, "DESC": true, "CREATE": true, "TABLE": true, "INDEX": true,
		"DROP": true, "ALTER": true, "ADD": true, "COLUMN": true, "PRIMARY": true, "KEY": true,
		"FOREIGN": true, "REFERENCES": true, "CONSTRAINT": true, "UNIQUE": true, "DEFAULT": true,
		"AUTO_INCREMENT": true, "CHECK": true, "WITH": true,
	}
	
	return &EnhancedSQLParser{
		keywords: keywords,
	}
}

// EnhancedSQLStatement 增强的SQL语句结构
type EnhancedSQLStatement struct {
	Type               SQLType                // SQL类型：SELECT, INSERT, UPDATE, DELETE
	Tables             []string               // 涉及的表
	Columns            []string               // 涉及的列
	Conditions         map[string]interface{} // WHERE条件
	JoinTables         []JoinInfo             // JOIN信息
	SubQueries         []*EnhancedSQLStatement // 子查询
	GroupBy            []string               // GROUP BY列
	OrderBy            []OrderByInfo          // ORDER BY信息
	Having             string                 // HAVING条件
	Limit              *LimitInfo             // LIMIT信息
	IsComplex          bool                   // 是否为复杂查询
	OriginalSQL        string                 // 原始SQL
	Parameters         []interface{}          // 参数
	PostgreSQLFeatures map[string]interface{} // PostgreSQL 特定功能
}

// JoinInfo JOIN信息
type JoinInfo struct {
	Type      string // INNER, LEFT, RIGHT, FULL
	Table     string // 表名
	Condition string // JOIN条件
}

// OrderByInfo ORDER BY信息
type OrderByInfo struct {
	Column    string // 列名
	Direction string // ASC, DESC
}

// LimitInfo LIMIT信息
type LimitInfo struct {
	Count  int // 限制数量
	Offset int // 偏移量
}

// Parse 解析SQL语句
func (p *EnhancedSQLParser) Parse(sql string) (*EnhancedSQLStatement, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, fmt.Errorf("empty SQL statement")
	}
	
	stmt := &EnhancedSQLStatement{
		OriginalSQL: sql,
		Conditions:  make(map[string]interface{}),
	}
	
	// 确定SQL类型
	sqlUpper := strings.ToUpper(sql)
	switch {
	case strings.HasPrefix(sqlUpper, "SELECT"):
		stmt.Type = SQLTypeSelect
		return p.parseSelect(sql, stmt)
	case strings.HasPrefix(sqlUpper, "INSERT"):
		stmt.Type = SQLTypeInsert
		return p.parseInsert(sql, stmt)
	case strings.HasPrefix(sqlUpper, "UPDATE"):
		stmt.Type = SQLTypeUpdate
		return p.parseUpdate(sql, stmt)
	case strings.HasPrefix(sqlUpper, "DELETE"):
		stmt.Type = SQLTypeDelete
		return p.parseDelete(sql, stmt)
	case strings.HasPrefix(sqlUpper, "CREATE"):
		stmt.Type = SQLTypeCreate
		return p.parseCreate(sql, stmt)
	case strings.HasPrefix(sqlUpper, "DROP"):
		stmt.Type = SQLTypeDrop
		return p.parseDrop(sql, stmt)
	case strings.HasPrefix(sqlUpper, "ALTER"):
		stmt.Type = SQLTypeAlter
		return p.parseAlter(sql, stmt)
	default:
		stmt.Type = SQLTypeOther
		return stmt, nil
	}
}

// parseSelect 解析SELECT语句
func (p *EnhancedSQLParser) parseSelect(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// 检查是否包含子查询
	if p.hasSubQuery(sql) {
		stmt.IsComplex = true
		subQueries, err := p.extractSubQueries(sql)
		if err != nil {
			return nil, fmt.Errorf("failed to extract sub queries: %w", err)
		}
		stmt.SubQueries = subQueries
	}
	
	// 检查是否包含JOIN
	if p.hasJoin(sql) {
		stmt.IsComplex = true
		joinInfo, err := p.extractJoinInfo(sql)
		if err != nil {
			return nil, fmt.Errorf("failed to extract join info: %w", err)
		}
		stmt.JoinTables = joinInfo
	}
	
	// 提取表名
	tables, err := p.extractTables(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tables: %w", err)
	}
	stmt.Tables = tables
	
	// 提取列名
	columns, err := p.extractColumns(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract columns: %w", err)
	}
	stmt.Columns = columns
	
	// 提取WHERE条件
	conditions, err := p.extractConditions(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract conditions: %w", err)
	}
	stmt.Conditions = conditions
	
	// 提取GROUP BY
	groupBy, err := p.extractGroupBy(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract group by: %w", err)
	}
	stmt.GroupBy = groupBy
	
	// 提取ORDER BY
	orderBy, err := p.extractOrderBy(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract order by: %w", err)
	}
	stmt.OrderBy = orderBy
	
	// 提取LIMIT
	limitInfo, err := p.extractLimit(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract limit: %w", err)
	}
	stmt.Limit = limitInfo
	
	return stmt, nil
}

// parseInsert 解析INSERT语句
func (p *EnhancedSQLParser) parseInsert(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// INSERT INTO table_name (columns) VALUES (values)
	insertRegex := regexp.MustCompile(`(?i)INSERT\s+INTO\s+(\w+)`)
	matches := insertRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	// 提取列名
	columnRegex := regexp.MustCompile(`(?i)\(([^)]+)\)\s+VALUES`)
	columnMatches := columnRegex.FindStringSubmatch(sql)
	if len(columnMatches) > 1 {
		columns := strings.Split(columnMatches[1], ",")
		for i, col := range columns {
			columns[i] = strings.TrimSpace(col)
		}
		stmt.Columns = columns
	}
	
	return stmt, nil
}

// parseUpdate 解析UPDATE语句
func (p *EnhancedSQLParser) parseUpdate(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// UPDATE table_name SET column = value WHERE condition
	updateRegex := regexp.MustCompile(`(?i)UPDATE\s+(\w+)`)
	matches := updateRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	// 提取SET子句中的列
	setRegex := regexp.MustCompile(`(?i)SET\s+(.+?)(?:\s+WHERE|$)`)
	setMatches := setRegex.FindStringSubmatch(sql)
	if len(setMatches) > 1 {
		setPairs := strings.Split(setMatches[1], ",")
		for _, pair := range setPairs {
			parts := strings.Split(pair, "=")
			if len(parts) >= 1 {
				column := strings.TrimSpace(parts[0])
				stmt.Columns = append(stmt.Columns, column)
			}
		}
	}
	
	// 提取WHERE条件
	conditions, err := p.extractConditions(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract conditions: %w", err)
	}
	stmt.Conditions = conditions
	
	return stmt, nil
}

// parseDelete 解析DELETE语句
func (p *EnhancedSQLParser) parseDelete(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// DELETE FROM table_name WHERE condition
	deleteRegex := regexp.MustCompile(`(?i)DELETE\s+FROM\s+(\w+)`)
	matches := deleteRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	// 提取WHERE条件
	conditions, err := p.extractConditions(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract conditions: %w", err)
	}
	stmt.Conditions = conditions
	
	return stmt, nil
}

// hasSubQuery 检查是否包含子查询
func (p *EnhancedSQLParser) hasSubQuery(sql string) bool {
	// 简单检查是否包含嵌套的SELECT
	selectCount := strings.Count(strings.ToUpper(sql), "SELECT")
	return selectCount > 1
}

// hasJoin 检查是否包含JOIN
func (p *EnhancedSQLParser) hasJoin(sql string) bool {
	sqlUpper := strings.ToUpper(sql)
	return strings.Contains(sqlUpper, " JOIN ") ||
		strings.Contains(sqlUpper, " INNER JOIN ") ||
		strings.Contains(sqlUpper, " LEFT JOIN ") ||
		strings.Contains(sqlUpper, " RIGHT JOIN ") ||
		strings.Contains(sqlUpper, " FULL JOIN ")
}

// extractSubQueries 提取子查询
func (p *EnhancedSQLParser) extractSubQueries(sql string) ([]*EnhancedSQLStatement, error) {
	var subQueries []*EnhancedSQLStatement
	
	// 使用正则表达式查找括号中的SELECT语句
	subQueryRegex := regexp.MustCompile(`\(([^()]*SELECT[^()]*)\)`)
	matches := subQueryRegex.FindAllStringSubmatch(sql, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			subSQL := strings.TrimSpace(match[1])
			subStmt, err := p.Parse(subSQL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse sub query: %w", err)
			}
			subQueries = append(subQueries, subStmt)
		}
	}
	
	return subQueries, nil
}

// extractJoinInfo 提取JOIN信息
func (p *EnhancedSQLParser) extractJoinInfo(sql string) ([]JoinInfo, error) {
	var joinInfo []JoinInfo
	
	joinRegex := regexp.MustCompile(`(?i)(INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|FULL\s+JOIN|JOIN)\s+(\w+)\s+ON\s+([^WHERE\s]+)`)
	matches := joinRegex.FindAllStringSubmatch(sql, -1)
	
	for _, match := range matches {
		if len(match) >= 4 {
			joinType := strings.ToUpper(strings.TrimSpace(match[1]))
			if joinType == "JOIN" {
				joinType = "INNER JOIN"
			}
			
			joinInfo = append(joinInfo, JoinInfo{
				Type:      joinType,
				Table:     strings.TrimSpace(match[2]),
				Condition: strings.TrimSpace(match[3]),
			})
		}
	}
	
	return joinInfo, nil
}

// extractGroupBy 提取GROUP BY
func (p *EnhancedSQLParser) extractGroupBy(sql string) ([]string, error) {
	groupByRegex := regexp.MustCompile(`(?i)GROUP\s+BY\s+([^HAVING\s]+)`)
	matches := groupByRegex.FindStringSubmatch(sql)
	
	if len(matches) > 1 {
		columns := strings.Split(matches[1], ",")
		for i, col := range columns {
			columns[i] = strings.TrimSpace(col)
		}
		return columns, nil
	}
	
	return nil, nil
}

// extractTables 提取表名
func (p *EnhancedSQLParser) extractTables(sql string) ([]string, error) {
	var tables []string
	
	// 提取FROM子句中的表
	fromRegex := regexp.MustCompile(`(?i)FROM\s+(\w+)`)
	matches := fromRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}
	
	// 提取JOIN中的表
	joinRegex := regexp.MustCompile(`(?i)JOIN\s+(\w+)`)
	joinMatches := joinRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range joinMatches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}
	
	return tables, nil
}

// extractColumns 提取列名
func (p *EnhancedSQLParser) extractColumns(sql string) ([]string, error) {
	var columns []string
	
	// 对于SELECT语句，提取SELECT子句中的列
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT") {
		selectRegex := regexp.MustCompile(`(?i)SELECT\s+(.+?)\s+FROM`)
		matches := selectRegex.FindStringSubmatch(sql)
		if len(matches) > 1 {
			columnStr := matches[1]
			if strings.TrimSpace(columnStr) != "*" {
				cols := strings.Split(columnStr, ",")
				for _, col := range cols {
					col = strings.TrimSpace(col)
					// 移除别名
					if strings.Contains(col, " AS ") {
						parts := strings.Split(col, " AS ")
						col = strings.TrimSpace(parts[0])
					}
					columns = append(columns, col)
				}
			}
		}
	}
	
	return columns, nil
}

// extractConditions 提取WHERE条件
func (p *EnhancedSQLParser) extractConditions(sql string) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	
	// 提取WHERE子句
	whereRegex := regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:\s+GROUP\s+BY|\s+ORDER\s+BY|\s+LIMIT|$)`)
	matches := whereRegex.FindStringSubmatch(sql)
	
	if len(matches) > 1 {
		whereClause := matches[1]
		
		// 简单的条件解析：column = value
		conditionRegex := regexp.MustCompile(`(\w+)\s*=\s*([^AND\s]+)`)
		condMatches := conditionRegex.FindAllStringSubmatch(whereClause, -1)
		
		for _, match := range condMatches {
			if len(match) >= 3 {
				column := strings.TrimSpace(match[1])
				value := strings.TrimSpace(match[2])
				// 移除引号
				value = strings.Trim(value, "'\"")
				conditions[column] = value
			}
		}
		
		// 处理IN条件
		inRegex := regexp.MustCompile(`(\w+)\s+IN\s*\(([^)]+)\)`)
		inMatches := inRegex.FindAllStringSubmatch(whereClause, -1)
		
		for _, match := range inMatches {
			if len(match) >= 3 {
				column := strings.TrimSpace(match[1])
				valuesStr := strings.TrimSpace(match[2])
				values := strings.Split(valuesStr, ",")
				var cleanValues []interface{}
				for _, v := range values {
					v = strings.TrimSpace(v)
					v = strings.Trim(v, "'\"")
					cleanValues = append(cleanValues, v)
				}
				conditions[column] = cleanValues
			}
		}
	}
	
	return conditions, nil
}

// extractOrderBy 提取ORDER BY
func (p *EnhancedSQLParser) extractOrderBy(sql string) ([]OrderByInfo, error) {
	var orderByInfo []OrderByInfo
	
	orderByRegex := regexp.MustCompile(`(?i)ORDER\s+BY\s+([^LIMIT\s]+)`)
	matches := orderByRegex.FindStringSubmatch(sql)
	
	if len(matches) > 1 {
		orderItems := strings.Split(matches[1], ",")
		for _, item := range orderItems {
			item = strings.TrimSpace(item)
			parts := strings.Fields(item)
			if len(parts) >= 1 {
				orderBy := OrderByInfo{
					Column:    parts[0],
					Direction: "ASC", // 默认
				}
				if len(parts) >= 2 && strings.ToUpper(parts[1]) == "DESC" {
					orderBy.Direction = "DESC"
				}
				orderByInfo = append(orderByInfo, orderBy)
			}
		}
	}
	
	return orderByInfo, nil
}

// extractLimit 提取LIMIT
func (p *EnhancedSQLParser) extractLimit(sql string) (*LimitInfo, error) {
	limitRegex := regexp.MustCompile(`(?i)LIMIT\s+(\d+)(?:\s+OFFSET\s+(\d+))?`)
	matches := limitRegex.FindStringSubmatch(sql)
	
	if len(matches) >= 2 {
		count := 0
		offset := 0
		
		if matches[1] != "" {
			fmt.Sscanf(matches[1], "%d", &count)
		}
		
		if len(matches) >= 3 && matches[2] != "" {
			fmt.Sscanf(matches[2], "%d", &offset)
		}
		
		return &LimitInfo{
			Count:  count,
			Offset: offset,
		}, nil
	}
	
	return nil, nil
}

// parseCreate 解析CREATE语句
func (p *EnhancedSQLParser) parseCreate(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// CREATE TABLE table_name (columns...)
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(\w+)`)
	matches := createTableRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	return stmt, nil
}

// parseDrop 解析DROP语句
func (p *EnhancedSQLParser) parseDrop(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// DROP TABLE table_name
	dropTableRegex := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(\w+)`)
	matches := dropTableRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	return stmt, nil
}

// ExtractTables 公共接口：提取表名
func (p *EnhancedSQLParser) ExtractTables(sql string) []string {
	tables, err := p.extractTables(sql)
	if err != nil {
		return []string{}
	}
	return tables
}

// parseAlter 解析ALTER语句
func (p *EnhancedSQLParser) parseAlter(sql string, stmt *EnhancedSQLStatement) (*EnhancedSQLStatement, error) {
	// ALTER TABLE table_name ...
	alterTableRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(\w+)`)
	matches := alterTableRegex.FindStringSubmatch(sql)
	if len(matches) > 1 {
		stmt.Tables = []string{matches[1]}
	}
	
	return stmt, nil
}