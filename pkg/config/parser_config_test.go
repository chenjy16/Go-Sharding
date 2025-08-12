package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserConfig(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectedConfig *ParserConfig
	}{
		{
			name: "complete parser config",
			yamlContent: `
parser:
  enable_tidb_parser: true
  enable_postgresql_parser: false
  fallback_to_original: true
  enable_benchmarking: true
  log_parsing_errors: true
dataSources:
  ds0:
    driverName: mysql
    url: "user:pass@tcp(localhost:3306)/db"
`,
			expectedConfig: &ParserConfig{
				EnableTiDBParser:       true,
				EnablePostgreSQLParser: false,
				FallbackToOriginal:     true,
				EnableBenchmarking:     true,
				LogParsingErrors:       true,
			},
		},
		{
			name: "minimal parser config",
			yamlContent: `
parser:
  enable_tidb_parser: true
dataSources:
  ds0:
    driverName: mysql
    url: "user:pass@tcp(localhost:3306)/db"
`,
			expectedConfig: &ParserConfig{
				EnableTiDBParser:       true,
				EnablePostgreSQLParser: false,
				FallbackToOriginal:     false,
				EnableBenchmarking:     false,
				LogParsingErrors:       false,
			},
		},
		{
			name: "no parser config",
			yamlContent: `
dataSources:
  ds0:
    driverName: mysql
    url: "user:pass@tcp(localhost:3306)/db"
`,
			expectedConfig: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时配置文件
			tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.yamlContent)
			require.NoError(t, err)
			tmpFile.Close()

			// 加载配置
			config, err := LoadFromYAML(tmpFile.Name())
			require.NoError(t, err)

			// 验证解析器配置
			if tt.expectedConfig == nil {
				assert.Nil(t, config.Parser)
			} else {
				require.NotNil(t, config.Parser)
				assert.Equal(t, tt.expectedConfig.EnableTiDBParser, config.Parser.EnableTiDBParser)
				assert.Equal(t, tt.expectedConfig.EnablePostgreSQLParser, config.Parser.EnablePostgreSQLParser)
				assert.Equal(t, tt.expectedConfig.FallbackToOriginal, config.Parser.FallbackToOriginal)
				assert.Equal(t, tt.expectedConfig.EnableBenchmarking, config.Parser.EnableBenchmarking)
				assert.Equal(t, tt.expectedConfig.LogParsingErrors, config.Parser.LogParsingErrors)
			}
		})
	}
}

func TestGetParserConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *ShardingConfig
		expectedConfig *ParserConfig
	}{
		{
			name: "with parser config",
			config: &ShardingConfig{
				Parser: &ParserConfig{
					EnableTiDBParser:       true,
					EnablePostgreSQLParser: false,
					FallbackToOriginal:     true,
					EnableBenchmarking:     true,
					LogParsingErrors:       true,
				},
			},
			expectedConfig: &ParserConfig{
				EnableTiDBParser:       true,
				EnablePostgreSQLParser: false,
				FallbackToOriginal:     true,
				EnableBenchmarking:     true,
				LogParsingErrors:       true,
			},
		},
		{
			name:   "without parser config",
			config: &ShardingConfig{},
			expectedConfig: &ParserConfig{
				EnableTiDBParser:       false,
				EnablePostgreSQLParser: false,
				FallbackToOriginal:     true,
				EnableBenchmarking:     false,
				LogParsingErrors:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetParserConfig()
			assert.Equal(t, tt.expectedConfig, result)
		})
	}
}

func TestSaveParserConfigToYAML(t *testing.T) {
	config := &ShardingConfig{
		DataSources: map[string]*DataSourceConfig{
			"ds0": {
				DriverName: "mysql",
				URL:        "user:pass@tcp(localhost:3306)/db",
			},
		},
		Parser: &ParserConfig{
			EnableTiDBParser:       true,
			EnablePostgreSQLParser: false,
			FallbackToOriginal:     true,
			EnableBenchmarking:     true,
			LogParsingErrors:       true,
		},
	}

	// 保存到临时文件
	tmpFile, err := os.CreateTemp("", "test_save_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = config.SaveToYAML(tmpFile.Name())
	require.NoError(t, err)

	// 重新加载并验证
	loadedConfig, err := LoadFromYAML(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, config.Parser, loadedConfig.Parser)
}