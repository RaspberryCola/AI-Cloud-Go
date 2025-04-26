package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/ollama/ollama/api"
)

var (
	defaultBaseURL = "http://localhost:11434"
	defaultTimeout = 10 * time.Minute
	defaultModel   = "nomic-embed-text:latest"
	defaultDim     = 1024
)

type OllamaEmbeddingConfig struct {
	Timeout    *time.Duration `json:"timeout"`
	HTTPClient *http.Client   `json:"http_client"`
	BaseURL    string         `json:"base_url"`
	Model      string         `json:"model"`
	Dimensions *int           `json:"dimensions,omitempty"`
}

var _ embedding.Embedder = (*OllamaEmbedder)(nil)

type OllamaEmbedder struct {
	cli  *api.Client
	conf *OllamaEmbeddingConfig
}

func NewOllamaEmbedder(ctx context.Context, config *OllamaEmbeddingConfig) (*OllamaEmbedder, error) {

	// options config
	if len(config.BaseURL) == 0 {
		config.BaseURL = defaultBaseURL
	}

	if config.Timeout == nil {
		config.Timeout = &defaultTimeout
	}

	if len(config.Model) == 0 {
		config.Model = defaultModel
	}

	if config.Dimensions == nil {
		config.Dimensions = &defaultDim
	}
	// client
	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else {
		httpClient = &http.Client{Timeout: *config.Timeout}
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	cli := api.NewClient(baseURL, httpClient)

	return &OllamaEmbedder{
		cli:  cli,
		conf: config,
	}, nil
}

func (e *OllamaEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) (
	embeddings [][]float64, err error) {
	req := &api.EmbedRequest{
		Model: e.conf.Model,
		Input: texts,
	}
	resp, err := e.cli.Embed(ctx, req)
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

const typ = "Ollama"

func (e *OllamaEmbedder) GetType() string {
	return typ
}

func (e *OllamaEmbedder) IsCallbacksEnabled() bool {
	return true
}
