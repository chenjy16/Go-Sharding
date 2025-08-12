package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalDialectRegistry(t *testing.T) {
	// 测试获取 MySQL 方言
	mysqlDialect, err := GlobalDialectRegistry.GetDialect(MySQL)
	assert.NoError(t, err)
	assert.NotNil(t, mysqlDialect)
	assert.Equal(t, "`", mysqlDialect.GetQuoteCharacter())
	assert.Equal(t, MySQL, mysqlDialect.GetDatabaseType())

	// 测试获取 PostgreSQL 方言
	pgDialect, err := GlobalDialectRegistry.GetDialect(PostgreSQL)
	assert.NoError(t, err)
	assert.NotNil(t, pgDialect)
	assert.Equal(t, "\"", pgDialect.GetQuoteCharacter())
	assert.Equal(t, PostgreSQL, pgDialect.GetDatabaseType())

	// 测试不支持的数据库类型
	_, err = GlobalDialectRegistry.GetDialect(DatabaseType("unsupported"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database type")
}

func TestGlobalDatabaseTypeRegistry(t *testing.T) {
	// 测试获取 MySQL 数据库类型
	dbType, err := GlobalDatabaseTypeRegistry.GetDatabaseType("mysql")
	assert.NoError(t, err)
	assert.Equal(t, MySQL, dbType)

	// 测试获取 PostgreSQL 数据库类型
	dbType, err = GlobalDatabaseTypeRegistry.GetDatabaseType("postgres")
	assert.NoError(t, err)
	assert.Equal(t, PostgreSQL, dbType)

	// 测试不支持的驱动
	_, err = GlobalDatabaseTypeRegistry.GetDatabaseType("unsupported")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

func TestMySQLDialect(t *testing.T) {
	dialect := &MySQLDialect{}

	// 测试引用字符
	assert.Equal(t, "`", dialect.GetQuoteCharacter())

	// 测试 LIMIT 语法
	assert.Equal(t, "LIMIT 10", dialect.GetLimitClause(0, 10))
	assert.Equal(t, "LIMIT 5, 10", dialect.GetLimitClause(5, 10))

	// 测试自增关键字
	assert.Equal(t, "AUTO_INCREMENT", dialect.GetAutoIncrementKeyword())

	// 测试当前时间戳函数
	assert.Equal(t, "NOW()", dialect.GetCurrentTimestampFunction())

	// 测试批量插入支持
	assert.True(t, dialect.SupportsBatchInsert())

	// 测试数据库类型
	assert.Equal(t, MySQL, dialect.GetDatabaseType())
}

func TestPostgreSQLDialect(t *testing.T) {
	dialect := &PostgreSQLDialect{}

	// 测试引用字符
	assert.Equal(t, "\"", dialect.GetQuoteCharacter())

	// 测试 LIMIT 语法
	assert.Equal(t, "LIMIT 10", dialect.GetLimitClause(0, 10))
	assert.Equal(t, "LIMIT 10 OFFSET 5", dialect.GetLimitClause(5, 10))

	// 测试自增关键字
	assert.Equal(t, "SERIAL", dialect.GetAutoIncrementKeyword())

	// 测试当前时间戳函数
	assert.Equal(t, "NOW()", dialect.GetCurrentTimestampFunction())

	// 测试批量插入支持
	assert.True(t, dialect.SupportsBatchInsert())

	// 测试数据库类型
	assert.Equal(t, PostgreSQL, dialect.GetDatabaseType())
}

func TestDialectRegistry_Register(t *testing.T) {
	registry := NewDialectRegistry()

	// 测试注册新方言
	customDialect := &MySQLDialect{}
	registry.Register("custom", customDialect)

	// 测试获取注册的方言
	dialect, err := registry.GetDialect("custom")
	assert.NoError(t, err)
	assert.Equal(t, customDialect, dialect)
}

func TestDatabaseTypeRegistry_Register(t *testing.T) {
	registry := NewDatabaseTypeRegistry()

	// 测试注册新数据库类型
	registry.Register("custom", "custom")

	// 测试获取注册的数据库类型
	dbType, err := registry.GetDatabaseType("custom")
	assert.NoError(t, err)
	assert.Equal(t, DatabaseType("custom"), dbType)
}

func TestDatabaseTypeConstants(t *testing.T) {
	// 测试数据库类型常量
	assert.Equal(t, DatabaseType("mysql"), MySQL)
	assert.Equal(t, DatabaseType("postgresql"), PostgreSQL)
}

// 基准测试
func BenchmarkDialectRegistry_GetDialect(b *testing.B) {
	registry := GlobalDialectRegistry

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetDialect(MySQL)
	}
}

func BenchmarkDatabaseTypeRegistry_GetDatabaseType(b *testing.B) {
	registry := GlobalDatabaseTypeRegistry

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetDatabaseType("mysql")
	}
}

func BenchmarkMySQLDialect_GetLimitClause(b *testing.B) {
	dialect := &MySQLDialect{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dialect.GetLimitClause(int64(i), 10)
	}
}

func BenchmarkPostgreSQLDialect_GetLimitClause(b *testing.B) {
	dialect := &PostgreSQLDialect{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dialect.GetLimitClause(int64(i), 10)
	}
}