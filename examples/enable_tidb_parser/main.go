package main

import (
	"fmt"
	"log"
	"os"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("=== Go-Sharding TiDB 解析器启用示例 ===")
	
	// 方法1: 使用默认配置自动启用TiDB解析器
	fmt.Println("\n方法1: 使用默认配置自动启用TiDB解析器")
	if err := parser.InitializeParser(nil); err != nil {
		log.Fatalf("初始化解析器失败: %v", err)
	}
	
	// 显示解析器信息
	parser.PrintParserInfo()
	
	// 方法2: 使用自定义配置
	fmt.Println("\n方法2: 使用自定义配置")
	customConfig := &parser.InitConfig{
		EnableTiDBParser:       true,
		EnablePostgreSQLParser: false,
		FallbackToOriginal:     true,
		EnableBenchmarking:     true,
		LogParsingErrors:       true,
		AutoEnableTiDB:         true,
	}
	
	if err := parser.InitializeParser(customConfig); err != nil {
		log.Fatalf("使用自定义配置初始化解析器失败: %v", err)
	}
	
	// 方法3: 从环境变量初始化（演示）
	fmt.Println("\n方法3: 从环境变量初始化")
	fmt.Println("设置环境变量示例:")
	fmt.Println("export ENABLE_TIDB_PARSER=true")
	fmt.Println("export AUTO_ENABLE_TIDB=true")
	fmt.Println("export FALLBACK_TO_ORIGINAL=true")
	
	// 设置一些示例环境变量
	os.Setenv("ENABLE_TIDB_PARSER", "true")
	os.Setenv("AUTO_ENABLE_TIDB", "true")
	os.Setenv("ENABLE_BENCHMARKING", "true")
	
	if err := parser.InitializeParserFromEnv(); err != nil {
		log.Fatalf("从环境变量初始化解析器失败: %v", err)
	}
	
	fmt.Println("✅ 从环境变量成功初始化解析器!")
	
	// 显示当前解析器状态
	fmt.Printf("\n当前默认解析器: %s\n", parser.GetDefaultParserType())

	// 测试解析一些 SQL 语句
	fmt.Println("\n=== 测试 SQL 解析 ===")
	testSQLs := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO orders (user_id, amount) VALUES (1, 100.50)",
		"UPDATE products SET price = 99.99 WHERE category = 'electronics'",
		"DELETE FROM logs WHERE created_at < '2023-01-01'",
		"SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id WHERE u.status = 'active'",
	}

	for i, sql := range testSQLs {
		fmt.Printf("\n测试 SQL %d: %s\n", i+1, sql)

		// 使用默认解析器（现在是 TiDB 解析器）解析
		stmt, err := parser.DefaultParserFactory.Parse(sql)
		if err != nil {
			fmt.Printf("❌ 解析失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ 解析成功!\n")
		fmt.Printf("   类型: %s\n", stmt.Type)
		fmt.Printf("   表名: %v\n", stmt.Tables)
		if len(stmt.Columns) > 0 {
			fmt.Printf("   列名: %v\n", stmt.Columns)
		}
		if len(stmt.Conditions) > 0 {
			fmt.Printf("   条件数量: %d\n", len(stmt.Conditions))
		}

		// 提取表名
		tables := parser.DefaultParserFactory.ExtractTables(sql)
		fmt.Printf("   提取的表名: %v\n", tables)
	}

	// 显示最终的解析器统计信息
	fmt.Println("\n=== 最终解析器统计 ===")
	finalStats := parser.GetParserFactoryStats()
	fmt.Printf("解析器统计: %+v\n", finalStats)

	// 演示如何回退到原始解析器（可选）
	fmt.Println("\n=== 演示回退功能 ===")
	fmt.Println("如需回退到原始解析器，可以调用:")
	fmt.Println("parser.DisableTiDBParser()")

	// 取消注释下面的代码来演示回退功能
	/*
		fmt.Println("正在回退到原始解析器...")
		if err := parser.DisableTiDBParser(); err != nil {
			log.Printf("回退失败: %v", err)
		} else {
			fmt.Printf("✅ 已回退到原始解析器: %s\n", parser.GetDefaultParserType())
		}
	*/
}
