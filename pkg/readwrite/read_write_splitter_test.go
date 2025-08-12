package readwrite

import (
	"context"
	"database/sql"
	"go-sharding/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 移除 MockDB，直接使用 sql.DB 进行测试

func TestNewReadWriteSplitter(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ReadWriteSplitConfig
		dataSources map[string]*sql.DB
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.ReadWriteSplitConfig{
				Name:                 "test_rw",
				MasterDataSource:     "master",
				SlaveDataSources:     []string{"slave1", "slave2"},
				LoadBalanceAlgorithm: "round_robin",
			},
			dataSources: map[string]*sql.DB{
				"master": &sql.DB{},
				"slave1": &sql.DB{},
				"slave2": &sql.DB{},
			},
			expectError: false,
		},
		{
			name: "missing master",
			config: &config.ReadWriteSplitConfig{
				Name:                 "test_rw",
				MasterDataSource:     "missing_master",
				SlaveDataSources:     []string{"slave1"},
				LoadBalanceAlgorithm: "round_robin",
			},
			dataSources: map[string]*sql.DB{
				"slave1": &sql.DB{},
			},
			expectError: true,
		},
		{
			name: "missing slave",
			config: &config.ReadWriteSplitConfig{
				Name:                 "test_rw",
				MasterDataSource:     "master",
				SlaveDataSources:     []string{"missing_slave"},
				LoadBalanceAlgorithm: "round_robin",
			},
			dataSources: map[string]*sql.DB{
				"master": &sql.DB{},
			},
			expectError: true,
		},
		{
			name: "no slaves",
			config: &config.ReadWriteSplitConfig{
				Name:                 "test_rw",
				MasterDataSource:     "master",
				SlaveDataSources:     []string{},
				LoadBalanceAlgorithm: "round_robin",
			},
			dataSources: map[string]*sql.DB{
				"master": &sql.DB{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter, err := NewReadWriteSplitter(tt.config, tt.dataSources)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, splitter)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, splitter)
				assert.Equal(t, tt.config, splitter.GetConfig())
			}
		})
	}
}

func TestReadWriteSplitter_isWriteSQL(t *testing.T) {
	config := &config.ReadWriteSplitConfig{
		Name:                 "test_rw",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1"},
		LoadBalanceAlgorithm: "round_robin",
	}
	dataSources := map[string]*sql.DB{
		"master": &sql.DB{},
		"slave1": &sql.DB{},
	}

	splitter, err := NewReadWriteSplitter(config, dataSources)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{"insert", "INSERT INTO users VALUES (1, 'test')", true},
		{"update", "UPDATE users SET name = 'test' WHERE id = 1", true},
		{"delete", "DELETE FROM users WHERE id = 1", true},
		{"create", "CREATE TABLE test (id INT)", true},
		{"drop", "DROP TABLE test", true},
		{"alter", "ALTER TABLE users ADD COLUMN email VARCHAR(255)", true},
		{"truncate", "TRUNCATE TABLE users", true},
		{"replace", "REPLACE INTO users VALUES (1, 'test')", true},
		{"select", "SELECT * FROM users", false},
		{"show", "SHOW TABLES", false},
		{"describe", "DESCRIBE users", false},
		{"explain", "EXPLAIN SELECT * FROM users", false},
		{"with spaces", "  SELECT * FROM users  ", false},
		{"lowercase insert", "insert into users values (1, 'test')", true},
		{"mixed case", "Select * From users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitter.isWriteSQL(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadWriteSplitter_Route(t *testing.T) {
	config := &config.ReadWriteSplitConfig{
		Name:                 "test_rw",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1", "slave2"},
		LoadBalanceAlgorithm: "round_robin",
	}
	
	masterDB := &sql.DB{}
	slave1DB := &sql.DB{}
	slave2DB := &sql.DB{}
	
	dataSources := map[string]*sql.DB{
		"master": masterDB,
		"slave1": slave1DB,
		"slave2": slave2DB,
	}

	splitter, err := NewReadWriteSplitter(config, dataSources)
	assert.NoError(t, err)

	// 测试写操作路由到主库
	writeDB := splitter.Route("INSERT INTO users VALUES (1, 'test')")
	assert.Equal(t, masterDB, writeDB)

	// 测试读操作路由到从库
	readDB1 := splitter.Route("SELECT * FROM users")
	readDB2 := splitter.Route("SELECT * FROM users")
	
	// 应该路由到从库
	assert.Contains(t, []*sql.DB{slave1DB, slave2DB}, readDB1)
	assert.Contains(t, []*sql.DB{slave1DB, slave2DB}, readDB2)
}

func TestReadWriteSplitter_RouteContext(t *testing.T) {
	config := &config.ReadWriteSplitConfig{
		Name:                 "test_rw",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1"},
		LoadBalanceAlgorithm: "round_robin",
	}
	
	masterDB := &sql.DB{}
	slave1DB := &sql.DB{}
	
	dataSources := map[string]*sql.DB{
		"master": masterDB,
		"slave1": slave1DB,
	}

	splitter, err := NewReadWriteSplitter(config, dataSources)
	assert.NoError(t, err)

	// 测试强制使用主库
	ctx := context.WithValue(context.Background(), "force_master", true)
	db := splitter.RouteContext(ctx, "SELECT * FROM users")
	assert.Equal(t, masterDB, db)

	// 测试事务中的读操作
	ctx = context.WithValue(context.Background(), "in_transaction", true)
	db = splitter.RouteContext(ctx, "SELECT * FROM users")
	assert.Equal(t, masterDB, db)

	// 测试正常读操作
	ctx = context.Background()
	db = splitter.RouteContext(ctx, "SELECT * FROM users")
	assert.Equal(t, slave1DB, db)
}

func TestReadWriteSplitter_LoadBalance(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
	}{
		{"round_robin", "round_robin"},
		{"random", "random"},
		{"weight", "weight"},
		{"unknown", "unknown_algorithm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config.ReadWriteSplitConfig{
				Name:                 "test_rw",
				MasterDataSource:     "master",
				SlaveDataSources:     []string{"slave1", "slave2", "slave3"},
				LoadBalanceAlgorithm: tt.algorithm,
			}
			
			dataSources := map[string]*sql.DB{
				"master": &sql.DB{},
				"slave1": &sql.DB{},
				"slave2": &sql.DB{},
				"slave3": &sql.DB{},
			}

			splitter, err := NewReadWriteSplitter(config, dataSources)
			assert.NoError(t, err)

			// 测试多次读操作，确保负载均衡工作
			dbCounts := make(map[*sql.DB]int)
			for i := 0; i < 30; i++ {
				db := splitter.Route("SELECT * FROM users")
				dbCounts[db]++
			}

			// 应该只路由到从库
			for db := range dbCounts {
				assert.Contains(t, splitter.GetSlaveDBS(), db)
			}

			// 对于轮询算法，应该相对均匀分布
			if tt.algorithm == "round_robin" {
				for _, count := range dbCounts {
					assert.Equal(t, 10, count) // 30次请求，3个从库，每个应该10次
				}
			}
		})
	}
}

func TestReadWriteSplitter_SingleSlave(t *testing.T) {
	config := &config.ReadWriteSplitConfig{
		Name:                 "test_rw",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1"},
		LoadBalanceAlgorithm: "round_robin",
	}
	
	masterDB := &sql.DB{}
	slave1DB := &sql.DB{}
	
	dataSources := map[string]*sql.DB{
		"master": masterDB,
		"slave1": slave1DB,
	}

	splitter, err := NewReadWriteSplitter(config, dataSources)
	assert.NoError(t, err)

	// 单个从库时，所有读操作都应该路由到该从库
	for i := 0; i < 10; i++ {
		db := splitter.Route("SELECT * FROM users")
		assert.Equal(t, slave1DB, db)
	}
}

func BenchmarkReadWriteSplitter_Route(b *testing.B) {
	config := &config.ReadWriteSplitConfig{
		Name:                 "test_rw",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1", "slave2"},
		LoadBalanceAlgorithm: "round_robin",
	}
	
	dataSources := map[string]*sql.DB{
		"master": &sql.DB{},
		"slave1": &sql.DB{},
		"slave2": &sql.DB{},
	}

	splitter, _ := NewReadWriteSplitter(config, dataSources)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			splitter.Route("SELECT * FROM users")
		} else {
			splitter.Route("INSERT INTO users VALUES (1, 'test')")
		}
	}
}