package main

import (
	"fmt"
	"go-sharding/pkg/id"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Go-Sharding ID生成器示例 ===")

	// 1. 创建生成器工厂
	factory := id.NewGeneratorFactory()

	// 2. 测试雪花算法ID生成器
	fmt.Println("\n--- 雪花算法ID生成器 ---")
	snowflakeGen, err := factory.CreateGenerator("snowflake", map[string]interface{}{
		"node_id": int64(1),
	})
	if err != nil {
		fmt.Printf("创建雪花算法生成器失败: %v\n", err)
	} else {
		fmt.Println("雪花算法生成器创建成功")
		
		// 生成多个ID
		fmt.Println("生成的雪花算法ID:")
		for i := 0; i < 10; i++ {
			id, err := snowflakeGen.NextID()
			if err != nil {
				fmt.Printf("生成ID失败: %v\n", err)
			} else {
				fmt.Printf("ID %d: %d\n", i+1, id)
			}
		}
	}

	// 3. 测试UUID生成器
	fmt.Println("\n--- UUID生成器 ---")
	uuidGen, err := factory.CreateGenerator("uuid", nil)
	if err != nil {
		fmt.Printf("创建UUID生成器失败: %v\n", err)
	} else {
		fmt.Println("UUID生成器创建成功")
		
		// 生成多个UUID
		fmt.Println("生成的UUID:")
		for i := 0; i < 5; i++ {
			id, err := uuidGen.NextID()
			if err != nil {
				fmt.Printf("生成UUID失败: %v\n", err)
			} else {
				fmt.Printf("UUID %d: %d\n", i+1, id)
			}
		}
	}

	// 4. 测试自增ID生成器
	fmt.Println("\n--- 自增ID生成器 ---")
	incrementGen, err := factory.CreateGenerator("increment", map[string]interface{}{
		"start": int64(1000),
		"step":  int64(1),
	})
	if err != nil {
		fmt.Printf("创建自增生成器失败: %v\n", err)
	} else {
		fmt.Println("自增生成器创建成功")
		
		// 生成多个自增ID
		fmt.Println("生成的自增ID:")
		for i := 0; i < 10; i++ {
			id, err := incrementGen.NextID()
			if err != nil {
				fmt.Printf("生成自增ID失败: %v\n", err)
			} else {
				fmt.Printf("自增ID %d: %d\n", i+1, id)
			}
		}
	}

	// 5. 并发测试
	fmt.Println("\n--- 并发测试 ---")
	testConcurrency(snowflakeGen, "雪花算法", 100, 10)
	testConcurrency(incrementGen, "自增算法", 100, 10)

	// 6. 性能测试
	fmt.Println("\n--- 性能测试 ---")
	testPerformance(snowflakeGen, "雪花算法", 10000)
	testPerformance(uuidGen, "UUID", 10000)
	testPerformance(incrementGen, "自增算法", 10000)

	// 7. 测试不同节点的雪花算法生成器
	fmt.Println("\n--- 多节点雪花算法测试 ---")
	testMultiNodeSnowflake(factory)

	// 8. 测试ID唯一性
	fmt.Println("\n--- ID唯一性测试 ---")
	testUniqueness(snowflakeGen, "雪花算法", 1000)
	testUniqueness(incrementGen, "自增算法", 1000)

	fmt.Println("\nID生成器示例完成")
}

// testConcurrency 测试并发生成ID
func testConcurrency(gen id.Generator, name string, goroutines, idsPerGoroutine int) {
	fmt.Printf("%s并发测试: %d个协程，每个生成%d个ID\n", name, goroutines, idsPerGoroutine)
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	ids := make([]int64, 0, goroutines*idsPerGoroutine)
	errorCount := 0
	
	start := time.Now()
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < idsPerGoroutine; j++ {
				id, err := gen.NextID()
				
				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					ids = append(ids, id)
				}
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	fmt.Printf("并发测试完成: 生成%d个ID，耗时%v，错误%d个\n", len(ids), duration, errorCount)
	
	// 检查唯一性
	uniqueIDs := make(map[int64]bool)
	duplicates := 0
	for _, id := range ids {
		if uniqueIDs[id] {
			duplicates++
		} else {
			uniqueIDs[id] = true
		}
	}
	
	if duplicates > 0 {
		fmt.Printf("警告: 发现%d个重复ID\n", duplicates)
	} else {
		fmt.Println("所有ID都是唯一的")
	}
}

// testPerformance 测试性能
func testPerformance(gen id.Generator, name string, count int) {
	fmt.Printf("%s性能测试: 生成%d个ID\n", name, count)
	
	start := time.Now()
	errorCount := 0
	
	for i := 0; i < count; i++ {
		_, err := gen.NextID()
		if err != nil {
			errorCount++
		}
	}
	
	duration := time.Since(start)
	qps := float64(count-errorCount) / duration.Seconds()
	
	fmt.Printf("性能测试完成: 耗时%v，QPS: %.2f，错误%d个\n", duration, qps, errorCount)
}

// testMultiNodeSnowflake 测试多节点雪花算法
func testMultiNodeSnowflake(factory *id.GeneratorFactory) {
	fmt.Println("创建多个节点的雪花算法生成器...")
	
	generators := make([]id.Generator, 0, 5)
	
	// 创建5个不同节点的生成器
	for i := 0; i < 5; i++ {
		gen, err := factory.CreateGenerator("snowflake", map[string]interface{}{
			"node_id": int64(i + 1),
		})
		if err != nil {
			fmt.Printf("创建节点%d的生成器失败: %v\n", i+1, err)
			continue
		}
		generators = append(generators, gen)
	}
	
	fmt.Printf("成功创建%d个节点的生成器\n", len(generators))
	
	// 每个节点生成一些ID
	allIDs := make([]int64, 0)
	for i, gen := range generators {
		fmt.Printf("节点%d生成的ID: ", i+1)
		for j := 0; j < 5; j++ {
			id, err := gen.NextID()
			if err != nil {
				fmt.Printf("错误 ")
			} else {
				fmt.Printf("%d ", id)
				allIDs = append(allIDs, id)
			}
		}
		fmt.Println()
	}
	
	// 检查所有ID的唯一性
	uniqueIDs := make(map[int64]bool)
	duplicates := 0
	for _, id := range allIDs {
		if uniqueIDs[id] {
			duplicates++
		} else {
			uniqueIDs[id] = true
		}
	}
	
	if duplicates > 0 {
		fmt.Printf("警告: 多节点间发现%d个重复ID\n", duplicates)
	} else {
		fmt.Println("多节点生成的所有ID都是唯一的")
	}
}

// testUniqueness 测试ID唯一性
func testUniqueness(gen id.Generator, name string, count int) {
	fmt.Printf("%s唯一性测试: 生成%d个ID\n", name, count)
	
	uniqueIDs := make(map[int64]bool)
	duplicates := 0
	errorCount := 0
	
	for i := 0; i < count; i++ {
		id, err := gen.NextID()
		if err != nil {
			errorCount++
			continue
		}
		
		if uniqueIDs[id] {
			duplicates++
			fmt.Printf("发现重复ID: %d\n", id)
		} else {
			uniqueIDs[id] = true
		}
	}
	
	fmt.Printf("唯一性测试完成: 生成%d个唯一ID，%d个重复，%d个错误\n", 
		len(uniqueIDs), duplicates, errorCount)
}