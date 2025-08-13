package main

import (
	"fmt"
	"log"
	"os"

	"go-sharding/pkg/parser"
)

func main() {
	fmt.Println("🔧 MySQL 和 PostgreSQL 数据库解析器配置使用方式演示")
	fmt.Println("========================================================")

	// 演示1: MySQL TiDB 解析器配置
	demonstrateMySQLTiDBParser()

	// 演示2: PostgreSQL 基础解析器配置
	demonstratePostgreSQLParser()

	// 演示3: PostgreSQL 增强解析器配置
	demonstratePostgreSQLEnhancedParser()

	// 演示4: 从配置文件初始化解析器
	demonstrateConfigFileParser()

	// 演示5: 从环境变量初始化解析器
	demonstrateEnvironmentVariableParser()

	// 演示6: 动态切换解析器
	demonstrateDynamicParserSwitching()

	// 演示7: 解析器性能对比
	demonstrateParserPerformanceComparison()

	fmt.Println("\n🎉 所有解析器配置演示完成！")
}

// 演示1: MySQL TiDB 解析器配置
func demonstrateMySQLTiDBParser() {
	fmt.Println("\n📊 演示1: MySQL TiDB 解析器配置")
	fmt.Println("================================")

	// 方法1: 使用默认配置启用TiDB解析器
	fmt.Println("\n方法1: 使用默认配置启用TiDB解析器")
	if err := parser.InitializeParser(nil); err != nil {
		log.Printf("❌ 初始化解析器失败: %v", err)
		return
	}
	fmt.Println("✅ TiDB解析器已启用")

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
		log.Printf("❌ 使用自定义配置初始化解析器失败: %v", err)
		return
	}
	fmt.Println("✅ 自定义配置TiDB解析器已启用")

	// 测试MySQL特有的SQL语句
	testMySQLSQLs := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO orders (user_id, amount) VALUES (1, 100.50)",
		"UPDATE products SET price = 99.99 WHERE category = 'electronics'",
		"SELECT * FROM users LIMIT 10, 20", // MySQL风格的LIMIT
		"SELECT * FROM `users` WHERE `name` = 'John'", // MySQL反引号
		"CREATE TABLE test (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))", // MySQL AUTO_INCREMENT
	}

	fmt.Println("\n🧪 测试MySQL特有SQL语句:")
	for i, sql := range testMySQLSQLs {
		fmt.Printf("\n测试 %d: %s\n", i+1, sql)
		stmt, err := parser.DefaultParserFactory.Parse(sql)
		if err != nil {
			fmt.Printf("  ❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 解析成功: %s\n", stmt.Type)
			tables := parser.DefaultParserFactory.ExtractTables(sql)
			if len(tables) > 0 {
				fmt.Printf("  📋 涉及表: %v\n", tables)
			}
		}
	}
}

// 演示2: PostgreSQL 基础解析器配置
func demonstratePostgreSQLParser() {
	fmt.Println("\n📊 演示2: PostgreSQL 基础解析器配置")
	fmt.Println("====================================")

	// 创建PostgreSQL解析器
	postgresParser := parser.NewPostgreSQLParser()
	fmt.Println("✅ PostgreSQL基础解析器已创建")

	// 测试PostgreSQL特有的SQL语句
	testPostgreSQLSQLs := []string{
		"SELECT * FROM users WHERE id = $1", // PostgreSQL参数占位符
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", // RETURNING子句
		"SELECT username, profile->>'age' as age FROM users WHERE profile @> '{\"city\": \"Beijing\"}'", // JSONB操作
		"UPDATE users SET tags = array_append(tags, $1) WHERE user_id = $2", // 数组操作
		"SELECT * FROM products WHERE price::numeric > 100.00", // 类型转换
		"SELECT username, ROW_NUMBER() OVER (ORDER BY created_at DESC) as rank FROM users", // 窗口函数
	}

	fmt.Println("\n🧪 测试PostgreSQL特有SQL语句:")
	for i, sql := range testPostgreSQLSQLs {
		fmt.Printf("\n测试 %d: %s\n", i+1, sql)
		stmt, err := postgresParser.ParsePostgreSQLSpecific(sql)
		if err != nil {
			fmt.Printf("  ❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 解析成功: %s\n", stmt.Type)
			if len(stmt.Tables) > 0 {
				fmt.Printf("  📋 涉及表: %v\n", stmt.Tables)
			}
			if len(stmt.PostgreSQLFeatures) > 0 {
				fmt.Printf("  🔧 PostgreSQL特性: %d项\n", len(stmt.PostgreSQLFeatures))
			}
		}
	}
}

// 演示3: PostgreSQL 增强解析器配置
func demonstratePostgreSQLEnhancedParser() {
	fmt.Println("\n📊 演示3: PostgreSQL 增强解析器配置")
	fmt.Println("=====================================")

	// 创建增强的PostgreSQL解析器
	enhancedParser := parser.NewPostgreSQLEnhancedParser()
	fmt.Println("✅ PostgreSQL增强解析器已创建")

	// 测试复杂的PostgreSQL SQL语句
	complexSQLs := []string{
		// CTE查询
		`WITH active_users AS (
			SELECT id, name FROM users WHERE active = true
		),
		user_stats AS (
			SELECT user_id, COUNT(*) as post_count FROM posts GROUP BY user_id
		)
		SELECT au.name, us.post_count FROM active_users au
		JOIN user_stats us ON au.id = us.user_id`,

		// 复杂JOIN查询
		`SELECT u.id, u.name, p.title, c.content 
		 FROM users u 
		 INNER JOIN posts p ON u.id = p.user_id 
		 LEFT JOIN comments c ON p.id = c.post_id 
		 WHERE u.active = true AND p.published_at > '2023-01-01'
		 ORDER BY p.published_at DESC`,

		// 带子查询的查询
		`SELECT u.name, 
				(SELECT COUNT(*) FROM posts WHERE user_id = u.id) as post_count,
				(SELECT AVG(rating) FROM reviews WHERE user_id = u.id) as avg_rating
		 FROM users u 
		 WHERE u.id IN (SELECT DISTINCT user_id FROM posts WHERE published = true)`,
	}

	fmt.Println("\n🧪 测试复杂PostgreSQL SQL语句:")
	for i, sql := range complexSQLs {
		fmt.Printf("\n测试 %d: 复杂查询\n", i+1)
		analysis, err := enhancedParser.AnalyzeSQL(sql)
		if err != nil {
			fmt.Printf("  ❌ 分析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 分析成功: %s\n", analysis.Type)
			fmt.Printf("  📋 涉及表: %v\n", analysis.Tables)
			fmt.Printf("  🔗 JOIN数量: %d\n", len(analysis.Joins))
			fmt.Printf("  📊 子查询数量: %d\n", len(analysis.Subqueries))
			fmt.Printf("  🔄 CTE数量: %d\n", len(analysis.CTEs))
			fmt.Printf("  🪟 窗口函数数量: %d\n", len(analysis.WindowFunctions))
			if len(analysis.Optimizations) > 0 {
				fmt.Printf("  💡 优化建议: %d条\n", len(analysis.Optimizations))
			}
		}
	}
}

// 演示4: 从配置文件初始化解析器
func demonstrateConfigFileParser() {
	fmt.Println("\n📊 演示4: 从配置文件初始化解析器")
	fmt.Println("==================================")

	// 尝试从配置文件初始化
	configFiles := []string{
		"mysql_parser_config.yaml",
		"postgresql_parser_config.yaml",
		"mixed_parser_config.yaml",
	}

	for _, configFile := range configFiles {
		fmt.Printf("\n📁 尝试从配置文件加载: %s\n", configFile)
		if err := parser.InitializeParserFromConfig(configFile); err != nil {
			fmt.Printf("  ❌ 从配置文件初始化失败: %v\n", err)
			fmt.Printf("  💡 提示: 请确保配置文件存在并格式正确\n")
		} else {
			fmt.Printf("  ✅ 从配置文件成功初始化\n")
			fmt.Printf("  📊 当前解析器: %s\n", parser.GetDefaultParserType())
		}
	}

	// 显示配置文件示例
	fmt.Println("\n📝 配置文件示例:")
	fmt.Println("MySQL解析器配置 (mysql_parser_config.yaml):")
	fmt.Println(`parser:
  enable_tidb_parser: true
  enable_postgresql_parser: false
  fallback_to_original: true
  enable_benchmarking: true
  log_parsing_errors: true`)

	fmt.Println("\nPostgreSQL解析器配置 (postgresql_parser_config.yaml):")
	fmt.Println(`parser:
  enable_tidb_parser: false
  enable_postgresql_parser: true
  fallback_to_original: true
  enable_benchmarking: true
  log_parsing_errors: true

postgresql:
  enableAdvancedFeatures: true
  enableEnhancedParser: true`)
}

// 演示5: 从环境变量初始化解析器
func demonstrateEnvironmentVariableParser() {
	fmt.Println("\n📊 演示5: 从环境变量初始化解析器")
	fmt.Println("===================================")

	// 设置环境变量示例
	fmt.Println("\n🔧 设置环境变量:")
	envVars := map[string]string{
		"ENABLE_TIDB_PARSER":       "true",
		"ENABLE_POSTGRESQL_PARSER": "false",
		"AUTO_ENABLE_TIDB":         "true",
		"FALLBACK_TO_ORIGINAL":     "true",
		"ENABLE_BENCHMARKING":      "true",
		"LOG_PARSING_ERRORS":       "true",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		fmt.Printf("  %s=%s\n", key, value)
	}

	// 从环境变量初始化
	fmt.Println("\n🚀 从环境变量初始化解析器...")
	if err := parser.InitializeParserFromEnv(); err != nil {
		fmt.Printf("❌ 从环境变量初始化失败: %v\n", err)
	} else {
		fmt.Println("✅ 从环境变量成功初始化")
		fmt.Printf("📊 当前解析器: %s\n", parser.GetDefaultParserType())
	}

	// 清理环境变量
	fmt.Println("\n🧹 清理环境变量...")
	for key := range envVars {
		os.Unsetenv(key)
	}
}

// 演示6: 动态切换解析器
func demonstrateDynamicParserSwitching() {
	fmt.Println("\n📊 演示6: 动态切换解析器")
	fmt.Println("=========================")

	// 测试SQL
	testSQL := "SELECT * FROM users WHERE id = 1"

	// 切换到TiDB解析器
	fmt.Println("\n🔄 切换到TiDB解析器...")
	if err := parser.EnableTiDBParserAsDefault(); err != nil {
		fmt.Printf("❌ 切换到TiDB解析器失败: %v\n", err)
	} else {
		fmt.Println("✅ 已切换到TiDB解析器")
		stmt, err := parser.DefaultParserFactory.Parse(testSQL)
		if err != nil {
			fmt.Printf("  ❌ TiDB解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ TiDB解析成功: %s\n", stmt.Type)
		}
	}

	// 切换到PostgreSQL解析器
	fmt.Println("\n🔄 切换到PostgreSQL解析器...")
	postgresConfig := &parser.ParserConfig{
		EnableTiDBParser:       false,
		EnablePostgreSQLParser: true,
		FallbackToOriginal:     true,
	}
	parser.DefaultParserFactory.UpdateConfig(postgresConfig)
	fmt.Println("✅ 已切换到PostgreSQL解析器")

	// 测试PostgreSQL特有语法
	postgresSQL := "SELECT * FROM users WHERE id = $1"
	stmt, err := parser.DefaultParserFactory.Parse(postgresSQL)
	if err != nil {
		fmt.Printf("  ❌ PostgreSQL解析失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ PostgreSQL解析成功: %s\n", stmt.Type)
	}
}

// 演示7: 解析器性能对比
func demonstrateParserPerformanceComparison() {
	fmt.Println("\n📊 演示7: 解析器性能对比")
	fmt.Println("=========================")

	// 启用性能基准测试
	benchmarkConfig := &parser.ParserConfig{
		EnableTiDBParser:       true,
		EnablePostgreSQLParser: true,
		FallbackToOriginal:     true,
		EnableBenchmarking:     true,
		LogParsingErrors:       true,
	}
	parser.DefaultParserFactory.UpdateConfig(benchmarkConfig)

	// 测试SQL语句
	testSQLs := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO orders (user_id, amount) VALUES (1, 100.50)",
		"UPDATE products SET price = 99.99 WHERE category = 'electronics'",
		"SELECT u.name, o.amount FROM users u JOIN orders o ON u.id = o.user_id",
	}

	fmt.Println("\n🏃‍♂️ 执行性能测试...")
	for i, sql := range testSQLs {
		fmt.Printf("\n测试 %d: %s\n", i+1, sql)
		
		// 解析SQL
		stmt, err := parser.DefaultParserFactory.Parse(sql)
		if err != nil {
			fmt.Printf("  ❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 解析成功: %s\n", stmt.Type)
		}
	}

	// 显示性能统计
	fmt.Println("\n📈 性能统计信息:")
	stats := parser.GetParserFactoryStats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 显示解析器信息
	fmt.Println("\n📊 解析器详细信息:")
	parser.PrintParserInfo()
}