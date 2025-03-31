package service

import (
	"ai-cloud/config"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"ai-cloud/internal/storage"

	"github.com/cloudwego/eino-ext/components/document/loader/url"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"

	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/openai"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/cloudwego/eino/components/document"
	//"github.com/cloudwego/eino/components/document/parser"
	//"github.com/cloudwego/eino/schema"
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
	kbDao         dao.KnowledgeBaseDao
	milvus        client.Client
	fileService   FileService
	storageDriver storage.Driver
	embedder      *openai.Embedder
}

func NewKBService(kbDao dao.KnowledgeBaseDao, fileService FileService) KBService {
	ctx := context.Background()
	// 初始化Milvus客户端
	milvusClient, err := client.NewClient(context.Background(), client.Config{
		Address:  "localhost:19530", // Milvus默认地址
		Username: "ai_cloud",
		Password: "aicloud666",
	})
	if err != nil {
		panic("无法连接到Milvus: " + err.Error())
	}
	cfg := config.AppConfigInstance.Storage
	driver, err := storage.NewDriver(cfg)
	if err != nil {
		panic("无法连接到存储服务: " + err.Error())
	}

	// embedder
	dimesion := 1024
	embedder, _ := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey:     "sk-98077dd2f6d74722ba818a4d52e6dee9",
		Model:      "text-embedding-v3",
		BaseURL:    "https://dashscope.aliyuncs.com/compatible-mode/v1",
		Timeout:    30 * time.Second,
		Dimensions: &dimesion,
	})

	return &kbService{
		kbDao:         kbDao,
		milvus:        milvusClient,
		fileService:   fileService,
		storageDriver: driver,
		embedder:      embedder,
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

	//_, fileData, err := ks.fileService.DownloadFile(doc.FileID)
	//if err != nil {
	//	return fmt.Errorf("下载文件失败: %w", err)
	//}

	//// 2. 根据文件类型选择合适的加载器
	//var docs []schema.Document
	//switch doc.DocType {
	//case "text/plain", "txt", "text/plain; charset=utf-8":
	//	loader := documentloaders.NewText(bytes.NewReader(fileData))
	//	docs, err = loader.Load(ctx)
	//case "application/pdf", "pdf":
	//	// 使用 UniDoc 处理 PDF
	//	//docs, err = ks.processPDFWithUniDoc(fileData)
	//default:
	//	return fmt.Errorf("不支持的文档类型: %s", doc.DocType)
	//}
	//
	//if err != nil {
	//	return fmt.Errorf("加载文档失败: %w", err)
	//}

	// 3. 文本分割
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

	//

	var chunks []model.Chunk

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
	if err := ks.saveChunksToMilvus(chunks); err != nil {
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
	embeddings, err := ks.embedder.EmbedStrings(ctx, queryList)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 3. Milvus向量检索
	collectionName := "text_chunks"

	// 构建搜索参数
	sp, _ := entity.NewIndexIvfFlatSearchParam(16) // nprobe=16
	float32Vector := ConvertFloat64ToFloat32Embeddings(embeddings)[0]

	// 构建搜索条件
	expr := fmt.Sprintf("kb_id == \"%s\"", kbID)

	// 执行搜索
	searchResult, err := ks.milvus.Search(
		ctx,
		collectionName,
		[]string{}, // 分区列表
		expr,       // 过滤表达式
		[]string{"id", "content", "document_id", "kb_id", "chunk_index"}, // 输出字段
		[]entity.Vector{entity.FloatVector(float32Vector)},               // 查询向量
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

		// 先检查类型再转换
		idCol, ok := result.IDs.(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for ID column: %T", result.IDs)
		}

		contentCol, ok := result.Fields.GetColumn("content").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for content column: %T", result.Fields[1])
		}

		docIDCol, ok := result.Fields.GetColumn("document_id").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for document ID column: %T", result.Fields[2])
		}

		kbIDCol, ok := result.Fields.GetColumn("kb_id").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for KB ID column: %T", result.Fields[3])
		}

		indexCol, ok := result.Fields.GetColumn("chunk_index").(*entity.ColumnInt32)
		if !ok {
			return nil, fmt.Errorf("unexpected type for index column: %T", result.Fields[4])
		}

		// 获取结果数量
		resultCount := idCol.Len()

		for i := 0; i < resultCount; i++ {
			id := idCol.Data()[i]
			content, err := contentCol.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取内容失败: %w", err)
			}

			docID, err := docIDCol.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取文档ID失败: %w", err)
			}

			kb_id, err := kbIDCol.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取知识库ID失败: %w", err)
			}

			index := indexCol.Data()[i]

			chunks = append(chunks, model.Chunk{
				ID:         id,
				Content:    content,
				KBID:       kb_id,
				DocumentID: docID,
				Index:      int(index),
			})
		}
	}

	return chunks, nil
}
