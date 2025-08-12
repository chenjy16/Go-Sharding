package algorithm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ComplexInlineShardingAlgorithm 复合内联分片算法
type ComplexInlineShardingAlgorithm struct {
	algorithmExpression string
	properties          map[string]interface{}
}

// NewComplexInlineShardingAlgorithm 创建复合内联分片算法
func NewComplexInlineShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	expression, ok := properties["algorithm-expression"].(string)
	if !ok {
		return nil, fmt.Errorf("algorithm-expression is required for complex inline sharding algorithm")
	}
	
	return &ComplexInlineShardingAlgorithm{
		algorithmExpression: expression,
		properties:          properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *ComplexInlineShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	// 复合算法需要多个分片值，这里提供单值的默认实现
	shardingValues := map[string]*ShardingValue{
		shardingValue.ColumnName: shardingValue,
	}
	return a.DoComplexSharding(availableTargetNames, shardingValues)
}

// DoComplexSharding 复合分片
func (a *ComplexInlineShardingAlgorithm) DoComplexSharding(availableTargetNames []string, shardingValues map[string]*ShardingValue) ([]string, error) {
	result, err := a.evaluateComplexExpression(shardingValues)
	if err != nil {
		return nil, err
	}
	
	if contains(availableTargetNames, result) {
		return []string{result}, nil
	}
	
	return nil, fmt.Errorf("calculated target %s not found in available targets", result)
}

// GetType 获取算法类型
func (a *ComplexInlineShardingAlgorithm) GetType() string {
	return "COMPLEX_INLINE"
}

// GetProperties 获取算法属性
func (a *ComplexInlineShardingAlgorithm) GetProperties() map[string]interface{} {
	return a.properties
}

// evaluateComplexExpression 计算复合表达式
func (a *ComplexInlineShardingAlgorithm) evaluateComplexExpression(shardingValues map[string]*ShardingValue) (string, error) {
	expression := a.algorithmExpression
	
	// 替换所有变量占位符
	for columnName, shardingValue := range shardingValues {
		placeholder := fmt.Sprintf("${%s}", columnName)
		valueStr := ConvertToString(shardingValue.Value)
		expression = strings.ReplaceAll(expression, placeholder, valueStr)
	}
	
	// 处理复合数学表达式，如 ds_${(user_id % 2) + (order_id % 4)}
	complexRegex := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := complexRegex.FindAllStringSubmatch(expression, -1)
	
	for _, match := range matches {
		if len(match) == 2 {
			mathExpr := match[1]
			result, err := a.evaluateComplexMath(mathExpr, shardingValues)
			if err != nil {
				return "", fmt.Errorf("failed to evaluate complex expression %s: %w", mathExpr, err)
			}
			expression = strings.ReplaceAll(expression, match[0], strconv.FormatInt(result, 10))
		}
	}
	
	return expression, nil
}

// evaluateComplexMath 计算复合数学表达式
func (a *ComplexInlineShardingAlgorithm) evaluateComplexMath(expression string, shardingValues map[string]*ShardingValue) (int64, error) {
	// 替换变量名为实际值
	for columnName, shardingValue := range shardingValues {
		intValue, err := ConvertToInt(shardingValue.Value)
		if err != nil {
			return 0, fmt.Errorf("failed to convert %s value to int: %w", columnName, err)
		}
		expression = strings.ReplaceAll(expression, columnName, strconv.FormatInt(intValue, 10))
	}
	
	// 处理取模运算
	modRegex := regexp.MustCompile(`(\d+)\s*%\s*(\d+)`)
	for modRegex.MatchString(expression) {
		matches := modRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.ParseInt(matches[1], 10, 64)
			right, _ := strconv.ParseInt(matches[2], 10, 64)
			if right == 0 {
				return 0, fmt.Errorf("division by zero in mod operation")
			}
			result := left % right
			expression = modRegex.ReplaceAllString(expression, strconv.FormatInt(result, 10))
		}
	}
	
	// 处理乘法
	mulRegex := regexp.MustCompile(`(\d+)\s*\*\s*(\d+)`)
	for mulRegex.MatchString(expression) {
		matches := mulRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.ParseInt(matches[1], 10, 64)
			right, _ := strconv.ParseInt(matches[2], 10, 64)
			result := left * right
			expression = mulRegex.ReplaceAllString(expression, strconv.FormatInt(result, 10))
		}
	}
	
	// 处理除法
	divRegex := regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)
	for divRegex.MatchString(expression) {
		matches := divRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.ParseInt(matches[1], 10, 64)
			right, _ := strconv.ParseInt(matches[2], 10, 64)
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			result := left / right
			expression = divRegex.ReplaceAllString(expression, strconv.FormatInt(result, 10))
		}
	}
	
	// 处理加法
	addRegex := regexp.MustCompile(`(\d+)\s*\+\s*(\d+)`)
	for addRegex.MatchString(expression) {
		matches := addRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.ParseInt(matches[1], 10, 64)
			right, _ := strconv.ParseInt(matches[2], 10, 64)
			result := left + right
			expression = addRegex.ReplaceAllString(expression, strconv.FormatInt(result, 10))
		}
	}
	
	// 处理减法
	subRegex := regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)
	for subRegex.MatchString(expression) {
		matches := subRegex.FindStringSubmatch(expression)
		if len(matches) == 3 {
			left, _ := strconv.ParseInt(matches[1], 10, 64)
			right, _ := strconv.ParseInt(matches[2], 10, 64)
			result := left - right
			expression = subRegex.ReplaceAllString(expression, strconv.FormatInt(result, 10))
		}
	}
	
	// 处理括号
	parenRegex := regexp.MustCompile(`\(([^()]+)\)`)
	for parenRegex.MatchString(expression) {
		matches := parenRegex.FindStringSubmatch(expression)
		if len(matches) == 2 {
			innerResult, err := a.evaluateComplexMath(matches[1], shardingValues)
			if err != nil {
				return 0, err
			}
			expression = parenRegex.ReplaceAllString(expression, strconv.FormatInt(innerResult, 10))
		}
	}
	
	// 最终应该是一个数字
	result, err := strconv.ParseInt(strings.TrimSpace(expression), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse final result: %s", expression)
	}
	
	return result, nil
}