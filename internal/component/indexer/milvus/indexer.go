package milvus

import (
	"ai-cloud/config"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/consts"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"strconv"
)

type MilvusIndexerConfig struct {
	Collection string
	Dimension  int
	Embedding  embedding.Embedder
	Client     client.Client
}

type MilvusIndexer struct {
	config MilvusIndexerConfig
}

func NewMilvusIndexer(ctx context.Context, conf *MilvusIndexerConfig) (*MilvusIndexer, error) {
	// 检查配置
	if err := conf.check(); err != nil {
		return nil, fmt.Errorf("[NewMilvusIndexer] invalid config: %w", err)
	}

	// 检查Collection是否存在
	exists, err := conf.Client.HasCollection(ctx, conf.Collection)
	if err != nil {
		return nil, fmt.Errorf("[NewMilvusIndexer] check milvus collection failed : %w", err)
	}
	if !exists {
		if err := conf.createCollection(ctx, conf.Collection, conf.Dimension); err != nil {
			return nil, fmt.Errorf("[NewMilvusIndexer] create collection failed: %w", err)
		}
	}

	// 加载Collection
	err = conf.Client.LoadCollection(ctx, conf.Collection, false)
	if err != nil {
		return nil, fmt.Errorf("[NewMilvusIndexer] failed to load collection: %w", err)
	}

	return &MilvusIndexer{
		config: *conf,
	}, nil
}

func (m *MilvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {

	// 如果有opts则用opts中的配置（允许在Store的时候更换Embedder配置）
	co := indexer.GetCommonOptions(&indexer.Options{ //提供默认值选项
		SubIndexes: nil,
		Embedding:  m.config.Embedding,
	}, opts...)

	embedder := co.Embedding
	if embedder == nil {
		return nil, fmt.Errorf("[Indexer.Store] embedding not provided")
	}
	// 获取文档内容部分
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Content)
	}
	// 向量化
	vectors := make([][]float64, len(texts)) // 预分配结果切片

	for i, text := range texts {
		// 每次只embed一个文本
		vec, err := embedder.EmbedStrings(ctx, []string{text})
		if err != nil {
			return nil, fmt.Errorf("[Indexer.Store] failed to embed text at index %d: %w", i, err)
		}

		// 确保返回的向量是我们期望的单个结果
		if len(vec) != 1 {
			return nil, fmt.Errorf("[Indexer.Store] unexpected number of vectors returned: %d", len(vec))
		}

		vectors[i] = vec[0]
	}
	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("[Indexer.Store] embedding vector length mismatch")
	}
	rows, err := DocumentConvert(ctx, docs, vectors)
	if err != nil {
		return nil, err
	}

	results, err := m.config.Client.InsertRows(ctx, m.config.Collection, "", rows)
	if err != nil {
		return nil, err
	}
	if err := m.config.Client.Flush(ctx, m.config.Collection, false); err != nil {
		return nil, err
	}
	ids = make([]string, results.Len())
	for idx := 0; idx < results.Len(); idx++ {
		ids[idx], err = results.GetAsString(idx)
		if err != nil {
			return nil, fmt.Errorf("[Indexer.Store] failed to get id: %w", err)
		}
	}
	return ids, nil

}
func (m *MilvusIndexer) GetType() string {
	return "Milvus"
}

func DocumentConvert(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {

	em := make([]defaultSchema, 0, len(docs))
	texts := make([]string, 0, len(docs))
	rows := make([]interface{}, 0, len(docs))

	for _, doc := range docs {
		kbID, ok := doc.MetaData["kb_id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for kb_id")
		}

		docID, ok := doc.MetaData["document_id"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to marshal metadata: %w", ok)
		}

		metadata, err := sonic.Marshal(doc.MetaData)
		if err != nil {
			return nil, fmt.Errorf("[MilvusIndexer.DocumentConvert] failed to marshal metadata: %w", err)
		}
		em = append(em, defaultSchema{
			ID:         doc.ID,
			Content:    doc.Content,
			Vector:     nil,
			KBID:       kbID,
			DocumentID: docID,
			Metadata:   metadata,
		})
		texts = append(texts, doc.Content)
	}

	// build embedding documents for storing
	for idx, vec := range vectors {
		em[idx].Vector = utils.ConvertFloat64ToFloat32Embedding(vec)
		rows = append(rows, &em[idx])
	}
	return rows, nil
}

func (m *MilvusIndexerConfig) createCollection(ctx context.Context, collectionName string, dimension int) error {
	// 获取 Milvus 配置
	milvusConfig := config.GetConfig().Milvus
	// 创建集合Schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "存储文档分块和向量",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       consts.FieldNameID,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					"max_length": milvusConfig.IDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameContent,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.ContentMaxLength,
				},
			},
			{
				Name:     consts.FieldNameDocumentID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.DocIDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameKBID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.KbIDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameVector,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": strconv.Itoa(dimension),
				},
			},
			{
				Name:     consts.FieldNameMetadata,
				DataType: entity.FieldTypeJSON,
			},
		},
	}

	// 创建集合
	if err := m.Client.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("[NewMilvusIndexer.createCollection] 创建集合失败: %w", err)
	}

	// 创建索引
	idx, err := milvusConfig.GetMilvusIndex()
	if err != nil {
		return fmt.Errorf("[NewMilvusIndexer.createCollection] 从配置中获取索引类型失败: %w", err)
	}

	if err := m.Client.CreateIndex(ctx, collectionName, consts.FieldNameVector, idx, false); err != nil {
		return fmt.Errorf("[NewMilvusIndexer.createCollection] 创建索引失败: %w", err)
	}
	return nil
}

func (m *MilvusIndexerConfig) check() error {
	if m.Client == nil {
		return fmt.Errorf("[NewMilvusIndexer] milvus client is nil")
	}
	if m.Embedding == nil {
		return fmt.Errorf("[NewMilvusIndexer] embedding is nil")
	}
	if m.Collection == "" {
		return fmt.Errorf("[NewMilvusIndexer] collection is empty")
	}
	if m.Dimension == 0 {
		return fmt.Errorf("[NewMilvusIndexer] embedding dimension is zero")
	}
	return nil
}

func (m *MilvusIndexerConfig) IsCallbacksEnabled() bool {
	return true
}
