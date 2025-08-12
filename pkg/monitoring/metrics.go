package monitoring

import (
	"context"
	"sync"
	"time"
)

// MetricType 指标类型
type MetricType int

const (
	// Counter 计数器
	Counter MetricType = iota
	// Gauge 仪表盘
	Gauge
	// Histogram 直方图
	Histogram
	// Summary 摘要
	Summary
)

// Metric 指标接口
type Metric interface {
	// GetName 获取指标名称
	GetName() string
	// GetType 获取指标类型
	GetType() MetricType
	// GetValue 获取指标值
	GetValue() interface{}
	// GetLabels 获取标签
	GetLabels() map[string]string
}

// CounterMetric 计数器指标
type CounterMetric struct {
	name   string
	value  int64
	labels map[string]string
	mu     sync.RWMutex
}

// NewCounterMetric 创建计数器指标
func NewCounterMetric(name string, labels map[string]string) *CounterMetric {
	return &CounterMetric{
		name:   name,
		labels: labels,
	}
}

// GetName 获取指标名称
func (c *CounterMetric) GetName() string {
	return c.name
}

// GetType 获取指标类型
func (c *CounterMetric) GetType() MetricType {
	return Counter
}

// GetValue 获取指标值
func (c *CounterMetric) GetValue() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// GetLabels 获取标签
func (c *CounterMetric) GetLabels() map[string]string {
	return c.labels
}

// Inc 增加计数
func (c *CounterMetric) Inc() {
	c.Add(1)
}

// Add 增加指定值
func (c *CounterMetric) Add(value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
}

// GaugeMetric 仪表盘指标
type GaugeMetric struct {
	name   string
	value  float64
	labels map[string]string
	mu     sync.RWMutex
}

// NewGaugeMetric 创建仪表盘指标
func NewGaugeMetric(name string, labels map[string]string) *GaugeMetric {
	return &GaugeMetric{
		name:   name,
		labels: labels,
	}
}

// GetName 获取指标名称
func (g *GaugeMetric) GetName() string {
	return g.name
}

// GetType 获取指标类型
func (g *GaugeMetric) GetType() MetricType {
	return Gauge
}

// GetValue 获取指标值
func (g *GaugeMetric) GetValue() interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// GetLabels 获取标签
func (g *GaugeMetric) GetLabels() map[string]string {
	return g.labels
}

// Set 设置值
func (g *GaugeMetric) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Inc 增加 1
func (g *GaugeMetric) Inc() {
	g.Add(1)
}

// Dec 减少 1
func (g *GaugeMetric) Dec() {
	g.Add(-1)
}

// Add 增加指定值
func (g *GaugeMetric) Add(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
}

// HistogramMetric 直方图指标
type HistogramMetric struct {
	name    string
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
	labels  map[string]string
	mu      sync.RWMutex
}

// NewHistogramMetric 创建直方图指标
func NewHistogramMetric(name string, buckets []float64, labels map[string]string) *HistogramMetric {
	return &HistogramMetric{
		name:    name,
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1), // +1 for +Inf bucket
		labels:  labels,
	}
}

// GetName 获取指标名称
func (h *HistogramMetric) GetName() string {
	return h.name
}

// GetType 获取指标类型
func (h *HistogramMetric) GetType() MetricType {
	return Histogram
}

// GetValue 获取指标值
func (h *HistogramMetric) GetValue() interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return map[string]interface{}{
		"buckets": h.buckets,
		"counts":  h.counts,
		"sum":     h.sum,
		"count":   h.count,
	}
}

// GetLabels 获取标签
func (h *HistogramMetric) GetLabels() map[string]string {
	return h.labels
}

// Observe 观察值
func (h *HistogramMetric) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.sum += value
	h.count++
	
	// 找到对应的桶（只增加第一个匹配的桶）
	bucketFound := false
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			bucketFound = true
			break
		}
	}
	
	// 如果没有找到匹配的桶，则放入 +Inf 桶
	if !bucketFound {
		h.counts[len(h.buckets)]++
	}
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]Metric),
	}
}

// RegisterMetric 注册指标
func (mc *MetricsCollector) RegisterMetric(metric Metric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.getMetricKey(metric.GetName(), metric.GetLabels())
	mc.metrics[key] = metric
}

// GetMetric 获取指标
func (mc *MetricsCollector) GetMetric(name string, labels map[string]string) Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	key := mc.getMetricKey(name, labels)
	return mc.metrics[key]
}

// GetAllMetrics 获取所有指标
func (mc *MetricsCollector) GetAllMetrics() map[string]Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	result := make(map[string]Metric)
	for k, v := range mc.metrics {
		result[k] = v
	}
	return result
}

// getMetricKey 生成指标键
func (mc *MetricsCollector) getMetricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += "_" + k + "_" + v
	}
	return key
}

// ShardingMetrics 分片相关指标
type ShardingMetrics struct {
	collector *MetricsCollector
	
	// 查询相关指标
	QueryTotal     *CounterMetric
	QueryDuration  *HistogramMetric
	QueryErrors    *CounterMetric
	
	// 连接相关指标
	ConnectionsActive *GaugeMetric
	ConnectionsTotal  *CounterMetric
	
	// 分片相关指标
	ShardingRoutes    *CounterMetric
	CrossShardQueries *CounterMetric
	
	// 事务相关指标
	TransactionTotal    *CounterMetric
	TransactionDuration *HistogramMetric
	TransactionErrors   *CounterMetric
}

// NewShardingMetrics 创建分片指标
func NewShardingMetrics() *ShardingMetrics {
	collector := NewMetricsCollector()
	
	// 创建各种指标
	queryTotal := NewCounterMetric("sharding_query_total", map[string]string{"type": "all"})
	queryDuration := NewHistogramMetric("sharding_query_duration_seconds", 
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0}, 
		map[string]string{})
	queryErrors := NewCounterMetric("sharding_query_errors_total", map[string]string{})
	
	connectionsActive := NewGaugeMetric("sharding_connections_active", map[string]string{})
	connectionsTotal := NewCounterMetric("sharding_connections_total", map[string]string{})
	
	shardingRoutes := NewCounterMetric("sharding_routes_total", map[string]string{})
	crossShardQueries := NewCounterMetric("sharding_cross_shard_queries_total", map[string]string{})
	
	transactionTotal := NewCounterMetric("sharding_transaction_total", map[string]string{})
	transactionDuration := NewHistogramMetric("sharding_transaction_duration_seconds",
		[]float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0}, 
		map[string]string{})
	transactionErrors := NewCounterMetric("sharding_transaction_errors_total", map[string]string{})
	
	// 注册指标
	collector.RegisterMetric(queryTotal)
	collector.RegisterMetric(queryDuration)
	collector.RegisterMetric(queryErrors)
	collector.RegisterMetric(connectionsActive)
	collector.RegisterMetric(connectionsTotal)
	collector.RegisterMetric(shardingRoutes)
	collector.RegisterMetric(crossShardQueries)
	collector.RegisterMetric(transactionTotal)
	collector.RegisterMetric(transactionDuration)
	collector.RegisterMetric(transactionErrors)
	
	return &ShardingMetrics{
		collector:           collector,
		QueryTotal:          queryTotal,
		QueryDuration:       queryDuration,
		QueryErrors:         queryErrors,
		ConnectionsActive:   connectionsActive,
		ConnectionsTotal:    connectionsTotal,
		ShardingRoutes:      shardingRoutes,
		CrossShardQueries:   crossShardQueries,
		TransactionTotal:    transactionTotal,
		TransactionDuration: transactionDuration,
		TransactionErrors:   transactionErrors,
	}
}

// RecordQuery 记录查询指标
func (sm *ShardingMetrics) RecordQuery(duration time.Duration, err error) {
	sm.QueryTotal.Inc()
	sm.QueryDuration.Observe(duration.Seconds())
	
	if err != nil {
		sm.QueryErrors.Inc()
	}
}

// RecordTransaction 记录事务指标
func (sm *ShardingMetrics) RecordTransaction(duration time.Duration, err error) {
	sm.TransactionTotal.Inc()
	sm.TransactionDuration.Observe(duration.Seconds())
	
	if err != nil {
		sm.TransactionErrors.Inc()
	}
}

// RecordShardingRoute 记录分片路由
func (sm *ShardingMetrics) RecordShardingRoute(shardCount int) {
	sm.ShardingRoutes.Inc()
	
	if shardCount > 1 {
		sm.CrossShardQueries.Inc()
	}
}

// RecordConnection 记录连接
func (sm *ShardingMetrics) RecordConnection(active int) {
	sm.ConnectionsTotal.Inc()
	sm.ConnectionsActive.Set(float64(active))
}

// GetCollector 获取指标收集器
func (sm *ShardingMetrics) GetCollector() *MetricsCollector {
	return sm.collector
}

// Monitor 监控器接口
type Monitor interface {
	// Start 启动监控
	Start(ctx context.Context) error
	// Stop 停止监控
	Stop() error
	// GetMetrics 获取指标
	GetMetrics() *ShardingMetrics
}

// MonitorImpl 监控器实现
type MonitorImpl struct {
	metrics   *ShardingMetrics
	running   bool
	stopCh    chan struct{}
	mu        sync.RWMutex
}

// NewMonitor 创建监控器
func NewMonitor() *MonitorImpl {
	return &MonitorImpl{
		metrics: NewShardingMetrics(),
		stopCh:  make(chan struct{}),
	}
}

// Start 启动监控
func (m *MonitorImpl) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		return nil
	}
	
	m.running = true
	
	// 启动指标收集协程
	go m.collectMetrics(ctx)
	
	return nil
}

// Stop 停止监控
func (m *MonitorImpl) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return nil
	}
	
	m.running = false
	close(m.stopCh)
	
	return nil
}

// GetMetrics 获取指标
func (m *MonitorImpl) GetMetrics() *ShardingMetrics {
	return m.metrics
}

// collectMetrics 收集指标
func (m *MonitorImpl) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			// 定期收集系统指标
			m.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics 收集系统指标
func (m *MonitorImpl) collectSystemMetrics() {
	// 这里可以添加系统指标收集逻辑
	// 例如：内存使用、CPU 使用、连接数等
}

// MetricsExporter 指标导出器接口
type MetricsExporter interface {
	// Export 导出指标
	Export(metrics map[string]Metric) error
}

// PrometheusExporter Prometheus 指标导出器
type PrometheusExporter struct {
	endpoint string
}

// NewPrometheusExporter 创建 Prometheus 导出器
func NewPrometheusExporter(endpoint string) *PrometheusExporter {
	return &PrometheusExporter{
		endpoint: endpoint,
	}
}

// Export 导出指标到 Prometheus
func (pe *PrometheusExporter) Export(metrics map[string]Metric) error {
	// 这里实现 Prometheus 格式的指标导出
	// 简化实现，实际需要使用 Prometheus 客户端库
	return nil
}