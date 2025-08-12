package algorithm

import (
	"fmt"
	"strconv"
	"strings"
)

// HintInlineShardingAlgorithm Hint内联分片算法
type HintInlineShardingAlgorithm struct {
	algorithmExpression string
	properties          map[string]interface{}
}

// NewHintInlineShardingAlgorithm 创建Hint内联分片算法
func NewHintInlineShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	expression, ok := properties["algorithm-expression"].(string)
	if !ok {
		return nil, fmt.Errorf("algorithm-expression is required for hint inline sharding algorithm")
	}
	
	return &HintInlineShardingAlgorithm{
		algorithmExpression: expression,
		properties:          properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *HintInlineShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	return a.DoHintSharding(availableTargetNames, shardingValue)
}

// DoHintSharding Hint分片
func (a *HintInlineShardingAlgorithm) DoHintSharding(availableTargetNames []string, hintValue *ShardingValue) ([]string, error) {
	if hintValue.Values != nil && len(hintValue.Values) > 0 {
		// 处理多个Hint值
		var results []string
		for _, value := range hintValue.Values {
			result, err := a.evaluateHintExpression(value)
			if err != nil {
				return nil, err
			}
			if contains(availableTargetNames, result) && !contains(results, result) {
				results = append(results, result)
			}
		}
		return results, nil
	}
	
	// 处理单个Hint值
	result, err := a.evaluateHintExpression(hintValue.Value)
	if err != nil {
		return nil, err
	}
	
	if contains(availableTargetNames, result) {
		return []string{result}, nil
	}
	
	return nil, fmt.Errorf("hint target %s not found in available targets", result)
}

// GetType 获取算法类型
func (a *HintInlineShardingAlgorithm) GetType() string {
	return "HINT_INLINE"
}

// GetProperties 获取算法属性
func (a *HintInlineShardingAlgorithm) GetProperties() map[string]interface{} {
	return a.properties
}

// evaluateHintExpression 计算Hint表达式
func (a *HintInlineShardingAlgorithm) evaluateHintExpression(value interface{}) (string, error) {
	// Hint算法支持直接指定目标或通过表达式计算
	
	// 如果是字符串，可能是直接的目标名称
	if strValue, ok := value.(string); ok {
		// 检查是否是直接的目标名称
		if !strings.Contains(strValue, "${") {
			return strValue, nil
		}
	}
	
	// 否则按照内联表达式处理
	expression := a.algorithmExpression
	
	// 替换 ${value} 占位符
	valueStr := ConvertToString(value)
	expression = strings.ReplaceAll(expression, "${value}", valueStr)
	
	// 处理特殊的Hint表达式
	if strings.Contains(expression, "direct") {
		// 直接路由模式：direct_${value}
		expression = strings.ReplaceAll(expression, "direct_${value}", valueStr)
		return expression, nil
	}
	
	// 处理索引模式：index_${value % count}
	if strings.Contains(expression, "index") {
		intValue, err := ConvertToInt(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert hint value to int: %w", err)
		}
		
		// 提取分片数量
		countStr := a.extractShardingCount()
		if countStr != "" {
			count, err := strconv.ParseInt(countStr, 10, 64)
			if err != nil {
				return "", fmt.Errorf("invalid sharding count: %s", countStr)
			}
			index := intValue % count
			expression = strings.ReplaceAll(expression, "index_${value % count}", fmt.Sprintf("index_%d", index))
		}
	}
	
	// 处理范围模式：range_${value / 1000}
	if strings.Contains(expression, "range") {
		intValue, err := ConvertToInt(value)
		if err != nil {
			return "", fmt.Errorf("failed to convert hint value to int: %w", err)
		}
		
		// 提取范围大小
		rangeSize := a.extractRangeSize()
		if rangeSize > 0 {
			rangeIndex := intValue / rangeSize
			expression = strings.ReplaceAll(expression, "range_${value / 1000}", fmt.Sprintf("range_%d", rangeIndex))
		}
	}
	
	return expression, nil
}

// extractShardingCount 提取分片数量
func (a *HintInlineShardingAlgorithm) extractShardingCount() string {
	if count, ok := a.properties["sharding-count"].(string); ok {
		return count
	}
	if count, ok := a.properties["sharding-count"].(int); ok {
		return strconv.Itoa(count)
	}
	return "2" // 默认值
}

// extractRangeSize 提取范围大小
func (a *HintInlineShardingAlgorithm) extractRangeSize() int64 {
	if size, ok := a.properties["range-size"].(string); ok {
		if intSize, err := strconv.ParseInt(size, 10, 64); err == nil {
			return intSize
		}
	}
	if size, ok := a.properties["range-size"].(int); ok {
		return int64(size)
	}
	if size, ok := a.properties["range-size"].(int64); ok {
		return size
	}
	return 1000 // 默认值
}

// HintManager Hint管理器
type HintManager struct {
	hints map[string]interface{}
}

// NewHintManager 创建Hint管理器
func NewHintManager() *HintManager {
	return &HintManager{
		hints: make(map[string]interface{}),
	}
}

// SetDatabaseShardingValue 设置数据库分片Hint值
func (hm *HintManager) SetDatabaseShardingValue(value interface{}) {
	hm.hints["database_sharding_value"] = value
}

// SetTableShardingValue 设置表分片Hint值
func (hm *HintManager) SetTableShardingValue(value interface{}) {
	hm.hints["table_sharding_value"] = value
}

// SetMasterRouteOnly 设置仅主库路由
func (hm *HintManager) SetMasterRouteOnly() {
	hm.hints["master_route_only"] = true
}

// GetDatabaseShardingValue 获取数据库分片Hint值
func (hm *HintManager) GetDatabaseShardingValue() interface{} {
	return hm.hints["database_sharding_value"]
}

// GetTableShardingValue 获取表分片Hint值
func (hm *HintManager) GetTableShardingValue() interface{} {
	return hm.hints["table_sharding_value"]
}

// IsMasterRouteOnly 是否仅主库路由
func (hm *HintManager) IsMasterRouteOnly() bool {
	if value, ok := hm.hints["master_route_only"].(bool); ok {
		return value
	}
	return false
}

// Clear 清除所有Hint
func (hm *HintManager) Clear() {
	hm.hints = make(map[string]interface{})
}

// GetAllHints 获取所有Hint
func (hm *HintManager) GetAllHints() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range hm.hints {
		result[k] = v
	}
	return result
}