package embedding

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	eino "github.com/cloudwego/eino/components/embedding"
	"sync"
)

const (
	ProviderOpenAI = "openai"
	ProviderOllama = "ollama"
)

// EmbeddingFactory 定义用于创建嵌入服务实例的工厂函数
type EmbeddingFactory func(ctx context.Context, cfg *model.Model, opts ...EmbeddingOption) (EmbeddingService, error)

var (
	factories   = make(map[string]EmbeddingFactory)
	registryMux sync.RWMutex
)

// EmbeddingService 定义向量嵌入服务的通用接口
type EmbeddingService interface {
	// EmbedStrings 将文本转换为向量表示
	EmbedStrings(ctx context.Context, texts []string, opts ...eino.Option) ([][]float64, error)
	// GetDimension 返回嵌入向量的维度
	GetDimension() int
}

// NewEmbeddingService 创建一个新的嵌入服务实例
func NewEmbeddingService(ctx context.Context, cfg *model.Model, opts ...EmbeddingOption) (EmbeddingService, error) {

	if cfg == nil {
		return nil, fmt.Errorf("embedding config is nil")
	}
	if cfg.Server == "" {
		return nil, fmt.Errorf("model config server is empty")
	}
	if cfg.Type != "embedding" {
		return nil, fmt.Errorf("model config type is not embedding")
	}

	registryMux.RLock()                      //  读锁
	factory, exists := factories[cfg.Server] //  从工厂中获取对应的工厂函数
	registryMux.RUnlock()                    //  释放读锁

	if !exists {
		return nil, fmt.Errorf("不支持的嵌入服务提供者: %s", cfg.Type)
	}
	return factory(ctx, cfg, opts...)
}

// Register 注册一个新的嵌入服务提供者
func Register(name string, factory EmbeddingFactory) {
	registryMux.Lock()
	defer registryMux.Unlock()

	if factory == nil {
		panic("embedding: Register factory is nil")
	}
	if _, exists := factories[name]; exists {
		panic("embedding: Register called twice for factory " + name)
	}

	factories[name] = factory
}
