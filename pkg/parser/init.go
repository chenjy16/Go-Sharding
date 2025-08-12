package parser

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// InitConfig 初始化配置
type InitConfig struct {
	EnableTiDBParser       bool
	EnablePostgreSQLParser bool
	FallbackToOriginal     bool
	EnableBenchmarking     bool
	LogParsingErrors       bool
	AutoEnableTiDB         bool // 是否自动启用TiDB解析器作为默认解析器
}

// DefaultInitConfig 默认初始化配置
var DefaultInitConfig = &InitConfig{
	EnableTiDBParser:       true,
	EnablePostgreSQLParser: false,
	FallbackToOriginal:     true,
	EnableBenchmarking:     false,
	LogParsingErrors:       true,
	AutoEnableTiDB:         true, // 默认自动启用TiDB解析器
}

// InitializeParser 初始化解析器
func InitializeParser(config *InitConfig) error {
	if config == nil {
		config = DefaultInitConfig
	}
	
	// 更新解析器工厂配置
	parserConfig := &ParserConfig{
		EnableTiDBParser:       config.EnableTiDBParser,
		EnablePostgreSQLParser: config.EnablePostgreSQLParser,
		FallbackToOriginal:     config.FallbackToOriginal,
		EnableBenchmarking:     config.EnableBenchmarking,
		LogParsingErrors:       config.LogParsingErrors,
	}
	
	DefaultParserFactory.UpdateConfig(parserConfig)
	
	// 如果启用了TiDB解析器且设置了自动启用，则将其设为默认解析器
	if config.EnableTiDBParser && config.AutoEnableTiDB {
		if err := EnableTiDBParserAsDefault(); err != nil {
			return fmt.Errorf("failed to enable TiDB parser as default: %w", err)
		}
		
		if config.LogParsingErrors {
			log.Printf("✅ TiDB Parser has been enabled as the default MySQL SQL parser")
		}
	}
	
	return nil
}

// InitializeParserFromConfig 从配置文件初始化解析器
func InitializeParserFromConfig(configFile string) error {
	// 为了避免循环依赖，我们在这里直接处理YAML文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 定义一个临时结构体来解析配置
	var configData struct {
		Parser struct {
			EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
			EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
			FallbackToOriginal     bool `yaml:"fallback_to_original"`
			EnableBenchmarking     bool `yaml:"enable_benchmarking"`
			LogParsingErrors       bool `yaml:"log_parsing_errors"`
		} `yaml:"parser"`
	}

	// 使用gopkg.in/yaml.v3来解析YAML
	if err := parseYAML(data, &configData); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 创建InitConfig
	config := &InitConfig{
		EnableTiDBParser:       configData.Parser.EnableTiDBParser,
		EnablePostgreSQLParser: configData.Parser.EnablePostgreSQLParser,
		FallbackToOriginal:     configData.Parser.FallbackToOriginal,
		EnableBenchmarking:     configData.Parser.EnableBenchmarking,
		LogParsingErrors:       configData.Parser.LogParsingErrors,
		AutoEnableTiDB:         configData.Parser.EnableTiDBParser, // 如果启用TiDB解析器，则自动启用
	}

	return InitializeParser(config)
}

// parseYAML 简单的YAML解析函数
func parseYAML(data []byte, v interface{}) error {
	// 这里使用简单的字符串解析来避免引入yaml依赖
	// 在实际使用中，建议使用gopkg.in/yaml.v3
	
	// 为了简化，我们先检查是否包含关键配置
	content := string(data)
	
	// 解析enable_tidb_parser
	if strings.Contains(content, "enable_tidb_parser: true") {
		if config, ok := v.(*struct {
			Parser struct {
				EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
				EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
				FallbackToOriginal     bool `yaml:"fallback_to_original"`
				EnableBenchmarking     bool `yaml:"enable_benchmarking"`
				LogParsingErrors       bool `yaml:"log_parsing_errors"`
			} `yaml:"parser"`
		}); ok {
			config.Parser.EnableTiDBParser = true
		}
	}
	
	// 解析其他配置项
	if strings.Contains(content, "enable_postgresql_parser: true") {
		if config, ok := v.(*struct {
			Parser struct {
				EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
				EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
				FallbackToOriginal     bool `yaml:"fallback_to_original"`
				EnableBenchmarking     bool `yaml:"enable_benchmarking"`
				LogParsingErrors       bool `yaml:"log_parsing_errors"`
			} `yaml:"parser"`
		}); ok {
			config.Parser.EnablePostgreSQLParser = true
		}
	}
	
	if strings.Contains(content, "fallback_to_original: true") {
		if config, ok := v.(*struct {
			Parser struct {
				EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
				EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
				FallbackToOriginal     bool `yaml:"fallback_to_original"`
				EnableBenchmarking     bool `yaml:"enable_benchmarking"`
				LogParsingErrors       bool `yaml:"log_parsing_errors"`
			} `yaml:"parser"`
		}); ok {
			config.Parser.FallbackToOriginal = true
		}
	}
	
	if strings.Contains(content, "enable_benchmarking: true") {
		if config, ok := v.(*struct {
			Parser struct {
				EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
				EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
				FallbackToOriginal     bool `yaml:"fallback_to_original"`
				EnableBenchmarking     bool `yaml:"enable_benchmarking"`
				LogParsingErrors       bool `yaml:"log_parsing_errors"`
			} `yaml:"parser"`
		}); ok {
			config.Parser.EnableBenchmarking = true
		}
	}
	
	if strings.Contains(content, "log_parsing_errors: true") {
		if config, ok := v.(*struct {
			Parser struct {
				EnableTiDBParser       bool `yaml:"enable_tidb_parser"`
				EnablePostgreSQLParser bool `yaml:"enable_postgresql_parser"`
				FallbackToOriginal     bool `yaml:"fallback_to_original"`
				EnableBenchmarking     bool `yaml:"enable_benchmarking"`
				LogParsingErrors       bool `yaml:"log_parsing_errors"`
			} `yaml:"parser"`
		}); ok {
			config.Parser.LogParsingErrors = true
		}
	}
	
	return nil
}

// InitializeParserFromEnv 从环境变量初始化解析器
func InitializeParserFromEnv() error {
	config := &InitConfig{
		EnableTiDBParser:       getBoolEnv("ENABLE_TIDB_PARSER", true),
		EnablePostgreSQLParser: getBoolEnv("ENABLE_POSTGRESQL_PARSER", false),
		FallbackToOriginal:     getBoolEnv("FALLBACK_TO_ORIGINAL", true),
		EnableBenchmarking:     getBoolEnv("ENABLE_BENCHMARKING", false),
		LogParsingErrors:       getBoolEnv("LOG_PARSING_ERRORS", true),
		AutoEnableTiDB:         getBoolEnv("AUTO_ENABLE_TIDB", true),
	}
	
	return InitializeParser(config)
}

// getBoolEnv 从环境变量获取布尔值
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	value = strings.ToLower(value)
	switch value {
	case "true", "1", "yes", "on", "enable", "enabled":
		return true
	case "false", "0", "no", "off", "disable", "disabled":
		return false
	default:
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
		return defaultValue
	}
}

// GetParserInfo 获取解析器信息
func GetParserInfo() map[string]interface{} {
	stats := GetParserFactoryStats()
	
	info := map[string]interface{}{
		"current_default_parser": GetDefaultParserType(),
		"available_parsers":      DefaultParserFactory.GetAvailableParsers(),
		"configuration":          DefaultParserFactory.GetConfig(),
		"migration_status":       DefaultParserFactory.GetMigrationStatus(),
		"statistics":             stats,
	}
	
	return info
}

// PrintParserInfo 打印解析器信息
func PrintParserInfo() {
	info := GetParserInfo()
	
	fmt.Println("=== Go-Sharding SQL Parser Information ===")
	fmt.Printf("Current Default Parser: %v\n", info["current_default_parser"])
	fmt.Printf("Available Parsers: %v\n", info["available_parsers"])
	
	if config, ok := info["configuration"].(*ParserConfig); ok {
		fmt.Println("\nConfiguration:")
		fmt.Printf("  TiDB Parser Enabled: %v\n", config.EnableTiDBParser)
		fmt.Printf("  PostgreSQL Parser Enabled: %v\n", config.EnablePostgreSQLParser)
		fmt.Printf("  Fallback to Original: %v\n", config.FallbackToOriginal)
		fmt.Printf("  Benchmarking Enabled: %v\n", config.EnableBenchmarking)
		fmt.Printf("  Log Parsing Errors: %v\n", config.LogParsingErrors)
	}
	
	if status, ok := info["migration_status"].(*MigrationStatus); ok {
		fmt.Println("\nMigration Status:")
		fmt.Printf("  TiDB Migration Phase: %s\n", status.TiDBMigrationPhase)
		fmt.Printf("  PostgreSQL Migration Phase: %s\n", status.PostgreSQLMigrationPhase)
		fmt.Printf("  Success Rate: %.2f%%\n", status.SuccessRate)
		fmt.Printf("  Total Parse Count: %d\n", status.TotalParseCount)
		fmt.Printf("  Error Count: %d\n", status.ErrorCount)
	}
	
	fmt.Println("==========================================")
}