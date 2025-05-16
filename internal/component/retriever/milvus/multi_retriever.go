package milvus

import (
	"ai-cloud/internal/component/embedding"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/database"
	"context"
	"fmt"
	eretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"log"
	"sort"
	"time"
)

// 自定义的多知识库Retriever
type MultiKBRetriever struct {
	KBIDs    []string
	UserID   uint
	KBDao    dao.KnowledgeBaseDao
	ModelDao dao.ModelDao
	Ctx      context.Context
	TopK     int
}

func (m MultiKBRetriever) Retrieve(ctx context.Context, query string, opts ...eretriever.Option) ([]*schema.Document, error) {
	// 如果没有提供知识库ID，则返回空结果
	if len(m.KBIDs) == 0 {
		return []*schema.Document{}, nil
	}

	// 保存所有文档结果
	allDocuments := []*schema.Document{}

	// 对每个知识库进行检索
	for _, kbID := range m.KBIDs {
		// 获取知识库信息
		kb, err := m.KBDao.GetKBByID(kbID)
		if err != nil {
			return nil, fmt.Errorf("knowledge base not found: %w", err)
		}
		if kb.UserID != m.UserID {
			return nil, fmt.Errorf("userID mismatch: %w", err)
		}

		// 获取Embedding模型
		embedModel, err := m.ModelDao.GetByID(m.Ctx, m.UserID, kb.EmbedModelID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve embedding model: %w", err)
		}

		// 创建Embedding服务
		embeddingService, err := embedding.NewEmbeddingService(
			m.Ctx,
			embedModel,
			embedding.WithTimeout(30*time.Second),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize embedding service: %w", err)
		}

		// 创建当前知识库的Retriever
		retrieverConf := &MilvusRetrieverConfig{
			Client:         database.GetMilvusClient(),
			Embedding:      embeddingService,
			Collection:     kb.MilvusCollection,
			KBIDs:          []string{kbID},
			SearchFields:   nil,
			TopK:           3,
			ScoreThreshold: 0,
		}

		retriever, err := NewMilvusRetriever(ctx, retrieverConf)
		if err != nil {
			return nil, fmt.Errorf("failed to create retriever for kb %s: %w", kbID, err)
		}

		// 执行检索
		docs, err := retriever.Retrieve(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve from kb %s: %w", kbID, err)
		}

		// 将结果添加到总结果中
		allDocuments = append(allDocuments, docs...)
	}

	// TODO：先简单按照分数返回。这是不合理的！需要用rerank！
	sort.Slice(allDocuments, func(i, j int) bool {
		scoreI, okI := allDocuments[i].MetaData["score"].(float64)
		scoreJ, okJ := allDocuments[j].MetaData["score"].(float64)

		// 如果score不存在或类型不正确，则将其视为最低优先级
		if !okI {
			return false
		}
		if !okJ {
			return true
		}

		return scoreI > scoreJ // 降序排序
	})

	if len(allDocuments) > m.TopK {
		allDocuments = allDocuments[:m.TopK]
	}
	log.Printf("[Multi Retriever] Retrieved %d documents from %d knowledge bases", len(allDocuments), len(m.KBIDs))
	return allDocuments, nil
}

func (m *MultiKBRetriever) GetType() string {
	return "MultiKnowledgeBaseRetriever"
}
