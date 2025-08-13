package main

import (
	"fmt"
	"log"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("=== CockroachDB 适配器示例 ===")

	// 创建 CockroachDB 适配器
	adapter := parser.NewCockroachDBAdapter()

	// 测试 SQL 语句
	testSQLs := []string{
		"SELECT * FROM users WHERE id = 1",
		"SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id",
		"INSERT INTO products (name, price) VALUES ('laptop', 999.99)",
		"UPDATE inventory SET quantity = 100 WHERE product_id = 1",
		"DELETE FROM logs WHERE created_at < '2023-01-01'",
		"CREATE TABLE customers (id SERIAL PRIMARY KEY, name VARCHAR(100))",
		"DROP TABLE temp_data",
		"ALTER TABLE users ADD COLUMN email VARCHAR(255)",
	}

	for i, sql := range testSQLs {
		fmt.Printf("\n--- 测试 %d ---\n", i+1)
		fmt.Printf("SQL: %s\n", sql)

		// 解析 SQL
		stmt, err := adapter.ParsePostgreSQLSpecific(sql)
		if err != nil {
			log.Printf("解析失败: %v", err)
			continue
		}

		fmt.Printf("类型: %s\n", stmt.Type)
		fmt.Printf("表名: %v\n", stmt.Tables)
		if len(stmt.Columns) > 0 {
			fmt.Printf("列名: %v\n", stmt.Columns)
		}

		// 提取表名
		tables := adapter.ExtractTables(sql)
		fmt.Printf("提取的表名: %v\n", tables)

		// 验证 SQL
		err = adapter.ValidatePostgreSQLSQL(sql)
		if err == nil {
			fmt.Println("✅ SQL 验证通过")
		} else {
			fmt.Printf("❌ SQL 验证失败: %v\n", err)
		}

		// 重写 SQL（如果需要）
		rewritten, err := adapter.RewriteForPostgreSQL(sql, "users", "public.users")
		if err == nil && rewritten != sql {
			fmt.Printf("重写后: %s\n", rewritten)
		}
	}

	// 获取方言信息
	fmt.Println("\n=== 数据库方言 ===")
	dialect := adapter.GetPostgreSQLDialect()
	fmt.Printf("方言信息: %+v\n", dialect)
}