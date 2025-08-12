package algorithm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// InlineShardingAlgorithm 内联分片算法
type InlineShardingAlgorithm struct {
	algorithmExpression string
	properties          map[string]interface{}
}

// NewInlineShardingAlgorithm 创建内联分片算法
func NewInlineShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	expression, ok := properties["algorithm-expression"].(string)
	if !ok {
		return nil, fmt.Errorf("algorithm-expression is required for inline sharding algorithm")
	}
	
	return &InlineShardingAlgorithm{
		algorithmExpression: expression,
		properties:          properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *InlineShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	if shardingValue.Values != nil && len(shardingValue.Values) > 0 {
		// 处理 IN 查询
		var results []string
		for _, value := range shardingValue.Values {
			result, err := a.evaluateExpression(value)
			if err != nil {
				return nil, err
			}
			if contains(availableTargetNames, result) && !contains(results, result) {
				results = append(results, result)
			}
		}
		return results, nil
	}
	
	// 处理单值查询
	result, err := a.evaluateExpression(shardingValue.Value)
	if err != nil {
		return nil, err
	}
	
	if contains(availableTargetNames, result) {
		return []string{result}, nil
	}
	
	return nil, fmt.Errorf("calculated target %s not found in available targets", result)
}

// GetType 获取算法类型
func (a *InlineShardingAlgorithm) GetType() string {
	return "INLINE"
}

// GetProperties 获取算法属性
func (a *InlineShardingAlgorithm) GetProperties() map[string]interface{} {
	return a.properties
}

// evaluateExpression 计算表达式
func (a *InlineShardingAlgorithm) evaluateExpression(value interface{}) (string, error) {
	expression := a.algorithmExpression
	
	// 替换 ${value} 占位符
	valueStr := ConvertToString(value)
	expression = strings.ReplaceAll(expression, "${value}", valueStr)
	
	// 处理数学表达式，如 ds_${value % 2}
	modRegex := regexp.MustCompile(`\$\{(\w+)\s*%\s*(\d+)\}`)
	matches := modRegex.FindAllStringSubmatch(expression, -1)
	
	for _, match := range matches {
		if len(match) == 3 {
			varName := match[1]
			modValue, err := strconv.Atoi(match[2])
			if err != nil {
				return "", fmt.Errorf("invalid mod value: %s", match[2])
			}
			
			var varValue int64
			if varName == "value" {
				varValue, err = ConvertToInt(value)
				if err != nil {
					return "", fmt.Errorf("failed to convert value to int: %w", err)
				}
			} else {
				return "", fmt.Errorf("unsupported variable: %s", varName)
			}
			
			result := varValue % int64(modValue)
			expression = strings.ReplaceAll(expression, match[0], strconv.FormatInt(result, 10))
		}
	}
	
	// 处理简单的数学运算
	expression = a.evaluateSimpleMath(expression)
	
	return expression, nil
}

// evaluateSimpleMath 计算简单数学表达式
func (a *InlineShardingAlgorithm) evaluateSimpleMath(expression string) string {
	// 处理加法
	addRegex := regexp.MustCompile(`(\d+)\s*\+\s*(\d+)`)
	for addRegex.MatchString(expression) {
		matches := addRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.Atoi(matches[1])
			right, _ := strconv.Atoi(matches[2])
			result := left + right
			expression = addRegex.ReplaceAllString(expression, strconv.Itoa(result))
		}
	}
	
	// 处理减法
	subRegex := regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)
	for subRegex.MatchString(expression) {
		matches := subRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.Atoi(matches[1])
			right, _ := strconv.Atoi(matches[2])
			result := left - right
			expression = subRegex.ReplaceAllString(expression, strconv.Itoa(result))
		}
	}
	
	return expression
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}