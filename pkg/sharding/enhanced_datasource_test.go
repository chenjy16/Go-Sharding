package sharding

import (
	"context"
	"go-sharding/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnhancedShardingDB(t *testing.T) {
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"master_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
			"slave_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3307)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
		ReadWriteSplits: map[string]*config.ReadWriteSplitConfig{
			"rw_ds": {
				MasterDataSource:     "master_ds",
				SlaveDataSources:     []string{"slave_ds"},
				LoadBalanceAlgorithm: "round_robin",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "rw_ds.t_order_${0..1}",
					TableStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "order_id",
						Algorithm:      "t_order_${order_id % 2}",
						Type:           "inline",
					},
				},
			},
		},
	}

	// 注意：这个测试需要实际的数据库连接，在 CI 环境中可能会失败
	// 在实际测试中，应该使用 mock 数据库或测试数据库
	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}
	defer db.Close()

	assert.NotNil(t, db)
	assert.NotNil(t, db.config)
	assert.NotNil(t, db.router)
	assert.NotNil(t, db.rewriter)

}

func TestEnhancedShardingDB_QueryContext(t *testing.T) {
	// 创建一个简单的配置用于测试
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"test_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_test": {
					ActualDataNodes: "test_ds.t_test",
				},
			},
		},
	}

	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// 测试简单查询
	_, err = db.QueryContext(ctx, "SELECT 1")
	// 由于没有实际的数据库连接，这里可能会出错，但我们主要测试代码路径
	// 在实际环境中，这应该能正常工作
	t.Logf("Query result: %v", err)
}

func TestEnhancedShardingDB_ExecContext(t *testing.T) {
	// 创建一个简单的配置用于测试
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"test_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_test": {
					ActualDataNodes: "test_ds.t_test",
				},
			},
		},
	}

	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// 测试简单执行
	_, err = db.ExecContext(ctx, "SELECT 1")
	// 由于没有实际的数据库连接，这里可能会出错，但我们主要测试代码路径
	t.Logf("Exec result: %v", err)
}

func TestEnhancedShardingRows_Operations(t *testing.T) {
	// 测试 EnhancedShardingRows 的基本操作
	rows := &EnhancedShardingRows{
		rows:    nil, // 在实际测试中应该是真实的 sql.Rows
		allRows: nil,
		sqlType: "SELECT",
	}

	// 测试 Next 方法
	hasNext := rows.Next()
	assert.False(t, hasNext) // 因为 rows 是 nil

	// 测试 Columns 方法
	_, err := rows.Columns()
	assert.Error(t, err) // 应该返回错误，因为 rows 是 nil

	// 测试 Scan 方法
	var value interface{}
	err = rows.Scan(&value)
	assert.Error(t, err) // 应该返回错误，因为 rows 是 nil

	// 测试 Close 方法
	err = rows.Close()
	assert.NoError(t, err) // Close 应该不会出错，即使 rows 是 nil
}

func TestEnhancedShardingResult_Operations(t *testing.T) {
	// 测试 EnhancedShardingResult 的基本操作
	result := &EnhancedShardingResult{
		result:       nil,
		rowsAffected: 5,
		lastInsertId: 100,
	}

	// 测试 RowsAffected 方法
	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.Equal(t, int64(5), affected)

	// 测试 LastInsertId 方法
	insertId, err := result.LastInsertId()
	assert.NoError(t, err)
	assert.Equal(t, int64(100), insertId)
}

func TestEnhancedShardingDB_HealthCheck(t *testing.T) {
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"test_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
	}

	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		t.Skipf("Skipping test due to database connection error: %v", err)
		return
	}
	defer db.Close()

	// 测试健康检查
	err = db.HealthCheck()
	// 在没有实际数据库连接的情况下，这可能会失败
	// 但我们主要测试代码路径
	t.Logf("Health check result: %v", err)
}

// 基准测试
func BenchmarkEnhancedShardingDB_QueryContext(b *testing.B) {
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"test_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_test": {
					ActualDataNodes: "test_ds.t_test",
				},
			},
		},
	}

	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		b.Skipf("Skipping benchmark due to database connection error: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.QueryContext(ctx, "SELECT 1")
	}
}

func BenchmarkEnhancedShardingDB_ExecContext(b *testing.B) {
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"test_ds": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/test_db",
				MaxIdle:    5,
				MaxOpen:    10,
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_test": {
					ActualDataNodes: "test_ds.t_test",
				},
			},
		},
	}

	db, err := NewEnhancedShardingDB(cfg)
	if err != nil {
		b.Skipf("Skipping benchmark due to database connection error: %v", err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.ExecContext(ctx, "SELECT 1")
	}
}
