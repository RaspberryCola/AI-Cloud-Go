package devtest

import (
	llmfactory "ai-cloud/internal/component/llm"
	"ai-cloud/internal/model"
	"context"
	"fmt"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	eino_model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func RunAgentDemo() error {
	baseURL, err := requiredEnv("DEVTEST_LLM_BASE_URL")
	if err != nil {
		return err
	}
	apiKey, err := requiredEnv("DEVTEST_LLM_API_KEY")
	if err != nil {
		return err
	}
	modelName := optionalEnv("DEVTEST_LLM_MODEL", "deepseek-chat")
	mcpURL, err := requiredEnv("DEVTEST_MCP_SSE_URL")
	if err != nil {
		return err
	}

	ctx := context.Background()

	modelCfg := &model.Model{
		ID:              "devtest-llm",
		UserID:          1,
		Type:            "llm",
		ShowName:        "devtest-llm",
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

	options := []eino_model.Option{
		eino_model.WithMaxTokens(8192),
		eino_model.WithTemperature(0.8),
		eino_model.WithTopP(0.9),
	}

	messages := []*schema.Message{
		{Role: schema.User, Content: "查询南京今天天气？"},
	}

	cli, err := client.NewSSEMCPClient(mcpURL)
	if err != nil {
		return err
	}
	if err := cli.Start(ctx); err != nil {
		return err
	}

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "devtest-client",
		Version: "1.0.0",
	}

	if _, err := cli.Initialize(ctx, initRequest); err != nil {
		return err
	}

	tools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
	if err != nil {
		return err
	}

	ragent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: llm,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: tools},
	})
	if err != nil {
		return err
	}

	msg, err := ragent.Generate(ctx, messages,
		agent.WithComposeOptions(
			compose.WithChatModelOption(
				eino_model.WithTemperature(0.8),
				eino_model.WithTopP(0.9),
			),
		))
	if err != nil {
		return err
	}

	fmt.Println("\nAgent输出：", msg.String())
	_ = options
	return nil
}
