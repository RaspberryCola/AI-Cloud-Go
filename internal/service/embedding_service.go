package service

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	"time"

	"ai-cloud/internal/component/embedding/ollama"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
)

// EmbeddingService 定义向量嵌入服务的通用接口
type EmbeddingService interface {
	// EmbedStrings 将文本转换为向量表示
	EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)

	// GetDimension 返回嵌入向量的维度
	GetDimension() int
}

// EmbeddingFactory 用于创建embedding服务实例
type EmbeddingFactory interface {
	CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error)
}

// 具体的工厂实现
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
		embedder:  embedder,
		dimension: modelConfig.Dimension,
	}, nil
}

// OpenAIEmbeddingService 使用OpenAI API的嵌入服务
type OpenAIEmbeddingService struct {
	embedder  *openai.Embedder
	dimension int
}

// OllamaEmbeddingService 使用Ollama的嵌入服务
type OllamaEmbeddingService struct {
	embedder  *ollama.OllamaEmbedder // 作为标准接口保留，但不直接使用
	dimension int
}

// EmbedStrings OpenAI实现的向量嵌入
func (s *OpenAIEmbeddingService) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return s.embedder.EmbedStrings(ctx, texts)
}

// GetDimension 返回OpenAI嵌入向量的维度
func (s *OpenAIEmbeddingService) GetDimension() int {
	return s.dimension
}

// CreateEmbedder 根据模型配置创建embedder实例
func (s *OpenAIEmbeddingService) CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error) {
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

// EmbedStrings Ollama实现的向量嵌入，使用自定义API调用
func (s *OllamaEmbeddingService) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return s.embedder.EmbedStrings(ctx, texts)
}

// GetDimension 返回Ollama嵌入向量的维度
func (s *OllamaEmbeddingService) GetDimension() int {
	return s.dimension
}

func (s *OllamaEmbeddingService) CreateEmbedder(ctx context.Context, modelConfig *model.Model) (EmbeddingService, error) {
	embedder, err := ollama.NewOllamaEmbedder(ctx, &ollama.OllamaEmbeddingConfig{
		BaseURL:    modelConfig.BaseURL,
		Model:      modelConfig.ModelName,
		Dimensions: &modelConfig.Dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("创建Ollama嵌入服务失败: %w", err)
	}

	return &OllamaEmbeddingService{
		embedder:  embedder,
		dimension: modelConfig.Dimension,
	}, nil
}
