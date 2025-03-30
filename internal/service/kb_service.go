package service

import (
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
)

type KBService interface {
	CreateDB(name, description string, userID uint) error // 创建知识库
	DeleteKB(userID uint, kbid string) error              // 删除知识库
	// TODO：修改知识库（名称、说明）
	PageList(userID uint, page int, size int) (int64, []model.KnowledgeBase, error)     // 获取知识库列表
	CreateDocument(userID uint, kbID string, file *model.File) (*model.Document, error) // 添加File到知识库
	ProcessDocument(doc *model.Document) error                                          // 解析嵌入文档（后续需要细化）
	Retrieve(userID uint, kbID string, query string, topK int) ([]model.Chunk, error)   // 新增检索方法
	// TODO: 移动Document到其他知识库

}

type kbService struct {
	kbDao    dao.KnowledgeBaseDao
	embedder *openai.LLM
	// weaviate    *weaviate.Client
	milvus      client.Client
	fileService FileService
}

func NewKBService(kbDao dao.KnowledgeBaseDao, embedder *openai.LLM, fileService FileService) KBService {
	// 初始化Milvus客户端
	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address:  "localhost:19530", // Milvus默认地址
		Username: "ai_cloud",
		Password: "aicloud666",
	})
	if err != nil {
		panic("无法连接到Milvus: " + err.Error())
	}

	return &kbService{
		kbDao:       kbDao,
		embedder:    embedder,
		milvus:      milvusClient,
		fileService: fileService,
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

func (ks *kbService) ProcessDocument(doc *model.Document) error {
	ctx := context.Background()

	// 1. 获取文件内容
	_, fileData, err := ks.fileService.DownloadFile(doc.FileID)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}

	// 2. 根据文件类型选择合适的加载器
	var loader documentloaders.Loader
	switch doc.DocType {
	case "text/plain", "txt", "text/plain; charset=utf-8":
		loader = documentloaders.NewText(bytes.NewReader(fileData))
	case "application/pdf", "pdf":
		reader := bytes.NewReader(fileData)
		loader = documentloaders.NewPDF(reader, int64(len(fileData)))
	default:
		return fmt.Errorf("不支持的文档类型: %s", doc.DocType)
	}

	// 3. 加载文档
	docs, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("加载文档失败: %w", err)
	}

	// 4. 文本分割
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(100),
	)

	// 从文档中提取文本并分割
	var chunks []model.Chunk

	for _, d := range docs {
		// 对每个文档的内容进行分割
		texts, err := splitter.SplitText(d.PageContent)
		if err != nil {
			return fmt.Errorf("分割文本失败: %w", err)
		}

		batchSize := 10

		for i := 0; i < len(texts); i += batchSize {
			end := i + batchSize
			if end > len(texts) {
				end = len(texts)
			}
			textBatch := texts[i:end]
			ebds, err := ks.embedder.CreateEmbedding(ctx, textBatch)
			if err != nil {
				return fmt.Errorf("创建嵌入失败: %w", err)
			}

			for j, embedding := range ebds {
				chunk := model.Chunk{
					ID:         GenerateUUID(),
					Content:    textBatch[j],
					KBID:       doc.KnowledgeBaseID,
					DocumentID: doc.ID,
					Index:      i + j,
					Embeddings: embedding,
				}
				chunks = append(chunks, chunk)
			}
		}
	}

	// 6. 将 chunks 存储到 Milvus
	if err := ks.saveChunksToMilvus(chunks); err != nil {
		return fmt.Errorf("存储向量到 Milvus 失败: %w", err)
	}

	// 7. 更新文档状态
	doc.Status = 2 // 已完成
	doc.UpdatedAt = time.Now()
	if err := ks.kbDao.UpdateDocument(doc); err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	return nil
}

func (ks *kbService) saveChunksToMilvus(chunks []model.Chunk) error {
	ctx := context.Background()
	collectionName := "text_chunks"

	// 检查集合是否存在
	exists, err := ks.milvus.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合存在失败: %w", err)
	}

	// 如果集合不存在，则创建
	if !exists {
		// 定义schema
		schema := &entity.Schema{
			CollectionName: collectionName,
			Description:    "存储文档分块和向量",
			AutoID:         false,
			Fields: []*entity.Field{
				{
					Name:       "id",
					DataType:   entity.FieldTypeVarChar,
					PrimaryKey: true,
					AutoID:     false,
					TypeParams: map[string]string{
						"max_length": "64",
					},
				},
				{
					Name:     "content",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "65535",
					},
				},
				{
					Name:     "document_id",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "64",
					},
				},
				{
					Name:     "kb_id",
					DataType: entity.FieldTypeVarChar,
					TypeParams: map[string]string{
						"max_length": "64",
					},
				},
				{
					Name:     "chunk_index",
					DataType: entity.FieldTypeInt32,
				},
				{
					Name:     "vector",
					DataType: entity.FieldTypeFloatVector,
					TypeParams: map[string]string{
						"dim": "1024",
					},
				},
			},
		}

		// 创建集合
		err = ks.milvus.CreateCollection(ctx, schema, 1) // 1是分片数
		if err != nil {
			return fmt.Errorf("创建集合失败: %w", err)
		}
	}

	// 准备插入数据
	var ids []string
	var contents []string
	var documentIDs []string
	var kbIDs []string
	var chunkIndices []int32
	var vectors [][]float32

	for _, chunk := range chunks {
		ids = append(ids, chunk.ID)
		contents = append(contents, chunk.Content)
		documentIDs = append(documentIDs, chunk.DocumentID)
		kbIDs = append(kbIDs, chunk.KBID)
		chunkIndices = append(chunkIndices, int32(chunk.Index))
		vectors = append(vectors, chunk.Embeddings)
	}

	// 创建插入的数据列
	idColumn := entity.NewColumnVarChar("id", ids)
	contentColumn := entity.NewColumnVarChar("content", contents)
	documentIDColumn := entity.NewColumnVarChar("document_id", documentIDs)
	kbIDColumn := entity.NewColumnVarChar("kb_id", kbIDs)
	indexColumn := entity.NewColumnInt32("chunk_index", chunkIndices)
	vectorColumn := entity.NewColumnFloatVector("vector", 1024, vectors)

	// 插入数据
	_, err = ks.milvus.Insert(ctx, collectionName, "", idColumn, contentColumn, documentIDColumn, kbIDColumn, indexColumn, vectorColumn)
	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	// 创建索引（如果不存在）
	index, err := ks.milvus.DescribeIndex(ctx, collectionName, "vector")
	if err != nil || index == nil {
		idx, err := entity.NewIndexIvfFlat(entity.L2, 128) // 使用IVF_FLAT索引
		if err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}

		err = ks.milvus.CreateIndex(ctx, collectionName, "vector", idx, false)
		if err != nil {
			return fmt.Errorf("创建索引失败: %w", err)
		}
	}

	// 加载集合到内存
	err = ks.milvus.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("加载集合失败: %w", err)
	}

	return nil
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
	embeddings, err := ks.embedder.CreateEmbedding(ctx, queryList)
	//query_embeds := embeddings[0]

	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 3. Milvus向量检索
	collectionName := "text_chunks"

	// 构建搜索参数
	sp, _ := entity.NewIndexIvfFlatSearchParam(16) // nprobe=16
	vector := embeddings[0]                        // 假设embeddings已经是[]float32

	// 构建搜索条件
	expr := fmt.Sprintf("kb_id == \"%s\"", kbID)

	// 执行搜索
	searchResult, err := ks.milvus.Search(
		ctx,
		collectionName,
		[]string{}, // 分区列表
		expr,       // 过滤表达式
		[]string{"id", "content", "document_id", "kb_id", "chunk_index"}, // 输出字段
		[]entity.Vector{entity.FloatVector(vector)},                      // 查询向量
		"vector",  // 向量字段名
		entity.L2, // 距离度量
		topK,      // topK
		sp,        // 搜索参数
	)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	// 4. 转换结果
	var chunks []model.Chunk
	for _, result := range searchResult {
		idColumn := result.IDs.(*entity.ColumnVarChar)
		contentColumn := result.Fields[1].(*entity.ColumnVarChar)
		docIDColumn := result.Fields[2].(*entity.ColumnVarChar)
		kbIDColumn := result.Fields[3].(*entity.ColumnVarChar)
		indexColumn := result.Fields[4].(*entity.ColumnInt32)

		// 获取结果数量
		resultCount := idColumn.Len()

		for i := 0; i < resultCount; i++ {
			id := idColumn.Data()[i]
			content, err := contentColumn.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取内容失败: %w", err)
			}

			docID, err := docIDColumn.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取文档ID失败: %w", err)
			}

			kb_ID, err := kbIDColumn.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取知识库ID失败: %w", err)
			}

			index := indexColumn.Data()[i]
			if err != nil {
				return nil, fmt.Errorf("获取索引失败: %w", err)
			}

			chunks = append(chunks, model.Chunk{
				ID:         id,
				Content:    content,
				KBID:       kb_ID,
				DocumentID: docID,
				Index:      int(index),
			})
		}
	}

	return chunks, nil
}
