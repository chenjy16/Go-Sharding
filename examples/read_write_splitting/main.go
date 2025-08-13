package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/readwrite"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("=== Go-Sharding 读写分离示例 ===")

	// 1. 配置数据源
	dataSources := make(map[string]*sql.DB)

	// 主库配置
	masterDB, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/master_db")
	if err != nil {
		log.Printf("警告: 无法连接主库: %v (使用模拟模式)\n", err)
		// 在实际环境中，这里应该是真实的数据库连接
		masterDB = createMockDB("master")
	}
	dataSources["master"] = masterDB

	// 从库配置
	slave1DB, err := sql.Open("mysql", "root:password@tcp(localhost:3307)/slave1_db")
	if err != nil {
		log.Printf("警告: 无法连接从库1: %v (使用模拟模式)\n", err)
		slave1DB = createMockDB("slave1")
	}
	dataSources["slave1"] = slave1DB

	slave2DB, err := sql.Open("mysql", "root:password@tcp(localhost:3308)/slave2_db")
	if err != nil {
		log.Printf("警告: 无法连接从库2: %v (使用模拟模式)\n", err)
		slave2DB = createMockDB("slave2")
	}
	dataSources["slave2"] = slave2DB

	// 2. 创建读写分离配置
	rwConfig := &config.ReadWriteSplitConfig{
		Name:                 "user_rw_split",
		MasterDataSource:     "master",
		SlaveDataSources:     []string{"slave1", "slave2"},
		LoadBalanceAlgorithm: "round_robin",
	}

	// 3. 创建读写分离器
	splitter, err := readwrite.NewReadWriteSplitter(rwConfig, dataSources)
	if err != nil {
		log.Fatalf("创建读写分离器失败: %v", err)
	}

	fmt.Println("读写分离器创建成功")
	fmt.Printf("主库: %s\n", rwConfig.MasterDataSource)
	fmt.Printf("从库: %v\n", rwConfig.SlaveDataSources)
	fmt.Printf("负载均衡算法: %s\n", rwConfig.LoadBalanceAlgorithm)

	// 4. 测试写操作路由
	fmt.Println("\n--- 测试写操作路由 ---")
	writeQueries := []string{
		"INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')",
		"UPDATE users SET email = 'alice.new@example.com' WHERE id = 1",
		"DELETE FROM users WHERE id = 2",
		"CREATE TABLE products (id INT PRIMARY KEY, name VARCHAR(100))",
		"ALTER TABLE users ADD COLUMN age INT",
		"DROP TABLE temp_table",
	}

	for _, query := range writeQueries {
		db := splitter.Route(query)
		fmt.Printf("写操作: %s\n", query[:50]+"...")
		fmt.Printf("路由到: %s\n\n", getDBName(db, dataSources))
	}

	// 5. 测试读操作路由
	fmt.Println("--- 测试读操作路由 ---")
	readQueries := []string{
		"SELECT * FROM users WHERE id = 1",
		"SELECT COUNT(*) FROM orders",
		"SELECT u.name, o.total FROM users u JOIN orders o ON u.id = o.user_id",
		"SELECT * FROM products ORDER BY name LIMIT 10",
		"SHOW TABLES",
		"DESCRIBE users",
	}

	for i, query := range readQueries {
		db := splitter.Route(query)
		fmt.Printf("读操作 %d: %s\n", i+1, query)
		fmt.Printf("路由到: %s\n\n", getDBName(db, dataSources))
	}

	// 6. 测试上下文路由
	fmt.Println("--- 测试上下文路由 ---")

	// 强制使用主库
	ctx := context.WithValue(context.Background(), "force_master", true)
	db := splitter.RouteContext(ctx, "SELECT * FROM users")
	fmt.Printf("强制主库读取: 路由到 %s\n", getDBName(db, dataSources))

	// 事务中的读操作
	txCtx := context.WithValue(context.Background(), "in_transaction", true)
	db = splitter.RouteContext(txCtx, "SELECT * FROM users")
	fmt.Printf("事务中读取: 路由到 %s\n", getDBName(db, dataSources))

	// 普通读操作
	normalCtx := context.Background()
	db = splitter.RouteContext(normalCtx, "SELECT * FROM users")
	fmt.Printf("普通读取: 路由到 %s\n", getDBName(db, dataSources))

	// 7. 测试负载均衡
	fmt.Println("\n--- 测试负载均衡 (轮询) ---")
	for i := 0; i < 6; i++ {
		db := splitter.Route("SELECT * FROM users")
		fmt.Printf("第 %d 次读取: 路由到 %s\n", i+1, getDBName(db, dataSources))
	}

	// 8. 健康检查
	fmt.Println("\n--- 健康检查 ---")
	if err := splitter.HealthCheck(); err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
	} else {
		fmt.Println("所有数据源健康状态良好")
	}

	// 9. 模拟实际业务场景
	fmt.Println("\n--- 模拟实际业务场景 ---")
	simulateBusinessScenario(splitter, dataSources)

	// 10. 清理资源
	fmt.Println("\n--- 清理资源 ---")
	if err := splitter.Close(); err != nil {
		fmt.Printf("关闭读写分离器失败: %v\n", err)
	} else {
		fmt.Println("读写分离器已关闭")
	}

	fmt.Println("\n读写分离示例完成")
}

// createMockDB 创建模拟数据库连接
func createMockDB(name string) *sql.DB {
	// 这里返回一个模拟的数据库连接
	// 在实际应用中，这应该是真实的数据库连接
	db, _ := sql.Open("mysql", "mock://"+name)
	return db
}

// getDBName 获取数据库名称
func getDBName(db *sql.DB, dataSources map[string]*sql.DB) string {
	for name, ds := range dataSources {
		if ds == db {
			return name
		}
	}
	return "unknown"
}

// simulateBusinessScenario 模拟实际业务场景
func simulateBusinessScenario(splitter *readwrite.ReadWriteSplitter, dataSources map[string]*sql.DB) {
	fmt.Println("模拟用户注册和查询场景...")

	// 用户注册 (写操作)
	registerSQL := "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)"
	db := splitter.Route(registerSQL)
	fmt.Printf("用户注册: 路由到 %s\n", getDBName(db, dataSources))

	// 等待一段时间模拟主从同步延迟
	time.Sleep(100 * time.Millisecond)

	// 用户查询 (读操作)
	querySQL := "SELECT * FROM users WHERE email = ?"
	db = splitter.Route(querySQL)
	fmt.Printf("用户查询: 路由到 %s\n", getDBName(db, dataSources))

	// 在事务中的操作
	ctx := context.WithValue(context.Background(), "in_transaction", true)
	txReadSQL := "SELECT balance FROM accounts WHERE user_id = ?"
	db = splitter.RouteContext(ctx, txReadSQL)
	fmt.Printf("事务中余额查询: 路由到 %s\n", getDBName(db, dataSources))

	txWriteSQL := "UPDATE accounts SET balance = balance - ? WHERE user_id = ?"
	db = splitter.RouteContext(ctx, txWriteSQL)
	fmt.Printf("事务中余额更新: 路由到 %s\n", getDBName(db, dataSources))

	// 强制从主库读取最新数据
	forceCtx := context.WithValue(context.Background(), "force_master", true)
	db = splitter.RouteContext(forceCtx, querySQL)
	fmt.Printf("强制主库查询: 路由到 %s\n", getDBName(db, dataSources))
}