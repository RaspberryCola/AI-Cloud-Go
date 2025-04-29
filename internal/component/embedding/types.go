package embedding

import (
	"fmt"
)

const (
	ProviderOpenAI = "openai"
	ProviderOllama = "ollama"
)

type EmbeddingModelConfig struct {
	Type   string                 `mapstructure:"type"`
	OpenAI *OpenAIEmbeddingConfig `mapstructure:"openai"`
	Ollama *OllamaEmbeddingConfig `mapstructure:"ollama"`
}

func (e *EmbeddingModelConfig) CheckCfg() error {
	if e.Type == "" {
		return fmt.Errorf("embedding model type is empty")
	}
	return nil
}
