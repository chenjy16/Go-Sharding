package main

import (
	"fmt"
	"go-sharding/pkg/optimizer"
)

func main() {
	fmt.Println("=== Go-Sharding SQL优化器示例 ===")

	// 1. 创建SQL优化器
	sqlOptimizer := optimizer.NewSQLOptimizer()

	// 2. 创建优化上下文
	ctx := &optimizer.OptimizationContext{
		TableStats: map[string]*optimizer.TableStatistics{
			"users": {
				RowCount:    100000,
				DataSize:    1024 * 1024 * 100, // 100MB
				IndexCount:  3,
				LastUpdated: 1640995200, // 2022-01-01
			},
			"orders": {
				RowCount:    500000,
				DataSize:    1024 * 1024 * 500, // 500MB
				IndexCount:  5,
				LastUpdated: 1640995200,
			},
			"products": {
				RowCount:    10000,
				DataSize:    1024 * 1024 * 10, // 10MB
				IndexCount:  2,
				LastUpdated: 1640995200,
			},
		},
		IndexInfo: map[string][]string{
			"users":    {"idx_id", "idx_email", "idx_age"},
			"orders":   {"idx_id", "idx_user_id", "idx_status", "idx_created_at", "idx_total"},
			"products": {"idx_id", "idx_name"},
		},
		ShardingInfo: map[string]*optimizer.ShardingInfo{
			"users": {
				ShardingColumn: "id",
				ShardingType:   "hash",
				ShardCount:     4,
			},
			"orders": {
				ShardingColumn: "user_id",
				ShardingType:   "hash",
				ShardCount:     8,
			},
		},
	}

	// 3. 创建基于成本的优化器
	costOptimizer := optimizer.NewCostBasedOptimizer(ctx)

	// 4. 注册优化规则
	fmt.Println("\n--- 注册优化规则 ---")
	predicateRule := &optimizer.PredicatePushdownRule{}
	columnRule := &optimizer.ColumnPruningRule{}
	indexRule := &optimizer.IndexHintRule{}
	joinRule := &optimizer.JoinReorderRule{}

	sqlOptimizer.RegisterRule(predicateRule)
	sqlOptimizer.RegisterRule(columnRule)
	sqlOptimizer.RegisterRule(indexRule)
	sqlOptimizer.RegisterRule(joinRule)

	fmt.Printf("已注册 4 个优化规则\n")

	// 4. 测试谓词下推优化
	fmt.Println("\n--- 谓词下推优化 ---")
	originalSQL1 := "SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 18 AND o.status = 'completed'"
	fmt.Printf("原始SQL: %s\n", originalSQL1)

	optimizedSQL1, err := sqlOptimizer.Optimize(originalSQL1)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
	} else {
		fmt.Printf("优化后SQL: %s\n", optimizedSQL1)
	}

	// 5. 测试列裁剪优化
	fmt.Println("\n--- 列裁剪优化 ---")
	originalSQL2 := "SELECT u.name FROM (SELECT * FROM users) u WHERE u.id = 1"
	fmt.Printf("原始SQL: %s\n", originalSQL2)

	optimizedSQL2, err := sqlOptimizer.Optimize(originalSQL2)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
	} else {
		fmt.Printf("优化后SQL: %s\n", optimizedSQL2)
	}

	// 6. 测试索引提示优化
	fmt.Println("\n--- 索引提示优化 ---")
	originalSQL3 := "SELECT * FROM users WHERE email = 'user@example.com'"
	fmt.Printf("原始SQL: %s\n", originalSQL3)

	optimizedSQL3, err := sqlOptimizer.Optimize(originalSQL3)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
	} else {
		fmt.Printf("优化后SQL: %s\n", optimizedSQL3)
	}

	// 7. 测试JOIN重排序优化
	fmt.Println("\n--- JOIN重排序优化 ---")
	originalSQL4 := "SELECT * FROM orders o JOIN users u ON o.user_id = u.id JOIN products p ON o.product_id = p.id WHERE u.age > 25"
	fmt.Printf("原始SQL: %s\n", originalSQL4)

	optimizedSQL4, err := sqlOptimizer.Optimize(originalSQL4)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
	} else {
		fmt.Printf("优化后SQL: %s\n", optimizedSQL4)
	}

	// 8. 测试基于成本的优化
	fmt.Println("\n--- 基于成本的优化 ---")
	complexSQL := "SELECT u.name, COUNT(o.id) as order_count, SUM(o.total) as total_amount FROM users u LEFT JOIN orders o ON u.id = o.user_id WHERE u.created_at > '2023-01-01' GROUP BY u.id, u.name HAVING COUNT(o.id) > 5 ORDER BY total_amount DESC LIMIT 100"
	fmt.Printf("复杂SQL: %s\n", complexSQL[:80]+"...")

	// 估算原始查询成本
	originalCost, err := costOptimizer.EstimateCost(complexSQL)
	if err != nil {
		fmt.Printf("成本估算失败: %v\n", err)
	} else {
		fmt.Printf("原始查询成本: %.2f\n", originalCost)
	}

	// 使用SQL优化器优化查询
	optimizedComplexSQL, err := sqlOptimizer.Optimize(complexSQL)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
	} else {
		fmt.Printf("优化后SQL: %s\n", optimizedComplexSQL[:80]+"...")
		
		// 估算优化后查询成本
		optimizedCost, err := costOptimizer.EstimateCost(optimizedComplexSQL)
		if err != nil {
			fmt.Printf("优化后成本估算失败: %v\n", err)
		} else {
			fmt.Printf("优化后查询成本: %.2f\n", optimizedCost)
			if originalCost > 0 {
				fmt.Printf("成本降低: %.2f%%\n", (originalCost-optimizedCost)/originalCost*100)
			}
		}
	}

	// 9. 测试多种SQL类型的优化
	fmt.Println("\n--- 多种SQL类型优化测试 ---")
	testSQLs := []string{
		"SELECT * FROM users WHERE age > 18 AND status = 'active'",
		"UPDATE users SET last_login = NOW() WHERE id IN (SELECT user_id FROM sessions WHERE active = 1)",
		"DELETE FROM logs WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY)",
		"INSERT INTO user_stats (user_id, login_count) SELECT user_id, COUNT(*) FROM login_logs GROUP BY user_id",
	}

	for i, sql := range testSQLs {
		fmt.Printf("\n测试 %d:\n", i+1)
		fmt.Printf("原始: %s\n", sql)
		
		optimized, err := sqlOptimizer.Optimize(sql)
		if err != nil {
			fmt.Printf("优化失败: %v\n", err)
		} else {
			fmt.Printf("优化: %s\n", optimized)
		}
		
		// 估算成本
		cost, err := costOptimizer.EstimateCost(sql)
		if err != nil {
			fmt.Printf("成本估算失败: %v\n", err)
		} else {
			fmt.Printf("成本: %.2f\n", cost)
		}
	}

	// 10. 测试优化规则的单独应用
	fmt.Println("\n--- 单独测试优化规则 ---")
	testSQL := "SELECT u.id, u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id WHERE u.age > 21 AND o.status = 'paid'"
	fmt.Printf("测试SQL: %s\n", testSQL)

	// 测试谓词下推
	predicateResult, err := predicateRule.Apply(testSQL)
	if err == nil {
		fmt.Printf("谓词下推结果: %s\n", predicateResult)
	} else {
		fmt.Printf("谓词下推失败: %v\n", err)
	}

	// 测试列裁剪
	columnResult, err := columnRule.Apply(testSQL)
	if err == nil {
		fmt.Printf("列裁剪结果: %s\n", columnResult)
	} else {
		fmt.Printf("列裁剪失败: %v\n", err)
	}

	// 测试索引提示
	indexResult, err := indexRule.Apply(testSQL)
	if err == nil {
		fmt.Printf("索引提示结果: %s\n", indexResult)
	} else {
		fmt.Printf("索引提示失败: %v\n", err)
	}

	// 测试JOIN重排序
	joinResult, err := joinRule.Apply(testSQL)
	if err == nil {
		fmt.Printf("JOIN重排序结果: %s\n", joinResult)
	} else {
		fmt.Printf("JOIN重排序失败: %v\n", err)
	}

	// 11. 性能对比测试
	fmt.Println("\n--- 性能对比测试 ---")
	performanceTestSQLs := []string{
		"SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 1000)",
		"SELECT u.*, o.* FROM users u, orders o WHERE u.id = o.user_id AND u.age > 25",
		"SELECT COUNT(*) FROM users u JOIN orders o ON u.id = o.user_id JOIN products p ON o.product_id = p.id",
	}

	for i, sql := range performanceTestSQLs {
		fmt.Printf("\n性能测试 %d:\n", i+1)
		fmt.Printf("SQL: %s\n", sql[:60]+"...")
		
		originalCost, err := costOptimizer.EstimateCost(sql)
		if err != nil {
			fmt.Printf("原始成本估算失败: %v\n", err)
			continue
		}
		
		optimized, err := sqlOptimizer.Optimize(sql)
		if err != nil {
			fmt.Printf("优化失败: %v\n", err)
			continue
		}
		
		optimizedCost, err := costOptimizer.EstimateCost(optimized)
		if err != nil {
			fmt.Printf("优化后成本估算失败: %v\n", err)
			continue
		}
		improvement := (originalCost - optimizedCost) / originalCost * 100
		
		fmt.Printf("原始成本: %.2f\n", originalCost)
		fmt.Printf("优化成本: %.2f\n", optimizedCost)
		fmt.Printf("性能提升: %.2f%%\n", improvement)
	}

	fmt.Println("\nSQL优化器示例完成")
}