package main

import (
	"fmt"
	"log"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("🔧 PostgreSQL 配置文件解析器示例")
	fmt.Println("================================")

	// 从配置文件初始化解析器
	configFile := "config.yaml"
	fmt.Printf("📁 从配置文件加载解析器设置: %s\n", configFile)

	err := parser.InitializeParserFromConfig(configFile)
	if err != nil {
		log.Fatalf("❌ 从配置文件初始化解析器失败: %v", err)
	}

	fmt.Println("✅ 解析器已从配置文件成功初始化")

	// 验证解析器状态
	fmt.Println("\n📊 解析器状态验证")
	fmt.Println("------------------")
	parserType := parser.GetDefaultParserType()
	fmt.Printf("当前默认解析器: %s\n", parserType)

	// 打印详细信息
	parser.PrintParserInfo()

	// 测试 PostgreSQL 特有的 SQL 语句
	fmt.Println("\n🧪 PostgreSQL SQL 解析功能测试")
	fmt.Println("--------------------------------")

	testSQLs := []string{
		// 基本查询
		"SELECT * FROM users WHERE id = $1",
		// JSONB 查询
		"SELECT username, profile->>'age' as age FROM users WHERE profile @> '{\"city\": \"Beijing\"}'",
		// 数组操作
		"UPDATE users SET tags = array_append(tags, $1) WHERE user_id = $2",
		// 窗口函数
		"SELECT username, ROW_NUMBER() OVER (ORDER BY created_at DESC) as rank FROM users",
		// RETURNING 子句
		"INSERT INTO orders (user_id, amount) VALUES ($1, $2) RETURNING order_id",
		// 简单的 JOIN
		"SELECT u.username, o.amount FROM users u JOIN orders o ON u.id = o.user_id",
		// PostgreSQL 特有的数据类型
		"SELECT * FROM products WHERE price::numeric > 100.00",
	}

	for i, sql := range testSQLs {
		fmt.Printf("\n测试 %d: %s\n", i+1, sql)
		
		stmt, err := parser.DefaultParserFactory.Parse(sql)
		if err != nil {
			fmt.Printf("  ❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 解析成功: %s\n", stmt.Type)
			
			// 提取表名
			tables := parser.DefaultParserFactory.ExtractTables(sql)
			if len(tables) > 0 {
				fmt.Printf("  📋 涉及表: %v\n", tables)
			}
		}
	}

	// 显示最终统计
	fmt.Println("\n📈 解析器统计信息")
	fmt.Println("------------------")
	stats := parser.GetParserFactoryStats()
	if totalParses, ok := stats["total_parses"].(int); ok {
		fmt.Printf("总解析次数: %d\n", totalParses)
		if successfulParses, ok := stats["successful_parses"].(int); ok {
			fmt.Printf("成功解析次数: %d\n", successfulParses)
			if totalParses > 0 {
				successRate := float64(successfulParses) / float64(totalParses) * 100
				fmt.Printf("成功率: %.1f%%\n", successRate)
			}
		}
	} else {
		fmt.Printf("统计信息: %+v\n", stats)
	}

	fmt.Println("\n🎉 PostgreSQL 配置文件解析器示例完成！")
}