package service

import (
	"context"
	"fmt"
	"time"

	"ai-cloud/internal/component/embedding/ollama"
	"ai-cloud/internal/model"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

// EmbeddingService 定义向量嵌入服务的通用接口
type EmbeddingService interface {
	// EmbedStrings 将文本转换为向量表示
	EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error)

	// GetDimension 返回嵌入向量的维度
	GetDimension() int
}

// EmbeddingFactory 用于创建embedding服务实例
type EmbeddingFactory interface {
	CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error)
}

// OpenAIEmbeddingFactory OpenAI嵌入服务工厂
type OpenAIEmbeddingFactory struct{}

func (f *OpenAIEmbeddingFactory) CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error) {
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     modelConfig.APIKey,
		Model:      modelConfig.ModelName,
		BaseURL:    modelConfig.BaseURL,
		Timeout:    30 * time.Second,
		Dimensions: &modelConfig.Dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("创建OpenAI嵌入服务失败: %w", err)
	}

	return &OpenAIEmbeddingService{
		embedder:  embedder,
		dimension: modelConfig.Dimension,
	}, nil
}

// OllamaEmbeddingFactory Ollama嵌入服务工厂
type OllamaEmbeddingFactory struct{}

func (f *OllamaEmbeddingFactory) CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error) {
	embedder, err := ollama.NewOllamaEmbedder(ctx, &ollama.OllamaEmbeddingConfig{
		BaseURL:    modelConfig.BaseURL,
		Model:      modelConfig.ModelName,
		Dimensions: &modelConfig.Dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("创建Ollama嵌入服务失败: %w", err)
	}

	return &OllamaEmbeddingService{
		OllamaEmbedder: embedder,
		dimension:      modelConfig.Dimension,
	}, nil
}

// OpenAIEmbeddingService 使用OpenAI API的嵌入服务实现
type OpenAIEmbeddingService struct {
	embedder  *openai.Embedder
	dimension int
}

func (s *OpenAIEmbeddingService) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	return s.embedder.EmbedStrings(ctx, texts, opts...)
}

func (s *OpenAIEmbeddingService) GetDimension() int {
	return s.dimension
}

// OllamaEmbeddingService 使用Ollama的嵌入服务实现
type OllamaEmbeddingService struct {
	*ollama.OllamaEmbedder
	dimension int
}

func (s *OllamaEmbeddingService) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	return s.OllamaEmbedder.EmbedStrings(ctx, texts, opts...)
}

func (s *OllamaEmbeddingService) GetDimension() int {
	return s.dimension
}
