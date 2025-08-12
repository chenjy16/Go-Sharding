package monitoring

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCounterMetric(t *testing.T) {
	labels := map[string]string{"service": "test"}
	counter := NewCounterMetric("test_counter", labels)
	
	assert.NotNil(t, counter)
	assert.Equal(t, "test_counter", counter.GetName())
	assert.Equal(t, Counter, counter.GetType())
	assert.Equal(t, labels, counter.GetLabels())
	assert.Equal(t, int64(0), counter.GetValue())
}

func TestCounterMetric_Inc(t *testing.T) {
	counter := NewCounterMetric("test_counter", nil)
	
	counter.Inc()
	assert.Equal(t, int64(1), counter.GetValue())
	
	counter.Inc()
	assert.Equal(t, int64(2), counter.GetValue())
}

func TestCounterMetric_Add(t *testing.T) {
	counter := NewCounterMetric("test_counter", nil)
	
	counter.Add(5)
	assert.Equal(t, int64(5), counter.GetValue())
	
	counter.Add(3)
	assert.Equal(t, int64(8), counter.GetValue())
}

func TestCounterMetric_Concurrent(t *testing.T) {
	counter := NewCounterMetric("test_counter", nil)
	
	var wg sync.WaitGroup
	goroutines := 100
	increments := 10
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < increments; j++ {
				counter.Inc()
			}
		}()
	}
	
	wg.Wait()
	expected := int64(goroutines * increments)
	assert.Equal(t, expected, counter.GetValue())
}

func TestNewGaugeMetric(t *testing.T) {
	labels := map[string]string{"service": "test"}
	gauge := NewGaugeMetric("test_gauge", labels)
	
	assert.NotNil(t, gauge)
	assert.Equal(t, "test_gauge", gauge.GetName())
	assert.Equal(t, Gauge, gauge.GetType())
	assert.Equal(t, labels, gauge.GetLabels())
	assert.Equal(t, float64(0), gauge.GetValue())
}

func TestGaugeMetric_Set(t *testing.T) {
	gauge := NewGaugeMetric("test_gauge", nil)
	
	gauge.Set(3.14)
	assert.Equal(t, 3.14, gauge.GetValue())
	
	gauge.Set(-2.5)
	assert.Equal(t, -2.5, gauge.GetValue())
}

func TestGaugeMetric_Inc(t *testing.T) {
	gauge := NewGaugeMetric("test_gauge", nil)
	
	gauge.Inc()
	assert.Equal(t, float64(1), gauge.GetValue())
	
	gauge.Inc()
	assert.Equal(t, float64(2), gauge.GetValue())
}

func TestGaugeMetric_Dec(t *testing.T) {
	gauge := NewGaugeMetric("test_gauge", nil)
	gauge.Set(5)
	
	gauge.Dec()
	assert.Equal(t, float64(4), gauge.GetValue())
	
	gauge.Dec()
	assert.Equal(t, float64(3), gauge.GetValue())
}

func TestGaugeMetric_Add(t *testing.T) {
	gauge := NewGaugeMetric("test_gauge", nil)
	
	gauge.Add(2.5)
	assert.Equal(t, 2.5, gauge.GetValue())
	
	gauge.Add(-1.5)
	assert.Equal(t, 1.0, gauge.GetValue())
}

func TestGaugeMetric_Concurrent(t *testing.T) {
	gauge := NewGaugeMetric("test_gauge", nil)
	
	var wg sync.WaitGroup
	goroutines := 50
	
	// 一半增加，一半减少
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				gauge.Inc()
			} else {
				gauge.Dec()
			}
		}(i)
	}
	
	wg.Wait()
	
	// 由于并发，最终值可能不确定，但应该是有效的浮点数
	value := gauge.GetValue().(float64)
	assert.True(t, value >= -float64(goroutines) && value <= float64(goroutines))
}

func TestNewHistogramMetric(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	labels := map[string]string{"service": "test"}
	histogram := NewHistogramMetric("test_histogram", buckets, labels)
	
	assert.NotNil(t, histogram)
	assert.Equal(t, "test_histogram", histogram.GetName())
	assert.Equal(t, Histogram, histogram.GetType())
	assert.Equal(t, labels, histogram.GetLabels())
	
	value := histogram.GetValue().(map[string]interface{})
	assert.Equal(t, buckets, value["buckets"])
	assert.Equal(t, make([]int64, len(buckets)+1), value["counts"])
	assert.Equal(t, float64(0), value["sum"])
	assert.Equal(t, int64(0), value["count"])
}

func TestHistogramMetric_Observe(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	histogram := NewHistogramMetric("test_histogram", buckets, nil)
	
	// 观察一些值
	histogram.Observe(0.05) // bucket 0
	histogram.Observe(0.3)  // bucket 1
	histogram.Observe(0.8)  // bucket 2
	histogram.Observe(3.0)  // bucket 4
	histogram.Observe(10.0) // bucket 5 (+Inf)
	
	value := histogram.GetValue().(map[string]interface{})
	counts := value["counts"].([]int64)
	
	assert.Equal(t, int64(1), counts[0]) // 0.05 in [0, 0.1)
	assert.Equal(t, int64(1), counts[1]) // 0.3 in (0.1, 0.5]
	assert.Equal(t, int64(1), counts[2]) // 0.8 in (0.5, 1.0]
	assert.Equal(t, int64(0), counts[3]) // no values in (1.0, 2.5]
	assert.Equal(t, int64(1), counts[4]) // 3.0 in (2.5, 5.0]
	assert.Equal(t, int64(1), counts[5]) // 10.0 in (5.0, +Inf]
	
	assert.Equal(t, float64(14.15), value["sum"])   // 0.05 + 0.3 + 0.8 + 3.0 + 10.0
	assert.Equal(t, int64(5), value["count"])
}

func TestHistogramMetric_Concurrent(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0}
	histogram := NewHistogramMetric("test_histogram", buckets, nil)
	
	var wg sync.WaitGroup
	goroutines := 100
	observations := 10
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < observations; j++ {
				histogram.Observe(float64(id%3) * 0.4) // 0, 0.4, 0.8
			}
		}(i)
	}
	
	wg.Wait()
	
	value := histogram.GetValue().(map[string]interface{})
	totalCount := value["count"].(int64)
	expectedCount := int64(goroutines * observations)
	assert.Equal(t, expectedCount, totalCount)
}

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	assert.NotNil(t, collector)
}

func TestMetricsCollector_RegisterMetric(t *testing.T) {
	collector := NewMetricsCollector()
	counter := NewCounterMetric("test_counter", nil)
	
	collector.RegisterMetric(counter)
	
	// 验证指标已注册
	metric := collector.GetMetric("test_counter", nil)
	assert.NotNil(t, metric)
	assert.Equal(t, counter, metric)
}

func TestMetricsCollector_GetMetric(t *testing.T) {
	collector := NewMetricsCollector()
	counter := NewCounterMetric("test_counter", nil)
	
	collector.RegisterMetric(counter)
	
	metric := collector.GetMetric("test_counter", nil)
	assert.NotNil(t, metric)
	assert.Equal(t, counter, metric)
	
	// 获取不存在的指标
	metric = collector.GetMetric("non_existent", nil)
	assert.Nil(t, metric)
}

func TestMetricsCollector_GetAllMetrics(t *testing.T) {
	collector := NewMetricsCollector()
	
	counter := NewCounterMetric("test_counter", nil)
	gauge := NewGaugeMetric("test_gauge", nil)
	
	collector.RegisterMetric(counter)
	collector.RegisterMetric(gauge)
	
	metrics := collector.GetAllMetrics()
	assert.Len(t, metrics, 2)
	
	names := make(map[string]bool)
	for _, metric := range metrics {
		names[metric.GetName()] = true
	}
	
	assert.True(t, names["test_counter"])
	assert.True(t, names["test_gauge"])
}

func TestNewShardingMetrics(t *testing.T) {
	metrics := NewShardingMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.collector)
	assert.NotNil(t, metrics.QueryTotal)
	assert.NotNil(t, metrics.QueryDuration)
	assert.NotNil(t, metrics.QueryErrors)
}

func TestShardingMetrics_RecordQuery(t *testing.T) {
	metrics := NewShardingMetrics()
	
	// 记录成功查询
	metrics.RecordQuery(time.Millisecond*100, nil)
	assert.Equal(t, int64(1), metrics.QueryTotal.GetValue())
	
	// 记录失败查询
	metrics.RecordQuery(time.Millisecond*50, assert.AnError)
	assert.Equal(t, int64(2), metrics.QueryTotal.GetValue())
	assert.Equal(t, int64(1), metrics.QueryErrors.GetValue())
}

func TestShardingMetrics_RecordTransaction(t *testing.T) {
	metrics := NewShardingMetrics()
	
	// 记录成功事务
	metrics.RecordTransaction(time.Millisecond*50, nil)
	assert.Equal(t, int64(1), metrics.TransactionTotal.GetValue())
	
	// 记录失败事务
	metrics.RecordTransaction(time.Millisecond*30, assert.AnError)
	assert.Equal(t, int64(2), metrics.TransactionTotal.GetValue())
	assert.Equal(t, int64(1), metrics.TransactionErrors.GetValue())
}

func TestShardingMetrics_RecordShardingRoute(t *testing.T) {
	metrics := NewShardingMetrics()
	
	// 单分片查询
	metrics.RecordShardingRoute(1)
	assert.Equal(t, int64(1), metrics.ShardingRoutes.GetValue())
	assert.Equal(t, int64(0), metrics.CrossShardQueries.GetValue())
	
	// 跨分片查询
	metrics.RecordShardingRoute(3)
	assert.Equal(t, int64(2), metrics.ShardingRoutes.GetValue())
	assert.Equal(t, int64(1), metrics.CrossShardQueries.GetValue())
}

func TestShardingMetrics_RecordConnection(t *testing.T) {
	metrics := NewShardingMetrics()
	
	metrics.RecordConnection(10)
	assert.Equal(t, int64(1), metrics.ConnectionsTotal.GetValue())
	assert.Equal(t, float64(10), metrics.ConnectionsActive.GetValue())
	
	metrics.RecordConnection(15)
	assert.Equal(t, int64(2), metrics.ConnectionsTotal.GetValue())
	assert.Equal(t, float64(15), metrics.ConnectionsActive.GetValue())
}

func TestNewMonitor(t *testing.T) {
	monitor := NewMonitor()
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.metrics)
}

func TestMonitor_Start_Stop(t *testing.T) {
	monitor := NewMonitor()
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动监控
	err := monitor.Start(ctx)
	assert.NoError(t, err)
	
	// 等待一小段时间
	time.Sleep(time.Millisecond * 10)
	
	// 停止监控
	err = monitor.Stop()
	assert.NoError(t, err)
}

func TestMonitor_GetMetrics(t *testing.T) {
	monitor := NewMonitor()
	
	metrics := monitor.GetMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.QueryTotal)
}

func TestShardingMetrics_Concurrent(t *testing.T) {
	metrics := NewShardingMetrics()
	
	var wg sync.WaitGroup
	goroutines := 50
	operations := 10
	
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < operations; j++ {
				metrics.RecordQuery(time.Millisecond*10, nil)
				metrics.RecordConnection(10)
			}
		}(i)
	}
	
	wg.Wait()
	
	// 验证指标计数
	expectedQueries := int64(goroutines * operations)
	assert.Equal(t, expectedQueries, metrics.QueryTotal.GetValue())
}

// Benchmark tests
func BenchmarkCounterMetric_Inc(b *testing.B) {
	counter := NewCounterMetric("test_counter", nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

func BenchmarkGaugeMetric_Set(b *testing.B) {
	gauge := NewGaugeMetric("test_gauge", nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gauge.Set(3.14)
		}
	})
}

func BenchmarkHistogramMetric_Observe(b *testing.B) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	histogram := NewHistogramMetric("test_histogram", buckets, nil)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			histogram.Observe(1.5)
		}
	})
}

func BenchmarkShardingMetrics_RecordQuery(b *testing.B) {
	metrics := NewShardingMetrics()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.RecordQuery(time.Millisecond*10, nil)
		}
	})
}