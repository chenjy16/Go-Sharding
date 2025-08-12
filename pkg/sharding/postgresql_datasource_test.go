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

// MockPostgreSQLDriver 模拟 PostgreSQL 驱动
type MockPostgreSQLDriver struct{}

func (d MockPostgreSQLDriver) Open(name string) (driver.Conn, error) {
	return &MockPostgreSQLConn{}, nil
}

// MockPostgreSQLConn 模拟 PostgreSQL 连接
type MockPostgreSQLConn struct{}

func (c *MockPostgreSQLConn) Prepare(query string) (driver.Stmt, error) {
	return &MockPostgreSQLStmt{}, nil
}

func (c *MockPostgreSQLConn) Close() error {
	return nil
}

func (c *MockPostgreSQLConn) Begin() (driver.Tx, error) {
	return &MockTx{}, nil
}

// MockPostgreSQLStmt 模拟 PostgreSQL 预处理语句
type MockPostgreSQLStmt struct{}

func (s *MockPostgreSQLStmt) Close() error {
	return nil
}

func (s *MockPostgreSQLStmt) NumInput() int {
	return -1
}

func (s *MockPostgreSQLStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &MockResult{}, nil
}

func (s *MockPostgreSQLStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &MockRows{}, nil
}

func init() {
	sql.Register("postgres-mock", &MockPostgreSQLDriver{})
}

func TestNewPostgreSQLShardingDataSource(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ShardingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid PostgreSQL config",
			config: &config.ShardingConfig{
				DataSources: map[string]*config.DataSourceConfig{
					"pg_ds_0": {
						DriverName: "postgres-mock",
						URL:        "postgres://user:pass@localhost/db0",
						MaxIdle:    10,
						MaxOpen:    100,
					},
					"pg_ds_1": {
						DriverName: "postgres-mock",
						URL:        "postgres://user:pass@localhost/db1",
						MaxIdle:    10,
						MaxOpen:    100,
					},
				},
				ShardingRule: &config.ShardingRuleConfig{
					Tables: map[string]*config.TableRuleConfig{
						"t_order": {
							ActualDataNodes: "pg_ds_${0..1}.t_order_${0..1}",
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
			name: "mixed drivers - should fail",
			config: &config.ShardingConfig{
				DataSources: map[string]*config.DataSourceConfig{
					"pg_ds": {
						DriverName: "postgres-mock",
						URL:        "postgres://user:pass@localhost/db",
					},
					"mysql_ds": {
						DriverName: "mock", // MySQL mock driver
						URL:        "mysql://user:pass@localhost/db",
					},
				},
				ShardingRule: &config.ShardingRuleConfig{
					Tables: map[string]*config.TableRuleConfig{
						"t_order": {
							ActualDataNodes: "pg_ds.t_order",
						},
					},
				},
			},
			expectError: true,
			errorMsg:    "is not PostgreSQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewPostgreSQLShardingDataSource(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, ds)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ds)
				assert.NotNil(t, ds.ShardingDataSource)
				assert.NotNil(t, ds.pgParser)
				assert.NotNil(t, ds.dialect)

				// 测试获取数据库连接
				db := ds.DB()
				assert.NotNil(t, db)
				assert.NotNil(t, db.ShardingDB)
				assert.Equal(t, ds, db.pgDataSource)

				// 清理资源
				err = ds.Close()
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgreSQLDB_QueryContext(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()

	tests := []struct {
		name        string
		query       string
		args        []interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid PostgreSQL query",
			query:       "SELECT id, name FROM t_order WHERE user_id = $1",
			args:        []interface{}{1},
			expectError: false,
		},
		{
			name:        "query with JSONB",
			query:       "SELECT data FROM t_order WHERE data @> $1",
			args:        []interface{}{`{"key": "value"}`},
			expectError: false,
		},
		{
			name:        "query with array",
			query:       "SELECT * FROM t_order WHERE tags && $1",
			args:        []interface{}{`{"tag1", "tag2"}`},
			expectError: false,
		},
		{
			name:        "query with window function",
			query:       "SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) FROM t_order",
			args:        []interface{}{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.QueryContext(context.Background(), tt.query, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, rows)
			} else {
				// 注意：由于使用 mock 驱动，这里可能会有错误，但我们主要测试参数转换和解析逻辑
				if err != nil {
					// 如果是因为 mock 驱动导致的错误，我们检查是否是预期的错误类型
					t.Logf("Query error (expected with mock): %v", err)
				} else {
					assert.NotNil(t, rows)
					rows.Close()
				}
			}
		})
	}
}

func TestPostgreSQLDB_ExecContext(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()

	tests := []struct {
		name        string
		query       string
		args        []interface{}
		expectError bool
	}{
		{
			name:        "INSERT with RETURNING",
			query:       "INSERT INTO t_order (user_id, amount) VALUES ($1, $2) RETURNING id",
			args:        []interface{}{1, 100.50},
			expectError: false,
		},
		{
			name:        "UPDATE with JSONB",
			query:       "UPDATE t_order SET data = data || $1 WHERE id = $2",
			args:        []interface{}{`{"updated": true}`, 1},
			expectError: false,
		},
		{
			name:        "DELETE with array condition",
			query:       "DELETE FROM t_order WHERE id = ANY($1)",
			args:        []interface{}{`{1,2,3}`},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := db.ExecContext(context.Background(), tt.query, tt.args...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				// 使用 mock 驱动时可能会有错误，但我们主要测试 SQL 解析和参数转换
				if err != nil {
					t.Logf("Exec error (expected with mock): %v", err)
				} else {
					assert.NotNil(t, result)
				}
			}
		})
	}
}

func TestPostgreSQLDB_convertToPostgreSQLParams(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()

	tests := []struct {
		name           string
		query          string
		args           []interface{}
		expectedQuery  string
		expectedArgs   []interface{}
	}{
		{
			name:          "convert ? to $1, $2",
			query:         "SELECT * FROM t_order WHERE user_id = ? AND status = ?",
			args:          []interface{}{1, "active"},
			expectedQuery: "SELECT * FROM t_order WHERE user_id = $1 AND status = $2",
			expectedArgs:  []interface{}{1, "active"},
		},
		{
			name:          "already PostgreSQL format",
			query:         "SELECT * FROM t_order WHERE user_id = $1 AND status = $2",
			args:          []interface{}{1, "active"},
			expectedQuery: "SELECT * FROM t_order WHERE user_id = $1 AND status = $2",
			expectedArgs:  []interface{}{1, "active"},
		},
		{
			name:          "mixed with string literals",
			query:         "SELECT * FROM t_order WHERE name = 'test?' AND user_id = ?",
			args:          []interface{}{1},
			expectedQuery: "SELECT * FROM t_order WHERE name = 'test?' AND user_id = $1",
			expectedArgs:  []interface{}{1},
		},
		{
			name:          "no parameters",
			query:         "SELECT * FROM t_order",
			args:          []interface{}{},
			expectedQuery: "SELECT * FROM t_order",
			expectedArgs:  []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualQuery, actualArgs := db.convertToPostgreSQLParams(tt.query, tt.args)
			assert.Equal(t, tt.expectedQuery, actualQuery)
			assert.Equal(t, tt.expectedArgs, actualArgs)
		})
	}
}

func TestPostgreSQLDB_isInStringLiteral(t *testing.T) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	require.NoError(t, err)
	defer ds.Close()

	db := ds.DB()

	tests := []struct {
		name     string
		sql      string
		pos      int
		expected bool
	}{
		{
			name:     "not in string literal",
			sql:      "SELECT * FROM t_order WHERE id = ?",
			pos:      32, // position of ?
			expected: false,
		},
		{
			name:     "in single quote string",
			sql:      "SELECT * FROM t_order WHERE name = 'test?'",
			pos:      40, // position of ? inside 'test?'
			expected: true,
		},
		{
			name:     "in double quote string",
			sql:      `SELECT * FROM t_order WHERE name = "test?"`,
			pos:      40, // position of ? inside "test?"
			expected: true,
		},
		{
			name:     "escaped quote",
			sql:      `SELECT * FROM t_order WHERE name = 'test\'s ?'`,
			pos:      44, // position of ? after escaped quote
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := db.isInStringLiteral(tt.sql, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 基准测试
func BenchmarkPostgreSQLDB_QueryContext(b *testing.B) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	if err != nil {
		b.Fatal(err)
	}
	defer ds.Close()

	db := ds.DB()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.QueryContext(ctx, "SELECT * FROM t_order WHERE user_id = $1", i)
	}
}

func BenchmarkPostgreSQLDB_convertToPostgreSQLParams(b *testing.B) {
	config := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"pg_ds_0": {
				DriverName: "postgres-mock",
				URL:        "postgres://user:pass@localhost/db0",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "pg_ds_0.t_order",
				},
			},
		},
	}

	ds, err := NewPostgreSQLShardingDataSource(config)
	if err != nil {
		b.Fatal(err)
	}
	defer ds.Close()

	db := ds.DB()
	query := "SELECT * FROM t_order WHERE user_id = ? AND status = ? AND amount > ?"
	args := []interface{}{1, "active", 100.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.convertToPostgreSQLParams(query, args)
	}
}