package routing

import (
	"fmt"
	"go-sharding/pkg/config"
	"regexp"
	"strconv"
	"strings"
)

// RouteResult 路由结果
type RouteResult struct {
	DataSource string
	Table      string
}

// Router 路由器接口
type Router interface {
	Route(logicTable string, shardingValues map[string]interface{}) ([]*RouteResult, error)
}

// ShardingRouter 分片路由器
type ShardingRouter struct {
	dataSources  map[string]*config.DataSourceConfig
	shardingRule *config.ShardingRuleConfig
}

// NewShardingRouter 创建分片路由器
func NewShardingRouter(dataSources map[string]*config.DataSourceConfig, shardingRule *config.ShardingRuleConfig) *ShardingRouter {
	return &ShardingRouter{
		dataSources:  dataSources,
		shardingRule: shardingRule,
	}
}

// Route 执行路由
func (r *ShardingRouter) Route(logicTable string, shardingValues map[string]interface{}) ([]*RouteResult, error) {
	tableRule, exists := r.shardingRule.Tables[logicTable]
	if !exists {
		return nil, fmt.Errorf("table rule not found for table: %s", logicTable)
	}

	// 解析实际数据节点
	dataNodes, err := r.parseActualDataNodes(tableRule.ActualDataNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse actual data nodes: %w", err)
	}

	var results []*RouteResult

	// 如果没有分片值，返回所有数据节点
	if len(shardingValues) == 0 {
		for _, node := range dataNodes {
			results = append(results, &RouteResult{
				DataSource: node.DataSource,
				Table:      node.Table,
			})
		}
		return results, nil
	}

	// 计算数据库分片
	var targetDataSources []string
	if tableRule.DatabaseStrategy != nil {
		ds, err := r.calculateSharding(tableRule.DatabaseStrategy, shardingValues)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate database sharding: %w", err)
		}
		targetDataSources = ds
	} else {
		// 如果没有数据库分片策略，使用所有数据源
		for _, node := range dataNodes {
			if !contains(targetDataSources, node.DataSource) {
				targetDataSources = append(targetDataSources, node.DataSource)
			}
		}
	}

	// 计算表分片
	var targetTables []string
	if tableRule.TableStrategy != nil {
		tables, err := r.calculateSharding(tableRule.TableStrategy, shardingValues)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate table sharding: %w", err)
		}
		targetTables = tables
	} else {
		// 如果没有表分片策略，使用目标数据源下的所有表
		for _, node := range dataNodes {
			if contains(targetDataSources, node.DataSource) && !contains(targetTables, node.Table) {
				targetTables = append(targetTables, node.Table)
			}
		}
	}

	// 组合结果
	for _, ds := range targetDataSources {
		for _, table := range targetTables {
			// 验证数据节点是否存在
			if r.isValidDataNode(dataNodes, ds, table) {
				results = append(results, &RouteResult{
					DataSource: ds,
					Table:      table,
				})
			}
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no valid route found for table %s with sharding values %v", logicTable, shardingValues)
	}

	return results, nil
}

// DataNode 数据节点
type DataNode struct {
	DataSource string
	Table      string
}

// parseActualDataNodes 解析实际数据节点表达式
func (r *ShardingRouter) parseActualDataNodes(expression string) ([]*DataNode, error) {
	var nodes []*DataNode

	// 支持格式: ds_${0..1}.t_order_${0..1}
	// 需要找到不在 ${} 内的点作为分隔符
	var dotIndices []int
	braceLevel := 0
	for i, char := range expression {
		if char == '{' {
			braceLevel++
		} else if char == '}' {
			braceLevel--
		} else if char == '.' && braceLevel == 0 {
			dotIndices = append(dotIndices, i)
		}
	}

	if len(dotIndices) != 1 {
		return nil, fmt.Errorf("invalid actual data nodes expression: %s", expression)
	}

	dotIndex := dotIndices[0]
	dsPattern := expression[:dotIndex]
	tablePattern := expression[dotIndex+1:]

	// 解析数据源范围
	dataSources, err := r.parseRangeExpression(dsPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data source pattern: %w", err)
	}

	// 解析表范围
	tables, err := r.parseRangeExpression(tablePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse table pattern: %w", err)
	}

	// 组合所有可能的数据节点
	for _, ds := range dataSources {
		for _, table := range tables {
			nodes = append(nodes, &DataNode{
				DataSource: ds,
				Table:      table,
			})
		}
	}

	return nodes, nil
}

// parseRangeExpression 解析范围表达式
func (r *ShardingRouter) parseRangeExpression(pattern string) ([]string, error) {
	// 匹配 ${0..1} 格式
	rangeRegex := regexp.MustCompile(`\$\{(\d+)\.\.(\d+)\}`)
	matches := rangeRegex.FindStringSubmatch(pattern)

	if len(matches) == 3 {
		start, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid range start: %s", matches[1])
		}
		end, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid range end: %s", matches[2])
		}

		var results []string
		for i := start; i <= end; i++ {
			result := rangeRegex.ReplaceAllString(pattern, strconv.Itoa(i))
			results = append(results, result)
		}
		return results, nil
	}

	// 匹配 ${[0, 1]} 格式
	listRegex := regexp.MustCompile(`\$\{\[([^\]]+)\]\}`)
	listMatches := listRegex.FindStringSubmatch(pattern)

	if len(listMatches) == 2 {
		values := strings.Split(listMatches[1], ",")
		var results []string
		for _, value := range values {
			value = strings.TrimSpace(value)
			result := listRegex.ReplaceAllString(pattern, value)
			results = append(results, result)
		}
		return results, nil
	}

	// 如果没有匹配到范围表达式，直接返回原始模式
	return []string{pattern}, nil
}

// calculateSharding 计算分片结果
func (r *ShardingRouter) calculateSharding(strategy *config.ShardingStrategyConfig, shardingValues map[string]interface{}) ([]string, error) {
	if strategy.Type == "" || strategy.Type == "inline" {
		return r.calculateInlineSharding(strategy, shardingValues)
	}

	return nil, fmt.Errorf("unsupported sharding strategy type: %s", strategy.Type)
}

// calculateInlineSharding 计算内联分片
func (r *ShardingRouter) calculateInlineSharding(strategy *config.ShardingStrategyConfig, shardingValues map[string]interface{}) ([]string, error) {
	value, exists := shardingValues[strategy.ShardingColumn]
	if !exists {
		return nil, fmt.Errorf("sharding column %s not found in sharding values", strategy.ShardingColumn)
	}

	// 计算表达式结果
	result, err := r.evaluateInlineExpression(strategy.Algorithm, strategy.ShardingColumn, value)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate sharding algorithm: %w", err)
	}

	return []string{result}, nil
}

// evaluateInlineExpression 计算内联表达式
func (r *ShardingRouter) evaluateInlineExpression(algorithm, shardingColumn string, shardingValue interface{}) (string, error) {
	// 将分片值转换为整数
	var intValue int
	switch v := shardingValue.(type) {
	case int:
		intValue = v
	case int32:
		intValue = int(v)
	case int64:
		intValue = int(v)
	case float32:
		intValue = int(v)
	case float64:
		intValue = int(v)
	case string:
		var err error
		intValue, err = strconv.Atoi(v)
		if err != nil {
			return "", fmt.Errorf("cannot convert sharding value to integer: %v", v)
		}
	default:
		return "", fmt.Errorf("unsupported sharding value type: %T", v)
	}

	// 支持模运算表达式: ds_${user_id % 2} 或 t_order_${order_id % 2}
	modRegex := regexp.MustCompile(`^(.+)\$\{` + regexp.QuoteMeta(shardingColumn) + `\s*%\s*(\d+)\}(.*)$`)
	matches := modRegex.FindStringSubmatch(algorithm)
	
	if len(matches) == 4 {
		prefix := matches[1]
		mod, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", fmt.Errorf("invalid modulo in expression: %s", matches[2])
		}
		suffix := matches[3]
		
		result := intValue % mod
		return prefix + strconv.Itoa(result) + suffix, nil
	}

	// 支持简单的变量替换: ds_${user_id}
	simpleRegex := regexp.MustCompile(`^(.+)\$\{` + regexp.QuoteMeta(shardingColumn) + `\}(.*)$`)
	simpleMatches := simpleRegex.FindStringSubmatch(algorithm)
	
	if len(simpleMatches) == 3 {
		prefix := simpleMatches[1]
		suffix := simpleMatches[2]
		return prefix + strconv.Itoa(intValue) + suffix, nil
	}

	// 如果没有匹配到任何表达式，直接返回算法字符串
	return algorithm, nil
}

// isValidDataNode 检查数据节点是否有效
func (r *ShardingRouter) isValidDataNode(nodes []*DataNode, dataSource, table string) bool {
	for _, node := range nodes {
		if node.DataSource == dataSource && node.Table == table {
			return true
		}
	}
	return false
}

// contains 检查字符串数组是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}