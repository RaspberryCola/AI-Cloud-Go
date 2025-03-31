package service

import (
	"ai-cloud/config"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"ai-cloud/internal/storage"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"context"
	"errors"
	"fmt"
	"io"
	"time"

	openaiEmbed "github.com/cloudwego/eino-ext/components/embedding/openai"

	"github.com/cloudwego/eino/components/document"
)

type KBService interface {
	CreateDB(name, description string, userID uint) error // 创建知识库
	DeleteKB(userID uint, kbid string) error              // 删除知识库
	// TODO：修改知识库（名称、说明）
	PageList(userID uint, page int, size int) (int64, []model.KnowledgeBase, error)     // 获取知识库列表
	CreateDocument(userID uint, kbID string, file *model.File) (*model.Document, error) // 添加File到知识库
	ProcessDocument(doc *model.Document) error                                          // 解析嵌入文档（后续需要细化）
	Retrieve(userID uint, kbID string, query string, topK int) ([]model.Chunk, error)   // 获取检索的Chunks
	RAGQuery(userID uint, query string, kbIDs []string) (*model.ChatResponse, error)    // 新增RAG查询方法
	RAGQueryStream(ctx context.Context, userID uint, query string, kbIDs []string) (<-chan *model.ChatStreamResponse, error)
	// TODO: 移动Document到其他知识库

}

type kbService struct {
	kbDao         dao.KnowledgeBaseDao
	milvusDao     dao.MilvusDao
	fileService   FileService
	storageDriver storage.Driver
	embedder      *openaiEmbed.Embedder
	llm           *openai.ChatModel
}

func NewKBService(kbDao dao.KnowledgeBaseDao, milvusDao dao.MilvusDao, fileService FileService) KBService {
	ctx := context.Background()

	cfg := config.AppConfigInstance.Storage
	driver, err := storage.NewDriver(cfg)
	if err != nil {
		panic("无法连接到存储服务: " + err.Error())
	}

	// embedder
	dimesion := 1024
	embedder, _ := openaiEmbed.NewEmbedder(ctx, &openaiEmbed.EmbeddingConfig{
		APIKey:     "sk-98077dd2f6d74722ba818a4d52e6dee9",
		Model:      "text-embedding-v3",
		BaseURL:    "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Timeout:    30 * time.Second,
		Dimensions: &dimesion,
	})

	// llm
	// 初始化LLM
	maxTokens := 4096
	var temp float32 = 0.7
	llm, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     "https://ark.cn-beijing.volces.com/api/v3",
		APIKey:      "8d58ce8d-bc3b-4d03-a090-d43da03040b2",
		Model:       "deepseek-v3-250324",
		MaxTokens:   &maxTokens,
		Temperature: &temp,
	})

	return &kbService{
		kbDao:         kbDao,
		milvusDao:     milvusDao,
		fileService:   fileService,
		storageDriver: driver,
		embedder:      embedder,
		llm:           llm,
	}
}
func (ks *kbService) CreateDB(name, description string, userID uint) error {
	if name == "" {
		return errors.New("知识库名称不能为空")
	}

	kb := &model.KnowledgeBase{
		ID:          GenerateUUID(),
		Name:        name,
		Description: description,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := ks.kbDao.CreateKB(kb); err != nil {
		return errors.New("知识库创建失败")
	}
	return nil
}

func (ks *kbService) DeleteKB(userID uint, kbid string) error {

	kb, err := ks.kbDao.GetKBByID(kbid)
	if err != nil {
		return errors.New("知识库不存在")
	}
	if kb.UserID != userID {
		return errors.New("无删除权限")
	}

	if err := ks.kbDao.DeleteKB(kbid); err != nil {
		return errors.New("知识库删除失败")
	}
	return nil
}

func (ks *kbService) PageList(userId uint, page int, size int) (int64, []model.KnowledgeBase, error) {
	total, err := ks.kbDao.CountKBs(userId)
	if err != nil {
		return 0, nil, err
	}
	kbs, err := ks.kbDao.ListKBs(userId, page, size)
	if err != nil {
		return 0, nil, err
	}
	return total, kbs, err
}

func (ks *kbService) CreateDocument(userID uint, kbID string, file *model.File) (*model.Document, error) {
	doc := &model.Document{
		ID:              GenerateUUID(),
		UserID:          userID,
		KnowledgeBaseID: kbID,
		FileID:          file.ID,
		Title:           file.Name,
		DocType:         file.MIMEType,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := ks.kbDao.CreateDocument(doc); err != nil {
		return nil, errors.New("知识库文档创建失败")
	}
	return doc, nil
}

// 解析一个文件
func (ks *kbService) ProcessDocument(doc *model.Document) error {
	ctx := context.Background()

	// 1. 获取File
	f := &model.File{}
	f, err := ks.fileService.GetFileByID(doc.FileID)
	if err != nil {
		return fmt.Errorf("获取文件失败: %w", err)
	}

	fURL, _ := ks.storageDriver.GetURL(f.StorageKey)

	// 2. Loader 加载文档，获取schema.Document
	loader, err := url.NewLoader(ctx, nil)
	if err != nil {
		return fmt.Errorf("创建Loader失败: %w", err)
	}
	docs, err := loader.Load(ctx, document.Source{
		URI: fURL,
	})
	if err != nil {
		return fmt.Errorf("加载文档失败: %w", err)
	}
	for _, d := range docs {
		d.ID = f.Name
	}

	// 3. 文本分割
	var chunks []model.Chunk

	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   500,
		OverlapSize: 300,
	})
	if err != nil {
		return fmt.Errorf("加载分块器失败: %w", err)
	}

	texts, err := splitter.Transform(ctx, docs)
	if err != nil {
		return fmt.Errorf("分块失败: %w", err)
	}

	for i, text := range texts {
		textString := []string{text.Content}
		vectors64, _ := ks.embedder.EmbedStrings(ctx, textString)
		float32Vectors := ConvertFloat64ToFloat32Embeddings(vectors64)
		chunk := model.Chunk{
			ID:         GenerateUUID(),
			Content:    text.Content,
			KBID:       doc.KnowledgeBaseID,
			DocumentID: doc.ID,
			Index:      i,
			Embeddings: float32Vectors[0],
		}
		chunks = append(chunks, chunk)
	}

	// 4. 将 chunks 存储到 Milvus
	if err := ks.milvusDao.SaveChunks(chunks); err != nil {
		return fmt.Errorf("存储向量到 Milvus 失败: %w", err)
	}

	// 5. 更新文档状态
	doc.Status = 2 // 已完成
	doc.UpdatedAt = time.Now()
	if err := ks.kbDao.UpdateDocument(doc); err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	return nil
}

func ConvertFloat64ToFloat32Embeddings(embeddings [][]float64) [][]float32 {
	float32Embeddings := make([][]float32, len(embeddings))
	for i, vec64 := range embeddings {
		vec32 := make([]float32, len(vec64))
		for j, v := range vec64 {
			vec32[j] = float32(v)
		}
		float32Embeddings[i] = vec32
	}
	return float32Embeddings
}

func (ks *kbService) Retrieve(userID uint, kbID string, query string, topK int) ([]model.Chunk, error) {

	ctx := context.Background()

	// 1. 权限校验
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}
	if kb.UserID != userID {
		return nil, errors.New("无访问权限")
	}

	queryList := []string{query}

	// 2. 向量化query
	embeddings, err := ks.embedder.EmbedStrings(ctx, queryList)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 3. Milvus向量检索
	float32Vector := ConvertFloat64ToFloat32Embeddings(embeddings)[0]

	return ks.milvusDao.Search(kbID, float32Vector, topK)
}

// RAGQuery 实现RAG查询
func (ks *kbService) RAGQuery(userID uint, query string, kbIDs []string) (*model.ChatResponse, error) {
	ctx := context.Background()

	// 1. 权限校验
	for _, kbID := range kbIDs {
		kb, err := ks.kbDao.GetKBByID(kbID)
		if err != nil {
			return nil, fmt.Errorf("知识库不存在: %w", err)
		}
		if kb.UserID != userID {
			return nil, errors.New("无访问权限")
		}
	}

	// 2. 从每个知识库检索相关内容
	var allChunks []model.Chunk
	for _, kbID := range kbIDs {
		// TODO：后续要改成从所有知识库中检索最相关的几个片段
		chunks, err := ks.Retrieve(userID, kbID, query, 3) // 每个知识库取top3相关内容
		if err != nil {
			return nil, err
		}
		allChunks = append(allChunks, chunks...)
	}
	// 3. 构建提示词
	var context string
	for _, chunk := range allChunks {
		context += chunk.Content + "\n"
	}

	systemPrompt := "你是一个知识库助手。请基于以下参考内容回答用户问题。如果无法从参考内容中得到答案，请明确告知。\n参考内容:\n" + context

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(query),
	}

	// 4. 调用LLM生成回答
	response, err := ks.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	return &model.ChatResponse{
		Response:   response.Content,
		References: allChunks,
	}, nil
}

// RAGQueryStream 实现流式RAG查询
func (ks *kbService) RAGQueryStream(ctx context.Context, userID uint, query string, kbIDs []string) (<-chan *model.ChatStreamResponse, error) {
	// 创建响应通道
	responseChan := make(chan *model.ChatStreamResponse)

	// 1. 权限校验
	for _, kbID := range kbIDs {
		kb, err := ks.kbDao.GetKBByID(kbID)
		if err != nil {
			return nil, fmt.Errorf("知识库不存在: %w", err)
		}
		if kb.UserID != userID {
			return nil, errors.New("无访问权限")
		}
	}

	// 2. 从每个知识库检索相关内容
	var allChunks []model.Chunk
	for _, kbID := range kbIDs {
		chunks, err := ks.Retrieve(userID, kbID, query, 3)
		if err != nil {
			return nil, err
		}
		allChunks = append(allChunks, chunks...)
	}

	// 3. 构建提示词
	var context string
	for _, chunk := range allChunks {
		context += chunk.Content + "\n"
	}

	systemPrompt := "你是一个知识库助手。请基于以下参考内容回答用户问题。如果无法从参考内容中得到答案，请明确告知。\n参考内容:\n" + context

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(query),
	}

	// 4. 启动goroutine处理流式响应
	go func() {
		defer close(responseChan)

		reader, err := ks.llm.Stream(ctx, messages)
		if err != nil {
			return
		}
		defer reader.Close()

		id := GenerateUUID()
		created := time.Now().Unix()
		for {
			chunk, err := reader.Recv()
			if err != nil {
				// Send a final message with finish_reason if it's EOF
				if err == io.EOF {
					stop := "stop"
					response := &model.ChatStreamResponse{
						ID:      id,
						Object:  "chat.completion.chunk",
						Created: created,
						Model:   "deepseek-v3-250324",
						Choices: []model.ChatStreamChoice{
							{
								Delta:        model.ChatStreamDelta{},
								Index:        0,
								FinishReason: &stop,
							},
						},
					}
					responseChan <- response
				}
				break
			}

			response := &model.ChatStreamResponse{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   "deepseek-v3-250324",
				Choices: []model.ChatStreamChoice{
					{
						Delta: model.ChatStreamDelta{
							Content: chunk.Content,
						},
						Index:        0,
						FinishReason: nil,
					},
				},
			}

			select {
			case <-ctx.Done():
				return
			case responseChan <- response:
			}
		}

	}()

	return responseChan, nil
}
