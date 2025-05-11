package service

import (
	"ai-cloud/internal/component/embedding"
	llmfactory "ai-cloud/internal/component/llm"
	mretriever "ai-cloud/internal/component/retriever/milvus"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"ai-cloud/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"time"
)

const (
	InputToQuery   = "InputToQuery"
	InputToHistory = "InputToHistory"
	ChatTemplate   = "ChatTemplate"
	ChatModel      = "ChatModel"
	Retriever      = "Retriever"
	Agent          = "Agent"
)

type AgentService interface {
	CreateAgent(ctx context.Context, agent *model.Agent) error
	UpdateAgent(ctx context.Context, agent *model.Agent) error
	DeleteAgent(ctx context.Context, userID uint, agentID string) error
	GetAgent(ctx context.Context, userID uint, agentID string) (*model.Agent, error)
	ListAgents(ctx context.Context, userID uint) ([]*model.Agent, error)
	PageAgents(ctx context.Context, userID uint, page, size int) ([]*model.Agent, int64, error)
	ExecuteAgent(ctx context.Context, userID uint, agentID string, msg model.UserMessage) (string, error)
}

type agentService struct {
	dao      dao.AgentDao
	modelSvc ModelService
	kbSvc    KBService
	kbDao    dao.KnowledgeBaseDao
	modelDao dao.ModelDao
}

func NewAgentService(dao dao.AgentDao, modelSvc ModelService, kbSvc KBService, kbDao dao.KnowledgeBaseDao, modelDao dao.ModelDao) AgentService {
	return &agentService{
		dao:      dao,
		modelSvc: modelSvc,
		kbSvc:    kbSvc,
		kbDao:    kbDao,
		modelDao: modelDao,
	}
}

func (s *agentService) CreateAgent(ctx context.Context, agent *model.Agent) error {
	return s.dao.Create(ctx, agent)
}

func (s *agentService) UpdateAgent(ctx context.Context, agent *model.Agent) error {
	return s.dao.Update(ctx, agent)
}

func (s *agentService) DeleteAgent(ctx context.Context, userID uint, agentID string) error {
	return s.dao.Delete(ctx, userID, agentID)
}

func (s *agentService) GetAgent(ctx context.Context, userID uint, agentID string) (*model.Agent, error) {
	return s.dao.GetByID(ctx, userID, agentID)
}

func (s *agentService) ListAgents(ctx context.Context, userID uint) ([]*model.Agent, error) {
	return s.dao.List(ctx, userID)
}

func (s *agentService) PageAgents(ctx context.Context, userID uint, page, size int) ([]*model.Agent, int64, error) {
	return s.dao.Page(ctx, userID, page, size)
}

func (s *agentService) ExecuteAgent(ctx context.Context, userID uint, agentID string, msg model.UserMessage) (string, error) {
	// Retrieve the agent
	agent, err := s.dao.GetByID(ctx, userID, agentID)
	if err != nil {
		return "", err
	}

	// Parse the agent schema
	var agentSchema model.AgentSchema
	if err := json.Unmarshal([]byte(agent.AgentSchema), &agentSchema); err != nil {
		return "", err
	}

	graph, err := s.buildGraph(ctx, userID, agentSchema)
	if err != nil {
		return "", err
	}

	runner, err := graph.Compile(ctx, compose.WithGraphName("EinoAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))

	if err != nil {
		return "", err
	}

	res, err := runner.Invoke(ctx, &msg)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func (s *agentService) buildGraph(ctx context.Context, userID uint, agentSchema model.AgentSchema) (*compose.Graph[*model.UserMessage, *schema.Message], error) {
	// 获取LLM模型
	llmModelCfg, err := s.modelSvc.GetModel(ctx, userID, agentSchema.LLMConfig.ModelID)
	if err != nil {
		return nil, err
	}
	llm, err := llmfactory.GetLLMClient(ctx, llmModelCfg)
	if err != nil {
		return nil, err
	}

	// TODO：支持多知识库
	// 暂时只取一个知识库
	kbIDs := agentSchema.Knowledge.KnowledgeIDs
	kbID := kbIDs[0]
	kb, err := s.kbDao.GetKBByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}
	if kb.UserID != userID {
		return nil, errors.New("无访问权限")
	}
	embedModel, err := s.modelDao.GetByID(ctx, userID, kb.EmbedModelID)
	if err != nil {
		return nil, fmt.Errorf("获取嵌入模型失败: %w", err)
	}
	// TODO: Timeout从配置中获取
	embeddingService, err := embedding.NewEmbeddingService(
		ctx,
		embedModel,
		embedding.WithTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("创建embedding服务实例失败: %w", err)
	}

	retrieverConf := &mretriever.MilvusRetrieverConfig{
		Client:         database.GetMilvusClient(),
		Embedding:      embeddingService,
		Collection:     kb.MilvusCollection,
		KBIDs:          []string{kbID}, //TODO:后续需要考虑到不同知识库用的嵌入模型是不同的！
		SearchFields:   nil,
		TopK:           3,
		ScoreThreshold: 0,
	}

	retriever, err := mretriever.NewMilvusRetriever(ctx, retrieverConf)

	// 加载MCPTools
	tools := []tool.BaseTool{}

	for _, serverURL := range agentSchema.MCP.Servers {
		cli, err := client.NewSSEMCPClient(serverURL)
		err = cli.Start(ctx)
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "example-client",
			Version: "1.0.0",
		}

		_, err = cli.Initialize(ctx, initRequest)

		if err != nil {
			return nil, err
		}
		// 获取 mcpp 工具
		mcppTools, err := mcpp.GetTools(ctx, &mcpp.Config{Cli: cli})
		if err != nil {
			return nil, fmt.Errorf("failed to get mcpp tools: %w", err)
		}
		tools = append(tools, mcppTools...)
	}

	// 构建Agent
	agt, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: llm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: tools,
		},
		MaxStep: 10,
	})

	agentLambda, _ := compose.AnyLambda(agt.Generate, agt.Stream, nil, nil)

	// 构建提示词

	promptTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(agentSchema.Prompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{content}\n 参考信息：{documents}"),
	)

	// 开始编排
	graph := compose.NewGraph[*model.UserMessage, *schema.Message]()
	_ = graph.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(inputToQueryLambda), compose.WithNodeName("UserMessageToQuery"))
	_ = graph.AddChatTemplateNode(ChatTemplate, promptTemplate)
	_ = graph.AddRetrieverNode(Retriever, retriever, compose.WithOutputKey("documents"))
	_ = graph.AddLambdaNode(Agent, agentLambda, compose.WithNodeName("Agent"))
	_ = graph.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(inputToHistoryLambda), compose.WithNodeName("UserMessageToHistory"))

	_ = graph.AddEdge(compose.START, InputToQuery)
	_ = graph.AddEdge(compose.START, InputToHistory)
	_ = graph.AddEdge(InputToQuery, Retriever)
	_ = graph.AddEdge(Retriever, ChatTemplate)
	_ = graph.AddEdge(InputToHistory, ChatTemplate)
	_ = graph.AddEdge(ChatTemplate, Agent)
	_ = graph.AddEdge(Agent, compose.END)
	return graph, nil
}

// inputToQueryLambda component initialization function of node 'InputToQuery' in graph 'EinoAgent'
func inputToQueryLambda(ctx context.Context, input *model.UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// inputToHistoryLambda component initialization function of node 'InputToHistory' in graph 'EinoAgent'
func inputToHistoryLambda(ctx context.Context, input *model.UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"content": input.Query,
		"history": input.History,
		"date":    time.Now().Format(time.DateTime),
	}, nil
}
