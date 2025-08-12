package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-sharding/pkg/config"
	"go-sharding/pkg/sharding"
)

func main() {
	// 加载 PostgreSQL 配置
	cfg, err := config.LoadFromYAML("../postgresql_config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 PostgreSQL 分片数据源
	pgDataSource, err := sharding.NewPostgreSQLShardingDataSource(cfg)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL sharding data source: %v", err)
	}
	defer pgDataSource.Close()

	// 验证连接
	if err := pgDataSource.ValidateConnection(); err != nil {
		log.Fatalf("Failed to validate PostgreSQL connections: %v", err)
	}
	fmt.Println("PostgreSQL connections validated successfully")

	// 获取数据库连接
	db := pgDataSource.DB()

	// 演示 PostgreSQL 特有功能
	demonstratePostgreSQLFeatures(db)

	// 演示基本 CRUD 操作
	demonstrateCRUDOperations(db)

	// 演示事务操作
	demonstrateTransactionOperations(db)

	// 演示高级查询
	demonstrateAdvancedQueries(db)

	fmt.Println("PostgreSQL sharding example completed successfully!")
}

// demonstratePostgreSQLFeatures 演示 PostgreSQL 特有功能
func demonstratePostgreSQLFeatures(db *sharding.PostgreSQLDB) {
	fmt.Println("\n=== PostgreSQL 特有功能演示 ===")

	ctx := context.Background()

	// 1. 创建表（使用 PostgreSQL 特有的数据类型）
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS user_profiles (
		user_id BIGSERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		profile_data JSONB,
		tags TEXT[],
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		search_vector TSVECTOR
	)`

	_, err := db.ExecContext(ctx, createTableSQL)
	if err != nil {
		log.Printf("Failed to create user_profiles table: %v", err)
	} else {
		fmt.Println("✓ 创建包含 PostgreSQL 特有数据类型的表")
	}

	// 2. 插入数据（使用 JSONB 和数组）
	insertSQL := `
	INSERT INTO user_profiles (username, email, profile_data, tags, search_vector) 
	VALUES ($1, $2, $3, $4, to_tsvector('english', $1 || ' ' || $2))
	RETURNING user_id, created_at`

	profileData := `{"age": 25, "city": "Beijing", "interests": ["programming", "music"]}`
	tags := `{"golang", "postgresql", "sharding"}`

	// 使用基础的 Query 方法
	rows, err := db.Query(insertSQL, "john_doe", "john@example.com", profileData, tags)
	if err != nil {
		log.Printf("Failed to insert user profile: %v", err)
	} else {
		if rows.Next() {
			var userID int64
			var createdAt time.Time
			if err := rows.Scan(&userID, &createdAt); err == nil {
				fmt.Printf("✓ 插入用户数据，ID: %d, 创建时间: %v\n", userID, createdAt)
			}
		}
		rows.Close()
	}

	// 3. 使用 JSONB 查询
	jsonbQuerySQL := `
	SELECT username, profile_data->>'city' as city, profile_data->'interests' as interests
	FROM user_profiles 
	WHERE profile_data->>'age' = '25'`

	rows, err = db.QueryContext(ctx, jsonbQuerySQL)
	if err != nil {
		log.Printf("Failed to query JSONB data: %v", err)
	} else {
		fmt.Println("✓ JSONB 查询结果:")
		for rows.Next() {
			var username, city, interests string
			if err := rows.Scan(&username, &city, &interests); err == nil {
				fmt.Printf("  用户: %s, 城市: %s, 兴趣: %s\n", username, city, interests)
			}
		}
		rows.Close()
	}

	// 4. 使用数组查询
	arrayQuerySQL := `
	SELECT username, tags 
	FROM user_profiles 
	WHERE 'postgresql' = ANY(tags)`

	rows, err = db.QueryContext(ctx, arrayQuerySQL)
	if err != nil {
		log.Printf("Failed to query array data: %v", err)
	} else {
		fmt.Println("✓ 数组查询结果:")
		for rows.Next() {
			var username string
			var tags []string
			if err := rows.Scan(&username, &tags); err == nil {
				fmt.Printf("  用户: %s, 标签: %v\n", username, tags)
			}
		}
		rows.Close()
	}

	// 5. 全文搜索
	fullTextSearchSQL := `
	SELECT username, email, ts_rank(search_vector, query) as rank
	FROM user_profiles, plainto_tsquery('english', $1) query
	WHERE search_vector @@ query
	ORDER BY rank DESC`

	rows, err = db.QueryContext(ctx, fullTextSearchSQL, "john programming")
	if err != nil {
		log.Printf("Failed to perform full-text search: %v", err)
	} else {
		fmt.Println("✓ 全文搜索结果:")
		for rows.Next() {
			var username, email string
			var rank float64
			if err := rows.Scan(&username, &email, &rank); err == nil {
				fmt.Printf("  用户: %s, 邮箱: %s, 相关度: %.4f\n", username, email, rank)
			}
		}
		rows.Close()
	}
}

// demonstrateCRUDOperations 演示基本 CRUD 操作
func demonstrateCRUDOperations(db *sharding.PostgreSQLDB) {
	fmt.Println("\n=== 基本 CRUD 操作演示 ===")

	ctx := context.Background()

	// 创建用户表
	createUserTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		user_id BIGSERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		age INTEGER,
		created_at TIMESTAMP DEFAULT NOW()
	)`

	_, err := db.ExecContext(ctx, createUserTableSQL)
	if err != nil {
		log.Printf("Failed to create users table: %v", err)
		return
	}

	// 创建订单表
	createOrderTableSQL := `
	CREATE TABLE IF NOT EXISTS orders (
		order_id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		product_name VARCHAR(100) NOT NULL,
		amount DECIMAL(10,2) NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT NOW()
	)`

	_, err = db.ExecContext(ctx, createOrderTableSQL)
	if err != nil {
		log.Printf("Failed to create orders table: %v", err)
		return
	}

	// 插入用户数据
	insertUserSQL := `
	INSERT INTO users (username, email, age) 
	VALUES ($1, $2, $3) 
	RETURNING user_id`

	users := []struct {
		username string
		email    string
		age      int
	}{
		{"alice", "alice@example.com", 28},
		{"bob", "bob@example.com", 32},
		{"charlie", "charlie@example.com", 25},
		{"diana", "diana@example.com", 30},
	}

	var userIDs []int64
	for _, user := range users {
		rows, err := db.Query(insertUserSQL, user.username, user.email, user.age)
		if err != nil {
			log.Printf("Failed to insert user %s: %v", user.username, err)
			continue
		}
		if rows.Next() {
			var userID int64
			if err := rows.Scan(&userID); err == nil {
				userIDs = append(userIDs, userID)
				fmt.Printf("✓ 插入用户: %s (ID: %d)\n", user.username, userID)
			}
		}
		rows.Close()
	}

	// 插入订单数据
	insertOrderSQL := `
	INSERT INTO orders (user_id, product_name, amount, status) 
	VALUES ($1, $2, $3, $4) 
	RETURNING order_id`

	orders := []struct {
		userID      int64
		productName string
		amount      float64
		status      string
	}{
		{userIDs[0], "Laptop", 999.99, "completed"},
		{userIDs[1], "Mouse", 29.99, "pending"},
		{userIDs[0], "Keyboard", 79.99, "shipped"},
		{userIDs[2], "Monitor", 299.99, "completed"},
	}

	for _, order := range orders {
		rows, err := db.Query(insertOrderSQL, order.userID, order.productName, order.amount, order.status)
		if err != nil {
			log.Printf("Failed to insert order for user %d: %v", order.userID, err)
			continue
		}
		if rows.Next() {
			var orderID int64
			if err := rows.Scan(&orderID); err == nil {
				fmt.Printf("✓ 插入订单: %s (ID: %d, 用户ID: %d)\n", order.productName, orderID, order.userID)
			}
		}
		rows.Close()
	}

	// 查询用户数据
	queryUsersSQL := `
	SELECT user_id, username, email, age, created_at 
	FROM users 
	WHERE age >= $1 
	ORDER BY age DESC`

	rows, err := db.QueryContext(ctx, queryUsersSQL, 25)
	if err != nil {
		log.Printf("Failed to query users: %v", err)
	} else {
		fmt.Println("✓ 查询年龄 >= 25 的用户:")
		for rows.Next() {
			var userID int64
			var username, email string
			var age int
			var createdAt time.Time
			if err := rows.Scan(&userID, &username, &email, &age, &createdAt); err == nil {
				fmt.Printf("  ID: %d, 用户名: %s, 邮箱: %s, 年龄: %d\n", userID, username, email, age)
			}
		}
		rows.Close()
	}

	// 更新用户数据
	updateUserSQL := `
	UPDATE users 
	SET age = age + 1 
	WHERE username = $1 
	RETURNING user_id, age`

	rows, err = db.Query(updateUserSQL, "alice")
	if err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		if rows.Next() {
			var updatedUserID int64
			var updatedAge int
			if err := rows.Scan(&updatedUserID, &updatedAge); err == nil {
				fmt.Printf("✓ 更新用户 alice 年龄: %d (ID: %d)\n", updatedAge, updatedUserID)
			}
		}
		rows.Close()
	}

	// 删除订单数据
	deleteOrderSQL := `
	DELETE FROM orders 
	WHERE status = $1 
	RETURNING order_id`

	rows, err = db.QueryContext(ctx, deleteOrderSQL, "pending")
	if err != nil {
		log.Printf("Failed to delete orders: %v", err)
	} else {
		fmt.Println("✓ 删除待处理订单:")
		for rows.Next() {
			var orderID int64
			if err := rows.Scan(&orderID); err == nil {
				fmt.Printf("  删除订单 ID: %d\n", orderID)
			}
		}
		rows.Close()
	}
}

// demonstrateTransactionOperations 演示事务操作
func demonstrateTransactionOperations(db *sharding.PostgreSQLDB) {
	fmt.Println("\n=== 事务操作演示 ===")

	ctx := context.Background()

	// 开始事务
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// 在事务中插入用户
	insertUserSQL := `
	INSERT INTO users (username, email, age) 
	VALUES ($1, $2, $3) 
	RETURNING user_id`

	rows, err := tx.Query(insertUserSQL, "transaction_user", "tx@example.com", 35)
	if err != nil {
		log.Printf("Failed to insert user in transaction: %v", err)
		tx.Rollback()
		return
	}

	var userID int64
	if rows.Next() {
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Failed to scan user ID: %v", err)
			rows.Close()
			tx.Rollback()
			return
		}
	}
	rows.Close()

	// 在事务中插入订单
	insertOrderSQL := `
	INSERT INTO orders (user_id, product_name, amount, status) 
	VALUES ($1, $2, $3, $4) 
	RETURNING order_id`

	rows, err = tx.Query(insertOrderSQL, userID, "Transaction Product", 199.99, "pending")
	if err != nil {
		log.Printf("Failed to insert order in transaction: %v", err)
		tx.Rollback()
		return
	}

	var orderID int64
	if rows.Next() {
		if err := rows.Scan(&orderID); err != nil {
			log.Printf("Failed to scan order ID: %v", err)
			rows.Close()
			tx.Rollback()
			return
		}
	}
	rows.Close()

	// 提交事务
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Printf("✓ 事务成功提交 - 用户ID: %d, 订单ID: %d\n", userID, orderID)

	// 演示事务回滚
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin rollback transaction: %v", err)
		return
	}

	// 插入一个用户
	rows, err = tx.Query(insertUserSQL, "rollback_user", "rollback@example.com", 40)
	if err != nil {
		log.Printf("Failed to insert user for rollback demo: %v", err)
		tx.Rollback()
		return
	}

	if rows.Next() {
		if err := rows.Scan(&userID); err != nil {
			log.Printf("Failed to scan user ID for rollback: %v", err)
		}
	}
	rows.Close()

	// 故意回滚事务
	err = tx.Rollback()
	if err != nil {
		log.Printf("Failed to rollback transaction: %v", err)
		return
	}

	fmt.Println("✓ 事务成功回滚")

	// 验证回滚用户不存在
	countRows, err := db.Query("SELECT COUNT(*) FROM users WHERE username = $1", "rollback_user")
	if err != nil {
		log.Printf("Failed to verify rollback: %v", err)
	} else {
		if countRows.Next() {
			var count int
			if err := countRows.Scan(&count); err == nil {
				fmt.Printf("✓ 验证回滚: rollback_user 记录数 = %d\n", count)
			}
		}
		countRows.Close()
	}
}

// demonstrateAdvancedQueries 演示高级查询
func demonstrateAdvancedQueries(db *sharding.PostgreSQLDB) {
	fmt.Println("\n=== 高级查询演示 ===")

	ctx := context.Background()

	// 1. 窗口函数查询
	windowFunctionSQL := `
	SELECT 
		username,
		age,
		ROW_NUMBER() OVER (ORDER BY age DESC) as age_rank,
		AVG(age) OVER () as avg_age
	FROM users 
	ORDER BY age DESC`

	rows, err := db.QueryContext(ctx, windowFunctionSQL)
	if err != nil {
		log.Printf("Failed to execute window function query: %v", err)
	} else {
		fmt.Println("✓ 窗口函数查询结果:")
		for rows.Next() {
			var username string
			var age, ageRank int
			var avgAge float64
			if err := rows.Scan(&username, &age, &ageRank, &avgAge); err == nil {
				fmt.Printf("  用户: %s, 年龄: %d, 排名: %d, 平均年龄: %.1f\n", username, age, ageRank, avgAge)
			}
		}
		rows.Close()
	}

	// 2. CTE (Common Table Expression) 查询
	cteSQL := `
	WITH user_order_stats AS (
		SELECT 
			u.user_id,
			u.username,
			COUNT(o.order_id) as order_count,
			COALESCE(SUM(o.amount), 0) as total_amount
		FROM users u
		LEFT JOIN orders o ON u.user_id = o.user_id
		GROUP BY u.user_id, u.username
	)
	SELECT 
		username,
		order_count,
		total_amount,
		CASE 
			WHEN total_amount > 500 THEN 'VIP'
			WHEN total_amount > 100 THEN 'Regular'
			ELSE 'New'
		END as customer_level
	FROM user_order_stats
	ORDER BY total_amount DESC`

	rows, err = db.QueryContext(ctx, cteSQL)
	if err != nil {
		log.Printf("Failed to execute CTE query: %v", err)
	} else {
		fmt.Println("✓ CTE 查询结果:")
		for rows.Next() {
			var username, customerLevel string
			var orderCount int
			var totalAmount float64
			if err := rows.Scan(&username, &orderCount, &totalAmount, &customerLevel); err == nil {
				fmt.Printf("  用户: %s, 订单数: %d, 总金额: %.2f, 等级: %s\n", username, orderCount, totalAmount, customerLevel)
			}
		}
		rows.Close()
	}

	// 3. 聚合查询
	aggregateSQL := `
	SELECT 
		DATE_TRUNC('day', created_at) as order_date,
		COUNT(*) as order_count,
		SUM(amount) as total_amount,
		AVG(amount) as avg_amount,
		MIN(amount) as min_amount,
		MAX(amount) as max_amount
	FROM orders
	GROUP BY DATE_TRUNC('day', created_at)
	ORDER BY order_date DESC`

	rows, err = db.QueryContext(ctx, aggregateSQL)
	if err != nil {
		log.Printf("Failed to execute aggregate query: %v", err)
	} else {
		fmt.Println("✓ 聚合查询结果:")
		for rows.Next() {
			var orderDate time.Time
			var orderCount int
			var totalAmount, avgAmount, minAmount, maxAmount float64
			if err := rows.Scan(&orderDate, &orderCount, &totalAmount, &avgAmount, &minAmount, &maxAmount); err == nil {
				fmt.Printf("  日期: %s, 订单数: %d, 总金额: %.2f, 平均: %.2f, 最小: %.2f, 最大: %.2f\n",
					orderDate.Format("2006-01-02"), orderCount, totalAmount, avgAmount, minAmount, maxAmount)
			}
		}
		rows.Close()
	}

	// 4. 子查询
	subquerySQL := `
	SELECT 
		u.username,
		u.age,
		(SELECT COUNT(*) FROM orders o WHERE o.user_id = u.user_id) as order_count,
		(SELECT MAX(amount) FROM orders o WHERE o.user_id = u.user_id) as max_order_amount
	FROM users u
	WHERE u.user_id IN (
		SELECT DISTINCT user_id 
		FROM orders 
		WHERE amount > 50
	)
	ORDER BY order_count DESC`

	rows, err = db.QueryContext(ctx, subquerySQL)
	if err != nil {
		log.Printf("Failed to execute subquery: %v", err)
	} else {
		fmt.Println("✓ 子查询结果:")
		for rows.Next() {
			var username string
			var age, orderCount int
			var maxOrderAmount *float64
			if err := rows.Scan(&username, &age, &orderCount, &maxOrderAmount); err == nil {
				maxAmount := "N/A"
				if maxOrderAmount != nil {
					maxAmount = fmt.Sprintf("%.2f", *maxOrderAmount)
				}
				fmt.Printf("  用户: %s, 年龄: %d, 订单数: %d, 最大订单金额: %s\n", username, age, orderCount, maxAmount)
			}
		}
		rows.Close()
	}
}