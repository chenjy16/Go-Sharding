package database

import (
	"fmt"
	"strings"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	Oracle     DatabaseType = "oracle"
	SQLServer  DatabaseType = "sqlserver"
)

// DatabaseTypeRegistry 数据库类型注册表
type DatabaseTypeRegistry struct {
	types map[string]DatabaseType
}

// NewDatabaseTypeRegistry 创建数据库类型注册表
func NewDatabaseTypeRegistry() *DatabaseTypeRegistry {
	registry := &DatabaseTypeRegistry{
		types: make(map[string]DatabaseType),
	}
	
	// 注册支持的数据库类型
	registry.Register("mysql", MySQL)
	registry.Register("postgres", PostgreSQL)
	registry.Register("postgresql", PostgreSQL)
	registry.Register("postgres-mock", PostgreSQL) // 用于测试的 mock 驱动
	registry.Register("mock", MySQL)               // 通用 mock 驱动，默认为 MySQL
	registry.Register("oracle", Oracle)
	registry.Register("sqlserver", SQLServer)
	registry.Register("mssql", SQLServer)
	
	return registry
}

// Register 注册数据库类型
func (r *DatabaseTypeRegistry) Register(driverName string, dbType DatabaseType) {
	r.types[strings.ToLower(driverName)] = dbType
}

// GetDatabaseType 根据驱动名获取数据库类型
func (r *DatabaseTypeRegistry) GetDatabaseType(driverName string) (DatabaseType, error) {
	dbType, exists := r.types[strings.ToLower(driverName)]
	if !exists {
		return "", fmt.Errorf("unsupported database driver: %s", driverName)
	}
	return dbType, nil
}

// GetSupportedDrivers 获取支持的驱动列表
func (r *DatabaseTypeRegistry) GetSupportedDrivers() []string {
	var drivers []string
	for driver := range r.types {
		drivers = append(drivers, driver)
	}
	return drivers
}

// DatabaseDialect 数据库方言接口
type DatabaseDialect interface {
	// GetQuoteCharacter 获取引用字符
	GetQuoteCharacter() string
	
	// GetLimitClause 获取分页子句
	GetLimitClause(offset, limit int64) string
	
	// GetAutoIncrementKeyword 获取自增关键字
	GetAutoIncrementKeyword() string
	
	// GetCurrentTimestampFunction 获取当前时间戳函数
	GetCurrentTimestampFunction() string
	
	// SupportsBatchInsert 是否支持批量插入
	SupportsBatchInsert() bool
	
	// GetDatabaseType 获取数据库类型
	GetDatabaseType() DatabaseType
}

// MySQLDialect MySQL 方言
type MySQLDialect struct{}

func (d *MySQLDialect) GetQuoteCharacter() string {
	return "`"
}

func (d *MySQLDialect) GetLimitClause(offset, limit int64) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d, %d", offset, limit)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) GetAutoIncrementKeyword() string {
	return "AUTO_INCREMENT"
}

func (d *MySQLDialect) GetCurrentTimestampFunction() string {
	return "NOW()"
}

func (d *MySQLDialect) SupportsBatchInsert() bool {
	return true
}

func (d *MySQLDialect) GetDatabaseType() DatabaseType {
	return MySQL
}

// PostgreSQLDialect PostgreSQL 方言
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) GetQuoteCharacter() string {
	return "\""
}

func (d *PostgreSQLDialect) GetLimitClause(offset, limit int64) string {
	if offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *PostgreSQLDialect) GetAutoIncrementKeyword() string {
	return "SERIAL"
}

func (d *PostgreSQLDialect) GetCurrentTimestampFunction() string {
	return "NOW()"
}

func (d *PostgreSQLDialect) SupportsBatchInsert() bool {
	return true
}

func (d *PostgreSQLDialect) GetDatabaseType() DatabaseType {
	return PostgreSQL
}

// DialectRegistry 方言注册表
type DialectRegistry struct {
	dialects map[DatabaseType]DatabaseDialect
}

// NewDialectRegistry 创建方言注册表
func NewDialectRegistry() *DialectRegistry {
	registry := &DialectRegistry{
		dialects: make(map[DatabaseType]DatabaseDialect),
	}
	
	// 注册方言
	registry.Register(MySQL, &MySQLDialect{})
	registry.Register(PostgreSQL, &PostgreSQLDialect{})
	
	return registry
}

// Register 注册方言
func (r *DialectRegistry) Register(dbType DatabaseType, dialect DatabaseDialect) {
	r.dialects[dbType] = dialect
}

// GetDialect 获取方言
func (r *DialectRegistry) GetDialect(dbType DatabaseType) (DatabaseDialect, error) {
	dialect, exists := r.dialects[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return dialect, nil
}

// GetSupportedTypes 获取支持的数据库类型
func (r *DialectRegistry) GetSupportedTypes() []DatabaseType {
	var types []DatabaseType
	for dbType := range r.dialects {
		types = append(types, dbType)
	}
	return types
}

// 全局注册表实例
var (
	GlobalDatabaseTypeRegistry = NewDatabaseTypeRegistry()
	GlobalDialectRegistry      = NewDialectRegistry()
)