package id

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnowflakeGenerator(t *testing.T) {
	tests := []struct {
		name         string
		workerID     int64
		datacenterID int64
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid parameters",
			workerID:     1,
			datacenterID: 1,
			expectError:  false,
		},
		{
			name:         "boundary values",
			workerID:     31,
			datacenterID: 31,
			expectError:  false,
		},
		{
			name:         "zero values",
			workerID:     0,
			datacenterID: 0,
			expectError:  false,
		},
		{
			name:         "invalid worker ID - negative",
			workerID:     -1,
			datacenterID: 1,
			expectError:  true,
			errorMsg:     "worker ID must be between 0 and 31",
		},
		{
			name:         "invalid worker ID - too large",
			workerID:     32,
			datacenterID: 1,
			expectError:  true,
			errorMsg:     "worker ID must be between 0 and 31",
		},
		{
			name:         "invalid datacenter ID - negative",
			workerID:     1,
			datacenterID: -1,
			expectError:  true,
			errorMsg:     "datacenter ID must be between 0 and 31",
		},
		{
			name:         "invalid datacenter ID - too large",
			workerID:     1,
			datacenterID: 32,
			expectError:  true,
			errorMsg:     "datacenter ID must be between 0 and 31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewSnowflakeGenerator(tt.workerID, tt.datacenterID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)
				assert.Equal(t, tt.workerID, generator.workerID)
				assert.Equal(t, tt.datacenterID, generator.datacenterID)
			}
		})
	}
}

func TestSnowflakeGenerator_NextID(t *testing.T) {
	generator, err := NewSnowflakeGenerator(1, 1)
	require.NoError(t, err)

	// 测试生成单个 ID
	id1, err := generator.NextID()
	assert.NoError(t, err)
	assert.Greater(t, id1, int64(0))

	// 测试生成多个 ID，确保唯一性
	ids := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id, err := generator.NextID()
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))
		assert.False(t, ids[id], "ID should be unique: %d", id)
		ids[id] = true
	}
}

func TestSnowflakeGenerator_Concurrency(t *testing.T) {
	generator, err := NewSnowflakeGenerator(1, 1)
	require.NoError(t, err)

	const numGoroutines = 10
	const numIDsPerGoroutine = 100

	var wg sync.WaitGroup
	idChan := make(chan int64, numGoroutines*numIDsPerGoroutine)

	// 启动多个 goroutine 并发生成 ID
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIDsPerGoroutine; j++ {
				id, err := generator.NextID()
				assert.NoError(t, err)
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 检查所有 ID 的唯一性
	ids := make(map[int64]bool)
	count := 0
	for id := range idChan {
		assert.False(t, ids[id], "ID should be unique: %d", id)
		ids[id] = true
		count++
	}

	assert.Equal(t, numGoroutines*numIDsPerGoroutine, count)
}

func TestSnowflakeGenerator_IDStructure(t *testing.T) {
	workerID := int64(5)
	datacenterID := int64(3)
	generator, err := NewSnowflakeGenerator(workerID, datacenterID)
	require.NoError(t, err)

	id, err := generator.NextID()
	require.NoError(t, err)

	// 验证 ID 结构
	// 提取各个部分
	extractedDatacenterID := (id >> 17) & 0x1F // 5 位
	extractedWorkerID := (id >> 12) & 0x1F     // 5 位
	sequence := id & 0xFFF                     // 12 位

	assert.Equal(t, datacenterID, extractedDatacenterID)
	assert.Equal(t, workerID, extractedWorkerID)
	assert.GreaterOrEqual(t, sequence, int64(0))
	assert.LessOrEqual(t, sequence, int64(4095))
}

func TestNewUUIDGenerator(t *testing.T) {
	generator := NewUUIDGenerator()
	assert.NotNil(t, generator)
}

func TestUUIDGenerator_NextID(t *testing.T) {
	generator := NewUUIDGenerator()

	// 测试生成单个 ID
	id1, err := generator.NextID()
	assert.NoError(t, err)
	assert.NotEqual(t, int64(0), id1)

	// 测试生成多个 ID，确保不同
	id2, err := generator.NextID()
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	// 测试生成多个 ID
	ids := make(map[int64]bool)
	for i := 0; i < 100; i++ {
		id, err := generator.NextID()
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), id)
		ids[id] = true
	}

	// UUID 生成器应该产生不同的 ID（虽然理论上可能重复，但概率极低）
	assert.Greater(t, len(ids), 90) // 允许少量重复
}

func TestUUIDGenerator_generateUUID(t *testing.T) {
	generator := NewUUIDGenerator()

	uuid1, err := generator.generateUUID()
	assert.NoError(t, err)
	assert.Len(t, uuid1, 36) // UUID 标准长度

	uuid2, err := generator.generateUUID()
	assert.NoError(t, err)
	assert.NotEqual(t, uuid1, uuid2)

	// 验证 UUID 格式 (xxxxxxxx-xxxx-4xxx-xxxx-xxxxxxxxxxxx)
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, uuid1)
}

func TestNewIncrementGenerator(t *testing.T) {
	start := int64(100)
	step := int64(5)

	generator := NewIncrementGenerator(start, step)
	assert.NotNil(t, generator)
	assert.Equal(t, start, generator.current)
	assert.Equal(t, step, generator.step)
}

func TestIncrementGenerator_NextID(t *testing.T) {
	start := int64(100)
	step := int64(5)
	generator := NewIncrementGenerator(start, step)

	// 测试递增
	id1, err := generator.NextID()
	assert.NoError(t, err)
	assert.Equal(t, start+step, id1)

	id2, err := generator.NextID()
	assert.NoError(t, err)
	assert.Equal(t, start+step*2, id2)

	id3, err := generator.NextID()
	assert.NoError(t, err)
	assert.Equal(t, start+step*3, id3)
}

func TestIncrementGenerator_Concurrency(t *testing.T) {
	generator := NewIncrementGenerator(0, 1)

	const numGoroutines = 10
	const numIDsPerGoroutine = 100

	var wg sync.WaitGroup
	idChan := make(chan int64, numGoroutines*numIDsPerGoroutine)

	// 启动多个 goroutine 并发生成 ID
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIDsPerGoroutine; j++ {
				id, err := generator.NextID()
				assert.NoError(t, err)
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 检查所有 ID 的唯一性
	ids := make(map[int64]bool)
	count := 0
	for id := range idChan {
		assert.False(t, ids[id], "ID should be unique: %d", id)
		ids[id] = true
		count++
	}

	assert.Equal(t, numGoroutines*numIDsPerGoroutine, count)
}

func TestNewGeneratorFactory(t *testing.T) {
	factory := NewGeneratorFactory()
	assert.NotNil(t, factory)
	assert.NotNil(t, factory.generators)
}

func TestGeneratorFactory_RegisterAndGetGenerator(t *testing.T) {
	factory := NewGeneratorFactory()
	generator := NewUUIDGenerator()

	// 注册生成器
	factory.RegisterGenerator("test-uuid", generator)

	// 获取生成器
	retrieved, exists := factory.GetGenerator("test-uuid")
	assert.True(t, exists)
	assert.Equal(t, generator, retrieved)

	// 获取不存在的生成器
	_, exists = factory.GetGenerator("nonexistent")
	assert.False(t, exists)
}

func TestGeneratorFactory_CreateGenerator(t *testing.T) {
	factory := NewGeneratorFactory()

	tests := []struct {
		name          string
		generatorType string
		config        map[string]interface{}
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "snowflake generator",
			generatorType: "snowflake",
			config: map[string]interface{}{
				"workerID":     int64(1),
				"datacenterID": int64(1),
			},
			expectError: false,
		},
		{
			name:          "snowflake generator with default values",
			generatorType: "snowflake",
			config:        map[string]interface{}{},
			expectError:   false,
		},
		{
			name:          "uuid generator",
			generatorType: "uuid",
			config:        map[string]interface{}{},
			expectError:   false,
		},
		{
			name:          "increment generator",
			generatorType: "increment",
			config: map[string]interface{}{
				"start": int64(100),
				"step":  int64(5),
			},
			expectError: false,
		},
		{
			name:          "increment generator with default values",
			generatorType: "increment",
			config:        map[string]interface{}{},
			expectError:   false,
		},
		{
			name:          "unsupported generator type",
			generatorType: "unsupported",
			config:        map[string]interface{}{},
			expectError:   true,
			errorMsg:      "unsupported generator type: unsupported",
		},
		{
			name:          "snowflake with invalid worker ID",
			generatorType: "snowflake",
			config: map[string]interface{}{
				"workerID":     int64(32),
				"datacenterID": int64(1),
			},
			expectError: true,
			errorMsg:    "worker ID must be between 0 and 31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := factory.CreateGenerator(tt.generatorType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)

				// 测试生成 ID
				id, err := generator.NextID()
				assert.NoError(t, err)
				assert.NotEqual(t, int64(0), id)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSnowflakeGenerator_NextID(b *testing.B) {
	generator, err := NewSnowflakeGenerator(1, 1)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.NextID()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUUIDGenerator_NextID(b *testing.B) {
	generator := NewUUIDGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.NextID()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIncrementGenerator_NextID(b *testing.B) {
	generator := NewIncrementGenerator(0, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.NextID()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSnowflakeGenerator_Concurrent(b *testing.B) {
	generator, err := NewSnowflakeGenerator(1, 1)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := generator.NextID()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}