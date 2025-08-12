package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgreSQLParser(t *testing.T) {
	parser := NewPostgreSQLParser()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.EnhancedSQLParser)
	assert.NotNil(t, parser.dialect)
}

func TestPostgreSQLParser_ParsePostgreSQLSpecific(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name        string
		sql         string
		expectError bool
		checkFunc   func(*testing.T, *EnhancedSQLStatement)
	}{
		{
			name:        "simple SELECT",
			sql:         "SELECT id, name FROM users WHERE id = $1",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.Contains(t, stmt.Tables, "users")
				assert.NotNil(t, stmt.PostgreSQLFeatures)
			},
		},
		{
			name:        "SELECT with JSONB operator",
			sql:         "SELECT data FROM users WHERE data @> $1",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				operators, ok := stmt.PostgreSQLFeatures["operators"].([]string)
				assert.True(t, ok)
				assert.Contains(t, operators, "@>")
			},
		},
		{
			name:        "SELECT with array operator",
			sql:         "SELECT * FROM users WHERE tags && $1",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				operators, ok := stmt.PostgreSQLFeatures["operators"].([]string)
				assert.True(t, ok)
				assert.Contains(t, operators, "&&")
			},
		},
		{
			name:        "SELECT with ILIKE",
			sql:         "SELECT * FROM users WHERE name ILIKE $1",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				operators, ok := stmt.PostgreSQLFeatures["operators"].([]string)
				assert.True(t, ok)
				assert.Contains(t, operators, "ILIKE")
			},
		},
		{
			name:        "SELECT with window function",
			sql:         "SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) FROM users",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				functions, ok := stmt.PostgreSQLFeatures["functions"].([]string)
				assert.True(t, ok)
				assert.Contains(t, functions, "ROW_NUMBER")
			},
		},
		{
			name:        "INSERT with RETURNING",
			sql:         "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeInsert, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				returning, ok := stmt.PostgreSQLFeatures["returning"].(string)
				assert.True(t, ok)
				assert.Equal(t, "id", returning)
			},
		},
		{
			name:        "SELECT with LIMIT OFFSET",
			sql:         "SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 20",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeSelect, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				limit, ok := stmt.PostgreSQLFeatures["limit"].(string)
				assert.True(t, ok)
				assert.Equal(t, "10", limit)
				offset, ok := stmt.PostgreSQLFeatures["offset"].(string)
				assert.True(t, ok)
				assert.Equal(t, "20", offset)
			},
		},
		{
			name:        "CREATE TABLE with PostgreSQL types",
			sql:         "CREATE TABLE test (id SERIAL, data JSONB, tags TEXT[])",
			expectError: false,
			checkFunc: func(t *testing.T, stmt *EnhancedSQLStatement) {
				assert.Equal(t, SQLTypeCreate, stmt.Type)
				assert.NotNil(t, stmt.PostgreSQLFeatures)
				dataTypes, ok := stmt.PostgreSQLFeatures["dataTypes"].([]string)
				assert.True(t, ok)
				assert.Contains(t, dataTypes, "SERIAL")
				assert.Contains(t, dataTypes, "JSONB")
				assert.Contains(t, dataTypes, "ARRAY")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.ParsePostgreSQLSpecific(tt.sql)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stmt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stmt)
				if tt.checkFunc != nil {
					tt.checkFunc(t, stmt)
				}
			}
		})
	}
}

func TestPostgreSQLParser_parsePostgreSQLLimit(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedLimit  string
		expectedOffset string
	}{
		{
			name:           "LIMIT only",
			sql:            "SELECT * FROM users LIMIT 10",
			expectedLimit:  "10",
			expectedOffset: "",
		},
		{
			name:           "LIMIT with OFFSET",
			sql:            "SELECT * FROM users LIMIT 10 OFFSET 20",
			expectedLimit:  "10",
			expectedOffset: "20",
		},
		{
			name:           "LIMIT with parameter",
			sql:            "SELECT * FROM users LIMIT $1 OFFSET $2",
			expectedLimit:  "$1",
			expectedOffset: "$2",
		},
		{
			name:           "no LIMIT",
			sql:            "SELECT * FROM users",
			expectedLimit:  "",
			expectedOffset: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedSQLStatement{
				PostgreSQLFeatures: make(map[string]interface{}),
			}
			parser.parsePostgreSQLLimit(result, tt.sql)

			if tt.expectedLimit != "" {
				limit, ok := result.PostgreSQLFeatures["limit"].(string)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedLimit, limit)
			} else {
				_, exists := result.PostgreSQLFeatures["limit"]
				assert.False(t, exists)
			}

			if tt.expectedOffset != "" {
				offset, ok := result.PostgreSQLFeatures["offset"].(string)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedOffset, offset)
			} else {
				_, exists := result.PostgreSQLFeatures["offset"]
				assert.False(t, exists)
			}
		})
	}
}

func TestPostgreSQLParser_parsePostgreSQLDataTypes(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name          string
		sql           string
		expectedTypes []string
	}{
		{
			name:          "SERIAL type",
			sql:           "CREATE TABLE test (id SERIAL PRIMARY KEY)",
			expectedTypes: []string{"SERIAL"},
		},
		{
			name:          "JSONB type",
			sql:           "CREATE TABLE test (data JSONB)",
			expectedTypes: []string{"JSONB"},
		},
		{
			name:          "UUID type",
			sql:           "CREATE TABLE test (id UUID)",
			expectedTypes: []string{"UUID"},
		},
		{
			name:          "ARRAY type",
			sql:           "CREATE TABLE test (tags TEXT[])",
			expectedTypes: []string{"ARRAY"},
		},
		{
			name:          "multiple types",
			sql:           "CREATE TABLE test (id SERIAL, data JSONB, tags TEXT[])",
			expectedTypes: []string{"SERIAL", "JSONB", "ARRAY"},
		},
		{
			name:          "no PostgreSQL types",
			sql:           "CREATE TABLE test (id INT, name VARCHAR(255))",
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedSQLStatement{
				PostgreSQLFeatures: make(map[string]interface{}),
			}
			parser.parsePostgreSQLDataTypes(result, tt.sql)

			if len(tt.expectedTypes) > 0 {
				dataTypes, ok := result.PostgreSQLFeatures["dataTypes"].([]string)
				assert.True(t, ok)
				for _, expectedType := range tt.expectedTypes {
					assert.Contains(t, dataTypes, expectedType)
				}
			} else {
				_, exists := result.PostgreSQLFeatures["dataTypes"]
				assert.False(t, exists)
			}
		})
	}
}

func TestPostgreSQLParser_parsePostgreSQLFunctions(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name              string
		sql               string
		expectedFunctions []string
	}{
		{
			name:              "COALESCE function",
			sql:               "SELECT COALESCE(name, 'Unknown') FROM users",
			expectedFunctions: []string{"COALESCE"},
		},
		{
			name:              "window function",
			sql:               "SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) FROM users",
			expectedFunctions: []string{"ROW_NUMBER"},
		},
		{
			name:              "aggregate function",
			sql:               "SELECT ARRAY_AGG(name) FROM users",
			expectedFunctions: []string{"ARRAY_AGG"},
		},
		{
			name:              "multiple functions",
			sql:               "SELECT COALESCE(name, 'Unknown'), ARRAY_AGG(tags) FROM users",
			expectedFunctions: []string{"COALESCE", "ARRAY_AGG"},
		},
		{
			name:              "no PostgreSQL functions",
			sql:               "SELECT COUNT(*) FROM users",
			expectedFunctions: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedSQLStatement{
				PostgreSQLFeatures: make(map[string]interface{}),
			}
			parser.parsePostgreSQLFunctions(result, tt.sql)

			if len(tt.expectedFunctions) > 0 {
				functions, ok := result.PostgreSQLFeatures["functions"].([]string)
				assert.True(t, ok)
				for _, expectedFunc := range tt.expectedFunctions {
					assert.Contains(t, functions, expectedFunc)
				}
			} else {
				_, exists := result.PostgreSQLFeatures["functions"]
				assert.False(t, exists)
			}
		})
	}
}

func TestPostgreSQLParser_parsePostgreSQLOperators(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name              string
		sql               string
		expectedOperators []string
	}{
		{
			name:              "ILIKE operator",
			sql:               "SELECT * FROM users WHERE name ILIKE '%john%'",
			expectedOperators: []string{"ILIKE"},
		},
		{
			name:              "JSONB contains operator",
			sql:               "SELECT * FROM users WHERE data @> '{\"key\": \"value\"}'",
			expectedOperators: []string{"@>"},
		},
		{
			name:              "array overlap operator",
			sql:               "SELECT * FROM users WHERE tags && ARRAY['tag1', 'tag2']",
			expectedOperators: []string{"&&"},
		},
		{
			name:              "regex operator",
			sql:               "SELECT * FROM users WHERE name ~ '^[A-Z]'",
			expectedOperators: []string{"~"},
		},
		{
			name:              "string concatenation",
			sql:               "SELECT first_name || ' ' || last_name FROM users",
			expectedOperators: []string{"||"},
		},
		{
			name:              "multiple operators",
			sql:               "SELECT * FROM users WHERE name ILIKE '%john%' AND data @> '{\"active\": true}'",
			expectedOperators: []string{"ILIKE", "@>"},
		},
		{
			name:              "no PostgreSQL operators",
			sql:               "SELECT * FROM users WHERE id = 1",
			expectedOperators: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedSQLStatement{
				PostgreSQLFeatures: make(map[string]interface{}),
			}
			parser.parsePostgreSQLOperators(result, tt.sql)

			if len(tt.expectedOperators) > 0 {
				operators, ok := result.PostgreSQLFeatures["operators"].([]string)
				assert.True(t, ok)
				for _, expectedOp := range tt.expectedOperators {
					assert.Contains(t, operators, expectedOp)
				}
			} else {
				_, exists := result.PostgreSQLFeatures["operators"]
				assert.False(t, exists)
			}
		})
	}
}

func TestPostgreSQLParser_parseReturningClause(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name              string
		sql               string
		expectedReturning string
	}{
		{
			name:              "INSERT with RETURNING id",
			sql:               "INSERT INTO users (name) VALUES ('John') RETURNING id",
			expectedReturning: "id",
		},
		{
			name:              "INSERT with RETURNING multiple columns",
			sql:               "INSERT INTO users (name, email) VALUES ('John', 'john@example.com') RETURNING id, created_at",
			expectedReturning: "id, created_at",
		},
		{
			name:              "UPDATE with RETURNING",
			sql:               "UPDATE users SET name = 'Jane' WHERE id = 1 RETURNING name, updated_at",
			expectedReturning: "name, updated_at",
		},
		{
			name:              "DELETE with RETURNING",
			sql:               "DELETE FROM users WHERE id = 1 RETURNING *",
			expectedReturning: "*",
		},
		{
			name:              "no RETURNING clause",
			sql:               "INSERT INTO users (name) VALUES ('John')",
			expectedReturning: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &EnhancedSQLStatement{
				PostgreSQLFeatures: make(map[string]interface{}),
			}
			parser.parseReturningClause(result, tt.sql)

			if tt.expectedReturning != "" {
				returning, ok := result.PostgreSQLFeatures["returning"].(string)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedReturning, returning)
			} else {
				_, exists := result.PostgreSQLFeatures["returning"]
				assert.False(t, exists)
			}
		})
	}
}

func TestPostgreSQLParser_RewriteForPostgreSQL(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name            string
		sql             string
		tableName       string
		actualTableName string
		expectedSQL     string
		expectError     bool
	}{
		{
			name:            "simple table replacement",
			sql:             "SELECT * FROM t_order WHERE id = $1",
			tableName:       "t_order",
			actualTableName: "t_order_0",
			expectedSQL:     "SELECT * FROM t_order_0 WHERE id = $1",
			expectError:     false,
		},
		{
			name:            "multiple table occurrences",
			sql:             "SELECT * FROM t_order o1 JOIN t_order o2 ON o1.id = o2.parent_id",
			tableName:       "t_order",
			actualTableName: "t_order_1",
			expectedSQL:     "SELECT * FROM t_order_1 o1 JOIN t_order_1 o2 ON o1.id = o2.parent_id",
			expectError:     false,
		},
		{
			name:            "INSERT with RETURNING",
			sql:             "INSERT INTO t_order (user_id, amount) VALUES ($1, $2) RETURNING id",
			tableName:       "t_order",
			actualTableName: "t_order_2",
			expectedSQL:     "INSERT INTO t_order_2 (user_id, amount) VALUES ($1, $2) RETURNING id",
			expectError:     false,
		},
		{
			name:            "UPDATE with PostgreSQL features",
			sql:             "UPDATE t_order SET data = data || $1 WHERE id = $2",
			tableName:       "t_order",
			actualTableName: "t_order_3",
			expectedSQL:     "UPDATE t_order_3 SET data = data || $1 WHERE id = $2",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.RewriteForPostgreSQL(tt.sql, tt.tableName, tt.actualTableName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSQL, result)
			}
		})
	}
}

func TestPostgreSQLParser_ValidatePostgreSQLSQL(t *testing.T) {
	parser := NewPostgreSQLParser()

	tests := []struct {
		name        string
		sql         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid SELECT",
			sql:         "SELECT id, name FROM users WHERE id = $1",
			expectError: false,
		},
		{
			name:        "valid INSERT with RETURNING",
			sql:         "INSERT INTO users (name) VALUES ($1) RETURNING id",
			expectError: false,
		},
		{
			name:        "valid JSONB query",
			sql:         "SELECT data FROM users WHERE data @> $1",
			expectError: false,
		},
		{
			name:        "empty SQL",
			sql:         "",
			expectError: true,
			errorMsg:    "empty SQL",
		},
		{
			name:        "whitespace only",
			sql:         "   \n\t  ",
			expectError: true,
			errorMsg:    "empty SQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidatePostgreSQLSQL(tt.sql)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 基准测试
func BenchmarkPostgreSQLParser_ParsePostgreSQLSpecific(b *testing.B) {
	parser := NewPostgreSQLParser()
	sql := "SELECT id, data FROM t_order WHERE user_id = $1 AND data @> $2 ORDER BY created_at LIMIT 10 OFFSET 20"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParsePostgreSQLSpecific(sql)
	}
}

func BenchmarkPostgreSQLParser_RewriteForPostgreSQL(b *testing.B) {
	parser := NewPostgreSQLParser()
	sql := "SELECT * FROM t_order WHERE user_id = $1"
	tableName := "t_order"
	actualTableName := "t_order_1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.RewriteForPostgreSQL(sql, tableName, actualTableName)
	}
}

func BenchmarkPostgreSQLParser_parsePostgreSQLOperators(b *testing.B) {
	parser := NewPostgreSQLParser()
	sql := "SELECT * FROM users WHERE name ILIKE '%john%' AND data @> '{\"active\": true}' AND tags && ARRAY['tag1']"
	result := &EnhancedSQLStatement{
		PostgreSQLFeatures: make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.parsePostgreSQLOperators(result, sql)
	}
}