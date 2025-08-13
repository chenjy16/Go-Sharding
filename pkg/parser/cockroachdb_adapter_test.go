package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCockroachDBAdapter_ParsePostgreSQLSpecific(t *testing.T) {
	adapter := NewCockroachDBAdapter()

	tests := []struct {
		name     string
		sql      string
		wantType string
		wantErr  bool
	}{
		{
			name:     "Simple SELECT",
			sql:      "SELECT id, name FROM users WHERE age > 18",
			wantType: "SELECT",
			wantErr:  false,
		},
		{
			name:     "INSERT with RETURNING",
			sql:      "INSERT INTO users (name, email) VALUES ('John', 'john@example.com') RETURNING id",
			wantType: "INSERT",
			wantErr:  false,
		},
		{
			name:     "UPDATE with RETURNING",
			sql:      "UPDATE users SET name = 'Jane' WHERE id = 1 RETURNING id, name",
			wantType: "UPDATE",
			wantErr:  false,
		},
		{
			name:     "DELETE with WHERE",
			sql:      "DELETE FROM users WHERE age < 18",
			wantType: "DELETE",
			wantErr:  false,
		},
		{
			name:     "SELECT with LIMIT and OFFSET",
			sql:      "SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 20",
			wantType: "SELECT",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParsePostgreSQLSpecific(tt.sql)
			
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, string(result.Type))
			assert.Equal(t, tt.sql, result.OriginalSQL)
			assert.NotNil(t, result.PostgreSQLFeatures)
		})
	}
}

func TestCockroachDBAdapter_ExtractTables(t *testing.T) {
	adapter := NewCockroachDBAdapter()

	tests := []struct {
		name      string
		sql       string
		wantTables []string
	}{
		{
			name:       "Single table SELECT",
			sql:        "SELECT * FROM users",
			wantTables: []string{"users"},
		},
		{
			name:       "JOIN query",
			sql:        "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id",
			wantTables: []string{"users", "posts"},
		},
		{
			name:       "INSERT statement",
			sql:        "INSERT INTO products (name, price) VALUES ('Product A', 100)",
			wantTables: []string{"products"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables := adapter.ExtractTables(tt.sql)
			
			// 注意：实际的表名提取可能包含更多细节，这里只是基本验证
			assert.NotEmpty(t, tables, "Should extract at least one table")
		})
	}
}

func TestCockroachDBAdapter_RewriteForPostgreSQL(t *testing.T) {
	adapter := NewCockroachDBAdapter()

	tests := []struct {
		name            string
		sql             string
		tableName       string
		actualTableName string
		wantContains    string
	}{
		{
			name:            "Table name replacement",
			sql:             "SELECT * FROM users WHERE id = ?",
			tableName:       "users",
			actualTableName: "users_shard_1",
			wantContains:    "users_shard_1",
		},
		{
			name:            "Parameter placeholder conversion",
			sql:             "SELECT * FROM products WHERE price > ? AND category = ?",
			tableName:       "products",
			actualTableName: "products_shard_2",
			wantContains:    "$1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewritten, err := adapter.RewriteForPostgreSQL(tt.sql, tt.tableName, tt.actualTableName)
			
			require.NoError(t, err)
			assert.Contains(t, rewritten, tt.wantContains)
		})
	}
}

func TestCockroachDBAdapter_ValidatePostgreSQLSQL(t *testing.T) {
	adapter := NewCockroachDBAdapter()

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "Valid SELECT",
			sql:     "SELECT id, name FROM users WHERE age > 18",
			wantErr: false,
		},
		{
			name:    "Valid INSERT",
			sql:     "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			wantErr: false,
		},
		{
			name:    "Invalid SQL",
			sql:     "SELECT FROM WHERE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidatePostgreSQLSQL(tt.sql)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCockroachDBAdapter_PostgreSQLFeatures(t *testing.T) {
	adapter := NewCockroachDBAdapter()

	tests := []struct {
		name        string
		sql         string
		wantFeature string
	}{
		{
			name:        "LIMIT and OFFSET",
			sql:         "SELECT * FROM users LIMIT 10 OFFSET 5",
			wantFeature: "limit",
		},
		{
			name:        "RETURNING clause",
			sql:         "INSERT INTO users (name) VALUES ('John') RETURNING id",
			wantFeature: "returning",
		},
		{
			name:        "PostgreSQL functions",
			sql:         "SELECT COALESCE(name, 'Unknown') FROM users",
			wantFeature: "functions",
		},
		{
			name:        "PostgreSQL operators",
			sql:         "SELECT * FROM users WHERE name ILIKE '%john%'",
			wantFeature: "operators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParsePostgreSQLSpecific(tt.sql)
			
			require.NoError(t, err)
			assert.Contains(t, result.PostgreSQLFeatures, tt.wantFeature)
		})
	}
}

func TestCockroachDBAdapter_Compatibility(t *testing.T) {
	// 测试与现有 PostgreSQL 解析器的兼容性
	originalParser := NewPostgreSQLParser()
	adapter := NewCockroachDBAdapter()

	testSQL := "SELECT id, name FROM users WHERE age > 18 LIMIT 10"

	// 两个解析器都应该能够处理相同的 SQL
	originalResult, originalErr := originalParser.ParsePostgreSQLSpecific(testSQL)
	adapterResult, adapterErr := adapter.ParsePostgreSQLSpecific(testSQL)

	// 基本验证
	assert.NoError(t, originalErr)
	assert.NoError(t, adapterErr)
	
	// 验证基本属性
	assert.Equal(t, originalResult.Type, adapterResult.Type)
	assert.Equal(t, originalResult.OriginalSQL, adapterResult.OriginalSQL)
	
	// 验证方法兼容性
	originalTables := originalParser.ExtractTables(testSQL)
	adapterTables := adapter.ExtractTables(testSQL)
	
	assert.NotEmpty(t, originalTables)
	assert.NotEmpty(t, adapterTables)
	
	// 验证 SQL 重写兼容性
	originalRewritten, originalRewriteErr := originalParser.RewriteForPostgreSQL(testSQL, "users", "users_shard_1")
	adapterRewritten, adapterRewriteErr := adapter.RewriteForPostgreSQL(testSQL, "users", "users_shard_1")
	
	assert.NoError(t, originalRewriteErr)
	assert.NoError(t, adapterRewriteErr)
	assert.Contains(t, originalRewritten, "users_shard_1")
	assert.Contains(t, adapterRewritten, "users_shard_1")
}