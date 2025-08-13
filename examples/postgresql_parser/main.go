package main

import (
	"fmt"
	"go-sharding/pkg/parser"
	"log"
)

func main() {
	fmt.Println("=== PostgreSQL Parser with CockroachDB AST Demo ===")

	// 创建 PostgreSQL 解析器
	postgresParser := parser.NewPostgreSQLParser()

	// 测试用例
	testCases := []struct {
		name string
		sql  string
	}{
		{
			name: "Simple SELECT",
			sql:  "SELECT id, name FROM users WHERE age > 18",
		},
		{
			name: "SELECT with LIMIT/OFFSET",
			sql:  "SELECT * FROM products ORDER BY price DESC LIMIT 10 OFFSET 20",
		},
		{
			name: "INSERT with RETURNING",
			sql:  "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at",
		},
		{
			name: "UPDATE with RETURNING",
			sql:  "UPDATE users SET name = $1 WHERE id = $2 RETURNING name, updated_at",
		},
		{
			name: "DELETE with RETURNING",
			sql:  "DELETE FROM users WHERE id = $1 RETURNING id",
		},
		{
			name: "SELECT with PostgreSQL functions",
			sql:  "SELECT COALESCE(name, 'Unknown'), ARRAY_AGG(tag) FROM users GROUP BY name",
		},
		{
			name: "SELECT with PostgreSQL operators",
			sql:  "SELECT * FROM users WHERE name ILIKE '%john%' AND data @> '{\"active\": true}'",
		},
		{
			name: "CREATE TABLE with PostgreSQL types",
			sql:  "CREATE TABLE test (id SERIAL, data JSONB, tags TEXT[], created_at TIMESTAMP)",
		},
		{
			name: "DROP TABLE",
			sql:  "DROP TABLE IF EXISTS test_table",
		},
		{
			name: "ALTER TABLE",
			sql:  "ALTER TABLE users ADD COLUMN phone VARCHAR(20)",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)
		fmt.Printf("SQL: %s\n", tc.sql)

		// 解析 SQL
		stmt, err := postgresParser.ParsePostgreSQLSpecific(tc.sql)
		if err != nil {
			log.Printf("解析错误: %v", err)
			continue
		}

		// 显示解析结果
		fmt.Printf("语句类型: %s\n", stmt.Type)
		fmt.Printf("表名: %v\n", stmt.Tables)

		// 显示 PostgreSQL 特性
		if len(stmt.PostgreSQLFeatures) > 0 {
			fmt.Println("PostgreSQL 特性:")
			for key, value := range stmt.PostgreSQLFeatures {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}

		// 测试表名提取
		tables := postgresParser.ExtractTables(tc.sql)
		fmt.Printf("提取的表名: %v\n", tables)

		// 测试 SQL 验证
		err = postgresParser.ValidatePostgreSQLSQL(tc.sql)
		if err != nil {
			fmt.Printf("SQL 验证失败: %v\n", err)
		} else {
			fmt.Println("SQL 验证通过")
		}

		// 测试 SQL 重写
		rewrittenSQL, err := postgresParser.RewriteForPostgreSQL(tc.sql, "users", "app_users")
		if err != nil {
			fmt.Printf("SQL 重写失败: %v\n", err)
		} else if rewrittenSQL != tc.sql {
			fmt.Printf("重写后的 SQL: %s\n", rewrittenSQL)
		}
	}

	fmt.Println("\n=== PostgreSQL 方言信息 ===")
	dialect := postgresParser.GetPostgreSQLDialect()
	fmt.Printf("引用字符: %s\n", dialect.GetQuoteCharacter())

	fmt.Println("\n=== 性能对比测试 ===")
	// 测试复杂 SQL
	complexSQL := `
		SELECT 
			u.id, 
			u.name, 
			COALESCE(p.title, 'No Title') as post_title,
			ARRAY_AGG(t.name) as tags,
			ROW_NUMBER() OVER (PARTITION BY u.id ORDER BY p.created_at DESC) as rn
		FROM users u
		LEFT JOIN posts p ON u.id = p.user_id
		LEFT JOIN post_tags pt ON p.id = pt.post_id
		LEFT JOIN tags t ON pt.tag_id = t.id
		WHERE u.active = true 
			AND p.published_at IS NOT NULL
			AND p.content ILIKE '%postgresql%'
			AND p.metadata @> '{"featured": true}'
		GROUP BY u.id, u.name, p.title, p.created_at
		HAVING COUNT(t.id) > 2
		ORDER BY p.created_at DESC
		LIMIT 50 OFFSET 100
	`

	fmt.Printf("\n复杂 SQL 解析测试:\n%s\n", complexSQL)
	stmt, err := postgresParser.ParsePostgreSQLSpecific(complexSQL)
	if err != nil {
		log.Printf("复杂 SQL 解析失败: %v", err)
	} else {
		fmt.Printf("解析成功!\n")
		fmt.Printf("语句类型: %s\n", stmt.Type)
		fmt.Printf("表名: %v\n", stmt.Tables)
		if len(stmt.PostgreSQLFeatures) > 0 {
			fmt.Println("PostgreSQL 特性:")
			for key, value := range stmt.PostgreSQLFeatures {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	fmt.Println("\n=== Demo 完成 ===")
}