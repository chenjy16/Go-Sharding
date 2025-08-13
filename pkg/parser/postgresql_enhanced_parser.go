package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
)

// PostgreSQLEnhancedParser 增强的 PostgreSQL 解析器
type PostgreSQLEnhancedParser struct {
	*PostgreSQLParser
	optimizer *SQLOptimizer
}

// SQLOptimizer SQL 优化器
type SQLOptimizer struct {
	suggestions []OptimizationSuggestion
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
	LineNumber  int    `json:"line_number,omitempty"`
}

// EnhancedSQLAnalysis 增强的 SQL 分析结果
type EnhancedSQLAnalysis struct {
	*EnhancedSQLStatement
	Subqueries         []SubqueryInfo         `json:"subqueries,omitempty"`
	CTEs               []CTEInfo              `json:"ctes,omitempty"`
	Joins              []EnhancedJoinInfo     `json:"joins,omitempty"`
	WindowFunctions    []WindowFunctionInfo   `json:"window_functions,omitempty"`
	Optimizations      []OptimizationSuggestion `json:"optimizations,omitempty"`
	Complexity         ComplexityMetrics      `json:"complexity"`
	TableDependencies  map[string][]string    `json:"table_dependencies,omitempty"`
}

// SubqueryInfo 子查询信息
type SubqueryInfo struct {
	Type       string   `json:"type"`       // EXISTS, IN, SCALAR, etc.
	Tables     []string `json:"tables"`
	Columns    []string `json:"columns"`
	Nested     bool     `json:"nested"`
	Correlated bool     `json:"correlated"`
}

// CTEInfo CTE (Common Table Expression) 信息
type CTEInfo struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	Tables     []string `json:"tables"`
	Recursive  bool     `json:"recursive"`
}

// EnhancedJoinInfo 增强的连接信息
type EnhancedJoinInfo struct {
	Type       string   `json:"type"`       // INNER, LEFT, RIGHT, FULL, CROSS
	LeftTable  string   `json:"left_table"`
	RightTable string   `json:"right_table"`
	Condition  string   `json:"condition"`
	Columns    []string `json:"columns"`
}

// WindowFunctionInfo 窗口函数信息
type WindowFunctionInfo struct {
	Function    string   `json:"function"`
	PartitionBy []string `json:"partition_by"`
	OrderBy     []string `json:"order_by"`
	Frame       string   `json:"frame,omitempty"`
}

// ComplexityMetrics 复杂度指标
type ComplexityMetrics struct {
	Score          int `json:"score"`
	TableCount     int `json:"table_count"`
	JoinCount      int `json:"join_count"`
	SubqueryCount  int `json:"subquery_count"`
	CTECount       int `json:"cte_count"`
	WindowFuncCount int `json:"window_func_count"`
	NestingLevel   int `json:"nesting_level"`
}

// NewPostgreSQLEnhancedParser 创建增强的 PostgreSQL 解析器
func NewPostgreSQLEnhancedParser() *PostgreSQLEnhancedParser {
	return &PostgreSQLEnhancedParser{
		PostgreSQLParser: NewPostgreSQLParser(),
		optimizer:        &SQLOptimizer{suggestions: make([]OptimizationSuggestion, 0)},
	}
}

// AnalyzeSQL 分析 SQL 语句（增强版本）
func (p *PostgreSQLEnhancedParser) AnalyzeSQL(sql string) (*EnhancedSQLAnalysis, error) {
	// 首先使用基础解析器解析
	baseResult, err := p.PostgreSQLParser.ParsePostgreSQLSpecific(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	// 创建增强分析结果
	analysis := &EnhancedSQLAnalysis{
		EnhancedSQLStatement: baseResult,
		Subqueries:          make([]SubqueryInfo, 0),
		CTEs:                make([]CTEInfo, 0),
		Joins:               make([]EnhancedJoinInfo, 0),
		WindowFunctions:     make([]WindowFunctionInfo, 0),
		Optimizations:       make([]OptimizationSuggestion, 0),
		TableDependencies:   make(map[string][]string),
	}

	// 使用 AST 进行深度分析
	stmts, err := p.parser.Parse(sql)
	if err == nil && len(stmts) > 0 {
		p.analyzeASTDeep(analysis, stmts[0].AST, sql)
	}

	// 计算复杂度指标
	p.calculateComplexity(analysis)

	// 生成优化建议
	p.generateOptimizationSuggestions(analysis, sql)

	return analysis, nil
}

// analyzeASTDeep 深度分析 AST
func (p *PostgreSQLEnhancedParser) analyzeASTDeep(analysis *EnhancedSQLAnalysis, stmt tree.Statement, sql string) {
	switch s := stmt.(type) {
	case *tree.Select:
		p.analyzeSelectStatement(analysis, s)
	case *tree.Insert:
		p.analyzeInsertStatement(analysis, s)
	case *tree.Update:
		p.analyzeUpdateStatement(analysis, s)
	case *tree.Delete:
		p.analyzeDeleteStatement(analysis, s)
	}
}

// analyzeSelectStatement 分析 SELECT 语句
func (p *PostgreSQLEnhancedParser) analyzeSelectStatement(analysis *EnhancedSQLAnalysis, sel *tree.Select) {
	// 分析 CTE
	if sel.With != nil {
		p.analyzeCTEs(analysis, sel.With)
	}

	// 分析主查询
	if selectClause, ok := sel.Select.(*tree.SelectClause); ok {
		// 分析 FROM 子句
		if selectClause.From.Tables != nil {
			p.analyzeFromClause(analysis, selectClause.From.Tables)
		}

		// 分析 SELECT 列表中的子查询和窗口函数
		p.analyzeSelectList(analysis, selectClause.Exprs)

		// 分析 WHERE 子句中的子查询
		if selectClause.Where != nil {
			p.analyzeWhereClause(analysis, selectClause.Where.Expr)
		}

		// 分析 GROUP BY
		if selectClause.GroupBy != nil {
			p.analyzeGroupBy(analysis, selectClause.GroupBy)
		}

		// 分析 HAVING
		if selectClause.Having != nil {
			p.analyzeHaving(analysis, selectClause.Having.Expr)
		}
	}

	// 分析 ORDER BY
	if sel.OrderBy != nil {
		p.analyzeOrderBy(analysis, sel.OrderBy)
	}
}

// analyzeCTEs 分析 CTE
func (p *PostgreSQLEnhancedParser) analyzeCTEs(analysis *EnhancedSQLAnalysis, with *tree.With) {
	for _, cte := range with.CTEList {
		cteInfo := CTEInfo{
			Name:      cte.Name.Alias.String(),
			Columns:   make([]string, 0),
			Tables:    make([]string, 0),
			Recursive: with.Recursive,
		}

		// 提取 CTE 中的列名
		if cte.Name.Cols != nil {
			for _, col := range cte.Name.Cols {
				cteInfo.Columns = append(cteInfo.Columns, col.Name.String())
			}
		}

		// 分析 CTE 的查询部分
		if selectStmt, ok := cte.Stmt.(*tree.Select); ok {
			cteInfo.Tables = p.extractTablesFromSelect(selectStmt)
		}

		analysis.CTEs = append(analysis.CTEs, cteInfo)
	}
}

// analyzeFromClause 分析 FROM 子句
func (p *PostgreSQLEnhancedParser) analyzeFromClause(analysis *EnhancedSQLAnalysis, tables tree.TableExprs) {
	for _, tableExpr := range tables {
		p.analyzeTableExpr(analysis, tableExpr)
	}
}

// analyzeTableExpr 分析表表达式
func (p *PostgreSQLEnhancedParser) analyzeTableExpr(analysis *EnhancedSQLAnalysis, expr tree.TableExpr) {
	switch e := expr.(type) {
	case *tree.AliasedTableExpr:
		switch e.Expr.(type) {
		case *tree.TableName:
			// 普通表
		case *tree.Subquery:
			// 子查询
			subqueryInfo := SubqueryInfo{
				Type:    "FROM",
				Tables:  make([]string, 0),
				Columns: make([]string, 0),
				Nested:  true,
			}
			// 简化处理，直接从原始SQL中提取
			analysis.Subqueries = append(analysis.Subqueries, subqueryInfo)
		}
	case *tree.JoinTableExpr:
		// 连接表达式
		p.analyzeJoinExpr(analysis, e)
	}
}

// analyzeJoinExpr 分析连接表达式
func (p *PostgreSQLEnhancedParser) analyzeJoinExpr(analysis *EnhancedSQLAnalysis, join *tree.JoinTableExpr) {
	joinInfo := EnhancedJoinInfo{
		Type:      string(join.JoinType),
		Condition: "",
		Columns:   make([]string, 0),
	}

	// 提取左表名
	if leftTables := p.extractTableNameFromTableExpr(join.Left); len(leftTables) > 0 {
		joinInfo.LeftTable = leftTables[0]
	}

	// 提取右表名
	if rightTables := p.extractTableNameFromTableExpr(join.Right); len(rightTables) > 0 {
		joinInfo.RightTable = rightTables[0]
	}

	// 提取连接条件
	if join.Cond != nil {
		switch cond := join.Cond.(type) {
		case *tree.OnJoinCond:
			joinInfo.Condition = tree.AsString(cond.Expr)
		case *tree.UsingJoinCond:
			for _, col := range cond.Cols {
				joinInfo.Columns = append(joinInfo.Columns, col.String())
			}
		}
	}

	analysis.Joins = append(analysis.Joins, joinInfo)

	// 递归分析连接的子表达式
	p.analyzeTableExpr(analysis, join.Left)
	p.analyzeTableExpr(analysis, join.Right)
}

// analyzeSelectList 分析 SELECT 列表
func (p *PostgreSQLEnhancedParser) analyzeSelectList(analysis *EnhancedSQLAnalysis, exprs tree.SelectExprs) {
	for _, expr := range exprs {
		p.analyzeSelectExpr(analysis, expr.Expr)
	}
}

// analyzeSelectExpr 分析 SELECT 表达式
func (p *PostgreSQLEnhancedParser) analyzeSelectExpr(analysis *EnhancedSQLAnalysis, expr tree.Expr) {
	switch e := expr.(type) {
	case *tree.Subquery:
		// 标量子查询
		subqueryInfo := SubqueryInfo{
			Type:    "SCALAR",
			Tables:  make([]string, 0),
			Columns: make([]string, 0),
			Nested:  true,
		}
		// 简化处理
		analysis.Subqueries = append(analysis.Subqueries, subqueryInfo)
	case *tree.FuncExpr:
		// 检查是否是窗口函数
		if e.WindowDef != nil {
			p.analyzeWindowFunction(analysis, e)
		}
	}
}

// analyzeWindowFunction 分析窗口函数
func (p *PostgreSQLEnhancedParser) analyzeWindowFunction(analysis *EnhancedSQLAnalysis, funcExpr *tree.FuncExpr) {
	windowInfo := WindowFunctionInfo{
		Function:    funcExpr.Func.String(),
		PartitionBy: make([]string, 0),
		OrderBy:     make([]string, 0),
	}

	// 分析窗口定义
	if funcExpr.WindowDef != nil {
		// 分析 PARTITION BY
		for _, partitionExpr := range funcExpr.WindowDef.Partitions {
			windowInfo.PartitionBy = append(windowInfo.PartitionBy, tree.AsString(partitionExpr))
		}

		// 分析 ORDER BY
		for _, orderExpr := range funcExpr.WindowDef.OrderBy {
			windowInfo.OrderBy = append(windowInfo.OrderBy, tree.AsString(orderExpr.Expr))
		}

		// 分析窗口框架
		if funcExpr.WindowDef.Frame != nil {
			windowInfo.Frame = tree.AsString(funcExpr.WindowDef.Frame)
		}
	}

	analysis.WindowFunctions = append(analysis.WindowFunctions, windowInfo)
}

// analyzeWhereClause 分析 WHERE 子句
func (p *PostgreSQLEnhancedParser) analyzeWhereClause(analysis *EnhancedSQLAnalysis, expr tree.Expr) {
	p.analyzeExprForSubqueries(analysis, expr)
}

// analyzeExprForSubqueries 分析表达式中的子查询
func (p *PostgreSQLEnhancedParser) analyzeExprForSubqueries(analysis *EnhancedSQLAnalysis, expr tree.Expr) {
	switch e := expr.(type) {
	case *tree.Subquery:
		subqueryInfo := SubqueryInfo{
			Type:    "WHERE",
			Tables:  make([]string, 0),
			Columns: make([]string, 0),
			Nested:  true,
		}
		// 简化处理
		analysis.Subqueries = append(analysis.Subqueries, subqueryInfo)
	case *tree.ComparisonExpr:
		// 检查比较表达式的两边是否有子查询
		p.analyzeExprForSubqueries(analysis, e.Left)
		p.analyzeExprForSubqueries(analysis, e.Right)
	case *tree.AndExpr:
		p.analyzeExprForSubqueries(analysis, e.Left)
		p.analyzeExprForSubqueries(analysis, e.Right)
	case *tree.OrExpr:
		p.analyzeExprForSubqueries(analysis, e.Left)
		p.analyzeExprForSubqueries(analysis, e.Right)
	}
}

// analyzeGroupBy 分析 GROUP BY 子句
func (p *PostgreSQLEnhancedParser) analyzeGroupBy(analysis *EnhancedSQLAnalysis, groupBy tree.GroupBy) {
	// 可以在这里添加 GROUP BY 相关的分析逻辑
}

// analyzeHaving 分析 HAVING 子句
func (p *PostgreSQLEnhancedParser) analyzeHaving(analysis *EnhancedSQLAnalysis, expr tree.Expr) {
	p.analyzeExprForSubqueries(analysis, expr)
}

// analyzeOrderBy 分析 ORDER BY 子句
func (p *PostgreSQLEnhancedParser) analyzeOrderBy(analysis *EnhancedSQLAnalysis, orderBy tree.OrderBy) {
	// 可以在这里添加 ORDER BY 相关的分析逻辑
}

// analyzeInsertStatement 分析 INSERT 语句
func (p *PostgreSQLEnhancedParser) analyzeInsertStatement(analysis *EnhancedSQLAnalysis, insert *tree.Insert) {
	// 分析 INSERT 语句中的子查询
	if insert.Rows != nil {
		// 简化处理，检查是否有子查询
		subqueryInfo := SubqueryInfo{
			Type:    "INSERT_SELECT",
			Tables:  make([]string, 0),
			Columns: make([]string, 0),
			Nested:  true,
		}
		analysis.Subqueries = append(analysis.Subqueries, subqueryInfo)
	}
}

// analyzeUpdateStatement 分析 UPDATE 语句
func (p *PostgreSQLEnhancedParser) analyzeUpdateStatement(analysis *EnhancedSQLAnalysis, update *tree.Update) {
	// 分析 WHERE 子句中的子查询
	if update.Where != nil {
		p.analyzeWhereClause(analysis, update.Where.Expr)
	}

	// 分析 SET 子句中的子查询
	for _, updateExpr := range update.Exprs {
		p.analyzeExprForSubqueries(analysis, updateExpr.Expr)
	}
}

// analyzeDeleteStatement 分析 DELETE 语句
func (p *PostgreSQLEnhancedParser) analyzeDeleteStatement(analysis *EnhancedSQLAnalysis, delete *tree.Delete) {
	// 分析 WHERE 子句中的子查询
	if delete.Where != nil {
		p.analyzeWhereClause(analysis, delete.Where.Expr)
	}
}

// calculateComplexity 计算复杂度指标
func (p *PostgreSQLEnhancedParser) calculateComplexity(analysis *EnhancedSQLAnalysis) {
	analysis.Complexity = ComplexityMetrics{
		TableCount:      len(analysis.Tables),
		JoinCount:       len(analysis.Joins),
		SubqueryCount:   len(analysis.Subqueries),
		CTECount:        len(analysis.CTEs),
		WindowFuncCount: len(analysis.WindowFunctions),
		NestingLevel:    p.calculateNestingLevel(analysis),
	}

	// 计算复杂度分数
	score := 0
	score += analysis.Complexity.TableCount * 2
	score += analysis.Complexity.JoinCount * 5
	score += analysis.Complexity.SubqueryCount * 10
	score += analysis.Complexity.CTECount * 8
	score += analysis.Complexity.WindowFuncCount * 6
	score += analysis.Complexity.NestingLevel * 15

	analysis.Complexity.Score = score
}

// calculateNestingLevel 计算嵌套级别
func (p *PostgreSQLEnhancedParser) calculateNestingLevel(analysis *EnhancedSQLAnalysis) int {
	maxLevel := 0
	for _, subquery := range analysis.Subqueries {
		if subquery.Nested {
			maxLevel = 1 // 简化实现，实际应该递归计算
		}
	}
	return maxLevel
}

// generateOptimizationSuggestions 生成优化建议
func (p *PostgreSQLEnhancedParser) generateOptimizationSuggestions(analysis *EnhancedSQLAnalysis, sql string) {
	suggestions := make([]OptimizationSuggestion, 0)

	// 检查是否使用了 SELECT *
	if strings.Contains(strings.ToUpper(sql), "SELECT *") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "performance",
			Severity:   "warning",
			Message:    "使用 SELECT * 可能影响性能",
			Suggestion: "明确指定需要的列名，避免使用 SELECT *",
		})
	}

	// 检查是否缺少 WHERE 子句
	if analysis.Type == SQLTypeSelect && !strings.Contains(strings.ToUpper(sql), "WHERE") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "performance",
			Severity:   "info",
			Message:    "查询缺少 WHERE 子句",
			Suggestion: "考虑添加 WHERE 子句来限制结果集大小",
		})
	}

	// 检查复杂的子查询
	if len(analysis.Subqueries) > 3 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "complexity",
			Severity:   "warning",
			Message:    "查询包含过多子查询",
			Suggestion: "考虑使用 JOIN 或 CTE 来简化查询结构",
		})
	}

	// 检查是否使用了过多的 JOIN
	if len(analysis.Joins) > 5 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "performance",
			Severity:   "warning",
			Message:    "查询包含过多 JOIN 操作",
			Suggestion: "考虑分解查询或优化表结构",
		})
	}

	// 检查是否使用了 DISTINCT
	if strings.Contains(strings.ToUpper(sql), "DISTINCT") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "performance",
			Severity:   "info",
			Message:    "使用了 DISTINCT 操作",
			Suggestion: "确保 DISTINCT 是必要的，考虑使用 GROUP BY 替代",
		})
	}

	// 检查是否使用了 ORDER BY 但没有 LIMIT
	if strings.Contains(strings.ToUpper(sql), "ORDER BY") && !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:       "performance",
			Severity:   "info",
			Message:    "使用了 ORDER BY 但没有 LIMIT",
			Suggestion: "如果不需要所有结果，考虑添加 LIMIT 子句",
		})
	}

	analysis.Optimizations = suggestions
}

// ExtractTablesEnhanced 增强的表名提取（支持复杂查询）
func (p *PostgreSQLEnhancedParser) ExtractTablesEnhanced(sql string) (map[string][]string, error) {
	analysis, err := p.AnalyzeSQL(sql)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)

	// 主查询表
	result["main"] = analysis.Tables

	// CTE 表
	cteNames := make([]string, 0)
	for _, cte := range analysis.CTEs {
		cteNames = append(cteNames, cte.Name)
	}
	if len(cteNames) > 0 {
		result["cte"] = cteNames
	}

	// 子查询表
	subqueryTables := make([]string, 0)
	for _, subquery := range analysis.Subqueries {
		subqueryTables = append(subqueryTables, subquery.Tables...)
	}
	if len(subqueryTables) > 0 {
		result["subquery"] = subqueryTables
	}

	return result, nil
}

// RewriteForSharding 为分片重写 SQL（增强版本）
func (p *PostgreSQLEnhancedParser) RewriteForSharding(sql string, shardingRules map[string]string) (string, error) {
	analysis, err := p.AnalyzeSQL(sql)
	if err != nil {
		return "", err
	}

	rewrittenSQL := sql

	// 重写主查询中的表名
	for _, table := range analysis.Tables {
		if shardedTable, exists := shardingRules[table]; exists {
			rewrittenSQL = p.replaceTableName(rewrittenSQL, table, shardedTable)
		}
	}

	// 重写子查询中的表名
	for _, subquery := range analysis.Subqueries {
		for _, table := range subquery.Tables {
			if shardedTable, exists := shardingRules[table]; exists {
				rewrittenSQL = p.replaceTableName(rewrittenSQL, table, shardedTable)
			}
		}
	}

	// 重写 CTE 中的表名
	for _, cte := range analysis.CTEs {
		for _, table := range cte.Tables {
			if shardedTable, exists := shardingRules[table]; exists {
				rewrittenSQL = p.replaceTableName(rewrittenSQL, table, shardedTable)
			}
		}
	}

	return rewrittenSQL, nil
}

// replaceTableName 智能替换表名
func (p *PostgreSQLEnhancedParser) replaceTableName(sql, oldTable, newTable string) string {
	// 使用正则表达式确保只替换完整的表名
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(oldTable))
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(sql, newTable)
}

// ValidateComplexSQL 验证复杂 SQL 语法
func (p *PostgreSQLEnhancedParser) ValidateComplexSQL(sql string) ([]ValidationError, error) {
	errors := make([]ValidationError, 0)

	// 基础语法验证
	if err := p.PostgreSQLParser.ValidatePostgreSQLSQL(sql); err != nil {
		errors = append(errors, ValidationError{
			Type:    "syntax",
			Message: err.Error(),
		})
	}

	// 分析 SQL 结构
	analysis, err := p.AnalyzeSQL(sql)
	if err != nil {
		return errors, err
	}

	// 检查复杂度
	if analysis.Complexity.Score > 100 {
		errors = append(errors, ValidationError{
			Type:    "complexity",
			Message: fmt.Sprintf("查询复杂度过高 (分数: %d)", analysis.Complexity.Score),
		})
	}

	// 检查嵌套级别
	if analysis.Complexity.NestingLevel > 3 {
		errors = append(errors, ValidationError{
			Type:    "nesting",
			Message: fmt.Sprintf("嵌套级别过深 (级别: %d)", analysis.Complexity.NestingLevel),
		})
	}

	return errors, nil
}

// ValidationError 验证错误
type ValidationError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// GetOptimizationSuggestions 获取优化建议
func (p *PostgreSQLEnhancedParser) GetOptimizationSuggestions(sql string) ([]OptimizationSuggestion, error) {
	analysis, err := p.AnalyzeSQL(sql)
	if err != nil {
		return nil, err
	}
	return analysis.Optimizations, nil
}

// AnalyzeTableDependencies 分析表依赖关系
func (p *PostgreSQLEnhancedParser) AnalyzeTableDependencies(sql string) (map[string][]string, error) {
	analysis, err := p.AnalyzeSQL(sql)
	if err != nil {
		return nil, err
	}

	dependencies := make(map[string][]string)

	// 分析 JOIN 依赖
	for _, join := range analysis.Joins {
		if join.LeftTable != "" && join.RightTable != "" {
			if dependencies[join.LeftTable] == nil {
				dependencies[join.LeftTable] = make([]string, 0)
			}
			dependencies[join.LeftTable] = append(dependencies[join.LeftTable], join.RightTable)
		}
	}

	// 分析子查询依赖
	for _, subquery := range analysis.Subqueries {
		for _, table := range subquery.Tables {
			for _, mainTable := range analysis.Tables {
				if table != mainTable {
					if dependencies[mainTable] == nil {
						dependencies[mainTable] = make([]string, 0)
					}
					dependencies[mainTable] = append(dependencies[mainTable], table)
				}
			}
		}
	}

	return dependencies, nil
}