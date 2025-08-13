package parser

import (
	"fmt"
	"sync"
	"time"
)

// ParserType 解析器类型
type ParserType string

const (
	ParserTypeOriginal   ParserType = "original"   // 原有解析器
	ParserTypeTiDB       ParserType = "tidb"       // TiDB 解析器
	ParserTypePostgreSQL ParserType = "postgresql" // PostgreSQL 解析器
	ParserTypeEnhanced   ParserType = "enhanced"   // 增强解析器
)

// ParserInterface 解析器接口
type ParserInterface interface {
	Parse(sql string) (*SQLStatement, error)
	ExtractTables(sql string) []string
}

// ParserFactory 解析器工厂
type ParserFactory struct {
	mu              sync.RWMutex
	defaultParser   ParserType
	parsers         map[ParserType]ParserInterface
	config          *ParserConfig
	migrationStatus *MigrationStatus
}

// ParserConfig 解析器配置
type ParserConfig struct {
	EnableTiDBParser       bool `json:"enable_tidb_parser"`
	EnablePostgreSQLParser bool `json:"enable_postgresql_parser"`
	FallbackToOriginal     bool `json:"fallback_to_original"`
	EnableBenchmarking     bool `json:"enable_benchmarking"`
	LogParsingErrors       bool `json:"log_parsing_errors"`
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	TiDBMigrationPhase       string  `json:"tidb_migration_phase"`        // "disabled", "testing", "partial", "full"
	PostgreSQLMigrationPhase string  `json:"postgresql_migration_phase"`  // "disabled", "testing", "partial", "full"
	SuccessRate              float64 `json:"success_rate"`                 // 解析成功率
	PerformanceImprovement   float64 `json:"performance_improvement"`      // 性能提升百分比
	ErrorCount               int64   `json:"error_count"`                  // 错误计数
	TotalParseCount          int64   `json:"total_parse_count"`            // 总解析次数
}

// NewParserFactory 创建解析器工厂
func NewParserFactory() *ParserFactory {
	factory := &ParserFactory{
		defaultParser: ParserTypeOriginal,
		parsers:       make(map[ParserType]ParserInterface),
		config: &ParserConfig{
			EnableTiDBParser:       false,
			EnablePostgreSQLParser: false,
			FallbackToOriginal:     true,
			EnableBenchmarking:     false,
			LogParsingErrors:       true,
		},
		migrationStatus: &MigrationStatus{
			TiDBMigrationPhase:       "disabled",
			PostgreSQLMigrationPhase: "disabled",
			SuccessRate:              0.0,
			PerformanceImprovement:   0.0,
			ErrorCount:               0,
			TotalParseCount:          0,
		},
	}
	
	// 初始化解析器
	factory.initializeParsers()
	
	return factory
}

// PostgreSQLParserAdapter PostgreSQL 解析器适配器
type PostgreSQLParserAdapter struct {
	parser *PostgreSQLParser
}

func (a *PostgreSQLParserAdapter) Parse(sql string) (*SQLStatement, error) {
	enhancedStmt, err := a.parser.ParsePostgreSQLSpecific(sql)
	if err != nil {
		return nil, err
	}
	// 转换为标准 SQLStatement
	conditions := make([]Condition, 0)
	for column, value := range enhancedStmt.Conditions {
		conditions = append(conditions, Condition{
			Column:   column,
			Operator: "=",
			Value:    value,
			Logic:    "AND",
		})
	}
	return &SQLStatement{
		Type:       enhancedStmt.Type,
		Tables:     enhancedStmt.Tables,
		Columns:    enhancedStmt.Columns,
		Conditions: conditions,
	}, nil
}

func (a *PostgreSQLParserAdapter) ExtractTables(sql string) []string {
	return a.parser.ExtractTables(sql)
}

// EnhancedParserAdapter 增强解析器适配器
type EnhancedParserAdapter struct {
	parser *EnhancedSQLParser
}

func (a *EnhancedParserAdapter) Parse(sql string) (*SQLStatement, error) {
	enhancedStmt, err := a.parser.Parse(sql)
	if err != nil {
		return nil, err
	}
	// 转换为标准 SQLStatement
	conditions := make([]Condition, 0)
	for column, value := range enhancedStmt.Conditions {
		conditions = append(conditions, Condition{
			Column:   column,
			Operator: "=",
			Value:    value,
			Logic:    "AND",
		})
	}
	return &SQLStatement{
		Type:       enhancedStmt.Type,
		Tables:     enhancedStmt.Tables,
		Columns:    enhancedStmt.Columns,
		Conditions: conditions,
	}, nil
}

func (a *EnhancedParserAdapter) ExtractTables(sql string) []string {
	return a.parser.ExtractTables(sql)
}

// initializeParsers 初始化所有解析器
func (f *ParserFactory) initializeParsers() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 原有解析器
	f.parsers[ParserTypeOriginal] = NewSQLParser()
	
	// TiDB 解析器
	f.parsers[ParserTypeTiDB] = NewTiDBParser()
	
	// PostgreSQL 解析器（使用适配器）
	f.parsers[ParserTypePostgreSQL] = &PostgreSQLParserAdapter{parser: NewPostgreSQLParser()}
	
	// 增强解析器（使用适配器）
	f.parsers[ParserTypeEnhanced] = &EnhancedParserAdapter{parser: NewEnhancedSQLParser()}
}

// GetParser 获取指定类型的解析器
func (f *ParserFactory) GetParser(parserType ParserType) (ParserInterface, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	parser, exists := f.parsers[parserType]
	if !exists {
		return nil, fmt.Errorf("parser type %s not found", parserType)
	}
	
	return parser, nil
}

// GetDefaultParser 获取默认解析器
func (f *ParserFactory) GetDefaultParser() ParserInterface {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	parser, exists := f.parsers[f.defaultParser]
	if !exists {
		// 回退到原有解析器
		return f.parsers[ParserTypeOriginal]
	}
	
	return parser
}

// SetDefaultParser 设置默认解析器
func (f *ParserFactory) SetDefaultParser(parserType ParserType) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if _, exists := f.parsers[parserType]; !exists {
		return fmt.Errorf("parser type %s not found", parserType)
	}
	
	f.defaultParser = parserType
	return nil
}

// Parse 使用默认解析器解析 SQL
func (f *ParserFactory) Parse(sql string) (*SQLStatement, error) {
	return f.ParseWithType(sql, f.defaultParser)
}

// ParseWithType 使用指定类型的解析器解析 SQL
func (f *ParserFactory) ParseWithType(sql string, parserType ParserType) (*SQLStatement, error) {
	// 线程安全地更新统计信息
	f.mu.Lock()
	f.migrationStatus.TotalParseCount++
	f.mu.Unlock()
	
	parser, err := f.GetParser(parserType)
	if err != nil {
		f.mu.Lock()
		f.migrationStatus.ErrorCount++
		f.mu.Unlock()
		if f.config.FallbackToOriginal && parserType != ParserTypeOriginal {
			return f.ParseWithType(sql, ParserTypeOriginal)
		}
		return nil, err
	}
	
	stmt, err := parser.Parse(sql)
	if err != nil {
		f.mu.Lock()
		f.migrationStatus.ErrorCount++
		f.mu.Unlock()
		if f.config.FallbackToOriginal && parserType != ParserTypeOriginal {
			return f.ParseWithType(sql, ParserTypeOriginal)
		}
		return nil, err
	}
	
	// 更新成功率
	f.updateSuccessRate()
	
	return stmt, nil
}

// ExtractTables 使用默认解析器提取表名
func (f *ParserFactory) ExtractTables(sql string) []string {
	return f.ExtractTablesWithType(sql, f.defaultParser)
}

// ExtractTablesWithType 使用指定类型的解析器提取表名
func (f *ParserFactory) ExtractTablesWithType(sql string, parserType ParserType) []string {
	parser, err := f.GetParser(parserType)
	if err != nil {
		if f.config.FallbackToOriginal && parserType != ParserTypeOriginal {
			return f.ExtractTablesWithType(sql, ParserTypeOriginal)
		}
		return []string{}
	}
	
	return parser.ExtractTables(sql)
}

// StartTiDBMigration 开始 TiDB Parser 迁移
func (f *ParserFactory) StartTiDBMigration() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	tidbParser, exists := f.parsers[ParserTypeTiDB]
	if !exists {
		return fmt.Errorf("TiDB parser not available")
	}
	
	// 启用 TiDB Parser
	if tp, ok := tidbParser.(*TiDBParser); ok {
		tp.EnableTiDBParser()
	}
	
	// 更新配置
	f.config.EnableTiDBParser = true
	f.migrationStatus.TiDBMigrationPhase = "testing"
	
	return nil
}

// CompleteTiDBMigration 完成 TiDB Parser 迁移
func (f *ParserFactory) CompleteTiDBMigration() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 设置 TiDB Parser 为默认解析器
	f.defaultParser = ParserTypeTiDB
	f.migrationStatus.TiDBMigrationPhase = "full"
	
	return nil
}

// RollbackTiDBMigration 回滚 TiDB Parser 迁移
func (f *ParserFactory) RollbackTiDBMigration() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	// 禁用 TiDB Parser
	if tidbParser, exists := f.parsers[ParserTypeTiDB]; exists {
		if tp, ok := tidbParser.(*TiDBParser); ok {
			tp.DisableTiDBParser()
		}
	}
	
	// 回退到原有解析器
	f.defaultParser = ParserTypeOriginal
	f.config.EnableTiDBParser = false
	f.migrationStatus.TiDBMigrationPhase = "disabled"
	
	return nil
}

// GetMigrationStatus 获取迁移状态
func (f *ParserFactory) GetMigrationStatus() *MigrationStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// 返回状态副本
	status := *f.migrationStatus
	return &status
}

// GetConfig 获取配置
func (f *ParserFactory) GetConfig() *ParserConfig {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	// 返回配置副本
	config := *f.config
	return &config
}

// UpdateConfig 更新配置
func (f *ParserFactory) UpdateConfig(config *ParserConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	f.config = config
}

// updateSuccessRate 更新成功率
// 注意：调用此方法前必须已经获取了锁
func (f *ParserFactory) updateSuccessRate() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	if f.migrationStatus.TotalParseCount > 0 {
		successCount := f.migrationStatus.TotalParseCount - f.migrationStatus.ErrorCount
		f.migrationStatus.SuccessRate = float64(successCount) / float64(f.migrationStatus.TotalParseCount) * 100
	}
}

// BenchmarkParsers 对比不同解析器的性能
func (f *ParserFactory) BenchmarkParsers(sql string, iterations int) map[ParserType]map[string]interface{} {
	if iterations <= 0 {
		iterations = 1000
	}
	
	results := make(map[ParserType]map[string]interface{})
	
	for parserType := range f.parsers {
		parser, err := f.GetParser(parserType)
		if err != nil {
			continue
		}
		
		// 如果解析器支持基准测试
		if benchmarkable, ok := parser.(interface {
			Benchmark(string, int) map[string]interface{}
		}); ok {
			results[parserType] = benchmarkable.Benchmark(sql, iterations)
		} else {
			// 简单的性能测试
			totalDuration := time.Duration(0)
			for i := 0; i < iterations; i++ {
				duration := f.simpleBenchmark(parser, sql)
				totalDuration += duration
			}
			avgDuration := totalDuration / time.Duration(iterations)
			
			results[parserType] = map[string]interface{}{
				"iterations":    iterations,
				"total_time_ns": totalDuration.Nanoseconds(),
				"avg_time_ns":   avgDuration.Nanoseconds(),
				"sql_length":    len(sql),
			}
		}
	}
	
	return results
}

// simpleBenchmark 简单基准测试
func (f *ParserFactory) simpleBenchmark(parser ParserInterface, sql string) time.Duration {
	start := time.Now()
	parser.Parse(sql)
	return time.Since(start)
}

// GetAvailableParsers 获取可用的解析器列表
func (f *ParserFactory) GetAvailableParsers() []ParserType {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	var parsers []ParserType
	for parserType := range f.parsers {
		parsers = append(parsers, parserType)
	}
	
	return parsers
}

// ValidateSQL 验证 SQL 语法
func (f *ParserFactory) ValidateSQL(sql string) error {
	_, err := f.Parse(sql)
	return err
}

// GetParserStats 获取解析器统计信息
func (f *ParserFactory) GetParserStats() map[string]interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	return map[string]interface{}{
		"default_parser":     f.defaultParser,
		"available_parsers":  f.GetAvailableParsers(),
		"migration_status":   f.migrationStatus,
		"config":             f.config,
		"total_parse_count":  f.migrationStatus.TotalParseCount,
		"error_count":        f.migrationStatus.ErrorCount,
		"success_rate":       f.migrationStatus.SuccessRate,
	}
}

// 全局解析器工厂实例
var DefaultParserFactory = NewParserFactory()

// EnableTiDBParserAsDefault 启用TiDB解析器作为默认解析器
func EnableTiDBParserAsDefault() error {
	// 启动TiDB迁移
	if err := DefaultParserFactory.StartTiDBMigration(); err != nil {
		return fmt.Errorf("failed to start TiDB migration: %w", err)
	}
	
	// 完成TiDB迁移，设置为默认解析器
	if err := DefaultParserFactory.CompleteTiDBMigration(); err != nil {
		return fmt.Errorf("failed to complete TiDB migration: %w", err)
	}
	
	return nil
}

// DisableTiDBParser 禁用TiDB解析器，回退到原始解析器
func DisableTiDBParser() error {
	return DefaultParserFactory.RollbackTiDBMigration()
}

// GetDefaultParserType 获取当前默认解析器类型
func GetDefaultParserType() ParserType {
	DefaultParserFactory.mu.RLock()
	defer DefaultParserFactory.mu.RUnlock()
	return DefaultParserFactory.defaultParser
}

// GetParserFactoryStats 获取全局解析器工厂统计信息
func GetParserFactoryStats() map[string]interface{} {
	return DefaultParserFactory.GetParserStats()
}