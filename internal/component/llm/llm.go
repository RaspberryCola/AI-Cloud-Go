package llmfactory

import (
	"ai-cloud/internal/model"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	eino_model "github.com/cloudwego/eino/components/model"
	eino_model_options "github.com/cloudwego/eino/components/model"
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

// GetLLMClient 使用 dbModel 配置基础，并允许通过 clientDefaultOpts 设置客户端级别的默认调用选项。
func GetLLMClient(ctx context.Context, cfg *model.Model, clientDefaultOpts ...eino_model_options.Option) (eino_model.ToolCallingChatModel, error) {
	// 检查Model配置
	// TODO: 考虑通过check函数实现
	if cfg == nil {
		return nil, errors.New("input model configuration is nil")
	}
	if cfg.Type != modelTypeLLM {
		return nil, fmt.Errorf("model type is '%s', but expected '%s'", cfg.Type, modelTypeLLM)
	}

	// 1. 解析传入的 clientDefaultOpts，形成客户端级别的默认调用参数
	// BaseOptions 的字段将从 clientDefaultOpts 中填充
	baseCallOpts := eino_model_options.GetCommonOptions(nil, clientDefaultOpts...)

	// cfg中的基础值，代表Model的一些限制参数，这些值可能在调用时被 baseCallOpts覆盖
	// 模型的初始化下的最大输出长度，默认通过模型cfg配置，在opts中可以覆盖（但是需要检查是否超过cfg）
	finalMaxOutputLength := cfg.MaxOutputLength // 这是 GORM 模型中的字段，对应 LLM 输出长度
	if baseCallOpts.MaxTokens != nil {
		optMaxTokens := *baseCallOpts.MaxTokens
		if optMaxTokens > 0 && optMaxTokens <= finalMaxOutputLength {
			finalMaxOutputLength = optMaxTokens
		}
	}

	// 2. 返回对应的server
	switch strings.ToLower(cfg.Server) {
	case serverOllama:
		// Ollama API 选项，先从 cfg (经过 baseCallOpts 可能的覆盖) 初始化
		ollamaAPIOpts := &ollama_api.Options{
			NumPredict: finalMaxOutputLength,
		}
		// 应用 baseCallOpts 中的 Temperature, TopP, Stop 等
		if baseCallOpts.Temperature != nil {
			ollamaAPIOpts.Temperature = *baseCallOpts.Temperature
		}
		if baseCallOpts.TopP != nil {
			ollamaAPIOpts.TopP = *baseCallOpts.TopP
		}
		if len(baseCallOpts.Stop) > 0 {
			ollamaAPIOpts.Stop = baseCallOpts.Stop
		}
		// 其他 ollama_api.Options 字段也可以类似处理

		ollamaCfg := &ollama.ChatModelConfig{
			BaseURL: cfg.BaseURL,
			Model:   cfg.ModelName, // 使用最终确定的模型名称
			Timeout: defaultLLMTimeout,
			Options: ollamaAPIOpts, // 设置包含默认调用参数的 Options
		}
		return ollama.NewChatModel(ctx, ollamaCfg)

	case serverOpenAI:
		openAICfg := &openai.ChatModelConfig{
			APIKey:    cfg.APIKey,
			Model:     cfg.ModelName,
			Timeout:   defaultLLMTimeout,
			BaseURL:   cfg.BaseURL,
			MaxTokens: &finalMaxOutputLength,
		}
		// 应用 baseCallOpts 中的 Temperature, TopP, MaxTokens (输出), Stop 等
		if baseCallOpts.Temperature != nil {
			openAICfg.Temperature = baseCallOpts.Temperature // OpenAI config 直接用指针
		}
		if baseCallOpts.TopP != nil {
			openAICfg.TopP = baseCallOpts.TopP
		}

		if len(baseCallOpts.Stop) > 0 {
			openAICfg.Stop = baseCallOpts.Stop
		}
		// TODO: 需要修改能够传递公共opts中没有的配置

		return openai.NewChatModel(ctx, openAICfg)

	default:
		return nil, fmt.Errorf("unsupported LLM server type: '%s'", cfg.Server)
	}
}
