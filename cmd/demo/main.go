package main

import (
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/sharding"
	"log"
	"time"
)

func main() {
	fmt.Println("=== Go-Sharding 演示程序 ===")
	fmt.Println("基于 Apache ShardingSphere 设计的 Go 语言分片数据库中间件")
	fmt.Println()

	// 创建数据源配置（模拟配置，用于演示）
	dataSources := map[string]*config.DataSourceConfig{
		"ds_0": {
			DriverName: "mock",
			URL:        "mock://ds_0",
			MaxIdle:    10,
			MaxOpen:    100,
		},
		"ds_1": {
			DriverName: "mock",
			URL:        "mock://ds_1",
			MaxIdle:    10,
			MaxOpen:    100,
		},
	}

	// 创建分片规则配置
	shardingRule := &config.ShardingRuleConfig{
		Tables: map[string]*config.TableRuleConfig{
			"t_user": {
				LogicTable:      "t_user",
				ActualDataNodes: "ds_${0..1}.t_user",
				DatabaseStrategy: &config.ShardingStrategyConfig{
					ShardingColumn: "user_id",
					Algorithm:      "ds_${user_id % 2}",
					Type:           "inline",
				},
				KeyGenerator: &config.KeyGeneratorConfig{
					Column: "user_id",
					Type:   "snowflake",
				},
			},
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

	// 演示分片配置和路由逻辑
	fmt.Println("正在演示分片配置和路由逻辑...")
	demoShardingLogic(shardingConfig)
	fmt.Println("分片演示完成！")
}

func demoShardingLogic(cfg *config.ShardingConfig) {
	fmt.Println("=== 分片配置演示 ===")
	fmt.Printf("配置的数据源数量: %d\n", len(cfg.DataSources))
	for name, ds := range cfg.DataSources {
		fmt.Printf("- 数据源 %s: %s\n", name, ds.URL)
	}
	fmt.Println()

	fmt.Println("=== 分片表配置演示 ===")
	fmt.Printf("配置的分片表数量: %d\n", len(cfg.ShardingRule.Tables))
	for tableName, tableRule := range cfg.ShardingRule.Tables {
		fmt.Printf("- 逻辑表 %s:\n", tableName)
		fmt.Printf("  实际数据节点: %s\n", tableRule.ActualDataNodes)
		if tableRule.DatabaseStrategy != nil {
			fmt.Printf("  数据库分片列: %s\n", tableRule.DatabaseStrategy.ShardingColumn)
			fmt.Printf("  数据库分片算法: %s\n", tableRule.DatabaseStrategy.Algorithm)
		}
		if tableRule.TableStrategy != nil {
			fmt.Printf("  表分片列: %s\n", tableRule.TableStrategy.ShardingColumn)
			fmt.Printf("  表分片算法: %s\n", tableRule.TableStrategy.Algorithm)
		}
		if tableRule.KeyGenerator != nil {
			fmt.Printf("  主键生成器: %s (%s)\n", tableRule.KeyGenerator.Type, tableRule.KeyGenerator.Column)
		}
		fmt.Println()
	}

	fmt.Println("=== 分片路由演示 ===")
	demoRouting()
}

func demoRouting() {
	fmt.Println("演示 SQL 路由逻辑:")
	
	testCases := []struct {
		sql    string
		params []interface{}
		desc   string
	}{
		{
			sql:    "SELECT * FROM t_user WHERE user_id = ?",
			params: []interface{}{2},
			desc:   "单分片查询 - 用户ID为2",
		},
		{
			sql:    "SELECT * FROM t_user WHERE user_id = ?",
			params: []interface{}{3},
			desc:   "单分片查询 - 用户ID为3",
		},
		{
			sql:    "SELECT COUNT(*) FROM t_user",
			params: []interface{}{},
			desc:   "跨分片聚合查询",
		},
		{
			sql:    "SELECT * FROM t_order WHERE user_id = ? AND order_id = ?",
			params: []interface{}{2, 1001},
			desc:   "复合分片查询 - 用户ID为2，订单ID为1001",
		},
	}

	for i, tc := range testCases {
		fmt.Printf("%d. %s\n", i+1, tc.desc)
		fmt.Printf("   SQL: %s\n", tc.sql)
		fmt.Printf("   参数: %v\n", tc.params)
		
		// 模拟路由逻辑
		if len(tc.params) > 0 {
			if userID, ok := tc.params[0].(int); ok {
				dsIndex := userID % 2
				fmt.Printf("   路由结果: ds_%d\n", dsIndex)
				
				if len(tc.params) > 1 {
					if orderID, ok := tc.params[1].(int); ok {
						tableIndex := orderID % 2
						fmt.Printf("   表路由结果: t_order_%d\n", tableIndex)
					}
				}
			}
		} else {
			fmt.Printf("   路由结果: 所有数据源 (ds_0, ds_1)\n")
		}
		fmt.Println()
	}
}

func runDemo(db *sharding.ShardingDB) {
	// 1. 演示单分片查询
	fmt.Println("=== 1. 单分片查询演示 ===")
	fmt.Println("查询用户 ID 为 2 的用户信息（路由到 ds_0）")
	
	querySQL := "SELECT user_id, user_name, user_email, user_age FROM t_user WHERE user_id = ?"
	rows, err := db.Query(querySQL, 2)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("查询列: %v\n", columns)
		
		rowCount := 0
		for rows.Next() {
			rowCount++
			fmt.Printf("找到用户记录 #%d\n", rowCount)
		}
		fmt.Printf("共查询到 %d 条记录\n", rowCount)
	}
	fmt.Println()

	// 2. 演示跨分片聚合查询
	fmt.Println("=== 2. 跨分片聚合查询演示 ===")
	fmt.Println("统计所有用户数量（需要查询 ds_0 和 ds_1）")
	
	countSQL := "SELECT COUNT(*) as user_count FROM t_user"
	rows, err = db.Query(countSQL)
	if err != nil {
		log.Printf("聚合查询失败: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("跨分片聚合查询执行成功")
		
		for rows.Next() {
			fmt.Println("统计结果已合并")
		}
	}
	fmt.Println()

	// 3. 演示复合分片查询
	fmt.Println("=== 3. 复合分片查询演示 ===")
	fmt.Println("查询用户 ID 为 2 的订单（数据库分片 + 表分片）")
	
	orderSQL := "SELECT order_id, user_id, order_status, order_amount FROM t_order WHERE user_id = ?"
	rows, err = db.Query(orderSQL, 2)
	if err != nil {
		log.Printf("复合分片查询失败: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("查询列: %v\n", columns)
		
		rowCount := 0
		for rows.Next() {
			rowCount++
			fmt.Printf("找到订单记录 #%d\n", rowCount)
		}
		fmt.Printf("共查询到 %d 条订单记录\n", rowCount)
	}
	fmt.Println()

	// 4. 演示 JOIN 查询
	fmt.Println("=== 4. JOIN 查询演示 ===")
	fmt.Println("查询用户及其订单信息（跨表 JOIN）")
	
	joinSQL := `
		SELECT u.user_id, u.user_name, o.order_id, o.order_status, o.order_amount 
		FROM t_user u 
		JOIN t_order o ON u.user_id = o.user_id 
		WHERE u.user_id = ?
	`
	rows, err = db.Query(joinSQL, 2)
	if err != nil {
		log.Printf("JOIN 查询失败: %v", err)
	} else {
		defer rows.Close()
		columns, _ := rows.Columns()
		fmt.Printf("JOIN 查询列: %v\n", columns)
		
		rowCount := 0
		for rows.Next() {
			rowCount++
			fmt.Printf("找到 JOIN 记录 #%d\n", rowCount)
		}
		fmt.Printf("共查询到 %d 条 JOIN 记录\n", rowCount)
	}
	fmt.Println()

	// 5. 演示插入操作
	fmt.Println("=== 5. 插入操作演示 ===")
	fmt.Println("插入新用户（自动生成 ID 和路由）")
	
	insertSQL := "INSERT INTO t_user (user_name, user_email, user_age) VALUES (?, ?, ?)"
	result, err := db.Exec(insertSQL, "新用户", "newuser@example.com", 25)
	if err != nil {
		log.Printf("插入失败: %v", err)
	} else {
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("成功插入 %d 条记录\n", affected)
		}
		if lastId, err := result.LastInsertId(); err == nil {
			fmt.Printf("生成的用户 ID: %d\n", lastId)
		}
	}
	fmt.Println()

	// 6. 演示更新操作
	fmt.Println("=== 6. 更新操作演示 ===")
	fmt.Println("更新指定用户的年龄")
	
	updateSQL := "UPDATE t_user SET user_age = ? WHERE user_id = ?"
	result, err = db.Exec(updateSQL, 26, 2)
	if err != nil {
		log.Printf("更新失败: %v", err)
	} else {
		if affected, err := result.RowsAffected(); err == nil {
			fmt.Printf("成功更新 %d 条记录\n", affected)
		}
	}
	fmt.Println()

	// 7. 演示事务操作
	fmt.Println("=== 7. 事务操作演示 ===")
	fmt.Println("插入订单和订单项（模拟事务）")
	
	// 插入订单
	insertOrderSQL := "INSERT INTO t_order (user_id, order_status, order_amount) VALUES (?, ?, ?)"
	orderResult, err := db.Exec(insertOrderSQL, 2, "PENDING", 199.99)
	if err != nil {
		log.Printf("插入订单失败: %v", err)
	} else {
		fmt.Println("订单插入成功")
		
		// 插入订单项
		insertItemSQL := "INSERT INTO t_order_item (order_id, user_id, item_name, item_price, item_quantity) VALUES (?, ?, ?, ?, ?)"
		_, err = db.Exec(insertItemSQL, 1006, 2, "演示商品", 199.99, 1)
		if err != nil {
			log.Printf("插入订单项失败: %v", err)
		} else {
			fmt.Println("订单项插入成功")
		}
		
		if affected, err := orderResult.RowsAffected(); err == nil {
			fmt.Printf("事务影响行数: %d\n", affected)
		}
	}
	fmt.Println()

	// 8. 演示性能测试
	fmt.Println("=== 8. 性能测试演示 ===")
	fmt.Println("执行批量查询测试...")
	
	start := time.Now()
	for i := 0; i < 10; i++ {
		testSQL := "SELECT COUNT(*) FROM t_user WHERE user_id > ?"
		rows, err := db.Query(testSQL, i)
		if err != nil {
			log.Printf("性能测试查询失败: %v", err)
			continue
		}
		rows.Close()
	}
	elapsed := time.Since(start)
	fmt.Printf("10 次查询耗时: %v\n", elapsed)
	fmt.Printf("平均每次查询耗时: %v\n", elapsed/10)
	fmt.Println()

	fmt.Println("=== 演示完成 ===")
	fmt.Println("Go-Sharding 分片数据库中间件演示结束")
	fmt.Println("主要功能:")
	fmt.Println("✓ 数据库分片")
	fmt.Println("✓ 表分片") 
	fmt.Println("✓ 复合分片")
	fmt.Println("✓ 跨分片查询")
	fmt.Println("✓ 结果合并")
	fmt.Println("✓ JOIN 查询")
	fmt.Println("✓ 聚合函数")
	fmt.Println("✓ 事务支持")
	fmt.Println("✓ ID 自动生成")
}