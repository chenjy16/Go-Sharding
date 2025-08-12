package main

import (
	"fmt"
	"log"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("🔧 从配置文件初始化解析器示例")
	fmt.Println("================================")

	// 方法1: 直接从配置文件初始化解析器
	configFile := "config.yaml"
	fmt.Printf("📁 从配置文件加载解析器设置: %s\n", configFile)
	
	if err := parser.InitializeParserFromConfig(configFile); err != nil {
		log.Fatalf("❌ 从配置文件初始化解析器失败: %v", err)
	}

	fmt.Println("✅ 解析器已从配置文件成功初始化")

	// 验证解析器状态
	fmt.Println("\n📊 解析器状态验证")
	fmt.Println("------------------")
	
	// 获取当前默认解析器类型
	parserType := parser.GetDefaultParserType()
	fmt.Printf("当前默认解析器: %s\n", parserType)

	// 打印详细的解析器信息
	parser.PrintParserInfo()

	// 测试解析功能
	fmt.Println("\n🧪 解析功能测试")
	fmt.Println("----------------")
	
	testSQLs := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO orders (user_id, amount) VALUES (1, 100.50)",
		"UPDATE products SET price = 99.99 WHERE category = 'electronics'",
		"DELETE FROM logs WHERE created_at < '2023-01-01'",
	}

	for i, sql := range testSQLs {
		fmt.Printf("测试 %d: %s\n", i+1, sql)
		
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
		fmt.Println()
	}

	// 显示最终统计信息
	fmt.Println("📈 解析器统计信息")
	fmt.Println("------------------")
	stats := parser.GetParserFactoryStats()
	fmt.Printf("统计信息: %+v\n", stats)

	fmt.Println("\n✨ 配置文件解析器初始化完成!")
}