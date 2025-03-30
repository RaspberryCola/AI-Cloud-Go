package service

import (
	"ai-cloud/internal/dao"
	"ai-cloud/internal/model"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/weaviate/weaviate/entities/models"
	"strings"
	"time"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

type KBService interface {
	CreateDB(name, description string, userID uint) error // 创建知识库
	DeleteKB(userID uint, kbid string) error              // 删除知识库
	// TODO：修改知识库（名称、说明）
	PageList(userID uint, page int, size int) (int64, []model.KnowledgeBase, error)     // 获取知识库列表
	CreateDocument(userID uint, kbID string, file *model.File) (*model.Document, error) // 添加File到知识库
	ProcessDocument(doc *model.Document) error                                          // 解析嵌入文档（后续需要细化）
	// TODO：检索知识库 Retrieve
	// TODO: 移动Document到其他知识库

}

type kbService struct {
	kbDao       dao.KnowledgeBaseDao
	embedder    *openai.LLM
	weaviate    *weaviate.Client
	fileService FileService
}

func NewKBService(kbDao dao.KnowledgeBaseDao, embedder *openai.LLM, fileService FileService) KBService {
	cfg := weaviate.Config{
		Host:   "localhost:8081", // Weaviate 默认地址
		Scheme: "http",
	}
	client := weaviate.New(cfg)

	return &kbService{
		kbDao:       kbDao,
		embedder:    embedder,
		weaviate:    client,
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
		textsplitter.WithChunkOverlap(50),
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

	// 6. 将 chunks 存储到 Weaviate
	if err := ks.saveChunksToWeaviate(chunks); err != nil {
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

func (ks *kbService) saveChunksToWeaviate(chunks []model.Chunk) error {
	ctx := context.Background()

	// 确保 class 存在
	classObj := &models.Class{
		Class: "TextChunk",
		Properties: []*models.Property{
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "documentId",
				DataType: []string{"string"},
			},
			{
				Name:     "index",
				DataType: []string{"int"},
			},
			{
				Name:     "kbId",
				DataType: []string{"string"},
			},
		},
		Vectorizer: "none", // 因为我们自己提供向量
	}

	// 创建 class（如果不存在）
	err := ks.weaviate.Schema().ClassCreator().WithClass(classObj).Do(ctx)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("创建 schema 失败: %w", err)
	}

	// 批量添加数据
	batcher := ks.weaviate.Batch().ObjectsBatcher()
	for _, chunk := range chunks {
		fmt.Printf("Chunk ID: %s, Embeddings: %v\n", chunk.ID, chunk.Embeddings)
		properties := map[string]interface{}{
			"content":    chunk.Content,
			"documentId": chunk.DocumentID,
			"index":      chunk.Index,
			"kbId":       chunk.KBID,
		}

		obj := &models.Object{
			Class:      "TextChunk",
			ID:         strfmt.UUID(chunk.ID),
			Properties: properties,
			Vector:     chunk.Embeddings,
		}

		batcher = batcher.WithObject(obj)
	}

	resp, err := batcher.Do(ctx)
	if err != nil {
		return fmt.Errorf("批量插入数据失败: %w", err)
	}

	// 检查是否所有对象都成功创建
	for _, result := range resp {
		if result.Result.Errors != nil {
			return fmt.Errorf("部分数据插入失败: %w", result.Result.Errors)
		}
	}
	return nil
}
