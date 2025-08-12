package main

import (
	"context"
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/sharding"
	"log"
)

func main() {
	// 创建增强的分片配置
	cfg := &config.ShardingConfig{
		DataSources: map[string]*config.DataSourceConfig{
			"master_ds_0": {
				DriverName: "mysql",
				URL:        "root:password@tcp(localhost:3306)/sharding_db_0",
				MaxIdle:    10,
				MaxOpen:    100,
			},
			"slave_ds_0_1": {
				DriverName: "mysql",
				URL:        "root:password@tcp(localhost:3307)/sharding_db_0",
				MaxIdle:    10,
				MaxOpen:    100,
			},
			"slave_ds_0_2": {
				DriverName: "mysql",
				URL:        "root:password@tcp(localhost:3308)/sharding_db_0",
				MaxIdle:    10,
				MaxOpen:    100,
			},
			"master_ds_1": {
				DriverName: "mysql",
				URL:        "root:password@tcp(localhost:3309)/sharding_db_1",
				MaxIdle:    10,
				MaxOpen:    100,
			},
			"slave_ds_1_1": {
				DriverName: "mysql",
				URL:        "root:password@tcp(localhost:3310)/sharding_db_1",
				MaxIdle:    10,
				MaxOpen:    100,
			},
		},
		ReadWriteSplits: map[string]*config.ReadWriteSplitConfig{
			"rw_ds_0": {
				MasterDataSource: "master_ds_0",
				SlaveDataSources: []string{"slave_ds_0_1", "slave_ds_0_2"},
				LoadBalanceAlgorithm: "round_robin",
			},
			"rw_ds_1": {
				MasterDataSource: "master_ds_1",
				SlaveDataSources: []string{"slave_ds_1_1"},
				LoadBalanceAlgorithm: "random",
			},
		},
		ShardingRule: &config.ShardingRuleConfig{
			Tables: map[string]*config.TableRuleConfig{
				"t_order": {
					ActualDataNodes: "rw_ds_${0..1}.t_order_${0..1}",
					DatabaseStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "user_id",
						Algorithm:      "rw_ds_${user_id % 2}",
						Type:           "inline",
					},
					TableStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "order_id",
						Algorithm:      "t_order_${order_id % 2}",
						Type:           "inline",
					},
				},
				"t_user": {
					ActualDataNodes: "rw_ds_${0..1}.t_user",
					DatabaseStrategy: &config.ShardingStrategyConfig{
						ShardingColumn: "user_id",
						Algorithm:      "rw_ds_${user_id % 2}",
						Type:           "inline",
					},
				},
			},
		},
	}

	// 创建增强的分片数据库实例
	db, err := sharding.NewEnhancedShardingDB(cfg)
	if err != nil {
		log.Fatalf("Failed to create enhanced sharding database: %v", err)
	}
	defer db.Close()

	// 健康检查
	if err := db.HealthCheck(); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		log.Println("Health check passed")
	}

	ctx := context.Background()

	// 示例 1: 插入数据（写操作，会路由到主库）
	fmt.Println("=== 示例 1: 插入数据 ===")
	insertSQL := "INSERT INTO t_order (order_id, user_id, amount, status) VALUES (?, ?, ?, ?)"
	result, err := db.ExecContext(ctx, insertSQL, 1001, 1, 99.99, "PENDING")
	if err != nil {
		log.Printf("Insert failed: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		lastInsertId, _ := result.LastInsertId()
		fmt.Printf("Insert successful: rows affected = %d, last insert id = %d\n", rowsAffected, lastInsertId)
	}

	// 示例 2: 查询数据（读操作，会路由到从库）
	fmt.Println("\n=== 示例 2: 查询数据 ===")
	querySQL := "SELECT order_id, user_id, amount, status FROM t_order WHERE user_id = ?"
	rows, err := db.QueryContext(ctx, querySQL, 1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		defer rows.Close()
		
		columns, _ := rows.Columns()
		fmt.Printf("Columns: %v\n", columns)
		
		for rows.Next() {
			var orderId, userId int
			var amount float64
			var status string
			
			if err := rows.Scan(&orderId, &userId, &amount, &status); err != nil {
				log.Printf("Scan failed: %v", err)
				continue
			}
			
			fmt.Printf("Order: ID=%d, UserID=%d, Amount=%.2f, Status=%s\n", 
				orderId, userId, amount, status)
		}
	}

	// 示例 3: 复杂查询（JOIN 查询）
	fmt.Println("\n=== 示例 3: 复杂查询 ===")
	complexSQL := `
		SELECT o.order_id, o.amount, u.username 
		FROM t_order o 
		JOIN t_user u ON o.user_id = u.user_id 
		WHERE o.user_id = ? AND o.status = ?
		ORDER BY o.order_id DESC 
		LIMIT 10
	`
	rows, err = db.QueryContext(ctx, complexSQL, 1, "PENDING")
	if err != nil {
		log.Printf("Complex query failed: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("Complex query executed successfully")
		
		for rows.Next() {
			var orderId int
			var amount float64
			var username string
			
			if err := rows.Scan(&orderId, &amount, &username); err != nil {
				log.Printf("Scan failed: %v", err)
				continue
			}
			
			fmt.Printf("Result: OrderID=%d, Amount=%.2f, Username=%s\n", 
				orderId, amount, username)
		}
	}

	// 示例 4: 更新数据（写操作，会路由到主库）
	fmt.Println("\n=== 示例 4: 更新数据 ===")
	updateSQL := "UPDATE t_order SET status = ? WHERE order_id = ? AND user_id = ?"
	result, err = db.ExecContext(ctx, updateSQL, "COMPLETED", 1001, 1)
	if err != nil {
		log.Printf("Update failed: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("Update successful: rows affected = %d\n", rowsAffected)
	}

	// 示例 5: 删除数据（写操作，会路由到主库）
	fmt.Println("\n=== 示例 5: 删除数据 ===")
	deleteSQL := "DELETE FROM t_order WHERE order_id = ? AND user_id = ?"
	result, err = db.ExecContext(ctx, deleteSQL, 1001, 1)
	if err != nil {
		log.Printf("Delete failed: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("Delete successful: rows affected = %d\n", rowsAffected)
	}

	// 示例 6: 事务操作（强制使用主库）
	fmt.Println("\n=== 示例 6: 事务操作 ===")
	// 注意：这里只是演示，实际的事务支持需要更复杂的实现
	transactionSQL := "SELECT COUNT(*) FROM t_order WHERE user_id = ? FOR UPDATE"
	rows, err = db.QueryContext(ctx, transactionSQL, 1)
	if err != nil {
		log.Printf("Transaction query failed: %v", err)
	} else {
		defer rows.Close()
		
		if rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				log.Printf("Scan failed: %v", err)
			} else {
				fmt.Printf("Transaction query result: count = %d\n", count)
			}
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}

// 辅助函数：演示如何使用强制主库的上下文
func demonstrateForceMainContext() context.Context {
	// 创建一个强制使用主库的上下文
	ctx := context.Background()
	return context.WithValue(ctx, "force_master", true)
}

// 辅助函数：演示如何使用事务上下文
func demonstrateTransactionContext() context.Context {
	// 创建一个事务上下文
	ctx := context.Background()
	return context.WithValue(ctx, "in_transaction", true)
}