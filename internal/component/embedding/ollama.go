package embedding

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	einoEmbedding "github.com/cloudwego/eino/components/embedding"
	"time"

	"github.com/ollama/ollama/api"
	"net/http"
	"net/url"
)

func init() {
	register(ProviderOllama, &ollamaEmbedder{})
}

var (
	defaultBaseURL = "http://localhost:11434"
	defaultTimeout = 10 * time.Minute
	defaultModel   = "nomic-embed-text:latest"
	defaultDim     = 1024
)

type OllamaEmbeddingConfig struct {
	BaseURL    string
	Model      string
	Dimension  *int
	Timeout    *time.Duration
	HTTPClient *http.Client
}

type ollamaEmbedder struct {
	cli  *api.Client
	conf *OllamaEmbeddingConfig
}

func (o *ollamaEmbedder) New(ctx context.Context, model *model.Model, opts ...EmbeddingOption) (EmbeddingService, error) {
	options := &EmbeddingOptions{}
	for _, opt := range opts {
		opt(options)
	}

	config := &OllamaEmbeddingConfig{
		BaseURL:   model.BaseURL,
		Model:     model.ModelName,
		Dimension: &model.Dimension,
		Timeout:   options.Timeout,
	}

	if err := config.CheckCfg(); err != nil {
		return nil, err
	}

	//  构造 client
	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else {
		httpClient = &http.Client{Timeout: *config.Timeout}
	}

	// 构造url
	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// 创建 client
	cli := api.NewClient(baseURL, httpClient)

	return &ollamaEmbedder{
		cli:  cli,
		conf: config,
	}, nil
}

func (o *ollamaEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...einoEmbedding.Option) (
	embeddings [][]float64, err error) {
	req := &api.EmbedRequest{
		Model: o.conf.Model,
		Input: texts,
	}
	resp, err := o.cli.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	embeddings = make([][]float64, len(resp.Embeddings))
	for i, d := range resp.Embeddings {
		res := make([]float64, len(d))
		for j, emb := range d {
			res[j] = float64(emb)
		}
		embeddings[i] = res
	}

	return embeddings, nil
}

func (o *ollamaEmbedder) GetType() string {
	return ProviderOllama
}

//func (e *ollamaEmbedder) IsCallbacksEnabled() bool {
//	return true
//}

func (o *ollamaEmbedder) GetDimension() int {
	return *o.conf.Dimension
}

func (o *OllamaEmbeddingConfig) CheckCfg() error {
	if o.BaseURL == "" {
		o.BaseURL = defaultBaseURL
	}
	if o.Model == "" {
		o.Model = defaultModel
	}
	if o.Dimension == nil {
		o.Dimension = &defaultDim
	}
	if o.Timeout == nil {
		o.Timeout = &defaultTimeout
	}
	return nil
}
