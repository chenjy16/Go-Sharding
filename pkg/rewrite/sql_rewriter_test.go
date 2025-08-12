package rewrite

import (
	"go-sharding/pkg/routing"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSQLRewriter(t *testing.T) {
	rewriter := NewSQLRewriter()
	assert.NotNil(t, rewriter)
}

func TestSQLRewriter_Rewrite(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name         string
		ctx          *RewriteContext
		expectedLen  int
		expectError  bool
		errorMsg     string
	}{
		{
			name: "single data source single table",
			ctx: &RewriteContext{
				OriginalSQL: "SELECT * FROM t_order WHERE user_id = ?",
				LogicTables: []string{"t_order"},
				RouteResults: []*routing.RouteResult{
					{DataSource: "ds_0", Table: "t_order_0"},
				},
				Parameters: []interface{}{1},
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "single data source multiple tables",
			ctx: &RewriteContext{
				OriginalSQL: "SELECT * FROM t_order WHERE user_id = ?",
				LogicTables: []string{"t_order"},
				RouteResults: []*routing.RouteResult{
					{DataSource: "ds_0", Table: "t_order_0"},
					{DataSource: "ds_0", Table: "t_order_1"},
				},
				Parameters: []interface{}{1},
			},
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "multiple data sources",
			ctx: &RewriteContext{
				OriginalSQL: "SELECT * FROM t_order WHERE user_id = ?",
				LogicTables: []string{"t_order"},
				RouteResults: []*routing.RouteResult{
					{DataSource: "ds_0", Table: "t_order_0"},
					{DataSource: "ds_1", Table: "t_order_1"},
				},
				Parameters: []interface{}{1},
			},
			expectedLen: 2,
			expectError: false,
		},
		{
			name: "empty route results",
			ctx: &RewriteContext{
				OriginalSQL:  "SELECT * FROM t_order WHERE user_id = ?",
				LogicTables:  []string{"t_order"},
				RouteResults: []*routing.RouteResult{},
				Parameters:   []interface{}{1},
			},
			expectedLen: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := rewriter.Rewrite(tt.ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectedLen)

				for _, result := range results {
					assert.NotEmpty(t, result.SQL)
					assert.NotEmpty(t, result.DataSource)
					assert.Equal(t, tt.ctx.Parameters, result.Parameters)
				}
			}
		})
	}
}

func TestSQLRewriter_groupByDataSource(t *testing.T) {
	rewriter := NewSQLRewriter()

	routes := []*routing.RouteResult{
		{DataSource: "ds_0", Table: "t_order_0"},
		{DataSource: "ds_0", Table: "t_order_1"},
		{DataSource: "ds_1", Table: "t_order_0"},
		{DataSource: "ds_1", Table: "t_order_1"},
	}

	groups := rewriter.groupByDataSource(routes)

	assert.Len(t, groups, 2)
	assert.Len(t, groups["ds_0"], 2)
	assert.Len(t, groups["ds_1"], 2)

	// 验证分组正确性
	assert.Equal(t, "t_order_0", groups["ds_0"][0].Table)
	assert.Equal(t, "t_order_1", groups["ds_0"][1].Table)
	assert.Equal(t, "t_order_0", groups["ds_1"][0].Table)
	assert.Equal(t, "t_order_1", groups["ds_1"][1].Table)
}

func TestSQLRewriter_rewriteSQLForDataSource(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name        string
		originalSQL string
		logicTables []string
		routes      []*routing.RouteResult
		expected    string
	}{
		{
			name:        "single table replacement",
			originalSQL: "SELECT * FROM t_order WHERE user_id = ?",
			logicTables: []string{"t_order"},
			routes: []*routing.RouteResult{
				{DataSource: "ds_0", Table: "t_order_0"},
			},
			expected: "SELECT * FROM t_order_0 WHERE user_id = ?",
		},
		{
			name:        "multiple tables union",
			originalSQL: "SELECT * FROM t_order WHERE user_id = ?",
			logicTables: []string{"t_order"},
			routes: []*routing.RouteResult{
				{DataSource: "ds_0", Table: "t_order_0"},
				{DataSource: "ds_0", Table: "t_order_1"},
			},
			expected: "(SELECT * FROM t_order_0 WHERE user_id = ?) UNION ALL (SELECT * FROM t_order_1 WHERE user_id = ?)",
		},
		{
			name:        "no logic tables",
			originalSQL: "SELECT 1",
			logicTables: []string{},
			routes:      []*routing.RouteResult{},
			expected:    "SELECT 1",
		},
		{
			name:        "complex SQL with multiple table references",
			originalSQL: "SELECT t_order.id, t_order.user_id FROM t_order WHERE t_order.user_id = ?",
			logicTables: []string{"t_order"},
			routes: []*routing.RouteResult{
				{DataSource: "ds_0", Table: "t_order_0"},
			},
			expected: "SELECT t_order_0.id, t_order_0.user_id FROM t_order_0 WHERE t_order_0.user_id = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rewriter.rewriteSQLForDataSource(tt.originalSQL, tt.logicTables, tt.routes)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_getActualTablesForLogicTable(t *testing.T) {
	rewriter := NewSQLRewriter()

	routes := []*routing.RouteResult{
		{DataSource: "ds_0", Table: "t_order_0"},
		{DataSource: "ds_0", Table: "t_order_1"},
		{DataSource: "ds_1", Table: "t_user_0"},
	}

	actualTables := rewriter.getActualTablesForLogicTable("t_order", routes)
	expected := []string{"t_order_0", "t_order_1", "t_user_0"}
	assert.Equal(t, expected, actualTables)
}

func TestSQLRewriter_replaceTableName(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name        string
		sql         string
		logicTable  string
		actualTable string
		expected    string
	}{
		{
			name:        "simple replacement",
			sql:         "SELECT * FROM t_order",
			logicTable:  "t_order",
			actualTable: "t_order_0",
			expected:    "SELECT * FROM t_order_0",
		},
		{
			name:        "multiple occurrences",
			sql:         "SELECT t_order.id FROM t_order WHERE t_order.user_id = 1",
			logicTable:  "t_order",
			actualTable: "t_order_0",
			expected:    "SELECT t_order_0.id FROM t_order_0 WHERE t_order_0.user_id = 1",
		},
		{
			name:        "no replacement needed",
			sql:         "SELECT * FROM t_user",
			logicTable:  "t_order",
			actualTable: "t_order_0",
			expected:    "SELECT * FROM t_user",
		},
		{
			name:        "partial match should not be replaced",
			sql:         "SELECT * FROM t_order_detail",
			logicTable:  "t_order",
			actualTable: "t_order_0",
			expected:    "SELECT * FROM t_order_detail",
		},
		{
			name:        "case sensitive",
			sql:         "SELECT * FROM T_ORDER",
			logicTable:  "t_order",
			actualTable: "t_order_0",
			expected:    "SELECT * FROM T_ORDER", // 不应该替换，因为大小写不匹配
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.replaceTableName(tt.sql, tt.logicTable, tt.actualTable)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_generateUnionSQL(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name         string
		originalSQL  string
		logicTable   string
		actualTables []string
		expected     string
	}{
		{
			name:         "two tables union",
			originalSQL:  "SELECT * FROM t_order WHERE user_id = ?",
			logicTable:   "t_order",
			actualTables: []string{"t_order_0", "t_order_1"},
			expected:     "(SELECT * FROM t_order_0 WHERE user_id = ?) UNION ALL (SELECT * FROM t_order_1 WHERE user_id = ?)",
		},
		{
			name:         "three tables union",
			originalSQL:  "SELECT id FROM t_order",
			logicTable:   "t_order",
			actualTables: []string{"t_order_0", "t_order_1", "t_order_2"},
			expected:     "(SELECT id FROM t_order_0) UNION ALL (SELECT id FROM t_order_1) UNION ALL (SELECT id FROM t_order_2)",
		},
		{
			name:         "single table",
			originalSQL:  "SELECT * FROM t_order",
			logicTable:   "t_order",
			actualTables: []string{"t_order_0"},
			expected:     "(SELECT * FROM t_order_0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.generateUnionSQL(tt.originalSQL, tt.logicTable, tt.actualTables)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_ExtractLogicTables(t *testing.T) {
	rewriter := NewSQLRewriter()

	configuredTables := map[string]bool{
		"t_order": true,
		"t_user":  true,
		"t_item":  true,
	}

	tests := []struct {
		name     string
		sql      string
		expected []string
	}{
		{
			name:     "simple select",
			sql:      "SELECT * FROM t_order WHERE user_id = 1",
			expected: []string{"t_order"},
		},
		{
			name:     "join query",
			sql:      "SELECT * FROM t_order o JOIN t_user u ON o.user_id = u.id",
			expected: []string{"t_order", "t_user"},
		},
		{
			name:     "insert statement",
			sql:      "INSERT INTO t_order (user_id, amount) VALUES (1, 100)",
			expected: []string{"t_order"},
		},
		{
			name:     "update statement",
			sql:      "UPDATE t_order SET amount = 200 WHERE id = 1",
			expected: []string{"t_order"},
		},
		{
			name:     "delete statement",
			sql:      "DELETE FROM t_order WHERE user_id = 1",
			expected: []string{"t_order"},
		},
		{
			name:     "no configured tables",
			sql:      "SELECT * FROM t_unknown WHERE id = 1",
			expected: []string{},
		},
		{
			name:     "multiple same tables",
			sql:      "SELECT * FROM t_order o1 JOIN t_order o2 ON o1.id = o2.parent_id",
			expected: []string{"t_order"},
		},
		{
			name:     "with comments",
			sql:      "SELECT * FROM t_order -- this is a comment\nWHERE user_id = 1",
			expected: []string{"t_order"},
		},
		{
			name:     "with string literals",
			sql:      "SELECT * FROM t_order WHERE name = 't_user'",
			expected: []string{"t_order"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.ExtractLogicTables(tt.sql, configuredTables)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_extractWords(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name     string
		sql      string
		contains []string
		notContains []string
	}{
		{
			name:        "simple select",
			sql:         "SELECT id, name FROM t_order",
			contains:    []string{"id", "name", "t_order"},
			notContains: []string{"SELECT", "FROM"},
		},
		{
			name:        "with comments",
			sql:         "SELECT id FROM t_order -- comment",
			contains:    []string{"id", "t_order"},
			notContains: []string{"comment"},
		},
		{
			name:        "with string literals",
			sql:         "SELECT * FROM t_order WHERE name = 'test'",
			contains:    []string{"t_order", "name"},
			notContains: []string{"test"},
		},
		{
			name:        "complex query",
			sql:         "SELECT o.id, u.name FROM t_order o JOIN t_user u ON o.user_id = u.id",
			contains:    []string{"o", "id", "u", "name", "t_order", "t_user", "user_id"},
			notContains: []string{"SELECT", "FROM", "JOIN", "ON"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := rewriter.extractWords(tt.sql)

			for _, word := range tt.contains {
				assert.Contains(t, words, word, "Expected word '%s' to be found", word)
			}

			for _, word := range tt.notContains {
				assert.NotContains(t, words, word, "Expected word '%s' to not be found", word)
			}
		})
	}
}

func TestSQLRewriter_removeComments(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "single line comment",
			sql:      "SELECT * FROM t_order -- this is a comment",
			expected: "SELECT * FROM t_order ",
		},
		{
			name:     "block comment",
			sql:      "SELECT * FROM t_order /* this is a block comment */",
			expected: "SELECT * FROM t_order ",
		},
		{
			name:     "multiple comments",
			sql:      "SELECT * FROM t_order -- comment1\n/* comment2 */ WHERE id = 1",
			expected: "SELECT * FROM t_order \n WHERE id = 1",
		},
		{
			name:     "no comments",
			sql:      "SELECT * FROM t_order WHERE id = 1",
			expected: "SELECT * FROM t_order WHERE id = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.removeComments(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_removeStringLiterals(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "single quote string",
			sql:      "SELECT * FROM t_order WHERE name = 'test'",
			expected: "SELECT * FROM t_order WHERE name = ''",
		},
		{
			name:     "double quote string",
			sql:      `SELECT * FROM t_order WHERE name = "test"`,
			expected: `SELECT * FROM t_order WHERE name = ""`,
		},
		{
			name:     "multiple strings",
			sql:      "SELECT * FROM t_order WHERE name = 'test' AND desc = \"description\"",
			expected: "SELECT * FROM t_order WHERE name = '' AND desc = \"\"",
		},
		{
			name:     "escaped quotes",
			sql:      "SELECT * FROM t_order WHERE name = 'test\\'s'",
			expected: "SELECT * FROM t_order WHERE name = ''",
		},
		{
			name:     "no strings",
			sql:      "SELECT * FROM t_order WHERE id = 1",
			expected: "SELECT * FROM t_order WHERE id = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.removeStringLiterals(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_isSQLKeyword(t *testing.T) {
	rewriter := NewSQLRewriter()

	tests := []struct {
		name     string
		word     string
		expected bool
	}{
		{
			name:     "SELECT keyword",
			word:     "SELECT",
			expected: true,
		},
		{
			name:     "FROM keyword",
			word:     "FROM",
			expected: true,
		},
		{
			name:     "WHERE keyword",
			word:     "WHERE",
			expected: true,
		},
		{
			name:     "table name",
			word:     "t_order",
			expected: false,
		},
		{
			name:     "column name",
			word:     "user_id",
			expected: false,
		},
		{
			name:     "lowercase keyword",
			word:     "select",
			expected: false, // 函数检查大写
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.isSQLKeyword(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSQLRewriter_contains(t *testing.T) {
	rewriter := NewSQLRewriter()

	slice := []string{"a", "b", "c"}

	assert.True(t, rewriter.contains(slice, "a"))
	assert.True(t, rewriter.contains(slice, "b"))
	assert.True(t, rewriter.contains(slice, "c"))
	assert.False(t, rewriter.contains(slice, "d"))
	assert.False(t, rewriter.contains([]string{}, "a"))
}

// Benchmark tests
func BenchmarkSQLRewriter_Rewrite(b *testing.B) {
	rewriter := NewSQLRewriter()
	ctx := &RewriteContext{
		OriginalSQL: "SELECT * FROM t_order WHERE user_id = ? AND order_id = ?",
		LogicTables: []string{"t_order"},
		RouteResults: []*routing.RouteResult{
			{DataSource: "ds_0", Table: "t_order_0"},
			{DataSource: "ds_1", Table: "t_order_1"},
		},
		Parameters: []interface{}{1, 100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rewriter.Rewrite(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSQLRewriter_ExtractLogicTables(b *testing.B) {
	rewriter := NewSQLRewriter()
	sql := "SELECT o.id, o.user_id, u.name FROM t_order o JOIN t_user u ON o.user_id = u.id WHERE o.status = 'active'"
	configuredTables := map[string]bool{
		"t_order": true,
		"t_user":  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rewriter.ExtractLogicTables(sql, configuredTables)
	}
}