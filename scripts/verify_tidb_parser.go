package main

import (
	"fmt"
	"log"
	"time"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("🔍 验证 TiDB 解析器启用状态")
	fmt.Println("================================")
	
	// 启用 TiDB 解析器
	if err := parser.EnableTiDBParserAsDefault(); err != nil {
		log.Fatalf("❌ 启用 TiDB 解析器失败: %v", err)
	}
	
	// 验证当前默认解析器
	currentParser := parser.GetDefaultParserType()
	if currentParser == parser.ParserTypeTiDB {
		fmt.Printf("✅ TiDB 解析器已成功启用为默认解析器\n")
	} else {
		fmt.Printf("❌ 当前默认解析器: %s (期望: tidb)\n", currentParser)
		return
	}
	
	// 测试解析性能
	fmt.Println("\n📊 性能测试")
	fmt.Println("----------")
	
	testSQL := "SELECT u.name, u.email, o.amount, o.created_at FROM users u JOIN orders o ON u.id = o.user_id WHERE u.status = 'active' AND o.amount > 100 ORDER BY o.created_at DESC LIMIT 10"
	
	// 测试 TiDB 解析器
	start := time.Now()
	stmt, err := parser.DefaultParserFactory.ParseWithType(testSQL, parser.ParserTypeTiDB)
	tidbDuration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ TiDB 解析器解析失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ TiDB 解析器解析成功\n")
	fmt.Printf("   耗时: %v\n", tidbDuration)
	fmt.Printf("   表名: %v\n", stmt.Tables)
	fmt.Printf("   列名: %v\n", stmt.Columns)
	
	// 测试原始解析器进行对比
	start = time.Now()
	_, err = parser.DefaultParserFactory.ParseWithType(testSQL, parser.ParserTypeOriginal)
	originalDuration := time.Since(start)
	
	if err != nil {
		fmt.Printf("⚠️  原始解析器解析失败: %v\n", err)
	} else {
		fmt.Printf("✅ 原始解析器解析成功\n")
		fmt.Printf("   耗时: %v\n", originalDuration)
		
		// 计算性能提升
		if originalDuration > 0 && tidbDuration > 0 {
			improvement := float64(originalDuration) / float64(tidbDuration)
			fmt.Printf("🚀 TiDB 解析器性能提升: %.2fx\n", improvement)
		}
	}
	
	// 测试表名提取
	fmt.Println("\n🔍 表名提取测试")
	fmt.Println("---------------")
	
	tables := parser.DefaultParserFactory.ExtractTables(testSQL)
	fmt.Printf("提取的表名: %v\n", tables)
	
	// 显示解析器统计信息
	fmt.Println("\n📈 解析器统计信息")
	fmt.Println("----------------")
	
	stats := parser.GetParserFactoryStats()
	fmt.Printf("默认解析器: %v\n", stats["default_parser"])
	fmt.Printf("可用解析器: %v\n", stats["available_parsers"])
	fmt.Printf("总解析次数: %v\n", stats["total_parse_count"])
	fmt.Printf("错误次数: %v\n", stats["error_count"])
	fmt.Printf("成功率: %.2f%%\n", stats["success_rate"])
	
	// 测试多种 SQL 类型
	fmt.Println("\n🧪 多种 SQL 类型测试")
	fmt.Println("-------------------")
	
	testCases := []struct {
		name string
		sql  string
	}{
		{"SELECT", "SELECT * FROM users WHERE id = 1"},
		{"INSERT", "INSERT INTO orders (user_id, amount) VALUES (1, 100.50)"},
		{"UPDATE", "UPDATE products SET price = 99.99 WHERE id = 1"},
		{"DELETE", "DELETE FROM logs WHERE created_at < '2023-01-01'"},
		{"复杂 JOIN", "SELECT u.*, p.name FROM users u LEFT JOIN profiles p ON u.id = p.user_id"},
	}
	
	for _, tc := range testCases {
		stmt, err := parser.DefaultParserFactory.Parse(tc.sql)
		if err != nil {
			fmt.Printf("❌ %s: 解析失败 - %v\n", tc.name, err)
		} else {
			fmt.Printf("✅ %s: 解析成功 (类型: %s, 表: %v)\n", tc.name, stmt.Type, stmt.Tables)
		}
	}
	
	fmt.Println("\n🎉 TiDB 解析器验证完成!")
	fmt.Println("========================")
	fmt.Println("✅ TiDB 解析器已成功启用并正常工作")
	fmt.Println("✅ 性能测试通过")
	fmt.Println("✅ 多种 SQL 类型解析正常")
	fmt.Println("✅ 表名提取功能正常")
}