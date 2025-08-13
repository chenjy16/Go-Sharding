package main

import (
	"fmt"
	"go-sharding/pkg/config"
	"go-sharding/pkg/readwrite"
)

func main() {
	fmt.Println("=== Go-Sharding 负载均衡示例 ===")

	// 1. 演示轮询负载均衡
	fmt.Println("\n--- 轮询负载均衡演示 ---")
	demonstrateRoundRobin()

	// 2. 演示随机负载均衡
	fmt.Println("\n--- 随机负载均衡演示 ---")
	demonstrateRandom()

	// 3. 演示加权负载均衡
	fmt.Println("\n--- 加权负载均衡演示 ---")
	demonstrateWeighted()

	// 4. 演示读写分离配置
	fmt.Println("\n--- 读写分离配置演示 ---")
	demonstrateReadWriteSplit()

	fmt.Println("\n负载均衡示例完成")
}

// demonstrateRoundRobin 演示轮询负载均衡
func demonstrateRoundRobin() {
	fmt.Println("轮询负载均衡算法特点:")
	fmt.Println("- 按顺序依次分配请求到各个从库")
	fmt.Println("- 确保负载均匀分布")
	fmt.Println("- 适用于各从库性能相近的场景")

	// 模拟轮询分配
	slaves := []string{"slave1", "slave2", "slave3"}
	index := 0

	fmt.Println("\n模拟12次请求的分配:")
	for i := 1; i <= 12; i++ {
		currentSlave := slaves[index%len(slaves)]
		fmt.Printf("请求%d -> %s\n", i, currentSlave)
		index++
	}

	// 创建轮询配置示例
	rwConfig := &config.ReadWriteSplitConfig{
		Name:                "round_robin_example",
		MasterDataSource:    "master",
		SlaveDataSources:    []string{"slave1", "slave2", "slave3"},
		LoadBalanceAlgorithm: string(readwrite.RoundRobin),
	}

	fmt.Printf("\n配置示例: %+v\n", rwConfig)
}

// demonstrateRandom 演示随机负载均衡
func demonstrateRandom() {
	fmt.Println("随机负载均衡算法特点:")
	fmt.Println("- 随机选择从库处理请求")
	fmt.Println("- 实现简单，性能较好")
	fmt.Println("- 长期来看负载相对均匀")

	// 模拟随机分配（使用固定种子以便演示）
	slaves := []string{"slave1", "slave2", "slave3"}
	pattern := []int{0, 2, 1, 0, 1, 2, 1, 0, 2, 1} // 模拟随机序列

	fmt.Println("\n模拟10次请求的随机分配:")
	for i, idx := range pattern {
		currentSlave := slaves[idx]
		fmt.Printf("请求%d -> %s\n", i+1, currentSlave)
	}

	// 创建随机配置示例
	rwConfig := &config.ReadWriteSplitConfig{
		Name:                "random_example",
		MasterDataSource:    "master",
		SlaveDataSources:    []string{"slave1", "slave2", "slave3"},
		LoadBalanceAlgorithm: string(readwrite.Random),
	}

	fmt.Printf("\n配置示例: %+v\n", rwConfig)
}

// demonstrateWeighted 演示加权负载均衡
func demonstrateWeighted() {
	fmt.Println("加权负载均衡算法特点:")
	fmt.Println("- 根据服务器性能分配不同权重")
	fmt.Println("- 性能强的服务器处理更多请求")
	fmt.Println("- 适用于服务器性能差异较大的场景")

	// 模拟加权分配
	type WeightedSlave struct {
		Name   string
		Weight int
	}

	slaves := []WeightedSlave{
		{"slave1", 5}, // 高性能服务器
		{"slave2", 3}, // 中等性能服务器
		{"slave3", 2}, // 低性能服务器
	}

	fmt.Println("\n权重配置:")
	totalWeight := 0
	for _, slave := range slaves {
		fmt.Printf("%s: 权重=%d\n", slave.Name, slave.Weight)
		totalWeight += slave.Weight
	}

	fmt.Printf("\n总权重: %d\n", totalWeight)
	fmt.Println("\n理论分配比例:")
	for _, slave := range slaves {
		percentage := float64(slave.Weight) / float64(totalWeight) * 100
		fmt.Printf("%s: %.1f%%\n", slave.Name, percentage)
	}

	// 模拟加权轮询分配
	fmt.Println("\n模拟10次请求的加权分配:")
	weightedPattern := []string{
		"slave1", "slave1", "slave2", "slave1", "slave3",
		"slave1", "slave2", "slave1", "slave2", "slave3",
	}

	for i, slave := range weightedPattern {
		fmt.Printf("请求%d -> %s\n", i+1, slave)
	}

	// 创建加权配置示例
	rwConfig := &config.ReadWriteSplitConfig{
		Name:                "weighted_example",
		MasterDataSource:    "master",
		SlaveDataSources:    []string{"slave1", "slave2", "slave3"},
		LoadBalanceAlgorithm: string(readwrite.Weight),
	}

	fmt.Printf("\n配置示例: %+v\n", rwConfig)
}

// demonstrateReadWriteSplit 演示读写分离配置
func demonstrateReadWriteSplit() {
	fmt.Println("读写分离核心概念:")
	fmt.Println("- 写操作路由到主库 (Master)")
	fmt.Println("- 读操作路由到从库 (Slave)")
	fmt.Println("- 通过负载均衡算法分配读请求")

	// 模拟SQL路由
	sqlExamples := []struct {
		SQL  string
		Type string
	}{
		{"SELECT * FROM users WHERE id = 1", "读操作 -> 从库"},
		{"INSERT INTO users (name) VALUES ('张三')", "写操作 -> 主库"},
		{"UPDATE users SET name = '李四' WHERE id = 1", "写操作 -> 主库"},
		{"SELECT COUNT(*) FROM orders", "读操作 -> 从库"},
		{"DELETE FROM users WHERE id = 1", "写操作 -> 主库"},
	}

	fmt.Println("\nSQL路由示例:")
	for _, example := range sqlExamples {
		fmt.Printf("SQL: %s\n", example.SQL)
		fmt.Printf("路由: %s\n\n", example.Type)
	}

	// 完整配置示例
	fmt.Println("完整的读写分离配置示例:")
	completeConfig := &config.ReadWriteSplitConfig{
		Name:                "production_rw_split",
		MasterDataSource:    "master_db",
		SlaveDataSources:    []string{"slave_db_1", "slave_db_2", "slave_db_3"},
		LoadBalanceAlgorithm: string(readwrite.RoundRobin),
	}

	fmt.Printf("%+v\n", completeConfig)

	fmt.Println("\n使用建议:")
	fmt.Println("1. 轮询算法: 适用于从库性能相近的场景")
	fmt.Println("2. 随机算法: 实现简单，适用于大多数场景")
	fmt.Println("3. 加权算法: 适用于从库性能差异较大的场景")
	fmt.Println("4. 配置健康检查以确保高可用性")
	fmt.Println("5. 监控各从库的负载分布情况")
}