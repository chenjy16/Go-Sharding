package optimizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLOptimizer(t *testing.T) {
	t.Run("create optimizer", func(t *testing.T) {
		optimizer := NewSQLOptimizer()
		assert.NotNil(t, optimizer)
		assert.Len(t, optimizer.rules, 4) // 默认注册4个规则
	})

	t.Run("register custom rule", func(t *testing.T) {
		optimizer := NewSQLOptimizer()
		initialRuleCount := len(optimizer.rules)
		
		customRule := &TestRule{}
		optimizer.RegisterRule(customRule)
		
		assert.Len(t, optimizer.rules, initialRuleCount+1)
	})

	t.Run("optimize simple sql", func(t *testing.T) {
		optimizer := NewSQLOptimizer()
		sql := "SELECT * FROM users WHERE id = 1"
		
		optimizedSQL, err := optimizer.Optimize(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, optimizedSQL)
	})

	t.Run("optimize with rule error", func(t *testing.T) {
		optimizer := NewSQLOptimizer()
		errorRule := &ErrorRule{}
		optimizer.RegisterRule(errorRule)
		
		sql := "SELECT * FROM users"
		optimizedSQL, err := optimizer.Optimize(sql)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ErrorRule")
		assert.Equal(t, sql, optimizedSQL) // 应该返回原始SQL
	})
}

func TestPredicatePushdownRule(t *testing.T) {
	rule := &PredicatePushdownRule{}

	t.Run("get rule name", func(t *testing.T) {
		assert.Equal(t, "PredicatePushdown", rule.GetRuleName())
	})

	t.Run("apply to simple sql", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = 1"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result) // 简单SQL不变
	})

	t.Run("apply to sql with subquery", func(t *testing.T) {
		sql := "SELECT * FROM (SELECT id, name FROM users) WHERE id = 1"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("apply to sql without subquery", func(t *testing.T) {
		sql := "UPDATE users SET name = 'test'"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})
}

func TestColumnPruningRule(t *testing.T) {
	rule := &ColumnPruningRule{}

	t.Run("get rule name", func(t *testing.T) {
		assert.Equal(t, "ColumnPruning", rule.GetRuleName())
	})

	t.Run("apply to select star", func(t *testing.T) {
		sql := "SELECT * FROM users"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result) // 当前实现不修改
	})

	t.Run("apply to specific columns", func(t *testing.T) {
		sql := "SELECT id, name FROM users"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})
}

func TestIndexHintRule(t *testing.T) {
	rule := &IndexHintRule{}

	t.Run("get rule name", func(t *testing.T) {
		assert.Equal(t, "IndexHint", rule.GetRuleName())
	})

	t.Run("apply to sql with id condition", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = 1"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("apply to sql with user_id condition", func(t *testing.T) {
		sql := "SELECT * FROM orders WHERE user_id = 123"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("apply to sql without where", func(t *testing.T) {
		sql := "SELECT * FROM users"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("apply to sql with non-sharding column", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE name = 'john'"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestJoinReorderRule(t *testing.T) {
	rule := &JoinReorderRule{}

	t.Run("get rule name", func(t *testing.T) {
		assert.Equal(t, "JoinReorder", rule.GetRuleName())
	})

	t.Run("apply to sql with join", func(t *testing.T) {
		sql := "SELECT * FROM users u JOIN orders o ON u.id = o.user_id"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("apply to sql without join", func(t *testing.T) {
		sql := "SELECT * FROM users"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("apply to sql with multiple joins", func(t *testing.T) {
		sql := "SELECT * FROM users u JOIN orders o ON u.id = o.user_id JOIN products p ON o.product_id = p.id"
		result, err := rule.Apply(sql)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

func TestOptimizationContext(t *testing.T) {
	t.Run("create context", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats:   make(map[string]*TableStatistics),
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		assert.NotNil(t, ctx)
		assert.NotNil(t, ctx.TableStats)
		assert.NotNil(t, ctx.IndexInfo)
		assert.NotNil(t, ctx.ShardingInfo)
	})

	t.Run("table statistics", func(t *testing.T) {
		stats := &TableStatistics{
			RowCount:    1000,
			DataSize:    1024 * 1024,
			IndexCount:  3,
			LastUpdated: 1234567890,
		}
		
		assert.Equal(t, int64(1000), stats.RowCount)
		assert.Equal(t, int64(1024*1024), stats.DataSize)
		assert.Equal(t, 3, stats.IndexCount)
		assert.Equal(t, int64(1234567890), stats.LastUpdated)
	})

	t.Run("sharding info", func(t *testing.T) {
		shardingInfo := &ShardingInfo{
			ShardingColumn: "user_id",
			ShardingType:   "MOD",
			ShardCount:     4,
		}
		
		assert.Equal(t, "user_id", shardingInfo.ShardingColumn)
		assert.Equal(t, "MOD", shardingInfo.ShardingType)
		assert.Equal(t, 4, shardingInfo.ShardCount)
	})
}

func TestCostBasedOptimizer(t *testing.T) {
	t.Run("create cost based optimizer", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats:   make(map[string]*TableStatistics),
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		optimizer := NewCostBasedOptimizer(ctx)
		assert.NotNil(t, optimizer)
		assert.Equal(t, ctx, optimizer.context)
	})

	t.Run("estimate cost simple query", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats: map[string]*TableStatistics{
				"users": {
					RowCount: 1000,
					DataSize: 1024 * 1024,
				},
			},
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		optimizer := NewCostBasedOptimizer(ctx)
		sql := "SELECT * FROM users WHERE id = 1"
		
		cost, err := optimizer.EstimateCost(sql)
		require.NoError(t, err)
		assert.Greater(t, cost, 0.0)
		assert.Equal(t, 0.5, cost) // 1000 * 0.001 * 0.5 (WHERE条件减少50%)
	})

	t.Run("estimate cost with join", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats: map[string]*TableStatistics{
				"users": {
					RowCount: 1000,
					DataSize: 1024 * 1024,
				},
				"orders": {
					RowCount: 5000,
					DataSize: 2 * 1024 * 1024,
				},
			},
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		optimizer := NewCostBasedOptimizer(ctx)
		sql := "SELECT * FROM users u JOIN orders o ON u.id = o.user_id WHERE u.id = 1"
		
		cost, err := optimizer.EstimateCost(sql)
		require.NoError(t, err)
		assert.Greater(t, cost, 0.0)
		// (1000 + 5000) * 0.001 + 10 (JOIN) = 6 + 10 = 16, 然后 * 0.5 (WHERE) = 8
		assert.Equal(t, 8.0, cost)
	})

	t.Run("estimate cost without table stats", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats:   make(map[string]*TableStatistics),
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		optimizer := NewCostBasedOptimizer(ctx)
		sql := "SELECT * FROM unknown_table"
		
		cost, err := optimizer.EstimateCost(sql)
		require.NoError(t, err)
		assert.Equal(t, 0.0, cost)
	})

	t.Run("extract tables from sql", func(t *testing.T) {
		ctx := &OptimizationContext{
			TableStats:   make(map[string]*TableStatistics),
			IndexInfo:    make(map[string][]string),
			ShardingInfo: make(map[string]*ShardingInfo),
		}
		
		optimizer := NewCostBasedOptimizer(ctx)
		
		tests := []struct {
			name     string
			sql      string
			expected []string
		}{
			{
				name:     "simple select",
				sql:      "SELECT * FROM users",
				expected: []string{"users"},
			},
			{
				name:     "join query",
				sql:      "SELECT * FROM users u JOIN orders o ON u.id = o.user_id",
				expected: []string{"users", "orders"},
			},
			{
				name:     "multiple joins",
				sql:      "SELECT * FROM users u JOIN orders o ON u.id = o.user_id JOIN products p ON o.product_id = p.id",
				expected: []string{"users", "orders", "products"},
			},
			{
				name:     "no tables",
				sql:      "SELECT 1",
				expected: []string{},
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tables := optimizer.extractTables(tt.sql)
				assert.ElementsMatch(t, tt.expected, tables)
			})
		}
	})
}

// 测试用的规则实现
type TestRule struct{}

func (r *TestRule) GetRuleName() string {
	return "TestRule"
}

func (r *TestRule) Apply(sql string) (string, error) {
	return sql, nil
}

type ErrorRule struct{}

func (r *ErrorRule) GetRuleName() string {
	return "ErrorRule"
}

func (r *ErrorRule) Apply(sql string) (string, error) {
	return "", assert.AnError
}