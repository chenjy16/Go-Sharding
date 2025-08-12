package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	DriverName string `yaml:"driverName" json:"driverName"`
	URL        string `yaml:"url" json:"url"`
	Username   string `yaml:"username" json:"username"`
	Password   string `yaml:"password" json:"password"`
	MaxIdle    int    `yaml:"maxIdle" json:"maxIdle"`
	MaxOpen    int    `yaml:"maxOpen" json:"maxOpen"`
}

// ShardingStrategyConfig 分片策略配置
type ShardingStrategyConfig struct {
	ShardingColumn string `yaml:"shardingColumn" json:"shardingColumn"`
	Algorithm      string `yaml:"algorithm" json:"algorithm"`
	Type           string `yaml:"type" json:"type"` // inline, standard, complex, hint
}

// TableRuleConfig 表规则配置
type TableRuleConfig struct {
	LogicTable       string                  `yaml:"logicTable" json:"logicTable"`
	ActualDataNodes  string                  `yaml:"actualDataNodes" json:"actualDataNodes"`
	DatabaseStrategy *ShardingStrategyConfig `yaml:"databaseStrategy" json:"databaseStrategy"`
	TableStrategy    *ShardingStrategyConfig `yaml:"tableStrategy" json:"tableStrategy"`
	KeyGenerator     *KeyGeneratorConfig     `yaml:"keyGenerator" json:"keyGenerator"`
}

// KeyGeneratorConfig 主键生成器配置
type KeyGeneratorConfig struct {
	Column string `yaml:"column" json:"column"`
	Type   string `yaml:"type" json:"type"` // snowflake, uuid, increment
}

// ShardingRuleConfig 分片规则配置
type ShardingRuleConfig struct {
	Tables                map[string]*TableRuleConfig `yaml:"tables" json:"tables"`
	DefaultDatabaseStrategy *ShardingStrategyConfig   `yaml:"defaultDatabaseStrategy" json:"defaultDatabaseStrategy"`
	DefaultTableStrategy    *ShardingStrategyConfig   `yaml:"defaultTableStrategy" json:"defaultTableStrategy"`
	DefaultKeyGenerator     *KeyGeneratorConfig       `yaml:"defaultKeyGenerator" json:"defaultKeyGenerator"`
}

// ReadWriteSplitConfig 读写分离配置
type ReadWriteSplitConfig struct {
	Name            string   `yaml:"name" json:"name"`
	MasterDataSource string   `yaml:"masterDataSource" json:"masterDataSource"`
	SlaveDataSources []string `yaml:"slaveDataSources" json:"slaveDataSources"`
	LoadBalanceAlgorithm string `yaml:"loadBalanceAlgorithm" json:"loadBalanceAlgorithm"`
}

// ShardingConfig 完整的分片配置
type ShardingConfig struct {
	DataSources      map[string]*DataSourceConfig    `yaml:"dataSources" json:"dataSources"`
	ShardingRule     *ShardingRuleConfig            `yaml:"shardingRule" json:"shardingRule"`
	ReadWriteSplits  map[string]*ReadWriteSplitConfig `yaml:"readWriteSplits" json:"readWriteSplits"`
}

// LoadFromYAML 从 YAML 文件加载配置
func LoadFromYAML(filename string) (*ShardingConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ShardingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveToYAML 保存配置到 YAML 文件
func (c *ShardingConfig) SaveToYAML(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate 验证配置的有效性
func (c *ShardingConfig) Validate() error {
	if len(c.DataSources) == 0 {
		return fmt.Errorf("at least one data source must be configured")
	}

	for name, ds := range c.DataSources {
		if ds.DriverName == "" {
			return fmt.Errorf("driver name is required for data source %s", name)
		}
		if ds.URL == "" {
			return fmt.Errorf("URL is required for data source %s", name)
		}
	}

	if c.ShardingRule != nil {
		for tableName, tableRule := range c.ShardingRule.Tables {
			if tableRule.LogicTable == "" {
				tableRule.LogicTable = tableName
			}
			if tableRule.ActualDataNodes == "" {
				return fmt.Errorf("actual data nodes is required for table %s", tableName)
			}
		}
	}

	return nil
}