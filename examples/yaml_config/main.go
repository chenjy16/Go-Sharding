package main

import (
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/sharding"
	"log"
)

func main() {
	// 从 YAML 文件加载配置
	shardingConfig, err := config.LoadFromYAML("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config from YAML: %v", err)
	}

	// 验证配置
	if err := shardingConfig.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	fmt.Println("=== 配置信息 ===")
	fmt.Printf("数据源数量: %d\n", len(shardingConfig.DataSources))
	for name := range shardingConfig.DataSources {
		fmt.Printf("- %s\n", name)
	}

	if shardingConfig.ShardingRule != nil {
		fmt.Printf("分片表数量: %d\n", len(shardingConfig.ShardingRule.Tables))
		for tableName := range shardingConfig.ShardingRule.Tables {
			fmt.Printf("- %s\n", tableName)
		}
	}

	fmt.Printf("读写分离配置数量: %d\n", len(shardingConfig.ReadWriteSplits))
	for name := range shardingConfig.ReadWriteSplits {
		fmt.Printf("- %s\n", name)
	}

	// 创建分片数据源
	dataSource, err := sharding.NewShardingDataSource(shardingConfig)
	if err != nil {
		log.Fatalf("Failed to create sharding data source: %v", err)
	}
	defer dataSource.Close()

	// 获取数据库连接
	db := dataSource.DB()

	fmt.Println("\n=== 测试数据库连接 ===")

	// 测试查询
	testSQL := "SELECT 1 as test_value"
	rows, err := db.Query(testSQL)
	if err != nil {
		log.Printf("Failed to execute test query: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("数据库连接测试成功")
	}

	// 测试分片表查询
	fmt.Println("\n=== 测试分片表操作 ===")
	
	// 插入测试数据
	insertSQL := "INSERT INTO t_user (user_name, user_email, user_age) VALUES (?, ?, ?)"
	result, err := db.Exec(insertSQL, "张三", "zhangsan@example.com", 25)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
	} else {
		fmt.Println("用户数据插入成功")
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("影响行数: %d\n", affected)
		}
	}

	// 查询测试数据
	querySQL := "SELECT * FROM t_user WHERE user_name = ?"
	rows, err = db.Query(querySQL, "张三")
	if err != nil {
		log.Printf("Failed to query user: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("查询结果列: %v\n", columns)
		
		rowCount := 0
		for rows.Next() {
			rowCount++
		}
		fmt.Printf("查询到 %d 行数据\n", rowCount)
	}

	// 测试跨分片聚合查询
	fmt.Println("\n=== 测试跨分片聚合查询 ===")
	aggregateSQL := "SELECT COUNT(*) as user_count FROM t_user"
	rows, err = db.Query(aggregateSQL)
	if err != nil {
		log.Printf("Failed to execute aggregate query: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("跨分片聚合查询执行成功")
	}

	// 测试事务（简化版本）
	fmt.Println("\n=== 测试事务操作 ===")
	
	// 开始事务
	insertOrder := "INSERT INTO t_order (user_id, order_status, order_amount) VALUES (?, ?, ?)"
	insertItem := "INSERT INTO t_order_item (order_id, item_name, item_price, item_quantity) VALUES (?, ?, ?, ?)"
	
	// 插入订单
	orderResult, err := db.Exec(insertOrder, 100, "PENDING", 299.99)
	if err != nil {
		log.Printf("Failed to insert order: %v", err)
	} else {
		fmt.Println("订单插入成功")
		
		// 插入订单项
		_, err = db.Exec(insertItem, 1001, "商品A", 99.99, 3)
		if err != nil {
			log.Printf("Failed to insert order item: %v", err)
		} else {
			fmt.Println("订单项插入成功")
		}
		
		if affected, err := orderResult.RowsAffected(); err == nil {
			fmt.Printf("订单影响行数: %d\n", affected)
		}
	}

	fmt.Println("\n=== YAML 配置示例执行完成 ===")
}