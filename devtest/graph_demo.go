package devtest

import (
	"ai-cloud/config"
	"ai-cloud/internal/component/embedding"
	"ai-cloud/internal/component/retriever/milvus"
	"ai-cloud/internal/model"
	"context"
	"fmt"
	"time"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

const (
	InputToQuery   = "InputToQuery"
	InputToHistory = "InputToHistory"
	ChatTemplate   = "ChatTemplate"
	ChatModel      = "ChatModel"
	Retriever      = "Retriever"
)

var systemPrompt = `
# Role: Eino Expert Assistant

## Core Competencies
- knowledge of Eino framework and ecosystem
- Project scaffolding and best practices consultation
- Documentation navigation and implementation guidance

## Context Information
- Current Date: {date}
- Related Documents: |-
==== doc start ====
  {documents}
==== doc end ====
`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

var cbHandler callbacks.Handler

func RunGraphDemo() error {
	config.InitConfig()
	ctx := context.Background()

	llmBaseURL, err := requiredEnv("DEVTEST_LLM_BASE_URL")
	if err != nil {
		return err
	}
	llmAPIKey, err := requiredEnv("DEVTEST_LLM_API_KEY")
	if err != nil {
		return err
	}
	llmModel := optionalEnv("DEVTEST_LLM_MODEL", "deepseek-chat")
	milvusAddress := optionalEnv("DEVTEST_MILVUS_ADDRESS", "localhost:19530")
	embeddingBaseURL := optionalEnv("DEVTEST_EMBEDDING_BASE_URL", "http://localhost:11434")
	embeddingModel := optionalEnv("DEVTEST_EMBEDDING_MODEL", "nomic-embed-text")
	collection := requiredOrFallback("DEVTEST_COLLECTION", "")
	kbID := requiredOrFallback("DEVTEST_KB_ID", "")
	if collection == "" || kbID == "" {
		return fmt.Errorf("DEVTEST_COLLECTION and DEVTEST_KB_ID are required")
	}

	ptConfig := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", true),
			schema.UserMessage("{content}"),
		},
	}
	pt := prompt.FromMessages(schema.FString, ptConfig.Templates...)

	llm, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
		BaseURL: llmBaseURL,
		APIKey:  llmAPIKey,
		Model:   llmModel,
	})
	if err != nil {
		return err
	}

	cli, err := client.NewClient(ctx, client.Config{Address: milvusAddress})
	if err != nil {
		return err
	}

	embedder, err := embedding.NewEmbeddingService(ctx, &model.Model{
		ID:              "devtest-embedding",
		UserID:          1,
		Type:            "embedding",
		ShowName:        embeddingModel,
		Server:          "ollama",
		BaseURL:         embeddingBaseURL,
		ModelName:       embeddingModel,
		Dimension:       768,
		MaxTokens:       8192,
		APIKey:          "",
		Function:        false,
		MaxOutputLength: 4096,
	}, embedding.WithTimeout(30*time.Second))
	if err != nil {
		return err
	}

	retriever, err := milvus.NewMilvusRetriever(ctx, &milvus.MilvusRetrieverConfig{
		Client:     cli,
		Embedding:  embedder,
		Collection: collection,
		KBIDs:      []string{kbID},
		TopK:       3,
	})
	if err != nil {
		return err
	}

	graph := compose.NewGraph[*UserMessage, *schema.Message]()
	_ = graph.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(newLambda), compose.WithNodeName("UserMessageToQuery"))
	_ = graph.AddChatTemplateNode(ChatTemplate, pt)
	_ = graph.AddChatModelNode(ChatModel, llm, compose.WithNodeName("ChatModel"))
	_ = graph.AddRetrieverNode(Retriever, retriever, compose.WithOutputKey("documents"))
	_ = graph.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(newLambda2), compose.WithNodeName("UserMessageToVariables"))

	_ = graph.AddEdge(compose.START, InputToQuery)
	_ = graph.AddEdge(compose.START, InputToHistory)
	_ = graph.AddEdge(ChatModel, compose.END)
	_ = graph.AddEdge(InputToQuery, Retriever)
	_ = graph.AddEdge(Retriever, ChatTemplate)
	_ = graph.AddEdge(InputToHistory, ChatTemplate)
	_ = graph.AddEdge(ChatTemplate, ChatModel)

	runner, err := graph.Compile(ctx, compose.WithGraphName("EinoAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return err
	}

	userMessage := &UserMessage{
		ID:    "1",
		Query: "你是谁？",
		History: []*schema.Message{
			{Role: schema.User, Content: "你好"},
			{Role: schema.Assistant, Content: "你好！有什么可以帮你的吗？"},
		},
	}

	res, err := runner.Invoke(ctx, userMessage)
	if err != nil {
		return err
	}
	fmt.Println("response:", res)
	return nil
}

func requiredOrFallback(key, fallback string) string {
	value := optionalEnv(key, fallback)
	return value
}
