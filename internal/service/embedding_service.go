package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-cloud/config"
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

// NewOpenAIEmbeddingService 创建OpenAI嵌入服务实例
func NewOpenAIEmbeddingService(ctx context.Context) (*OpenAIEmbeddingService, error) {
	cfg := config.GetConfig().Embedding.Remote

	apiKey := cfg.APIKey
	model := cfg.Model
	baseURL := cfg.BaseURL

	dimension := cfg.Dimension // 默认维度

	fmt.Println("创建OpenAI嵌入服务:", baseURL, "模型:", model)

	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     apiKey,
		Model:      model,
		BaseURL:    baseURL,
		Timeout:    30 * time.Second,
		Dimensions: &dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("创建OpenAI嵌入服务失败: %w", err)
	}

	return &OpenAIEmbeddingService{
		embedder:  embedder,
		dimension: dimension,
	}, nil
}

// EmbedStrings OpenAI实现的向量嵌入
func (s *OpenAIEmbeddingService) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	return s.embedder.EmbedStrings(ctx, texts)
}

// GetDimension 返回OpenAI嵌入向量的维度
func (s *OpenAIEmbeddingService) GetDimension() int {
	return s.dimension
}

// NewOllamaEmbeddingService 创建Ollama嵌入服务实例
func NewOllamaEmbeddingService(ctx context.Context) (*OllamaEmbeddingService, error) {
	cfg := config.GetConfig().Embedding.Ollama

	ollamaURL := cfg.URL
	ollamaModel := cfg.Model
	dimension := cfg.Dimension

	fmt.Println("创建Ollama嵌入服务:", ollamaURL, "模型:", ollamaModel)

	embedder, err := ollama.NewOllamaEmbedderEmbedder(ctx, &ollama.OllamaEmbeddingConfig{
		BaseURL:    ollamaURL,
		Model:      ollamaModel,
		Dimensions: &dimension,
	})

	if err != nil {
		return nil, fmt.Errorf("创建Ollama嵌入服务失败: %w", err)
	}

	return &OllamaEmbeddingService{
		embedder:  embedder,
		dimension: dimension,
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

// NewEmbeddingService 工厂函数，根据配置创建合适的嵌入服务
func NewEmbeddingService(ctx context.Context) (EmbeddingService, error) {
	embeddingType := strings.ToLower(config.GetConfig().Embedding.Service)

	switch embeddingType {
	case "ollama":
		return NewOllamaEmbeddingService(ctx)
	case "remote", "openai", "":
		return NewOpenAIEmbeddingService(ctx)
	default:
		return NewOpenAIEmbeddingService(ctx)
	}
}
