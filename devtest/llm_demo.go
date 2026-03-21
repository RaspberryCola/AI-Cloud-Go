package devtest

import (
	llmfactory "ai-cloud/internal/component/llm"
	"ai-cloud/internal/model"
	"context"
	"fmt"

	eino_model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func RunLLMDemo() error {
	baseURL, err := requiredEnv("DEVTEST_LLM_BASE_URL")
	if err != nil {
		return err
	}
	apiKey, err := requiredEnv("DEVTEST_LLM_API_KEY")
	if err != nil {
		return err
	}
	modelName := optionalEnv("DEVTEST_LLM_MODEL", "deepseek-chat")

	ctx := context.Background()

	modelCfg := &model.Model{
		ID:              "devtest-llm",
		UserID:          1,
		Type:            "llm",
		ShowName:        modelName,
		Server:          "openai",
		BaseURL:         baseURL,
		ModelName:       modelName,
		APIKey:          apiKey,
		MaxOutputLength: 8192,
		Function:        true,
		MaxTokens:       8192,
	}

	llm, err := llmfactory.GetLLMClient(ctx, modelCfg)
	if err != nil {
		return err
	}

	messages := []*schema.Message{
		{Role: schema.User, Content: "你好，请写一首关于春天的短诗。"},
	}

	options := []eino_model.Option{
		eino_model.WithMaxTokens(8192),
		eino_model.WithTemperature(0.8),
		eino_model.WithTopP(0.9),
	}

	res, err := llm.Generate(ctx, messages, options...)
	if err != nil {
		return err
	}

	fmt.Println(res)
	return nil
}
