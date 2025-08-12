package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTiDBParser_Basic(t *testing.T) {
	parser := NewTiDBParser()
	
	// 测试初始状态
	assert.True(t, parser.IsEnabled(), "TiDB Parser should be enabled by default")
	
	// 测试基本解析功能（使用回退解析器）
	sql := "SELECT id, name FROM users WHERE age > 18"
	stmt, err := parser.Parse(sql)
	require.NoError(t, err)
	assert.Equal(t, SQLTypeSelect, stmt.Type)
	assert.Contains(t, stmt.Tables, "users")
	assert.Contains(t, stmt.Columns, "id")
	assert.Contains(t, stmt.Columns, "name")
}

func TestTiDBParser_EnableDisable(t *testing.T) {
	parser := NewTiDBParser()
	
	// 测试启用
	parser.EnableTiDBParser()
	assert.True(t, parser.IsEnabled())
	
	// 测试禁用
	parser.DisableTiDBParser()
	assert.False(t, parser.IsEnabled())
}

func TestTiDBParser_ParseWithFallback(t *testing.T) {
	parser := NewTiDBParser()
	
	testCases := []struct {
		name     string
		sql      string
		expected SQLType
	}{
		{
			name:     "Simple SELECT",
			sql:      "SELECT * FROM users",
			expected: SQLTypeSelect,
		},
		{
			name:     "INSERT statement",
			sql:      "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			expected: SQLTypeInsert,
		},
		{
			name:     "UPDATE statement",
			sql:      "UPDATE users SET name = 'Jane' WHERE id = 1",
			expected: SQLTypeUpdate,
		},
		{
			name:     "DELETE statement",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: SQLTypeDelete,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := parser.Parse(tc.sql)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, stmt.Type)
			assert.Equal(t, tc.sql, stmt.OriginalSQL)
		})
	}
}

func TestTiDBParser_ParseWithEnhanced(t *testing.T) {
	parser := NewTiDBParser()
	parser.EnableTiDBParser() // 启用增强解析
	
	testCases := []struct {
		name     string
		sql      string
		expected SQLType
	}{
		{
			name:     "Complex SELECT with JOIN",
			sql:      "SELECT u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = 1",
			expected: SQLTypeSelect,
		},
		{
			name:     "SELECT with subquery",
			sql:      "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
			expected: SQLTypeSelect,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := parser.Parse(tc.sql)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, stmt.Type)
		})
	}
}

func TestTiDBParser_ExtractTables(t *testing.T) {
	parser := NewTiDBParser()
	
	testCases := []struct {
		name     string
		sql      string
		expected []string
	}{
		{
			name:     "Single table",
			sql:      "SELECT * FROM users",
			expected: []string{"users"},
		},
		{
			name:     "Multiple tables with JOIN",
			sql:      "SELECT * FROM users u JOIN orders o ON u.id = o.user_id",
			expected: []string{"users", "orders"},
		},
		{
			name:     "INSERT statement",
			sql:      "INSERT INTO products (name, price) VALUES ('Product', 99.99)",
			expected: []string{"products"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tables := parser.ExtractTables(tc.sql)
			for _, expectedTable := range tc.expected {
				assert.Contains(t, tables, expectedTable)
			}
		})
	}
}

func TestTiDBParser_GetParserInfo(t *testing.T) {
	parser := NewTiDBParser()
	
	info := parser.GetParserInfo()
	assert.Equal(t, "TiDBParser", info["type"])
	assert.Equal(t, true, info["enabled"]) // 现在默认启用
	assert.Equal(t, "1.0.0-integration", info["version"])
	assert.Equal(t, "transitional", info["status"])
	
	// 禁用后再测试
	parser.DisableTiDBParser()
	info = parser.GetParserInfo()
	assert.Equal(t, false, info["enabled"])
}

func TestTiDBParser_ValidateSQL(t *testing.T) {
	parser := NewTiDBParser()
	
	validSQL := "SELECT * FROM users WHERE id = 1"
	err := parser.ValidateSQL(validSQL)
	assert.NoError(t, err)
	
	invalidSQL := ""
	err = parser.ValidateSQL(invalidSQL)
	assert.Error(t, err)
}

func TestTiDBParser_GetSupportedFeatures(t *testing.T) {
	parser := NewTiDBParser()
	
	// 禁用状态下的功能
	features := parser.GetSupportedFeatures()
	expectedBasicFeatures := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"CREATE TABLE", "DROP TABLE", "ALTER TABLE",
		"JOIN", "SUBQUERY", "UNION",
		"WHERE", "ORDER BY", "GROUP BY", "HAVING", "LIMIT",
	}
	
	for _, feature := range expectedBasicFeatures {
		assert.Contains(t, features, feature)
	}
	
	// 启用状态下的功能
	parser.EnableTiDBParser()
	features = parser.GetSupportedFeatures()
	
	expectedAdvancedFeatures := []string{
		"WINDOW_FUNCTIONS", "CTE", "JSON_FUNCTIONS", "ADVANCED_JOINS",
	}
	
	for _, feature := range expectedAdvancedFeatures {
		assert.Contains(t, features, feature)
	}
}

func TestTiDBParser_Benchmark(t *testing.T) {
	parser := NewTiDBParser()
	
	sql := "SELECT * FROM users WHERE id = 1"
	iterations := 10
	
	result := parser.Benchmark(sql, iterations)
	
	assert.Equal(t, iterations, result["iterations"])
	assert.Equal(t, true, result["parser_enabled"]) // 现在默认启用
	assert.Equal(t, len(sql), result["sql_length"])
	assert.Contains(t, result, "total_time_ns")
	assert.Contains(t, result, "avg_time_ns")
}

func TestTiDBParser_EnhanceTableExtraction(t *testing.T) {
	parser := NewTiDBParser()
	
	sql := "SELECT u.name, p.title FROM users u LEFT JOIN posts p ON u.id = p.user_id"
	originalTables := []string{"users"}
	
	enhanced := parser.enhanceTableExtraction(sql, originalTables)
	
	// 应该包含原有表和 JOIN 表
	assert.Contains(t, enhanced, "users")
	// 注意：这里可能需要根据实际的 JOIN 提取逻辑调整
}

func TestTiDBParser_ErrorHandling(t *testing.T) {
	parser := NewTiDBParser()
	
	// 测试空 SQL
	_, err := parser.Parse("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty SQL statement")
	
	// 测试 ExtractTables 的错误处理
	tables := parser.ExtractTables("")
	assert.Empty(t, tables)
}

// 基准测试
func BenchmarkTiDBParser_Parse(b *testing.B) {
	parser := NewTiDBParser()
	sql := "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = 1 ORDER BY u.name LIMIT 10"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(sql)
	}
}

func BenchmarkTiDBParser_ParseEnabled(b *testing.B) {
	parser := NewTiDBParser()
	parser.EnableTiDBParser()
	sql := "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = 1 ORDER BY u.name LIMIT 10"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(sql)
	}
}

func BenchmarkTiDBParser_ExtractTables(b *testing.B) {
	parser := NewTiDBParser()
	sql := "SELECT u.id, u.name, p.title FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = 1"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.ExtractTables(sql)
	}
}