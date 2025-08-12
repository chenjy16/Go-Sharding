package readwrite

import (
	"context"
	"database/sql"
	"fmt"
	"go-sharding/pkg/config"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// LoadBalanceAlgorithm 负载均衡算法类型
type LoadBalanceAlgorithm string

const (
	RoundRobin LoadBalanceAlgorithm = "round_robin"
	Random     LoadBalanceAlgorithm = "random"
	Weight     LoadBalanceAlgorithm = "weight"
)

// ReadWriteSplitter 读写分离器
type ReadWriteSplitter struct {
	config       *config.ReadWriteSplitConfig
	dataSources  map[string]*sql.DB
	masterDB     *sql.DB
	slaveDBS     []*sql.DB
	roundRobinIndex int
	mutex        sync.RWMutex
	rand         *rand.Rand
}

// NewReadWriteSplitter 创建读写分离器
func NewReadWriteSplitter(cfg *config.ReadWriteSplitConfig, dataSources map[string]*sql.DB) (*ReadWriteSplitter, error) {
	splitter := &ReadWriteSplitter{
		config:      cfg,
		dataSources: dataSources,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// 获取主库连接
	masterDB, exists := dataSources[cfg.MasterDataSource]
	if !exists {
		return nil, fmt.Errorf("master data source %s not found", cfg.MasterDataSource)
	}
	splitter.masterDB = masterDB

	// 获取从库连接
	for _, slaveName := range cfg.SlaveDataSources {
		slaveDB, exists := dataSources[slaveName]
		if !exists {
			return nil, fmt.Errorf("slave data source %s not found", slaveName)
		}
		splitter.slaveDBS = append(splitter.slaveDBS, slaveDB)
	}

	if len(splitter.slaveDBS) == 0 {
		return nil, fmt.Errorf("at least one slave data source is required")
	}

	return splitter, nil
}

// Route 根据 SQL 类型路由到相应的数据库
func (rws *ReadWriteSplitter) Route(sql string) *sql.DB {
	if rws.isWriteSQL(sql) {
		return rws.masterDB
	}
	return rws.selectSlaveDB()
}

// RouteContext 根据 SQL 类型和上下文路由到相应的数据库
func (rws *ReadWriteSplitter) RouteContext(ctx context.Context, sql string) *sql.DB {
	// 检查上下文中是否强制使用主库
	if forceMaster, ok := ctx.Value("force_master").(bool); ok && forceMaster {
		return rws.masterDB
	}

	// 检查是否在事务中（事务中的读操作也应该路由到主库）
	if inTransaction, ok := ctx.Value("in_transaction").(bool); ok && inTransaction {
		return rws.masterDB
	}

	return rws.Route(sql)
}

// isWriteSQL 判断是否为写操作 SQL
func (rws *ReadWriteSplitter) isWriteSQL(sql string) bool {
	sql = strings.TrimSpace(strings.ToUpper(sql))
	
	writeKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER",
		"TRUNCATE", "REPLACE", "MERGE", "CALL", "EXEC",
	}

	for _, keyword := range writeKeywords {
		if strings.HasPrefix(sql, keyword) {
			return true
		}
	}

	return false
}

// selectSlaveDB 根据负载均衡算法选择从库
func (rws *ReadWriteSplitter) selectSlaveDB() *sql.DB {
	if len(rws.slaveDBS) == 1 {
		return rws.slaveDBS[0]
	}

	switch LoadBalanceAlgorithm(rws.config.LoadBalanceAlgorithm) {
	case RoundRobin:
		return rws.selectByRoundRobin()
	case Random:
		return rws.selectByRandom()
	case Weight:
		// 简化实现，暂时使用轮询
		return rws.selectByRoundRobin()
	default:
		return rws.selectByRoundRobin()
	}
}

// selectByRoundRobin 轮询选择从库
func (rws *ReadWriteSplitter) selectByRoundRobin() *sql.DB {
	rws.mutex.Lock()
	defer rws.mutex.Unlock()

	db := rws.slaveDBS[rws.roundRobinIndex]
	rws.roundRobinIndex = (rws.roundRobinIndex + 1) % len(rws.slaveDBS)
	return db
}

// selectByRandom 随机选择从库
func (rws *ReadWriteSplitter) selectByRandom() *sql.DB {
	rws.mutex.RLock()
	defer rws.mutex.RUnlock()

	index := rws.rand.Intn(len(rws.slaveDBS))
	return rws.slaveDBS[index]
}

// GetMasterDB 获取主库连接
func (rws *ReadWriteSplitter) GetMasterDB() *sql.DB {
	return rws.masterDB
}

// GetSlaveDBS 获取所有从库连接
func (rws *ReadWriteSplitter) GetSlaveDBS() []*sql.DB {
	return rws.slaveDBS
}

// GetConfig 获取配置
func (rws *ReadWriteSplitter) GetConfig() *config.ReadWriteSplitConfig {
	return rws.config
}

// HealthCheck 健康检查
func (rws *ReadWriteSplitter) HealthCheck() error {
	// 检查主库
	if err := rws.masterDB.Ping(); err != nil {
		return fmt.Errorf("master database health check failed: %w", err)
	}

	// 检查从库
	for i, slaveDB := range rws.slaveDBS {
		if err := slaveDB.Ping(); err != nil {
			return fmt.Errorf("slave database %d health check failed: %w", i, err)
		}
	}

	return nil
}

// Close 关闭所有数据库连接
func (rws *ReadWriteSplitter) Close() error {
	var errors []string

	if err := rws.masterDB.Close(); err != nil {
		errors = append(errors, fmt.Sprintf("failed to close master DB: %v", err))
	}

	for i, slaveDB := range rws.slaveDBS {
		if err := slaveDB.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close slave DB %d: %v", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %s", strings.Join(errors, "; "))
	}

	return nil
}