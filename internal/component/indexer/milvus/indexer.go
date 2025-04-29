package milvus

import (
	"ai-cloud/internal/database/milvus"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type MilvusIndexerConfig struct {
	Collection string
	Dimension  int
	Embedding  embedding.Embedder
}

type MilvusIndexer struct {
	client client.Client
	config MilvusIndexerConfig
}

func NewMilvusIndexer(indexerConfig *MilvusIndexerConfig) *MilvusIndexer {

	return &MilvusIndexer{
		client: milvus.GetMilvusClient(),
		config: *indexerConfig,
	}
}

func (m *MilvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	co := indexer.GetCommonOptions(&indexer.Options{
		SubIndexes: nil,
		Embedding:  m.config.Embedding,
	})

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
	//TODO: makeEmbeddingCtx?
	vectors, err := embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("[Indexer.Store] embedding vector length mismatch")
	}

	rows, err := DocumentConvert(ctx, docs, vectors)
	if err != nil {
		return nil, err
	}

	results, err := m.client.InsertRows(ctx, m.config.Collection, "", rows)
	if err != nil {
		return nil, err
	}
	if err := m.client.Flush(ctx, m.config.Collection, false); err != nil {
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
		kbID, err := doc.MetaData["kb_id"].(string) // 假设 kb_id 是 string 类型
		if !err {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		docID, err := doc.MetaData["document_id"].(string)
		if !err {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		docName, err := doc.MetaData["document_name"].(string)
		if !err {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		chunkIndex, err := doc.MetaData["chunk_index"].(int)
		if !err {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}

		em = append(em, defaultSchema{
			ID:           doc.ID,
			Content:      doc.Content,
			Vector:       nil,
			KBID:         kbID,
			DocumentID:   docID,
			DocumentName: docName,
			ChunkIndex:   chunkIndex,
		})
		texts = append(texts, doc.Content)
	}

	// build embedding documents for storing
	for idx, vec := range vectors {
		em[idx].Vector = ConvertFloat64ToFloat32Embedding(vec)
		rows = append(rows, &em[idx])
	}
	return rows, nil
}

func ConvertFloat64ToFloat32Embedding(embedding []float64) []float32 {
	float32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		float32Embedding[i] = float32(v)
	}
	return float32Embedding
}
