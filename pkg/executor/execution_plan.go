package executor

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	ID          string
	SQL         string
	Steps       []ExecutionStep
	EstimatedCost float64
	CreatedAt   time.Time
	Metadata    map[string]interface{}
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	ID          string
	Type        StepType
	Operation   string
	Target      string
	Condition   string
	Cost        float64
	Parallelizable bool
	Dependencies []string
	Metadata    map[string]interface{}
}

// StepType 步骤类型
type StepType string

const (
	StepTypeTableScan    StepType = "TABLE_SCAN"
	StepTypeIndexScan    StepType = "INDEX_SCAN"
	StepTypeFilter       StepType = "FILTER"
	StepTypeJoin         StepType = "JOIN"
	StepTypeAggregate    StepType = "AGGREGATE"
	StepTypeSort         StepType = "SORT"
	StepTypeLimit        StepType = "LIMIT"
	StepTypeUnion        StepType = "UNION"
	StepTypeSubquery     StepType = "SUBQUERY"
	StepTypeSharding     StepType = "SHARDING"
)

// ExecutionPlanGenerator 执行计划生成器
type ExecutionPlanGenerator struct {
	shardingRules map[string]*ShardingRule
	indexInfo     map[string][]IndexInfo
}

// ShardingRule 分片规则
type ShardingRule struct {
	TableName      string
	ShardingColumn string
	ShardingType   string
	ShardCount     int
	DataSources    []string
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name    string
	Columns []string
	Type    string
	Unique  bool
}

// NewExecutionPlanGenerator 创建执行计划生成器
func NewExecutionPlanGenerator() *ExecutionPlanGenerator {
	return &ExecutionPlanGenerator{
		shardingRules: make(map[string]*ShardingRule),
		indexInfo:     make(map[string][]IndexInfo),
	}
}

// RegisterShardingRule 注册分片规则
func (g *ExecutionPlanGenerator) RegisterShardingRule(rule *ShardingRule) {
	g.shardingRules[rule.TableName] = rule
}

// RegisterIndexInfo 注册索引信息
func (g *ExecutionPlanGenerator) RegisterIndexInfo(tableName string, indexes []IndexInfo) {
	g.indexInfo[tableName] = indexes
}

// GeneratePlan 生成执行计划
func (g *ExecutionPlanGenerator) GeneratePlan(sql string) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		ID:        generatePlanID(),
		SQL:       sql,
		Steps:     make([]ExecutionStep, 0),
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// 解析SQL类型
	sqlType := g.getSQLType(sql)
	plan.Metadata["sql_type"] = sqlType
	
	switch sqlType {
	case "SELECT":
		return g.generateSelectPlan(sql, plan)
	case "INSERT":
		return g.generateInsertPlan(sql, plan)
	case "UPDATE":
		return g.generateUpdatePlan(sql, plan)
	case "DELETE":
		return g.generateDeletePlan(sql, plan)
	default:
		return nil, fmt.Errorf("unsupported SQL type: %s", sqlType)
	}
}

// generateSelectPlan 生成SELECT执行计划
func (g *ExecutionPlanGenerator) generateSelectPlan(sql string, plan *ExecutionPlan) (*ExecutionPlan, error) {
	stepID := 1
	
	// 1. 分片步骤
	tables := g.extractTables(sql)
	for _, table := range tables {
		if rule, exists := g.shardingRules[table]; exists {
			step := ExecutionStep{
				ID:          fmt.Sprintf("step_%d", stepID),
				Type:        StepTypeSharding,
				Operation:   "ROUTE_SHARDS",
				Target:      table,
				Cost:        1.0,
				Parallelizable: true,
				Metadata:    map[string]interface{}{
					"sharding_column": rule.ShardingColumn,
					"shard_count":     rule.ShardCount,
					"data_sources":    rule.DataSources,
				},
			}
			plan.Steps = append(plan.Steps, step)
			stepID++
		}
	}
	
	// 2. 表扫描或索引扫描
	for _, table := range tables {
		scanStep := g.generateScanStep(table, sql, stepID)
		plan.Steps = append(plan.Steps, scanStep)
		stepID++
	}
	
	// 3. 过滤步骤
	if strings.Contains(strings.ToUpper(sql), "WHERE") {
		filterStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeFilter,
			Operation:   "APPLY_WHERE",
			Cost:        2.0,
			Parallelizable: true,
		}
		plan.Steps = append(plan.Steps, filterStep)
		stepID++
	}
	
	// 4. JOIN步骤
	if strings.Contains(strings.ToUpper(sql), "JOIN") {
		joinStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeJoin,
			Operation:   "HASH_JOIN",
			Cost:        10.0,
			Parallelizable: false,
		}
		plan.Steps = append(plan.Steps, joinStep)
		stepID++
	}
	
	// 5. 聚合步骤
	if strings.Contains(strings.ToUpper(sql), "GROUP BY") {
		aggregateStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeAggregate,
			Operation:   "GROUP_AGGREGATE",
			Cost:        5.0,
			Parallelizable: true,
		}
		plan.Steps = append(plan.Steps, aggregateStep)
		stepID++
	}
	
	// 6. 排序步骤
	if strings.Contains(strings.ToUpper(sql), "ORDER BY") {
		sortStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeSort,
			Operation:   "EXTERNAL_SORT",
			Cost:        8.0,
			Parallelizable: false,
		}
		plan.Steps = append(plan.Steps, sortStep)
		stepID++
	}
	
	// 7. 限制步骤
	if strings.Contains(strings.ToUpper(sql), "LIMIT") {
		limitStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeLimit,
			Operation:   "APPLY_LIMIT",
			Cost:        0.5,
			Parallelizable: false,
		}
		plan.Steps = append(plan.Steps, limitStep)
		stepID++
	}
	
	// 计算总成本
	totalCost := 0.0
	for _, step := range plan.Steps {
		totalCost += step.Cost
	}
	plan.EstimatedCost = totalCost
	
	return plan, nil
}

// generateInsertPlan 生成INSERT执行计划
func (g *ExecutionPlanGenerator) generateInsertPlan(sql string, plan *ExecutionPlan) (*ExecutionPlan, error) {
	stepID := 1
	
	// 提取目标表
	table := g.extractInsertTable(sql)
	
	// 1. 分片路由步骤
	if rule, exists := g.shardingRules[table]; exists {
		step := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeSharding,
			Operation:   "ROUTE_INSERT",
			Target:      table,
			Cost:        1.0,
			Parallelizable: true,
			Metadata:    map[string]interface{}{
				"sharding_column": rule.ShardingColumn,
				"shard_count":     rule.ShardCount,
			},
		}
		plan.Steps = append(plan.Steps, step)
		stepID++
	}
	
	// 2. 插入步骤
	insertStep := ExecutionStep{
		ID:          fmt.Sprintf("step_%d", stepID),
		Type:        StepTypeTableScan, // 复用类型
		Operation:   "INSERT_ROWS",
		Target:      table,
		Cost:        3.0,
		Parallelizable: true,
	}
	plan.Steps = append(plan.Steps, insertStep)
	
	plan.EstimatedCost = 4.0
	return plan, nil
}

// generateUpdatePlan 生成UPDATE执行计划
func (g *ExecutionPlanGenerator) generateUpdatePlan(sql string, plan *ExecutionPlan) (*ExecutionPlan, error) {
	stepID := 1
	
	// 提取目标表
	table := g.extractUpdateTable(sql)
	
	// 1. 分片路由步骤
	if rule, exists := g.shardingRules[table]; exists {
		step := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeSharding,
			Operation:   "ROUTE_UPDATE",
			Target:      table,
			Cost:        1.0,
			Parallelizable: true,
			Metadata:    map[string]interface{}{
				"sharding_column": rule.ShardingColumn,
				"shard_count":     rule.ShardCount,
			},
		}
		plan.Steps = append(plan.Steps, step)
		stepID++
	}
	
	// 2. 查找步骤
	scanStep := g.generateScanStep(table, sql, stepID)
	plan.Steps = append(plan.Steps, scanStep)
	stepID++
	
	// 3. 过滤步骤
	if strings.Contains(strings.ToUpper(sql), "WHERE") {
		filterStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeFilter,
			Operation:   "APPLY_WHERE",
			Cost:        2.0,
			Parallelizable: true,
		}
		plan.Steps = append(plan.Steps, filterStep)
		stepID++
	}
	
	// 4. 更新步骤
	updateStep := ExecutionStep{
		ID:          fmt.Sprintf("step_%d", stepID),
		Type:        StepTypeTableScan, // 复用类型
		Operation:   "UPDATE_ROWS",
		Target:      table,
		Cost:        5.0,
		Parallelizable: true,
	}
	plan.Steps = append(plan.Steps, updateStep)
	
	// 计算总成本
	totalCost := 0.0
	for _, step := range plan.Steps {
		totalCost += step.Cost
	}
	plan.EstimatedCost = totalCost
	
	return plan, nil
}

// generateDeletePlan 生成DELETE执行计划
func (g *ExecutionPlanGenerator) generateDeletePlan(sql string, plan *ExecutionPlan) (*ExecutionPlan, error) {
	stepID := 1
	
	// 提取目标表
	table := g.extractDeleteTable(sql)
	
	// 1. 分片路由步骤
	if rule, exists := g.shardingRules[table]; exists {
		step := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeSharding,
			Operation:   "ROUTE_DELETE",
			Target:      table,
			Cost:        1.0,
			Parallelizable: true,
			Metadata:    map[string]interface{}{
				"sharding_column": rule.ShardingColumn,
				"shard_count":     rule.ShardCount,
			},
		}
		plan.Steps = append(plan.Steps, step)
		stepID++
	}
	
	// 2. 查找步骤
	scanStep := g.generateScanStep(table, sql, stepID)
	plan.Steps = append(plan.Steps, scanStep)
	stepID++
	
	// 3. 过滤步骤
	if strings.Contains(strings.ToUpper(sql), "WHERE") {
		filterStep := ExecutionStep{
			ID:          fmt.Sprintf("step_%d", stepID),
			Type:        StepTypeFilter,
			Operation:   "APPLY_WHERE",
			Cost:        2.0,
			Parallelizable: true,
		}
		plan.Steps = append(plan.Steps, filterStep)
		stepID++
	}
	
	// 4. 删除步骤
	deleteStep := ExecutionStep{
		ID:          fmt.Sprintf("step_%d", stepID),
		Type:        StepTypeTableScan, // 复用类型
		Operation:   "DELETE_ROWS",
		Target:      table,
		Cost:        4.0,
		Parallelizable: true,
	}
	plan.Steps = append(plan.Steps, deleteStep)
	
	// 计算总成本
	totalCost := 0.0
	for _, step := range plan.Steps {
		totalCost += step.Cost
	}
	plan.EstimatedCost = totalCost
	
	return plan, nil
}

// generateScanStep 生成扫描步骤
func (g *ExecutionPlanGenerator) generateScanStep(table, sql string, stepID int) ExecutionStep {
	// 检查是否可以使用索引
	if indexes, exists := g.indexInfo[table]; exists {
		for _, index := range indexes {
			// 简化的索引选择逻辑
			if g.canUseIndex(sql, index) {
				return ExecutionStep{
					ID:          fmt.Sprintf("step_%d", stepID),
					Type:        StepTypeIndexScan,
					Operation:   "INDEX_SCAN",
					Target:      table,
					Cost:        2.0,
					Parallelizable: true,
					Metadata:    map[string]interface{}{
						"index_name": index.Name,
						"index_columns": index.Columns,
					},
				}
			}
		}
	}
	
	// 默认使用表扫描
	return ExecutionStep{
		ID:          fmt.Sprintf("step_%d", stepID),
		Type:        StepTypeTableScan,
		Operation:   "TABLE_SCAN",
		Target:      table,
		Cost:        10.0,
		Parallelizable: true,
	}
}

// 辅助方法
func (g *ExecutionPlanGenerator) getSQLType(sql string) string {
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(upperSQL, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(upperSQL, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(upperSQL, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(upperSQL, "DELETE") {
		return "DELETE"
	}
	return "UNKNOWN"
}

func (g *ExecutionPlanGenerator) extractTables(sql string) []string {
	// 简化的表名提取逻辑
	var tables []string
	// 这里应该使用更复杂的SQL解析逻辑
	return tables
}

func (g *ExecutionPlanGenerator) extractInsertTable(sql string) string {
	// 提取INSERT语句的目标表
	return "table"
}

func (g *ExecutionPlanGenerator) extractUpdateTable(sql string) string {
	// 提取UPDATE语句的目标表
	return "table"
}

func (g *ExecutionPlanGenerator) extractDeleteTable(sql string) string {
	// 提取DELETE语句的目标表
	return "table"
}

func (g *ExecutionPlanGenerator) canUseIndex(sql string, index IndexInfo) bool {
	// 简化的索引使用判断逻辑
	return false
}

func generatePlanID() string {
	return fmt.Sprintf("plan_%d", time.Now().UnixNano())
}