package algorithm

import (
	"fmt"
	"reflect"
)

// ShardingValue 分片值
type ShardingValue struct {
	ColumnName string
	Value      interface{}
	Values     []interface{} // 用于 IN 查询
	Range      *Range        // 用于范围查询
}

// Range 范围值
type Range struct {
	Start interface{}
	End   interface{}
}

// ShardingAlgorithm 分片算法接口
type ShardingAlgorithm interface {
	// DoSharding 执行分片计算
	DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error)
	// GetType 获取算法类型
	GetType() string
	// GetProperties 获取算法属性
	GetProperties() map[string]interface{}
}

// PreciseShardingAlgorithm 精确分片算法接口（用于 = 和 IN）
type PreciseShardingAlgorithm interface {
	ShardingAlgorithm
	// DoPreciseSharding 精确分片
	DoPreciseSharding(availableTargetNames []string, shardingValue *ShardingValue) (string, error)
}

// RangeShardingAlgorithm 范围分片算法接口（用于 BETWEEN）
type RangeShardingAlgorithm interface {
	ShardingAlgorithm
	// DoRangeSharding 范围分片
	DoRangeSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error)
}

// ComplexKeysShardingAlgorithm 复合分片算法接口（多分片键）
type ComplexKeysShardingAlgorithm interface {
	ShardingAlgorithm
	// DoComplexSharding 复合分片
	DoComplexSharding(availableTargetNames []string, shardingValues map[string]*ShardingValue) ([]string, error)
}

// HintShardingAlgorithm Hint分片算法接口（强制路由）
type HintShardingAlgorithm interface {
	ShardingAlgorithm
	// DoHintSharding Hint分片
	DoHintSharding(availableTargetNames []string, hintValue *ShardingValue) ([]string, error)
}

// AlgorithmFactory 算法工厂
type AlgorithmFactory struct {
	algorithms map[string]func(properties map[string]interface{}) (ShardingAlgorithm, error)
}

// NewAlgorithmFactory 创建算法工厂
func NewAlgorithmFactory() *AlgorithmFactory {
	factory := &AlgorithmFactory{
		algorithms: make(map[string]func(properties map[string]interface{}) (ShardingAlgorithm, error)),
	}
	
	// 注册内置算法
	factory.RegisterAlgorithm("INLINE", NewInlineShardingAlgorithm)
	factory.RegisterAlgorithm("MOD", NewModShardingAlgorithm)
	factory.RegisterAlgorithm("HASH_MOD", NewHashModShardingAlgorithm)
	factory.RegisterAlgorithm("RANGE", NewRangeShardingAlgorithm)
	factory.RegisterAlgorithm("COMPLEX_INLINE", NewComplexInlineShardingAlgorithm)
	factory.RegisterAlgorithm("HINT_INLINE", NewHintInlineShardingAlgorithm)
	
	return factory
}

// RegisterAlgorithm 注册算法
func (f *AlgorithmFactory) RegisterAlgorithm(name string, creator func(properties map[string]interface{}) (ShardingAlgorithm, error)) {
	f.algorithms[name] = creator
}

// CreateAlgorithm 创建算法实例
func (f *AlgorithmFactory) CreateAlgorithm(algorithmType string, properties map[string]interface{}) (ShardingAlgorithm, error) {
	creator, exists := f.algorithms[algorithmType]
	if !exists {
		return nil, fmt.Errorf("unsupported sharding algorithm: %s", algorithmType)
	}
	
	return creator(properties)
}

// GetAvailableAlgorithms 获取可用算法列表
func (f *AlgorithmFactory) GetAvailableAlgorithms() []string {
	var algorithms []string
	for name := range f.algorithms {
		algorithms = append(algorithms, name)
	}
	return algorithms
}

// ConvertToInt 转换为整数
func ConvertToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case string:
		// 尝试解析字符串为数字
		if len(v) == 0 {
			return 0, fmt.Errorf("empty string cannot be converted to int")
		}
		// 简单的字符串转数字逻辑
		var result int64
		for _, char := range v {
			if char >= '0' && char <= '9' {
				result = result*10 + int64(char-'0')
			} else {
				return 0, fmt.Errorf("invalid character in string: %c", char)
			}
		}
		return result, nil
	default:
		return 0, fmt.Errorf("unsupported type for conversion to int: %v", reflect.TypeOf(value))
	}
}

// ConvertToString 转换为字符串
func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}