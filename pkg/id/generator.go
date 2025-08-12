package id

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// Generator ID 生成器接口
type Generator interface {
	NextID() (int64, error)
}

// SnowflakeGenerator 雪花算法 ID 生成器
type SnowflakeGenerator struct {
	mutex       sync.Mutex
	epoch       int64 // 起始时间戳 (毫秒)
	workerID    int64 // 工作节点 ID
	datacenterID int64 // 数据中心 ID
	sequence    int64 // 序列号
	lastTime    int64 // 上次生成 ID 的时间戳
}

// NewSnowflakeGenerator 创建雪花算法生成器
func NewSnowflakeGenerator(workerID, datacenterID int64) (*SnowflakeGenerator, error) {
	if workerID < 0 || workerID > 31 {
		return nil, fmt.Errorf("worker ID must be between 0 and 31")
	}
	if datacenterID < 0 || datacenterID > 31 {
		return nil, fmt.Errorf("datacenter ID must be between 0 and 31")
	}

	// 使用 2020-01-01 00:00:00 UTC 作为起始时间
	epoch := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano() / 1e6

	return &SnowflakeGenerator{
		epoch:        epoch,
		workerID:     workerID,
		datacenterID: datacenterID,
		sequence:     0,
		lastTime:     0,
	}, nil
}

// NextID 生成下一个 ID
func (g *SnowflakeGenerator) NextID() (int64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now().UnixNano() / 1e6 // 毫秒时间戳

	if now < g.lastTime {
		return 0, fmt.Errorf("clock moved backwards, refusing to generate ID")
	}

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & 4095 // 12 位序列号
		if g.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			for now <= g.lastTime {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	// 组装 64 位 ID
	// 1 位符号位(0) + 41 位时间戳 + 5 位数据中心 ID + 5 位工作节点 ID + 12 位序列号
	id := ((now - g.epoch) << 22) | (g.datacenterID << 17) | (g.workerID << 12) | g.sequence

	return id, nil
}

// UUIDGenerator UUID 生成器
type UUIDGenerator struct {
}

// NewUUIDGenerator 创建 UUID 生成器
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// NextID 生成下一个 ID (返回 UUID 的哈希值)
func (g *UUIDGenerator) NextID() (int64, error) {
	uuid, err := g.generateUUID()
	if err != nil {
		return 0, err
	}
	
	// 将 UUID 转换为 int64 (简化处理)
	hash := int64(0)
	for i, b := range []byte(uuid) {
		if i >= 8 {
			break
		}
		hash = hash<<8 + int64(b)
	}
	
	return hash, nil
}

// generateUUID 生成 UUID
func (g *UUIDGenerator) generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	// 设置版本 (4) 和变体位
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// IncrementGenerator 自增 ID 生成器
type IncrementGenerator struct {
	mutex   sync.Mutex
	current int64
	step    int64
}

// NewIncrementGenerator 创建自增生成器
func NewIncrementGenerator(start, step int64) *IncrementGenerator {
	return &IncrementGenerator{
		current: start,
		step:    step,
	}
}

// NextID 生成下一个 ID
func (g *IncrementGenerator) NextID() (int64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	
	g.current += g.step
	return g.current, nil
}

// GeneratorFactory ID 生成器工厂
type GeneratorFactory struct {
	generators map[string]Generator
	mutex      sync.RWMutex
}

// NewGeneratorFactory 创建生成器工厂
func NewGeneratorFactory() *GeneratorFactory {
	return &GeneratorFactory{
		generators: make(map[string]Generator),
	}
}

// RegisterGenerator 注册生成器
func (f *GeneratorFactory) RegisterGenerator(name string, generator Generator) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.generators[name] = generator
}

// GetGenerator 获取生成器
func (f *GeneratorFactory) GetGenerator(name string) (Generator, bool) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	generator, exists := f.generators[name]
	return generator, exists
}

// CreateGenerator 根据类型创建生成器
func (f *GeneratorFactory) CreateGenerator(generatorType string, config map[string]interface{}) (Generator, error) {
	switch generatorType {
	case "snowflake":
		workerID := int64(0)
		datacenterID := int64(0)
		
		if val, ok := config["workerID"]; ok {
			if id, ok := val.(int64); ok {
				workerID = id
			}
		}
		if val, ok := config["datacenterID"]; ok {
			if id, ok := val.(int64); ok {
				datacenterID = id
			}
		}
		
		return NewSnowflakeGenerator(workerID, datacenterID)
		
	case "uuid":
		return NewUUIDGenerator(), nil
		
	case "increment":
		start := int64(1)
		step := int64(1)
		
		if val, ok := config["start"]; ok {
			if s, ok := val.(int64); ok {
				start = s
			}
		}
		if val, ok := config["step"]; ok {
			if s, ok := val.(int64); ok {
				step = s
			}
		}
		
		return NewIncrementGenerator(start, step), nil
		
	default:
		return nil, fmt.Errorf("unsupported generator type: %s", generatorType)
	}
}

// DefaultGeneratorFactory 默认生成器工厂实例
var DefaultGeneratorFactory = NewGeneratorFactory()

// GetDefaultGenerator 获取默认生成器
func GetDefaultGenerator() Generator {
	generator, exists := DefaultGeneratorFactory.GetGenerator("default")
	if !exists {
		// 创建默认的雪花算法生成器
		snowflake, _ := NewSnowflakeGenerator(1, 1)
		DefaultGeneratorFactory.RegisterGenerator("default", snowflake)
		return snowflake
	}
	return generator
}