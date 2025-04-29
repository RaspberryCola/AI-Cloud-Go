package embedding

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"net/url"
	"time"
)

func init() {
	register(ProviderOpenAI, &openaiEmbedder{})
}

type OpenAIEmbeddingConfig struct {
	BaseURL   string
	APIKey    string
	Model     string
	Timeout   *time.Duration
	Dimension *int
}
type openaiEmbedder struct {
	conf     *OpenAIEmbeddingConfig
	embedder *openai.Embedder
}

func (o *openaiEmbedder) New(ctx context.Context, cfg *model.Model, opts ...EmbeddingOption) (EmbeddingService, error) {
	if err := checkOpenAICfg(cfg); err != nil {
		return nil, err
	}

	options := &EmbeddingOptions{}
	for _, opt := range opts {
		opt(options)
	}

	config := &OpenAIEmbeddingConfig{
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.ModelName,
		Timeout:   options.Timeout,
		Dimension: &cfg.Dimension,
	}

	embeder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     config.APIKey,
		BaseURL:    config.BaseURL,
		Model:      config.Model,
		Timeout:    *options.Timeout,
		Dimensions: &cfg.Dimension,
	})
	if err != nil {
		return nil, err
	}
	return &openaiEmbedder{
		conf:     config,
		embedder: embeder,
	}, nil
}

func (s *openaiEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	return s.embedder.EmbedStrings(ctx, texts, opts...)
}

func (o *openaiEmbedder) GetDimension() int {
	return *o.conf.Dimension
}

func (o *openaiEmbedder) GetType() string {
	return ProviderOpenAI
}

func checkOpenAICfg(cfg *model.Model) error {
	if cfg.BaseURL == "" {
		return fmt.Errorf("opnai base URL cannot be empty")
	}

	if _, err := url.Parse(cfg.BaseURL); err != nil {
		return fmt.Errorf("invalid openai base URL: %w", err)
	}

	if cfg.ModelName == "" {
		return fmt.Errorf("openai model name cannot be empty")
	}

	if cfg.Dimension <= 0 {
		return fmt.Errorf("openai embedding dimension must be positive")
	}

	return nil
}
