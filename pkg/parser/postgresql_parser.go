package parser

import (
	"fmt"
	"go-sharding/pkg/database"
	"regexp"
	"strings"
)

// PostgreSQLParser PostgreSQL 特定的 SQL 解析器
type PostgreSQLParser struct {
	*EnhancedSQLParser
	dialect database.DatabaseDialect
}

// NewPostgreSQLParser 创建 PostgreSQL 解析器
func NewPostgreSQLParser() *PostgreSQLParser {
	dialect, _ := database.GlobalDialectRegistry.GetDialect(database.PostgreSQL)
	return &PostgreSQLParser{
		EnhancedSQLParser: NewEnhancedSQLParser(),
		dialect:           dialect,
	}
}

// ParsePostgreSQLSpecific 解析 PostgreSQL 特定语法
func (p *PostgreSQLParser) ParsePostgreSQLSpecific(sql string) (*EnhancedSQLStatement, error) {
	// 首先使用基础解析器
	result, err := p.EnhancedSQLParser.Parse(sql)
	if err != nil {
		return nil, err
	}

	// 增强 PostgreSQL 特定功能
	p.enhancePostgreSQLFeatures(result, sql)
	
	return result, nil
}

// enhancePostgreSQLFeatures 增强 PostgreSQL 特定功能
func (p *PostgreSQLParser) enhancePostgreSQLFeatures(result *EnhancedSQLStatement, sql string) {
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

// GetPostgreSQLDialect 获取 PostgreSQL 方言
func (p *PostgreSQLParser) GetPostgreSQLDialect() database.DatabaseDialect {
	return p.dialect
}