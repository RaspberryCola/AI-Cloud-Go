package service

import (
	llmfactory "ai-cloud/internal/component/llm"
	mretriever "ai-cloud/internal/component/retriever/milvus"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mcpp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
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
	StreamExecuteAgent(ctx context.Context, userID uint, agentID string, msg model.UserMessage) (*schema.StreamReader[*schema.Message], error)
}

type agentService struct {
	dao        dao.AgentDao
	modelSvc   ModelService
	kbSvc      KBService
	kbDao      dao.KnowledgeBaseDao
	modelDao   dao.ModelDao
	historySvc HistoryService
}

func NewAgentService(dao dao.AgentDao, modelSvc ModelService, kbSvc KBService, kbDao dao.KnowledgeBaseDao, modelDao dao.ModelDao, historySvc HistoryService) AgentService {
	return &agentService{
		dao:        dao,
		modelSvc:   modelSvc,
		kbSvc:      kbSvc,
		kbDao:      kbDao,
		modelDao:   modelDao,
		historySvc: historySvc,
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
		return "", fmt.Errorf("buildGraph失败：%w", err)
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

func (s *agentService) StreamExecuteAgent(ctx context.Context, userID uint, agentID string, msg model.UserMessage) (*schema.StreamReader[*schema.Message], error) {
	// 1.获取Agent配置
	agent, err := s.dao.GetByID(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}

	var agentSchema model.AgentSchema
	if err := json.Unmarshal([]byte(agent.AgentSchema), &agentSchema); err != nil {
		return nil, err
	}

	// 2.构建Graph
	graph, err := s.buildGraph(ctx, userID, agentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent graph：%w", err)
	}

	// 3.构建runner
	runner, err := graph.Compile(ctx, compose.WithGraphName("EinoAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, fmt.Errorf("failed to compile agent graph: %w", err)
	}

	// 执行stream
	sr, err := runner.Stream(ctx, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	return sr, nil
}

func (s *agentService) buildGraph(ctx context.Context, userID uint, agentSchema model.AgentSchema) (*compose.Graph[*model.UserMessage, *schema.Message], error) {
	// 1. 创建LLM
	llmModelCfg, err := s.modelSvc.GetModel(ctx, userID, agentSchema.LLMConfig.ModelID)
	if err != nil {
		return nil, fmt.Errorf("failed to create get model:%w", err)
	}
	llm, err := llmfactory.GetLLMClient(ctx, llmModelCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm client:%w", err)
	}

	// 2. 构建 Retriever
	multiRetriever := mretriever.MultiKBRetriever{
		KBIDs:    agentSchema.Knowledge.KnowledgeIDs,
		UserID:   userID,
		KBDao:    s.kbDao,
		ModelDao: s.modelDao,
		Ctx:      ctx,
		TopK:     agentSchema.Knowledge.TopK, // 默认返回前5个最相关的文档
	}

	// 3. 构建Tools
	tools := []tool.BaseTool{}
	// 3.1 加载MCPTools
	for _, serverURL := range agentSchema.MCP.Servers {
		cli, err := client.NewSSEMCPClient(serverURL)
		err = cli.Start(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create mcp client: %w", err)
		}
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
	// 3.2 加载系统和用户自定义Tools

	// 3.3 校验Tools

	// 4. 构建Agent
	agentConfig := &react.AgentConfig{
		ToolCallingModel: llm,
		MaxStep:          10,
	}
	// 只有在tools不为空时才绑定ToolsConfig
	if len(tools) > 0 {
		agentConfig.ToolsConfig = compose.ToolsNodeConfig{
			Tools: tools,
		}
	}

	agt, err := react.NewAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	if agt == nil {
		return nil, errors.New("react.NewAgent returned a nil agent instance")
	}

	agentLambda, _ := compose.AnyLambda(agt.Generate, agt.Stream, nil, nil)

	// 5. 构建提示词
	// TODO：优化提示词设计
	promptTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(agentSchema.Prompt),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("用户消息：{query}\n 参考信息：{documents}"),
	)

	// 6. 实现图编排
	graph := compose.NewGraph[*model.UserMessage, *schema.Message]()
	_ = graph.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(inputToQueryLambda), compose.WithNodeName("UserMessageToQuery"))
	_ = graph.AddChatTemplateNode(ChatTemplate, promptTemplate)
	_ = graph.AddRetrieverNode(Retriever, multiRetriever, compose.WithOutputKey("documents"))
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
		"query":   input.Query,
		"history": input.History,
		"date":    time.Now().Format(time.DateTime),
	}, nil
}
