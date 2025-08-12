package main

import (
	"fmt"
	"log"
	"os"

	"go-sharding/pkg/config"
	"go-sharding/pkg/parser"
)

// Application 应用程序结构体
type Application struct {
	Config *config.ShardingConfig
}

// NewApplication 创建新的应用程序实例
func NewApplication(configFile string) (*Application, error) {
	// 1. 加载配置文件
	cfg, err := config.LoadFromYAML(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 3. 应用解析器配置
	if err := applyParserConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to apply parser config: %w", err)
	}

	return &Application{
		Config: cfg,
	}, nil
}

// applyParserConfig 应用解析器配置
func applyParserConfig(cfg *config.ShardingConfig) error {
	parserCfg := cfg.GetParserConfig()
	
	// 创建解析器初始化配置
	initConfig := &parser.InitConfig{
		EnableTiDBParser:       parserCfg.EnableTiDBParser,
		EnablePostgreSQLParser: parserCfg.EnablePostgreSQLParser,
		FallbackToOriginal:     parserCfg.FallbackToOriginal,
		EnableBenchmarking:     parserCfg.EnableBenchmarking,
		LogParsingErrors:       parserCfg.LogParsingErrors,
		AutoEnableTiDB:         parserCfg.EnableTiDBParser,
	}

	return parser.InitializeParser(initConfig)
}

// Start 启动应用程序
func (app *Application) Start() error {
	fmt.Println("🚀 应用程序启动中...")
	
	// 显示解析器信息
	fmt.Println("\n📊 解析器配置信息:")
	parser.PrintParserInfo()

	// 模拟一些SQL操作
	fmt.Println("\n💼 模拟业务操作:")
	return app.simulateBusinessOperations()
}

// simulateBusinessOperations 模拟业务操作
func (app *Application) simulateBusinessOperations() error {
	operations := []struct {
		name string
		sql  string
	}{
		{"用户查询", "SELECT u.*, p.name as profile_name FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.status = 'active'"},
		{"订单创建", "INSERT INTO orders (user_id, product_id, quantity, price, created_at) VALUES (1001, 2001, 2, 199.99, NOW())"},
		{"库存更新", "UPDATE inventory SET quantity = quantity - 2, updated_at = NOW() WHERE product_id = 2001 AND warehouse_id = 1"},
		{"日志清理", "DELETE FROM access_logs WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)"},
		{"复杂统计", "SELECT DATE(created_at) as date, COUNT(*) as order_count, SUM(total_amount) as revenue FROM orders WHERE created_at >= '2024-01-01' GROUP BY DATE(created_at) ORDER BY date DESC"},
	}

	for i, op := range operations {
		fmt.Printf("\n%d. %s\n", i+1, op.name)
		fmt.Printf("   SQL: %s\n", op.sql)
		
		// 解析SQL
		stmt, err := parser.DefaultParserFactory.Parse(op.sql)
		if err != nil {
			fmt.Printf("   ❌ 解析失败: %v\n", err)
			continue
		}
		
		fmt.Printf("   ✅ 解析成功: %s\n", stmt.Type)
		
		// 提取表名
		tables := parser.DefaultParserFactory.ExtractTables(op.sql)
		if len(tables) > 0 {
			fmt.Printf("   📋 涉及表: %v\n", tables)
		}
	}

	return nil
}

// Stop 停止应用程序
func (app *Application) Stop() {
	fmt.Println("\n🛑 应用程序停止中...")
	
	// 显示最终统计信息
	fmt.Println("\n📈 最终解析器统计:")
	stats := parser.GetParserFactoryStats()
	fmt.Printf("总解析次数: %v\n", stats["total_parse_count"])
	fmt.Printf("成功率: %v%%\n", stats["success_rate"])
	fmt.Printf("错误次数: %v\n", stats["error_count"])
	
	fmt.Println("✅ 应用程序已停止")
}

func runApp() {
	fmt.Println("🏢 Go-Sharding 应用程序示例")
	fmt.Println("============================")

	// 检查配置文件参数
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	fmt.Printf("📁 使用配置文件: %s\n", configFile)

	// 创建应用程序实例
	app, err := NewApplication(configFile)
	if err != nil {
		log.Fatalf("❌ 创建应用程序失败: %v", err)
	}

	// 启动应用程序
	if err := app.Start(); err != nil {
		log.Fatalf("❌ 启动应用程序失败: %v", err)
	}

	// 停止应用程序
	app.Stop()

	fmt.Println("\n🎉 示例程序执行完成!")
}