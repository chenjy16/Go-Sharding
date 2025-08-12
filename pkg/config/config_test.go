package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromYAML(t *testing.T) {
	// 创建测试配置文件
	yamlContent := `
dataSources:
  ds_0:
    driverName: mysql
    url: "root:@tcp(localhost:3306)/ds_0"
    maxIdle: 10
    maxOpen: 100
  ds_1:
    driverName: mysql
    url: "root:@tcp(localhost:3306)/ds_1"
    maxIdle: 10
    maxOpen: 100

shardingRule:
  tables:
    t_order:
      logicTable: t_order
      actualDataNodes: "ds_${0..1}.t_order_${0..1}"
      databaseStrategy:
        type: inline
        shardingColumn: user_id
        algorithm: "ds_${user_id % 2}"
      tableStrategy:
        type: inline
        shardingColumn: order_id
        algorithm: "t_order_${order_id % 2}"
      keyGenerator:
        column: order_id
        type: snowflake
`

	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 测试加载配置
	config, err := LoadFromYAML(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, config)

	// 验证数据源配置
	assert.Len(t, config.DataSources, 2)
	assert.Contains(t, config.DataSources, "ds_0")
	assert.Contains(t, config.DataSources, "ds_1")

	ds0 := config.DataSources["ds_0"]
	assert.Equal(t, "mysql", ds0.DriverName)
	assert.Equal(t, "root:@tcp(localhost:3306)/ds_0", ds0.URL)
	assert.Equal(t, 10, ds0.MaxIdle)
	assert.Equal(t, 100, ds0.MaxOpen)

	// 验证分片规则配置
	assert.NotNil(t, config.ShardingRule)
	assert.Len(t, config.ShardingRule.Tables, 1)
	assert.Contains(t, config.ShardingRule.Tables, "t_order")

	tableRule := config.ShardingRule.Tables["t_order"]
	assert.Equal(t, "t_order", tableRule.LogicTable)
	assert.Equal(t, "ds_${0..1}.t_order_${0..1}", tableRule.ActualDataNodes)

	// 验证数据库分片策略
	assert.NotNil(t, tableRule.DatabaseStrategy)
	assert.Equal(t, "inline", tableRule.DatabaseStrategy.Type)
	assert.Equal(t, "user_id", tableRule.DatabaseStrategy.ShardingColumn)
	assert.Equal(t, "ds_${user_id % 2}", tableRule.DatabaseStrategy.Algorithm)

	// 验证表分片策略
	assert.NotNil(t, tableRule.TableStrategy)
	assert.Equal(t, "inline", tableRule.TableStrategy.Type)
	assert.Equal(t, "order_id", tableRule.TableStrategy.ShardingColumn)
	assert.Equal(t, "t_order_${order_id % 2}", tableRule.TableStrategy.Algorithm)

	// 验证主键生成器
	assert.NotNil(t, tableRule.KeyGenerator)
	assert.Equal(t, "order_id", tableRule.KeyGenerator.Column)
	assert.Equal(t, "snowflake", tableRule.KeyGenerator.Type)
}

func TestSaveToYAML(t *testing.T) {
	// 创建测试配置
	config := &ShardingConfig{
		DataSources: map[string]*DataSourceConfig{
			"ds_0": {
				DriverName: "mysql",
				URL:        "root:@tcp(localhost:3306)/ds_0",
				MaxIdle:    10,
				MaxOpen:    100,
			},
		},
		ShardingRule: &ShardingRuleConfig{
			Tables: map[string]*TableRuleConfig{
				"t_order": {
					LogicTable:      "t_order",
					ActualDataNodes: "ds_0.t_order_${0..1}",
					DatabaseStrategy: &ShardingStrategyConfig{
						Type:           "inline",
						ShardingColumn: "user_id",
						Algorithm:      "ds_${user_id % 2}",
					},
				},
			},
		},
	}

	// 保存到临时文件
	tmpFile, err := os.CreateTemp("", "test_save_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = config.SaveToYAML(tmpFile.Name())
	require.NoError(t, err)

	// 重新加载验证
	loadedConfig, err := LoadFromYAML(tmpFile.Name())
	require.NoError(t, err)

	assert.Len(t, loadedConfig.DataSources, 1)
	assert.Contains(t, loadedConfig.DataSources, "ds_0")
	assert.Equal(t, "mysql", loadedConfig.DataSources["ds_0"].DriverName)

	assert.NotNil(t, loadedConfig.ShardingRule)
	assert.Len(t, loadedConfig.ShardingRule.Tables, 1)
	assert.Contains(t, loadedConfig.ShardingRule.Tables, "t_order")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ShardingConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &ShardingConfig{
				DataSources: map[string]*DataSourceConfig{
					"ds_0": {
						DriverName: "mysql",
						URL:        "root:@tcp(localhost:3306)/ds_0",
					},
				},
				ShardingRule: &ShardingRuleConfig{
					Tables: map[string]*TableRuleConfig{
						"t_order": {
							ActualDataNodes: "ds_0.t_order",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "no data sources",
			config: &ShardingConfig{
				DataSources: map[string]*DataSourceConfig{},
			},
			expectError: true,
			errorMsg:    "at least one data source must be configured",
		},
		{
			name: "missing driver name",
			config: &ShardingConfig{
				DataSources: map[string]*DataSourceConfig{
					"ds_0": {
						URL: "root:@tcp(localhost:3306)/ds_0",
					},
				},
			},
			expectError: true,
			errorMsg:    "driver name is required for data source ds_0",
		},
		{
			name: "missing URL",
			config: &ShardingConfig{
				DataSources: map[string]*DataSourceConfig{
					"ds_0": {
						DriverName: "mysql",
					},
				},
			},
			expectError: true,
			errorMsg:    "URL is required for data source ds_0",
		},
		{
			name: "missing actual data nodes",
			config: &ShardingConfig{
				DataSources: map[string]*DataSourceConfig{
					"ds_0": {
						DriverName: "mysql",
						URL:        "root:@tcp(localhost:3306)/ds_0",
					},
				},
				ShardingRule: &ShardingRuleConfig{
					Tables: map[string]*TableRuleConfig{
						"t_order": {},
					},
				},
			},
			expectError: true,
			errorMsg:    "actual data nodes is required for table t_order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}