package main

import (
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/sharding"
	"log"
)

func main() {
	// 创建数据源配置
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {
			DriverName: "mysql",
			URL:        "root:password@tcp(localhost:3306)/ds_0?charset=utf8mb4&parseTime=True&loc=Local",
			MaxIdle:    10,
			MaxOpen:    100,
		},
		"ds_1": {
			DriverName: "mysql",
			URL:        "root:password@tcp(localhost:3306)/ds_1?charset=utf8mb4&parseTime=True&loc=Local",
			MaxIdle:    10,
			MaxOpen:    100,
		},
	}

	// 创建分片规则配置
	shardingRule := &config.ShardingRuleConfig{
		Tables: map[string]*config.TableRuleConfig{
			"t_order": {
				LogicTable:      "t_order",
				ActualDataNodes: "ds_${0..1}.t_order_${0..1}",
				DatabaseStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "user_id",
					Algorithm:      "ds_${user_id % 2}",
					Type:           "inline",
				},
				TableStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "order_id",
					Algorithm:      "t_order_${order_id % 2}",
					Type:           "inline",
				},
				KeyGenerator: &config.KeyGeneratorConfig{
					Column: "order_id",
					Type:   "snowflake",
				},
			},
			"t_order_item": {
				LogicTable:      "t_order_item",
				ActualDataNodes: "ds_${0..1}.t_order_item_${0..1}",
				DatabaseStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "user_id",
					Algorithm:      "ds_${user_id % 2}",
					Type:           "inline",
				},
				TableStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "order_id",
					Algorithm:      "t_order_item_${order_id % 2}",
					Type:           "inline",
				},
				KeyGenerator: &config.KeyGeneratorConfig{
					Column: "item_id",
					Type:   "snowflake",
				},
			},
		},
	}

	// 创建分片配置
	shardingConfig := &config.ShardingConfig{
		DataSources:  dataSources,
		ShardingRule: shardingRule,
	}

	// 创建分片数据源
	dataSource, err := sharding.NewShardingDataSource(shardingConfig)
	if err != nil {
		log.Fatalf("Failed to create sharding data source: %v", err)
	}
	defer dataSource.Close()

	// 获取数据库连接
	db := dataSource.DB()

	// 示例 1: 插入数据
	fmt.Println("=== 插入数据示例 ===")
	insertSQL := "INSERT INTO t_order (user_id, order_status, order_amount) VALUES (?, ?, ?)"
	result, err := db.Exec(insertSQL, 10, "PENDING", 100.50)
	if err != nil {
		log.Printf("Failed to insert order: %v", err)
	} else {
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("Inserted %d rows\n", affected)
		}
		if lastId, err := result.LastInsertId(); err == nil {
			fmt.Printf("Last insert ID: %d\n", lastId)
		}
	}

	// 示例 2: 查询单个分片
	fmt.Println("\n=== 单分片查询示例 ===")
	querySQL := "SELECT * FROM t_order WHERE user_id = ? AND order_id = ?"
	rows, err := db.Query(querySQL, 10, 1001)
	if err != nil {
		log.Printf("Failed to query orders: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("Columns: %v\n", columns)
		
		for rows.Next() {
			fmt.Println("Found matching order")
		}
	}

	// 示例 3: 跨分片查询
	fmt.Println("\n=== 跨分片查询示例 ===")
	crossShardSQL := "SELECT COUNT(*) as total FROM t_order WHERE order_status = ?"
	rows, err = db.Query(crossShardSQL, "PENDING")
	if err != nil {
		log.Printf("Failed to query cross-shard: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			fmt.Println("Cross-shard query executed successfully")
		}
	}

	// 示例 4: JOIN 查询
	fmt.Println("\n=== JOIN 查询示例 ===")
	joinSQL := `
		SELECT o.order_id, o.user_id, o.order_amount, i.item_name, i.item_price 
		FROM t_order o 
		JOIN t_order_item i ON o.order_id = i.order_id 
		WHERE o.user_id = ?
	`
	rows, err = db.Query(joinSQL, 10)
	if err != nil {
		log.Printf("Failed to execute join query: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("Join query columns: %v\n", columns)
	}

	// 示例 5: 更新数据
	fmt.Println("\n=== 更新数据示例 ===")
	updateSQL := "UPDATE t_order SET order_status = ? WHERE user_id = ? AND order_id = ?"
	result, err = db.Exec(updateSQL, "COMPLETED", 10, 1001)
	if err != nil {
		log.Printf("Failed to update order: %v", err)
	} else {
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("Updated %d rows\n", affected)
		}
	}

	// 示例 6: 删除数据
	fmt.Println("\n=== 删除数据示例 ===")
	deleteSQL := "DELETE FROM t_order WHERE user_id = ? AND order_status = ?"
	result, err = db.Exec(deleteSQL, 10, "CANCELLED")
	if err != nil {
		log.Printf("Failed to delete orders: %v", err)
	} else {
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("Deleted %d rows\n", affected)
		}
	}

	fmt.Println("\n=== 示例执行完成 ===")
}