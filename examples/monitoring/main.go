package main

import (
	"context"
	"fmt"
	"go-sharding/pkg/monitoring"
	"time"
)

func main() {
	fmt.Println("=== Go-Sharding 监控示例 ===")

	// 1. 创建监控器
	monitor := monitoring.NewMonitor()

	// 2. 获取分片指标
	shardingMetrics := monitor.GetMetrics()

	// 3. 创建指标收集器
	collector := monitoring.NewMetricsCollector()

	// 4. 创建自定义指标并注册
	queryCounter := monitoring.NewCounterMetric("custom_query_count", map[string]string{"type": "custom"})
	connectionGauge := monitoring.NewGaugeMetric("custom_connections", map[string]string{"db": "user_db"})
	queryHistogram := monitoring.NewHistogramMetric("custom_query_duration", 
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0}, 
		map[string]string{"operation": "select"})

	collector.RegisterMetric(queryCounter)
	collector.RegisterMetric(connectionGauge)
	collector.RegisterMetric(queryHistogram)

	// 5. 模拟业务操作并记录指标
	fmt.Println("\n--- 模拟业务操作 ---")
	for i := 0; i < 10; i++ {
		// 记录查询
		start := time.Now()
		time.Sleep(time.Millisecond * time.Duration(10+i*5)) // 模拟查询耗时
		duration := time.Since(start)
		
		// 使用内置分片指标
		shardingMetrics.RecordQuery(duration, nil)
		shardingMetrics.RecordShardingRoute(2) // 模拟跨2个分片
		shardingMetrics.RecordConnection(5 + i)
		
		// 使用自定义指标
		queryCounter.Inc()
		connectionGauge.Set(float64(5 + i))
		queryHistogram.Observe(duration.Seconds())
		
		fmt.Printf("执行查询 %d，耗时: %v\n", i+1, duration)
	}

	// 6. 记录错误情况
	fmt.Println("\n--- 模拟错误情况 ---")
	errorDuration := 100 * time.Millisecond
	errorExample := fmt.Errorf("connection timeout")
	shardingMetrics.RecordQuery(errorDuration, errorExample)
	shardingMetrics.RecordTransaction(errorDuration, errorExample)
	fmt.Println("记录了查询和事务错误")

	// 7. 获取指标统计
	fmt.Println("\n--- 指标统计 ---")
	allMetrics := collector.GetAllMetrics()
	fmt.Printf("收集器中共有 %d 个指标:\n", len(allMetrics))
	for name, metric := range allMetrics {
		fmt.Printf("- %s: %v (类型: %d)\n", name, metric.GetValue(), metric.GetType())
	}

	// 8. 启动监控服务
	fmt.Println("\n--- 启动监控服务 ---")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动监控
	go func() {
		if err := monitor.Start(ctx); err != nil {
			fmt.Printf("监控启动失败: %v\n", err)
		}
	}()

	// 9. 模拟持续的业务操作
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	fmt.Println("监控服务运行中，模拟持续业务操作...")
	for {
		select {
		case <-ticker.C:
			// 模拟新的业务操作
			duration := time.Duration(50+time.Now().UnixNano()%100) * time.Millisecond
			shardingMetrics.RecordQuery(duration, nil)
			shardingMetrics.RecordTransaction(duration*2, nil)
			
			queryCounter.Add(2)
			connectionGauge.Add(1)
			queryHistogram.Observe(duration.Seconds())
			
			fmt.Printf("持续监控中... 查询耗时: %v\n", duration)
			
		case <-ctx.Done():
			fmt.Println("\n--- 停止监控服务 ---")
			if err := monitor.Stop(); err != nil {
				fmt.Printf("监控停止失败: %v\n", err)
			} else {
				fmt.Println("监控服务已停止")
			}
			
			// 最终统计
			fmt.Println("\n--- 最终指标统计 ---")
			finalMetrics := collector.GetAllMetrics()
			for name, metric := range finalMetrics {
				fmt.Printf("%s: %v\n", name, metric.GetValue())
			}
			
			fmt.Println("\n监控示例完成")
			return
		}
	}
}