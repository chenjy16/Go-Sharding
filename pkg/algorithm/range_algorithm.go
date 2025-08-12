package algorithm

import (
	"fmt"
	"strconv"
	"strings"
)

// RangeShardingAlgorithmImpl 范围分片算法实现
type RangeShardingAlgorithmImpl struct {
	rangeMap   map[string]*RangeDefinition
	properties map[string]interface{}
}

// RangeDefinition 范围定义
type RangeDefinition struct {
	Start int64
	End   int64
	Target string
}

// NewRangeShardingAlgorithm 创建范围分片算法
func NewRangeShardingAlgorithm(properties map[string]interface{}) (ShardingAlgorithm, error) {
	rangeMapValue, ok := properties["range-map"].(string)
	if !ok {
		return nil, fmt.Errorf("range-map is required for range sharding algorithm")
	}
	
	rangeMap, err := parseRangeMap(rangeMapValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range-map: %w", err)
	}
	
	return &RangeShardingAlgorithmImpl{
		rangeMap:   rangeMap,
		properties: properties,
	}, nil
}

// DoSharding 执行分片计算
func (a *RangeShardingAlgorithmImpl) DoSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	if shardingValue.Range != nil {
		// 处理范围查询
		return a.doRangeSharding(availableTargetNames, shardingValue.Range)
	}
	
	if shardingValue.Values != nil && len(shardingValue.Values) > 0 {
		// 处理 IN 查询
		var results []string
		for _, value := range shardingValue.Values {
			targets, err := a.findTargetsForValue(value, availableTargetNames)
			if err != nil {
				return nil, err
			}
			for _, target := range targets {
				if !contains(results, target) {
					results = append(results, target)
				}
			}
		}
		return results, nil
	}
	
	// 处理单值查询
	return a.findTargetsForValue(shardingValue.Value, availableTargetNames)
}

// DoRangeSharding 范围分片
func (a *RangeShardingAlgorithmImpl) DoRangeSharding(availableTargetNames []string, shardingValue *ShardingValue) ([]string, error) {
	if shardingValue.Range == nil {
		return nil, fmt.Errorf("range value is required for range sharding")
	}
	
	return a.doRangeSharding(availableTargetNames, shardingValue.Range)
}

// GetType 获取算法类型
func (a *RangeShardingAlgorithmImpl) GetType() string {
	return "RANGE"
}

// GetProperties 获取算法属性
func (a *RangeShardingAlgorithmImpl) GetProperties() map[string]interface{} {
	return a.properties
}

// doRangeSharding 执行范围分片
func (a *RangeShardingAlgorithmImpl) doRangeSharding(availableTargetNames []string, rangeValue *Range) ([]string, error) {
	startValue, err := ConvertToInt(rangeValue.Start)
	if err != nil {
		return nil, fmt.Errorf("failed to convert range start to int: %w", err)
	}
	
	endValue, err := ConvertToInt(rangeValue.End)
	if err != nil {
		return nil, fmt.Errorf("failed to convert range end to int: %w", err)
	}
	
	var results []string
	for _, rangeDef := range a.rangeMap {
		// 检查范围是否有重叠
		if startValue <= rangeDef.End && endValue >= rangeDef.Start {
			if contains(availableTargetNames, rangeDef.Target) && !contains(results, rangeDef.Target) {
				results = append(results, rangeDef.Target)
			}
		}
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("no target found for range [%d, %d]", startValue, endValue)
	}
	
	return results, nil
}

// findTargetsForValue 为单个值查找目标
func (a *RangeShardingAlgorithmImpl) findTargetsForValue(value interface{}, availableTargetNames []string) ([]string, error) {
	intValue, err := ConvertToInt(value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to int: %w", err)
	}
	
	for _, rangeDef := range a.rangeMap {
		if intValue >= rangeDef.Start && intValue <= rangeDef.End {
			if contains(availableTargetNames, rangeDef.Target) {
				return []string{rangeDef.Target}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no target found for value %v", value)
}

// parseRangeMap 解析范围映射
// 格式: "0-100:ds_0,101-200:ds_1,201-300:ds_2"
func parseRangeMap(rangeMapStr string) (map[string]*RangeDefinition, error) {
	rangeMap := make(map[string]*RangeDefinition)
	
	ranges := strings.Split(rangeMapStr, ",")
	for i, rangeStr := range ranges {
		rangeStr = strings.TrimSpace(rangeStr)
		parts := strings.Split(rangeStr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", rangeStr)
		}
		
		rangePart := strings.TrimSpace(parts[0])
		target := strings.TrimSpace(parts[1])
		
		rangeBounds := strings.Split(rangePart, "-")
		if len(rangeBounds) != 2 {
			return nil, fmt.Errorf("invalid range bounds: %s", rangePart)
		}
		
		start, err := strconv.ParseInt(strings.TrimSpace(rangeBounds[0]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range start: %s", rangeBounds[0])
		}
		
		end, err := strconv.ParseInt(strings.TrimSpace(rangeBounds[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid range end: %s", rangeBounds[1])
		}
		
		if start > end {
			return nil, fmt.Errorf("range start %d cannot be greater than end %d", start, end)
		}
		
		key := fmt.Sprintf("range_%d", i)
		rangeMap[key] = &RangeDefinition{
			Start:  start,
			End:    end,
			Target: target,
		}
	}
	
	return rangeMap, nil
}