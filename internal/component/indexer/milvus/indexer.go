package milvus

import (
	"ai-cloud/internal/database/milvus"
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"log"
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

func NewMilvusIndexer(indexerConfig *MilvusIndexerConfig) (*MilvusIndexer, error) {
	return &MilvusIndexer{
		client: milvus.GetMilvusClient(),
		config: *indexerConfig,
	}, nil
}

func (m *MilvusIndexer) Store(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) (ids []string, err error) {
	// 如果有opts则用opts中的配置（允许在Store的时候更换Embedder配置）
	co := indexer.GetCommonOptions(&indexer.Options{ //提供默认值选项
		SubIndexes: nil,
		Embedding:  m.config.Embedding,
	}, opts...)

	log.Println("[INFO] 获取embeder")
	embedder := co.Embedding
	if embedder == nil {
		return nil, fmt.Errorf("[Indexer.Store] embedding not provided")
	}
	log.Println("[INFO] 获取texts")
	// 获取文档内容部分
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.Content)
	}

	log.Println("[INFO] 开始embedding")
	// 向量化
	//TODO: makeEmbeddingCtx?
	vectors := make([][]float64, len(texts)) // 预分配结果切片

	for i, text := range texts {
		// 每次只embed一个文本
		vec, err := embedder.EmbedStrings(ctx, []string{text})
		if err != nil {
			return nil, fmt.Errorf("failed to embed text at index %d: %w", i, err)
		}

		// 确保返回的向量是我们期望的单个结果
		if len(vec) != 1 {
			return nil, fmt.Errorf("unexpected number of vectors returned: %d", len(vec))
		}

		vectors[i] = vec[0]
	}
	//vectors, err := embedder.EmbedStrings(ctx, texts)
	//if err != nil {
	//	return nil, err
	//}

	if len(vectors) != len(docs) {
		return nil, fmt.Errorf("[Indexer.Store] embedding vector length mismatch")
	}
	log.Println("[INFO] 开始convert")
	rows, err := DocumentConvert(ctx, docs, vectors)
	if err != nil {
		return nil, err
	}
	fmt.Println(rows)

	results, err := m.client.InsertRows(ctx, m.config.Collection, "", rows)
	if err != nil {
		return nil, err
	}
	//if err := m.client.Flush(ctx, m.config.Collection, false); err != nil {
	//	return nil, err
	//}
	fmt.Println(results)
	log.Println("[INFO] 存储完成")
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
		docName, ok := doc.MetaData["document_name"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to marshal metadata: %w", ok)
		}
		chunkIndex, ok := doc.MetaData["chunk_index"].(int)
		if !ok {
			return nil, fmt.Errorf("failed to marshal metadata: %w", ok)
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
