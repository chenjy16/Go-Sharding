package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLParser_determineSQLType(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		sql      string
		expected SQLType
	}{
		{"select", "SELECT * FROM users", SQLTypeSelect},
		{"insert", "INSERT INTO users VALUES (1, 'test')", SQLTypeInsert},
		{"update", "UPDATE users SET name = 'test'", SQLTypeUpdate},
		{"delete", "DELETE FROM users WHERE id = 1", SQLTypeDelete},
		{"create", "CREATE TABLE users (id INT)", SQLTypeCreate},
		{"drop", "DROP TABLE users", SQLTypeDrop},
		{"alter", "ALTER TABLE users ADD COLUMN email VARCHAR(255)", SQLTypeAlter},
		{"show", "SHOW TABLES", SQLTypeShow},
		{"other", "EXPLAIN SELECT * FROM users", SQLTypeOther},
		{"lowercase", "select * from users", SQLTypeSelect},
		{"with spaces", "  SELECT * FROM users  ", SQLTypeSelect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.determineSQLType(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_parseSelect(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedTables []string
		expectedCols   []string
	}{
		{
			name:           "simple select",
			sql:            "SELECT id, name FROM users",
			expectedTables: []string{"users"},
			expectedCols:   []string{"id", "name"},
		},
		{
			name:           "select with alias",
			sql:            "SELECT u.id, u.name FROM users u",
			expectedTables: []string{"users"},
			expectedCols:   []string{"id", "name"},
		},
		{
			name:           "select all",
			sql:            "SELECT * FROM users",
			expectedTables: []string{"users"},
			expectedCols:   []string{"*"},
		},
		{
			name:           "select with join",
			sql:            "SELECT u.id, p.title FROM users u JOIN posts p ON u.id = p.user_id",
			expectedTables: []string{"users"},
			expectedCols:   []string{"id", "title"},
		},
		{
			name:           "select multiple tables",
			sql:            "SELECT * FROM users, posts WHERE users.id = posts.user_id",
			expectedTables: []string{"users", "posts"},
			expectedCols:   []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.parseSelect(tt.sql, &SQLStatement{Type: SQLTypeSelect})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTables, stmt.Tables)
			assert.Equal(t, tt.expectedCols, stmt.Columns)
		})
	}
}

func TestSQLParser_parseInsert(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedTables []string
		expectedCols   []string
	}{
		{
			name:           "insert with columns",
			sql:            "INSERT INTO users (id, name, email) VALUES (1, 'test', 'test@example.com')",
			expectedTables: []string{"users"},
			expectedCols:   []string{"id", "name", "email"},
		},
		{
			name:           "insert without columns",
			sql:            "INSERT INTO users VALUES (1, 'test', 'test@example.com')",
			expectedTables: []string{"users"},
			expectedCols:   []string{},
		},
		{
			name:           "insert with schema",
			sql:            "INSERT INTO db.users (name) VALUES ('test')",
			expectedTables: []string{"users"},
			expectedCols:   []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.parseInsert(tt.sql, &SQLStatement{Type: SQLTypeInsert})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTables, stmt.Tables)
			assert.Equal(t, tt.expectedCols, stmt.Columns)
		})
	}
}

func TestSQLParser_parseUpdate(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedTables []string
		expectedCols   []string
	}{
		{
			name:           "simple update",
			sql:            "UPDATE users SET name = 'test', email = 'test@example.com' WHERE id = 1",
			expectedTables: []string{"users"},
			expectedCols:   []string{"name", "email"},
		},
		{
			name:           "update with schema",
			sql:            "UPDATE db.users SET name = 'test' WHERE id = 1",
			expectedTables: []string{"users"},
			expectedCols:   []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.parseUpdate(tt.sql, &SQLStatement{Type: SQLTypeUpdate})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTables, stmt.Tables)
			assert.Equal(t, tt.expectedCols, stmt.Columns)
		})
	}
}

func TestSQLParser_parseDelete(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedTables []string
	}{
		{
			name:           "simple delete",
			sql:            "DELETE FROM users WHERE id = 1",
			expectedTables: []string{"users"},
		},
		{
			name:           "delete with schema",
			sql:            "DELETE FROM db.users WHERE id = 1",
			expectedTables: []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.parseDelete(tt.sql, &SQLStatement{Type: SQLTypeDelete})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTables, stmt.Tables)
		})
	}
}

func TestSQLParser_extractJoinTables(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		sql      string
		expected []JoinTable
	}{
		{
			name: "inner join",
			sql:  "SELECT * FROM users u INNER JOIN posts p ON u.id = p.user_id",
			expected: []JoinTable{
				{Type: "INNER JOIN", Table: "posts", Condition: "u.id = p.user_id"},
			},
		},
		{
			name: "left join",
			sql:  "SELECT * FROM users u LEFT JOIN posts p ON u.id = p.user_id",
			expected: []JoinTable{
				{Type: "LEFT JOIN", Table: "posts", Condition: "u.id = p.user_id"},
			},
		},
		{
			name: "multiple joins",
			sql:  "SELECT * FROM users u INNER JOIN posts p ON u.id = p.user_id LEFT JOIN comments c ON p.id = c.post_id",
			expected: []JoinTable{
				{Type: "INNER JOIN", Table: "posts", Condition: "u.id = p.user_id"},
				{Type: "LEFT JOIN", Table: "comments", Condition: "p.id = c.post_id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractJoinTables(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_extractOrderBy(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		sql      string
		expected []OrderByClause
	}{
		{
			name: "single column asc",
			sql:  "SELECT * FROM users ORDER BY name",
			expected: []OrderByClause{
				{Column: "name", Direction: "ASC"},
			},
		},
		{
			name: "single column desc",
			sql:  "SELECT * FROM users ORDER BY name DESC",
			expected: []OrderByClause{
				{Column: "name", Direction: "DESC"},
			},
		},
		{
			name: "multiple columns",
			sql:  "SELECT * FROM users ORDER BY name ASC, created_at DESC",
			expected: []OrderByClause{
				{Column: "name", Direction: "ASC"},
				{Column: "created_at", Direction: "DESC"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractOrderBy(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_extractLimit(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		sql      string
		expected *LimitClause
	}{
		{
			name:     "limit only",
			sql:      "SELECT * FROM users LIMIT 10",
			expected: &LimitClause{Count: 10, Offset: 0},
		},
		{
			name:     "limit with offset",
			sql:      "SELECT * FROM users LIMIT 10 OFFSET 20",
			expected: &LimitClause{Count: 10, Offset: 20},
		},
		{
			name:     "mysql style limit",
			sql:      "SELECT * FROM users LIMIT 20, 10",
			expected: &LimitClause{Count: 10, Offset: 20},
		},
		{
			name:     "no limit",
			sql:      "SELECT * FROM users",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.extractLimit(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_cleanSQL(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "remove line comments",
			sql:      "SELECT * FROM users -- this is a comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "remove block comments",
			sql:      "SELECT * FROM users /* this is a comment */",
			expected: "SELECT * FROM users",
		},
		{
			name:     "normalize spaces",
			sql:      "SELECT   *    FROM     users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "complex sql",
			sql:      "SELECT   u.id,  u.name  FROM users u -- get user info\n/* WHERE clause */ WHERE u.active = 1",
			expected: "SELECT u.id, u.name FROM users u WHERE u.active = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.cleanSQL(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_Parse(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name           string
		sql            string
		expectedType   SQLType
		expectedTables []string
		expectError    bool
	}{
		{
			name:           "complex select",
			sql:            "SELECT u.id, u.name, p.title FROM users u INNER JOIN posts p ON u.id = p.user_id WHERE u.active = 1 ORDER BY u.name LIMIT 10",
			expectedType:   SQLTypeSelect,
			expectedTables: []string{"users"},
			expectError:    false,
		},
		{
			name:           "insert statement",
			sql:            "INSERT INTO users (name, email) VALUES ('test', 'test@example.com')",
			expectedType:   SQLTypeInsert,
			expectedTables: []string{"users"},
			expectError:    false,
		},
		{
			name:           "update statement",
			sql:            "UPDATE users SET name = 'updated' WHERE id = 1",
			expectedType:   SQLTypeUpdate,
			expectedTables: []string{"users"},
			expectError:    false,
		},
		{
			name:           "delete statement",
			sql:            "DELETE FROM users WHERE id = 1",
			expectedType:   SQLTypeDelete,
			expectedTables: []string{"users"},
			expectError:    false,
		},
		{
			name:        "empty sql",
			sql:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := parser.Parse(tt.sql)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stmt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stmt)
				assert.Equal(t, tt.expectedType, stmt.Type)
				assert.Equal(t, tt.expectedTables, stmt.Tables)
				assert.Equal(t, tt.sql, stmt.OriginalSQL)
			}
		})
	}
}

func TestSQLParser_IsKeyword(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		{"select keyword", "SELECT", true},
		{"from keyword", "FROM", true},
		{"where keyword", "WHERE", true},
		{"table name", "users", false},
		{"column name", "id", false},
		{"lowercase keyword", "select", true},
		{"mixed case keyword", "Select", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsKeyword(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLParser_GetTables(t *testing.T) {
	parser := NewSQLParser()

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
			sql:      "SELECT * FROM users u JOIN posts p ON u.id = p.user_id",
			expected: []string{"users"},
		},
		{
			name:     "insert query",
			sql:      "INSERT INTO users VALUES (1, 'test')",
			expected: []string{"users"},
		},
		{
			name:     "update query",
			sql:      "UPDATE users SET name = 'test'",
			expected: []string{"users"},
		},
		{
			name:     "delete query",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetTables(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkSQLParser_Parse(b *testing.B) {
	parser := NewSQLParser()
	sql := "SELECT u.id, u.name, p.title FROM users u INNER JOIN posts p ON u.id = p.user_id WHERE u.active = 1 ORDER BY u.name LIMIT 10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(sql)
	}
}

func BenchmarkSQLParser_GetTables(b *testing.B) {
	parser := NewSQLParser()
	sql := "SELECT u.id, u.name, p.title FROM users u INNER JOIN posts p ON u.id = p.user_id WHERE u.active = 1 ORDER BY u.name LIMIT 10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.GetTables(sql)
	}
}