package sharding

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"go-sharding/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDriver 模拟数据库驱动
type MockDriver struct{}

func (d MockDriver) Open(name string) (driver.Conn, error) {
	return &MockConn{}, nil
}

// MockConn 模拟数据库连接
type MockConn struct{}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return &MockStmt{}, nil
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) Begin() (driver.Tx, error) {
	return &MockTx{}, nil
}

// MockStmt 模拟预处理语句
type MockStmt struct{}

func (s *MockStmt) Close() error {
	return nil
}

func (s *MockStmt) NumInput() int {
	return -1 // -1 表示接受任意数量的参数
}

func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &MockResult{}, nil
}

func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &MockRows{}, nil
}

// MockResult 模拟执行结果
type MockResult struct{}

func (r *MockResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (r *MockResult) RowsAffected() (int64, error) {
	return 1, nil
}

// MockRows 模拟查询结果
type MockRows struct {
	closed bool
}

func (r *MockRows) Columns() []string {
	return []string{"id", "name"}
}

func (r *MockRows) Close() error {
	r.closed = true
	return nil
}

func (r *MockRows) Next(dest []driver.Value) error {
	if r.closed {
		return sql.ErrNoRows
	}
	dest[0] = int64(1)
	dest[1] = "test"
	r.closed = true
	return nil
}

// MockTx 模拟事务
type MockTx struct{}

func (t *MockTx) Commit() error {
	return nil
}

func (t *MockTx) Rollback() error {
	return nil
}

func init() {
	sql.Register("mock", &MockDriver{})
}

func TestNewShardingDataSource(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ShardingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &config.ShardingConfig{
				DataSources: map[string]*config.DataSourceConfig{
					"ds_0": {
						DriverName: "mock",
						URL:        "mock://test",
						MaxIdle:    10,
						MaxOpen:    100,
					},
					"ds_1": {
						DriverName: "mock",
						URL:        "mock://test",
						MaxIdle:    10,
						MaxOpen:    100,
					},
				},
				ShardingRule: &config.ShardingRuleConfig{
					Tables: map[string]*config.TableRuleConfig{
						"t_order": {
							ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
							DatabaseStrategy: &config.ShardingStrategyConfig{
								ShardingColumn: "user_id",
								Algorithm:      "inline",
								Type:           "inline",
							},
							TableStrategy: &config.ShardingStrategyConfig{
								ShardingColumn: "order_id",
								Algorithm:      "inline",
								Type:           "inline",
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid driver",
			config: &config.ShardingConfig{
				DataSources: map[string]*config.DataSourceConfig{
					"ds_0": {
						DriverName: "invalid_driver",
						URL:        "invalid://test",
					},
				},
				ShardingRule: &config.ShardingRuleConfig{
					Tables: map[string]*config.TableRuleConfig{
						"t_order": {
							ActualDataNodes: "ds_0.t_order",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "failed to open database ds_0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewShardingDataSource(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, ds)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ds)
				assert.NotNil(t, ds.dataSources)
				assert.NotNil(t, ds.shardingRule)
				assert.NotNil(t, ds.router)
				assert.NotNil(t, ds.rewriter)
				assert.NotNil(t, ds.merger)
				assert.NotNil(t, ds.idGenerator)

				// 测试连接池设置
				for name, dsConfig := range tt.config.DataSources {
					db := ds.GetConnection(name)
					assert.NotNil(t, db)
					
					// 验证连接池参数设置
					stats := db.Stats()
					assert.Equal(t, dsConfig.MaxOpen, stats.MaxOpenConnections)
				}

				// 清理资源
				err = ds.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestShardingDataSource_GetConnection(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	// 测试获取存在的连接
	conn := ds.GetConnection("ds_0")
	assert.NotNil(t, conn)

	// 测试获取不存在的连接
	conn = ds.GetConnection("ds_999")
	assert.Nil(t, conn)
}

func TestShardingDataSource_GetShardingRule(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	rule := ds.GetShardingRule()
	assert.NotNil(t, rule)
	assert.Equal(t, config.ShardingRule, rule)
}

func TestShardingDataSource_GetConfiguredTables(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
				"t_user": {
					ActualDataNodes: "ds_0.t_user",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	tables := ds.GetConfiguredTables()
	assert.NotNil(t, tables)
	assert.Len(t, tables, 2)
	assert.Contains(t, tables, "t_order")
	assert.Contains(t, tables, "t_user")
}

func TestShardingDB_Query(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()
	assert.NotNil(t, db)

	// 测试查询
	rows, err := db.Query("SELECT * FROM t_order WHERE id = ?", 1)
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	
	if rows != nil && rows.rows != nil {
		rows.Close()
	}
}

func TestShardingDB_QueryContext(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()
	ctx := context.Background()

	// 测试带上下文的查询
	rows, err := db.QueryContext(ctx, "SELECT * FROM t_order WHERE id = ?", 1)
	assert.NoError(t, err)
	assert.NotNil(t, rows)
	
	if rows != nil && rows.rows != nil {
		rows.Close()
	}
}

func TestShardingDataSource_Close(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
			"ds_1": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_${0..1}.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(t, err)

	// 测试关闭所有连接
	err = ds.Close()
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkShardingDataSource_Query(b *testing.B) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"ds_0": {
				DriverName: "mock",
				URL:        "mock://test",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewShardingDataSource(config)
	require.NoError(b, err)
	defer ds.Close()

	db := ds.DB()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query("SELECT * FROM t_order WHERE id = ?", i)
		if err == nil && rows != nil && rows.rows != nil {
			rows.Close()
		}
	}
}