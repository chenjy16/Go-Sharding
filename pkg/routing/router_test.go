package routing

import (
	"go-sharding/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShardingRouter(t *testing.T) {
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {
			DriverName: "mysql",
			URL:        "root:@tcp(localhost:3306)/ds_0",
		},
		"ds_1": {
			DriverName: "mysql",
			URL:        "root:@tcp(localhost:3306)/ds_1",
		},
	}

	shardingRule := &config.ShardingRuleConfig{
		Tables: map[string]*config.TableRuleConfig{
			"t_order": {
				ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
			},
		},
	}

	router := NewShardingRouter(dataSources, shardingRule)
	assert.NotNil(t, router)
	assert.Equal(t, dataSources, router.dataSources)
	assert.Equal(t, shardingRule, router.shardingRule)
}

func TestShardingRouter_Route(t *testing.T) {
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_0"},
		"ds_1": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_1"},
	}

	tests := []struct {
		name           string
		shardingRule   *config.ShardingRuleConfig
		logicTable     string
		shardingValues map[string]interface{}
		expectedCount  int
		expectError    bool
		errorMsg       string
	}{
		{
			name: "table not found",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{},
			},
			logicTable:     "t_nonexistent",
			shardingValues: map[string]interface{}{},
			expectError:    true,
			errorMsg:       "table rule not found for table: t_nonexistent",
		},
		{
			name: "no sharding values - return all nodes",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{
					"t_order": {
						ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
					},
				},
			},
			logicTable:     "t_order",
			shardingValues: map[string]interface{}{},
			expectedCount:  4, // ds_0.t_order_0, ds_0.t_order_1, ds_1.t_order_0, ds_1.t_order_1
			expectError:    false,
		},
		{
			name: "with database strategy",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{
					"t_order": {
						ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
						DatabaseStrategy: &config.ShardingStrategyConfig{
							ShardingColumn: "user_id",
							Algorithm:      "ds_${user_id % 2}",
							Type:           "inline",
						},
					},
				},
			},
			logicTable: "t_order",
			shardingValues: map[string]interface{}{
				"user_id": 1,
			},
			expectedCount: 2, // ds_1.t_order_0, ds_1.t_order_1
			expectError:   false,
		},
		{
			name: "with table strategy",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{
					"t_order": {
						ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
						TableStrategy: &config.ShardingStrategyConfig{
							ShardingColumn: "order_id",
							Algorithm:      "t_order_${order_id % 2}",
							Type:           "inline",
						},
					},
				},
			},
			logicTable: "t_order",
			shardingValues: map[string]interface{}{
				"order_id": 2,
			},
			expectedCount: 2, // ds_0.t_order_0, ds_1.t_order_0
			expectError:   false,
		},
		{
			name: "with both strategies",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{
					"t_order": {
						ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
						DatabaseStrategy: &config.ShardingStrategyConfig{
							ShardingColumn: "user_id",
							Algorithm:      "ds_${user_id % 2}",
							Type:           "inline",
						},
						TableStrategy: &config.ShardingStrategyConfig{
							ShardingColumn: "order_id",
							Algorithm:      "t_order_${order_id % 2}",
							Type:           "inline",
						},
					},
				},
			},
			logicTable: "t_order",
			shardingValues: map[string]interface{}{
				"user_id":  1,
				"order_id": 2,
			},
			expectedCount: 1, // ds_1.t_order_0
			expectError:   false,
		},
		{
			name: "invalid actual data nodes",
			shardingRule: &config.ShardingRuleConfig{
				Tables: map[string]*config.TableRuleConfig{
					"t_order": {
						ActualDataNodes: "invalid_format",
					},
				},
			},
			logicTable:     "t_order",
			shardingValues: map[string]interface{}{},
			expectError:    true,
			errorMsg:       "failed to parse actual data nodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewShardingRouter(dataSources, tt.shardingRule)
			results, err := router.Route(tt.logicTable, tt.shardingValues)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				assert.Len(t, results, tt.expectedCount)

				// 验证结果的有效性
				for _, result := range results {
					assert.NotEmpty(t, result.DataSource)
					assert.NotEmpty(t, result.Table)
				}
			}
		})
	}
}

func TestShardingRouter_parseActualDataNodes(t *testing.T) {
	router := &ShardingRouter{}

	tests := []struct {
		name           string
		expression     string
		expectedCount  int
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "range expression",
			expression:     "ds_${0..1}.t_order_${0..1}",
			expectedCount:  4,
			expectError:    false,
		},
		{
			name:           "list expression",
			expression:     "ds_${[0, 1]}.t_order_${[0, 1]}",
			expectedCount:  4,
			expectError:    false,
		},
		{
			name:           "single value",
			expression:     "ds_0.t_order_0",
			expectedCount:  1,
			expectError:    false,
		},
		{
			name:           "invalid format - no dot",
			expression:     "ds_0_t_order_0",
			expectError:    true,
			errorMsg:       "invalid actual data nodes expression",
		},
		{
			name:           "invalid format - multiple dots",
			expression:     "ds_0.t_order.0",
			expectError:    true,
			errorMsg:       "invalid actual data nodes expression",
		},
		{
			name:           "invalid range format",
			expression:     "ds_${0..}.t_order_${0..1}",
			expectedCount:  2,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := router.parseActualDataNodes(tt.expression)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, nodes)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, nodes)
				assert.Len(t, nodes, tt.expectedCount)

				// 验证节点的有效性
				for _, node := range nodes {
					assert.NotEmpty(t, node.DataSource)
					assert.NotEmpty(t, node.Table)
				}
			}
		})
	}
}

func TestShardingRouter_parseRangeExpression(t *testing.T) {
	router := &ShardingRouter{}

	tests := []struct {
		name           string
		pattern        string
		expectedCount  int
		expectedValues []string
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "range pattern",
			pattern:        "ds_${0..2}",
			expectedCount:  3,
			expectedValues: []string{"ds_0", "ds_1", "ds_2"},
			expectError:    false,
		},
		{
			name:           "list pattern",
			pattern:        "ds_${[0, 2, 4]}",
			expectedCount:  3,
			expectedValues: []string{"ds_0", "ds_2", "ds_4"},
			expectError:    false,
		},
		{
			name:           "single value pattern",
			pattern:        "ds_0",
			expectedCount:  1,
			expectedValues: []string{"ds_0"},
			expectError:    false,
		},
		{
			name:           "invalid range start",
			pattern:        "ds_${a..1}",
			expectedCount:  1,
			expectedValues: []string{"ds_${a..1}"},
			expectError:    false,
		},
		{
			name:           "invalid range end",
			pattern:        "ds_${0..b}",
			expectedCount:  1,
			expectedValues: []string{"ds_${0..b}"},
			expectError:    false,
		},
		{
			name:           "invalid list format",
			pattern:        "ds_${[0, a]}",
			expectedCount:  2,
			expectedValues: []string{"ds_0", "ds_a"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := router.parseRangeExpression(tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				assert.Len(t, results, tt.expectedCount)
				assert.Equal(t, tt.expectedValues, results)
			}
		})
	}
}

func TestShardingRouter_calculateSharding(t *testing.T) {
	router := &ShardingRouter{}

	tests := []struct {
		name           string
		strategy       *config.ShardingStrategyConfig
		shardingValues map[string]interface{}
		expectedCount  int
		expectError    bool
		errorMsg       string
	}{
		{
			name: "inline strategy with modulo",
			strategy: &config.ShardingStrategyConfig{
				ShardingColumn: "user_id",
				Algorithm:      "ds_${user_id % 2}",
				Type:           "inline",
			},
			shardingValues: map[string]interface{}{
				"user_id": 5,
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "missing sharding column",
			strategy: &config.ShardingStrategyConfig{
				ShardingColumn: "user_id",
				Algorithm:      "ds_${user_id % 2}",
				Type:           "inline",
			},
			shardingValues: map[string]interface{}{
				"order_id": 5,
			},
			expectError: true,
			errorMsg:    "sharding column user_id not found in sharding values",
		},
		{
			name: "unsupported strategy type",
			strategy: &config.ShardingStrategyConfig{
				ShardingColumn: "user_id",
				Algorithm:      "ds_${user_id % 2}",
				Type:           "unsupported",
			},
			shardingValues: map[string]interface{}{
				"user_id": 5,
			},
			expectError: true,
			errorMsg:    "unsupported sharding strategy type: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := router.calculateSharding(tt.strategy, tt.shardingValues)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				assert.Len(t, results, tt.expectedCount)
			}
		})
	}
}

func TestShardingRouter_isValidDataNode(t *testing.T) {
	router := &ShardingRouter{}

	dataNodes := []*DataNode{
		{DataSource: "ds_0", Table: "t_order_0"},
		{DataSource: "ds_0", Table: "t_order_1"},
		{DataSource: "ds_1", Table: "t_order_0"},
		{DataSource: "ds_1", Table: "t_order_1"},
	}

	tests := []struct {
		name       string
		dataSource string
		table      string
		expected   bool
	}{
		{
			name:       "valid node",
			dataSource: "ds_0",
			table:      "t_order_0",
			expected:   true,
		},
		{
			name:       "invalid data source",
			dataSource: "ds_2",
			table:      "t_order_0",
			expected:   false,
		},
		{
			name:       "invalid table",
			dataSource: "ds_0",
			table:      "t_order_2",
			expected:   false,
		},
		{
			name:       "both invalid",
			dataSource: "ds_2",
			table:      "t_order_2",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.isValidDataNode(dataNodes, tt.dataSource, tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShardingRouter_EdgeCases(t *testing.T) {
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_0"},
		"ds_1": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_1"},
	}

	t.Run("empty sharding rule", func(t *testing.T) {
		shardingRule := &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{},
		}
		router := NewShardingRouter(dataSources, shardingRule)
		
		results, err := router.Route("t_order", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table rule not found")
		assert.Nil(t, results)
	})

	t.Run("complex sharding values", func(t *testing.T) {
		shardingRule := &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
					DatabaseStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "user_id",
						Algorithm:      "ds_${user_id % 2}",
						Type:           "inline",
					},
					TableStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "order_id",
						Algorithm:      "t_order_${order_id % 2}",
						Type:           "inline",
					},
				},
			},
		}
		router := NewShardingRouter(dataSources, shardingRule)
		
		shardingValues := map[string]interface{}{
			"user_id":  int64(123),
			"order_id": int32(456),
			"amount":   99.99,
		}
		
		results, err := router.Route("t_order", shardingValues)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 1)
		if len(results) > 0 {
			assert.Equal(t, "ds_1", results[0].DataSource)
			assert.Equal(t, "t_order_0", results[0].Table)
		}
	})
}

// 基准测试
func BenchmarkShardingRouter_Route(b *testing.B) {
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_0"},
		"ds_1": {DriverName: "mysql", URL: "root:@tcp(localhost:3306)/ds_1"},
	}

	shardingRule := &config.ShardingRuleConfig{
		Tables: map[string]*config.TableRuleConfig{
			"t_order": {
				ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
				DatabaseStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "user_id",
					Algorithm:      "ds_${user_id % 2}",
					Type:           "inline",
				},
				TableStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "order_id",
					Algorithm:      "t_order_${order_id % 2}",
					Type:           "inline",
				},
			},
		},
	}

	router := NewShardingRouter(dataSources, shardingRule)
	shardingValues := map[string]interface{}{
		"user_id":  1,
		"order_id": 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.Route("t_order", shardingValues)
	}
}

func BenchmarkShardingRouter_parseActualDataNodes(b *testing.B) {
	router := &ShardingRouter{}
	expression := "ds_${0..1}.t_order_${0..1}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = router.parseActualDataNodes(expression)
	}
}