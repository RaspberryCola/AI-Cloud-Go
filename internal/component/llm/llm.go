package llmfactory

import (
	"ai-cloud/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	eino_model "github.com/cloudwego/eino/components/model"
	ollama_api "github.com/ollama/ollama/api"

	// 假设这些是你的 Ollama 和 OpenAI 客户端包
	"ai-cloud/internal/component/llm/ollama"
	"ai-cloud/internal/component/llm/openai"
)

const (
	defaultLLMTimeout = 60 * time.Second
	serverOllama      = "ollama"
	serverOpenAI      = "openai"
	modelTypeLLM      = "llm"
)

// GetLLMClient 使用 传入的 配置基础，并允许通过 clientDefaultOpts 设置客户端级别的默认调用选项。
func GetLLMClient(ctx context.Context, cfg *model.Model) (eino_model.ToolCallingChatModel, error) {
	// 检查Model配置
	// TODO: 考虑通过check函数实现
	if cfg == nil {
		return nil, errors.New("input model configuration is nil")
	}
	if cfg.Type != modelTypeLLM {
		return nil, fmt.Errorf("model type is '%s', but expected '%s'", cfg.Type, modelTypeLLM)
	}

	// 2. 返回对应的server
	switch strings.ToLower(cfg.Server) {
	case serverOllama:
		ollamaCfg := &ollama.ChatModelConfig{
			BaseURL: cfg.BaseURL,
			Model:   cfg.ModelName, // 使用最终确定的模型名称
			Timeout: defaultLLMTimeout,
			Options: &ollama_api.Options{}, // 设置包含默认调用参数的 Options
		}
		return ollama.NewChatModel(ctx, ollamaCfg)

	case serverOpenAI:
		openAICfg := &openai.ChatModelConfig{
			APIKey:  cfg.APIKey,
			Model:   cfg.ModelName,
			Timeout: defaultLLMTimeout,
			BaseURL: cfg.BaseURL,
		}
		return openai.NewChatModel(ctx, openAICfg)

	default:
		return nil, fmt.Errorf("unsupported LLM server type: '%s'", cfg.Server)
	}
}
