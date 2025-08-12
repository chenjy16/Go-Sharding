package algorithm

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

// ModShardingAlgorithm 取模分片算法
type ModShardingAlgorithm struct {
	shardingCount int
	properties    map[string]interface{}
}

// NewModShardingAlgorithm 创建取模分片算法
func NewModShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	shardingCountValue, ok := properties["sharding-count"]
	if !ok {
		return nil, fmt.Errorf("sharding-count is required for mod sharding algorithm")
	}
	
	var shardingCount int
	switch v := shardingCountValue.(type) {
	case int:
		shardingCount = v
	case string:
		var err error
		shardingCount, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid sharding-count: %s", v)
		}
	default:
		return nil, fmt.Errorf("sharding-count must be int or string")
	}
	
	if shardingCount <= 0 {
		return nil, fmt.Errorf("sharding-count must be positive")
	}
	
	return &ModShardingAlgorithm{
		shardingCount: shardingCount,
		properties:    properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *ModShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	if shardingValue.Values != nil && len(shardingValue.Values) > 0 {
		// 处理 IN 查询
		var results []string
		for _, value := range shardingValue.Values {
			target, err := a.calculateTarget(value, availableTargetNames)
			if err != nil {
				return nil, err
			}
			if !contains(results, target) {
				results = append(results, target)
			}
		}
		return results, nil
	}
	
	// 处理单值查询
	target, err := a.calculateTarget(shardingValue.Value, availableTargetNames)
	if err != nil {
		return nil, err
	}
	
	return []string{target}, nil
}

// DoPreciseSharding 精确分片
func (a *ModShardingAlgorithm) DoPreciseSharding(availableTargetNames []string, shardingValue *ShardingValue) (string, error) {
	return a.calculateTarget(shardingValue.Value, availableTargetNames)
}

// GetType 获取算法类型
func (a *ModShardingAlgorithm) GetType() string {
	return "MOD"
}

// GetProperties 获取算法属性
func (a *ModShardingAlgorithm) GetProperties() map[string]interface{} {
	return a.properties
}

// calculateTarget 计算目标
func (a *ModShardingAlgorithm) calculateTarget(value interface{}, availableTargetNames []string) (string, error) {
	intValue, err := ConvertToInt(value)
	if err != nil {
		return "", fmt.Errorf("failed to convert value to int: %w", err)
	}
	
	index := int(intValue) % a.shardingCount
	if index < 0 {
		index = -index
	}
	
	if index >= len(availableTargetNames) {
		index = index % len(availableTargetNames)
	}
	
	return availableTargetNames[index], nil
}

// HashModShardingAlgorithm 哈希取模分片算法
type HashModShardingAlgorithm struct {
	shardingCount int
	properties    map[string]interface{}
}

// NewHashModShardingAlgorithm 创建哈希取模分片算法
func NewHashModShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	shardingCountValue, ok := properties["sharding-count"]
	if !ok {
		return nil, fmt.Errorf("sharding-count is required for hash mod sharding algorithm")
	}
	
	var shardingCount int
	switch v := shardingCountValue.(type) {
	case int:
		shardingCount = v
	case string:
		var err error
		shardingCount, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid sharding-count: %s", v)
		}
	default:
		return nil, fmt.Errorf("sharding-count must be int or string")
	}
	
	if shardingCount <= 0 {
		return nil, fmt.Errorf("sharding-count must be positive")
	}
	
	return &HashModShardingAlgorithm{
		shardingCount: shardingCount,
		properties:    properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *HashModShardingAlgorithm) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	if shardingValue.Values != nil && len(shardingValue.Values) > 0 {
		// 处理 IN 查询
		var results []string
		for _, value := range shardingValue.Values {
			target, err := a.calculateTarget(value, availableTargetNames)
			if err != nil {
				return nil, err
			}
			if !contains(results, target) {
				results = append(results, target)
			}
		}
		return results, nil
	}
	
	// 处理单值查询
	target, err := a.calculateTarget(shardingValue.Value, availableTargetNames)
	if err != nil {
		return nil, err
	}
	
	return []string{target}, nil
}

// DoPreciseSharding 精确分片
func (a *HashModShardingAlgorithm) DoPreciseSharding(availableTargetNames []string, shardingValue *ShardingValue) (string, error) {
	return a.calculateTarget(shardingValue.Value, availableTargetNames)
}

// GetType 获取算法类型
func (a *HashModShardingAlgorithm) GetType() string {
	return "HASH_MOD"
}

// GetProperties 获取算法属性
func (a *HashModShardingAlgorithm) GetProperties() map[string]interface{} {
	return a.properties
}

// calculateTarget 计算目标
func (a *HashModShardingAlgorithm) calculateTarget(value interface{}, availableTargetNames []string) (string, error) {
	valueStr := ConvertToString(value)
	hash := crc32.ChecksumIEEE([]byte(valueStr))
	index := int(hash) % a.shardingCount
	
	if index >= len(availableTargetNames) {
		index = index % len(availableTargetNames)
	}
	
	return availableTargetNames[index], nil
}