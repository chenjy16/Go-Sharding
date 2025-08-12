package algorithm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardingValue(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		sv := &ShardingValue{
			ColumnName: "user_id",
			Value:      123,
		}
		assert.Equal(t, "user_id", sv.ColumnName)
		assert.Equal(t, 123, sv.Value)
		assert.Nil(t, sv.Values)
		assert.Nil(t, sv.Range)
	})

	t.Run("multiple values", func(t *testing.T) {
		sv := &ShardingValue{
			ColumnName: "user_id",
			Values:     []interface{}{1, 2, 3},
		}
		assert.Equal(t, "user_id", sv.ColumnName)
		assert.Equal(t, []interface{}{1, 2, 3}, sv.Values)
	})

	t.Run("range value", func(t *testing.T) {
		sv := &ShardingValue{
			ColumnName: "user_id",
			Range: &Range{
				Start: 1,
				End:   100,
			},
		}
		assert.Equal(t, "user_id", sv.ColumnName)
		assert.NotNil(t, sv.Range)
		assert.Equal(t, 1, sv.Range.Start)
		assert.Equal(t, 100, sv.Range.End)
	})
}

func TestAlgorithmFactory(t *testing.T) {
	t.Run("create factory", func(t *testing.T) {
		factory := NewAlgorithmFactory()
		assert.NotNil(t, factory)
		
		algorithms := factory.GetAvailableAlgorithms()
		assert.Contains(t, algorithms, "INLINE")
		assert.Contains(t, algorithms, "MOD")
		assert.Contains(t, algorithms, "HASH_MOD")
		assert.Contains(t, algorithms, "RANGE")
		assert.Contains(t, algorithms, "COMPLEX_INLINE")
		assert.Contains(t, algorithms, "HINT_INLINE")
	})

	t.Run("register custom algorithm", func(t *testing.T) {
		factory := NewAlgorithmFactory()
		
		factory.RegisterAlgorithm("CUSTOM", func(properties map[string]interface{}) (ShardingAlgorithm, error) {
			return &ModShardingAlgorithm{
				shardingCount: 2,
				properties:    properties,
			}, nil
		})
		
		algorithms := factory.GetAvailableAlgorithms()
		assert.Contains(t, algorithms, "CUSTOM")
	})

	t.Run("create mod algorithm", func(t *testing.T) {
		factory := NewAlgorithmFactory()
		
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := factory.CreateAlgorithm("MOD", properties)
		require.NoError(t, err)
		assert.NotNil(t, algorithm)
		assert.Equal(t, "MOD", algorithm.GetType())
	})

	t.Run("create unsupported algorithm", func(t *testing.T) {
		factory := NewAlgorithmFactory()
		
		algorithm, err := factory.CreateAlgorithm("UNSUPPORTED", nil)
		assert.Error(t, err)
		assert.Nil(t, algorithm)
		assert.Contains(t, err.Error(), "unsupported sharding algorithm")
	})
}

func TestModShardingAlgorithm(t *testing.T) {
	t.Run("create with valid properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		assert.NotNil(t, algorithm)
		assert.Equal(t, "MOD", algorithm.GetType())
		assert.Equal(t, properties, algorithm.GetProperties())
	})

	t.Run("create with string sharding count", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": "4",
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		assert.NotNil(t, algorithm)
	})

	t.Run("create without sharding count", func(t *testing.T) {
		properties := map[string]interface{}{}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		assert.Error(t, err)
		assert.Nil(t, algorithm)
		assert.Contains(t, err.Error(), "sharding-count is required")
	})

	t.Run("create with invalid sharding count", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": "invalid",
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		assert.Error(t, err)
		assert.Nil(t, algorithm)
	})

	t.Run("create with zero sharding count", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 0,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		assert.Error(t, err)
		assert.Nil(t, algorithm)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("do sharding with single value", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_id",
			Value:      123,
		}
		
		results, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "ds3", results[0]) // 123 % 4 = 3
	})

	t.Run("do sharding with multiple values", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_id",
			Values:     []interface{}{0, 1, 2, 3}, // 这样会映射到所有4个数据源
		}
		
		results, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Contains(t, results, "ds0") // 0 % 4 = 0
		assert.Contains(t, results, "ds1") // 1 % 4 = 1
		assert.Contains(t, results, "ds2") // 2 % 4 = 2
		assert.Contains(t, results, "ds3") // 3 % 4 = 3
	})

	t.Run("do precise sharding", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		modAlgorithm := algorithm.(*ModShardingAlgorithm)
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_id",
			Value:      123,
		}
		
		result, err := modAlgorithm.DoPreciseSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Equal(t, "ds3", result)
	})

	t.Run("handle negative values", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_id",
			Value:      -123,
		}
		
		results, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "ds3", results[0]) // abs(-123) % 4 = 3
	})
}

func TestHashModShardingAlgorithm(t *testing.T) {
	t.Run("create with valid properties", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewHashModShardingAlgorithm(properties)
		require.NoError(t, err)
		assert.NotNil(t, algorithm)
		assert.Equal(t, "HASH_MOD", algorithm.GetType())
	})

	t.Run("do sharding with string value", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewHashModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_name",
			Value:      "john_doe",
		}
		
		results, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Contains(t, availableTargets, results[0])
	})

	t.Run("do precise sharding", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewHashModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		hashModAlgorithm := algorithm.(*HashModShardingAlgorithm)
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_name",
			Value:      "john_doe",
		}
		
		result, err := hashModAlgorithm.DoPreciseSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		assert.Contains(t, availableTargets, result)
	})

	t.Run("consistent hashing", func(t *testing.T) {
		properties := map[string]interface{}{
			"sharding-count": 4,
		}
		
		algorithm, err := NewHashModShardingAlgorithm(properties)
		require.NoError(t, err)
		
		availableTargets := []string{"ds0", "ds1", "ds2", "ds3"}
		shardingValue := &ShardingValue{
			ColumnName: "user_name",
			Value:      "john_doe",
		}
		
		// 多次执行应该得到相同结果
		result1, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		
		result2, err := algorithm.DoSharding(availableTargets, shardingValue)
		require.NoError(t, err)
		
		assert.Equal(t, result1, result2)
	})
}

func TestConvertToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		hasError bool
	}{
		{"int", 123, 123, false},
		{"int32", int32(123), 123, false},
		{"int64", int64(123), 123, false},
		{"uint", uint(123), 123, false},
		{"uint32", uint32(123), 123, false},
		{"uint64", uint64(123), 123, false},
		{"string number", "123", 123, false},
		{"empty string", "", 0, true},
		{"invalid string", "abc", 0, true},
		{"string with invalid char", "12a3", 0, true},
		{"float", 123.45, 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToInt(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"int32", int32(123), "123"},
		{"int64", int64(123), "123"},
		{"uint", uint(123), "123"},
		{"uint32", uint32(123), "123"},
		{"uint64", uint64(123), "123"},
		{"float", 123.45, "123.45"},
		{"bool", true, "true"},
		{"nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}