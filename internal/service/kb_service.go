package service

import (
	"ai-cloud/config"
	"ai-cloud/internal/component/embedding"
	mindexer "ai-cloud/internal/component/indexer/milvus"
	"ai-cloud/internal/component/parser/pdf"
	mretriever "ai-cloud/internal/component/retriever/milvus"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"ai-cloud/internal/model"
	"ai-cloud/internal/storage"
	"context"
	"errors"
	"fmt"
	einoRetriever "github.com/cloudwego/eino/components/retriever"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	//"github.com/cloudwego/eino-ext/components/embedding"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
)

type KBService interface {
	// 知识库
	CreateKB(userID uint, name, description, embedModelID string) error             // 创建知识库
	DeleteKB(userID uint, kbID string) error                                        // 删除知识库
	PageList(userID uint, page int, size int) (int64, []model.KnowledgeBase, error) // 获取知识库列表
	GetKBDetail(userID uint, kbID string) (*model.KnowledgeBase, error)             // 获取知识库详情

	// 文档
	CreateDocument(userID uint, kbID string, file *model.File) (*model.Document, error) // 添加File到知识库
	ProcessDocument(ctx context.Context, userID uint, kbID string, doc *model.Document) error
	DocList(userID uint, kbID string, page int, size int) (int64, []model.Document, error) // 获取知识库下的文件列表
	DeleteDocs(userID uint, kbID string, docs []string) error                              // 批量删除文件

	// RAG
	RAGQuery(ctx context.Context, userID uint, query string, kbIDs []string) (*model.ChatResponse, error)                    // 新增RAG查询方法
	RAGQueryStream(ctx context.Context, userID uint, query string, kbIDs []string) (<-chan *model.ChatStreamResponse, error) // 流式对话
	Retrieve(ctx context.Context, userID uint, kbID string, query string, topK int) ([]*schema.Document, error)
	// TODO: 移动Document到其他知识库
	// TODO：修改知识库（名称、说明）
}

type kbService struct {
	kbDao    dao.KnowledgeBaseDao
	modelDao dao.ModelDao
	// milvusDao     dao.MilvusDao
	fileService   FileService
	storageDriver storage.Driver
	llm           *openai.ChatModel
	//embeddingService embedding.EmbeddingService
}

func NewKBService(kbDao dao.KnowledgeBaseDao, fileService FileService, modelDao dao.ModelDao) KBService {
	ctx := context.Background()

	cfg := config.AppConfigInstance.Storage
	driver, err := storage.NewDriver(cfg)
	if err != nil {
		panic("无法连接到存储服务: " + err.Error())
	}

	// 从配置文件获取LLM配置
	llmCfg := config.GetConfig().LLM
	maxTokens := llmCfg.MaxTokens
	temp := llmCfg.Temperature

	// 初始化LLM
	llm, _ := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     llmCfg.BaseURL,
		APIKey:      llmCfg.APIKey,
		Model:       llmCfg.Model,
		MaxTokens:   &maxTokens,
		Temperature: &temp,
	})

	return &kbService{
		kbDao: kbDao,
		//milvusDao:     milvusDao,
		modelDao:      modelDao,
		fileService:   fileService,
		storageDriver: driver,
		llm:           llm,
	}
}

func (ks *kbService) CreateKB(userID uint, name, description, embedModelID string) error {
	collectionName := fmt.Sprintf("embed_%s", embedModelID)
	collectionName = strings.ReplaceAll(collectionName, "-", "_")

	kb := &model.KnowledgeBase{
		ID:               GenerateUUID(),
		Name:             name,
		Description:      description,
		UserID:           userID,
		EmbedModelID:     embedModelID,
		MilvusCollection: collectionName,
	}

	// 保存知识库记录
	if err := ks.kbDao.CreateKB(kb); err != nil {
		return errors.New("知识库创建失败")
	}
	return nil
}

func (ks *kbService) DeleteKB(userID uint, kbID string) error {
	// 1. 获取知识库并验证权限
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return fmt.Errorf("获取待删除的知识库失败: %w", err)
	}
	if kb.UserID != userID {
		return errors.New("无权限删除该知识库")
	}

	collectionName := kb.MilvusCollection

	// 2. 获取知识库下的所有文档
	docs, err := ks.kbDao.GetAllDocsByKBID(kbID)
	if err != nil {
		return fmt.Errorf("获取知识库下的文档失败: %w", err)
	}
	// 3. 提取文档id
	var docIDs []string
	for _, doc := range docs {
		docIDs = append(docIDs, doc.ID)
	}

	// 4. 开启事务
	tx := ks.kbDao.GetDB().Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败：%w", tx.Error)
	}

	// 5.事务删除
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	mClient := database.GetMilvusClient()

	if len(docIDs) > 0 {
		if err := mindexer.DeleteDos(mClient, docIDs, collectionName); err != nil {
			tx.Rollback()
			return fmt.Errorf("删除向量数据失败: %w", err)
		}
	}

	// 5.2 删除文档记录
	if err := ks.kbDao.DeleteDocsByKBID(kbID); err != nil {
		tx.Rollback()
		return err
	}

	// 5.3 删除知识库记录
	if err := ks.kbDao.DeleteKB(kbID); err != nil {
		tx.Rollback()
		return err
	}
	// 6. 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败: %w", err)
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

func (ks *kbService) ProcessDocument(ctx context.Context, userID uint, kbID string, doc *model.Document) error {
	// 获取知识库信息
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return fmt.Errorf("获取知识库失败: %w", err)
	}

	// 获取model，构建EmbeddingService实例
	embedModel, err := ks.modelDao.GetByID(ctx, userID, kb.EmbedModelID)
	if err != nil {
		return fmt.Errorf("获取嵌入模型失败: %w", err)
	}

	// TODO: Timeout从配置中获取
	embeddingService, err := embedding.NewEmbeddingService(
		ctx,
		embedModel,
		embedding.WithTimeout(30*time.Second),
	)

	if err != nil {
		return fmt.Errorf("创建embedding服务实例失败: %w", err)
	}

	// 获取文件元信息（*model.File）
	f, err := ks.fileService.GetFileByID(doc.FileID)
	if err != nil {
		return fmt.Errorf("获取文件失败: %w", err)
	}
	// 获取文件在存储中的URL
	fURL, _ := ks.storageDriver.GetURL(f.StorageKey)
	fmt.Println("处理文档:", f.Name, "URL:", fURL)

	ext := strings.ToLower(filepath.Ext(f.Name))
	fmt.Println("文件扩展名:", ext)

	// Loader 加载文档，获取schema.Document
	var p parser.Parser
	switch ext {
	case ".pdf":
		p, err = pdf.NewDocconvPDFParser(ctx, nil)
		if err != nil {
			return fmt.Errorf("获取pdfparser失败：%v", err)
		}
	default:
		fmt.Println("未找到适合扩展名的解析器:", ext)
		p = nil
	}

	l, err := url.NewLoader(ctx, &url.LoaderConfig{Parser: p})

	if err != nil {
		return fmt.Errorf("创建Loader失败: %w", err)
	}
	docs, err := l.Load(ctx, document.Source{
		URI: fURL,
	})
	if err != nil {
		return fmt.Errorf("加载文档失败: %w", err)
	}
	fmt.Printf("文档加载成功，共%d个文档部分\n", len(docs))

	// Splitter 文本分割
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   config.AppConfigInstance.RAG.ChunkSize,
		OverlapSize: config.AppConfigInstance.RAG.OverlapSize,
	})
	if err != nil {
		return fmt.Errorf("加载分块器失败: %w", err)
	}

	texts, err := splitter.Transform(ctx, docs)
	if err != nil {
		return fmt.Errorf("分块失败: %w", err)
	}
	fmt.Printf("文本分块成功，共%d个文本块\n", len(texts))
	if len(texts) == 0 {
		return fmt.Errorf("文档解析未生成有效文本块，请检查文档内容或格式")
	}

	for i, d := range texts {
		d.ID = GenerateUUID()
		d.MetaData["kb_id"] = kbID
		d.MetaData["document_id"] = doc.ID
		d.MetaData["document_name"] = f.Name
		d.MetaData["chunk_index"] = i
	}

	// Indexer
	fmt.Println("开始构建Indexer")
	milvusIndexer, err := mindexer.NewMilvusIndexer(ctx, &mindexer.MilvusIndexerConfig{
		Client:     database.GetMilvusClient(),
		Collection: kb.MilvusCollection,
		Dimension:  embeddingService.GetDimension(),
		Embedding:  embeddingService,
	})
	fmt.Println("构建Indexer成功")
	if err != nil {
		return fmt.Errorf("创建milvus索引器失败: %w", err)
	}
	fmt.Println("开始Store")
	ids, err := milvusIndexer.Store(ctx, texts)
	if err != nil {
		return fmt.Errorf("向量索引失败: %w", err)
	}

	fmt.Printf("向量索引成功，共%d个向量\n", len(ids))
	println(ids)

	// 更新文档状态
	doc.Status = 2 // 已完成
	doc.UpdatedAt = time.Now()
	if err := ks.kbDao.UpdateDocument(doc); err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	return nil
}

func (ks *kbService) Retrieve(ctx context.Context, userID uint, kbID string, query string, topK int) ([]*schema.Document, error) {
	// 1. 权限校验
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}
	if kb.UserID != userID {
		return nil, errors.New("无访问权限")
	}

	// 2. 向量化query，使用抽象的嵌入服务接口
	embedModel, err := ks.modelDao.GetByID(ctx, userID, kb.EmbedModelID)
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
		TopK:           topK,
		ScoreThreshold: 0,
	}

	retriever, err := mretriever.NewMilvusRetriever(ctx, retrieverConf)
	if err != nil {
		return nil, fmt.Errorf("创建retriever实例失败: %w", err)
	}
	docs, err := retriever.Retrieve(ctx, query, einoRetriever.WithTopK(topK))
	if err != nil {
		return nil, fmt.Errorf("检索失败: %w", err)
	}
	return docs, nil
}

// RAGQuery 实现RAG查询
func (ks *kbService) RAGQuery(ctx context.Context, userID uint, query string, kbIDs []string) (*model.ChatResponse, error) {
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
	var allDocs []*schema.Document
	for _, kbID := range kbIDs {
		// TODO：后续要改成从所有知识库中检索最相关的几个片段
		doc, err := ks.Retrieve(ctx, userID, kbID, query, 3) // 每个知识库取top3相关内容
		if err != nil {
			return nil, err
		}
		allDocs = append(allDocs, doc...)
	}
	// 3. 构建提示词
	var content string
	for _, chunk := range allDocs {
		content += chunk.Content + "\n"
	}

	systemPrompt := "你是一个知识库助手。请基于以下参考内容回答用户问题。如果无法从参考内容中得到答案，请明确告知。\n参考内容:\n" + content

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
		References: allDocs,
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
	var allChunks []*schema.Document
	for _, kbID := range kbIDs {
		chunks, err := ks.Retrieve(ctx, userID, kbID, query, 5)
		if err != nil {
			return nil, err
		}
		allChunks = append(allChunks, chunks...)
	}

	// 3. 构建提示词
	var content string
	for _, chunk := range allChunks {
		content += chunk.Content + "\n"
	}

	systemPrompt := "你是一个有用的助手，你可以获取外部知识来回答用户问题，以下是可利用的知识内容。\n外部知识库内容:\n" + content
	query = "用户提问：" + query
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(query),
	}

	log.Printf("messages: %v", messages)

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

func (ks *kbService) DocList(userID uint, kbID string, page int, size int) (int64, []model.Document, error) {
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return 0, nil, fmt.Errorf("获取知识库失败：%v", err)
	}

	if kb.UserID != userID {
		return 0, nil, fmt.Errorf("无查看知识库权限: %v", err)
	}

	total, err := ks.kbDao.CountDocs(kbID)
	if err != nil {
		return 0, nil, fmt.Errorf("获取count错误：%v", err)
	}

	docs, err := ks.kbDao.ListDocs(kbID, page, size)
	if err != nil {
		return 0, nil, fmt.Errorf("获取docs错误：%v", err)
	}

	return total, docs, nil
}

func (ks *kbService) GetKBDetail(userID uint, kbID string) (*model.KnowledgeBase, error) {
	// 获取知识库信息
	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("获取知识库失败：%v", err)
	}

	// 验证权限
	if kb.UserID != userID {
		return nil, fmt.Errorf("无权限访问该知识库")
	}

	return kb, nil
}

func (ks *kbService) DeleteDocs(userID uint, kbID string, docIDs []string) error {
	mClient := database.GetMilvusClient()

	kb, err := ks.kbDao.GetKBByID(kbID)
	if err != nil {
		return fmt.Errorf("获取知识库失败：%v", err)
	}
	collectionName := kb.MilvusCollection
	// 开启事务
	tx := ks.kbDao.GetDB().Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启事务失败：%w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(docIDs) > 0 {
		if err := mindexer.DeleteDos(mClient, docIDs, collectionName); err != nil {
			tx.Rollback()
			return fmt.Errorf("删除向量数据失败：%w", err)
		}
	}

	if err := ks.kbDao.BatchDeleteDocs(userID, docIDs); err != nil {
		tx.Rollback()
		return fmt.Errorf("批量删除Doc失败：%w", err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("提交事务失败：%w", err)
	}
	return nil
}
