package milvus

import (
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/consts"
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusRetrieverConfig struct {
	Client     client.Client
	Embedder   embedding.Embedder
	Collection string
	Index      string
	TopK       int
}

type MilvusRetriever struct {
	client client.Client
	config MilvusRetrieverConfig
}

func NewMilvusRetriver(config *MilvusRetrieverConfig) (*MilvusRetriever, error) {
	return &MilvusRetriever{
		client: config.Client,
		config: *config,
	}, nil
}

var FieldNameVector = "vector"

func (m *MilvusRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	co := retriever.GetCommonOptions(&retriever.Options{
		Index:     &m.config.Index,
		TopK:      &m.config.TopK,
		Embedding: m.config.Embedder,
	})

	emb := co.Embedding
	vectors, err := emb.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("[milvus retriever] embedding has error: %w", err)
	}
	// check the embedding result
	if len(vectors) != 1 {
		return nil, fmt.Errorf("[milvus retriever] invalid return length of vector, got=%d, expected=1", len(vectors))
	}

	vector := utils.ConvertFloat64ToFloat32Embedding(vectors[0])

	//vec := make([]entity.FloatVector, 0, len(vectors))
	//for _, vector := range vectors {
	//	v := utils.ConvertFloat64ToFloat32Embedding(vector)
	//
	//	vec = append(vec, utils.ConvertFloat64ToFloat32Embedding(vector))
	//}

	var results []client.SearchResult
	sp, _ := entity.NewIndexIvfFlatSearchParam(16)

	results, err = m.client.Search(
		ctx,
		m.config.Collection,
		[]string{},
		"",
		consts.SearchFields,
		[]entity.Vector{entity.FloatVector(vector)},
		consts.FieldNameVector,
		entity.COSINE,
		m.config.TopK,
		sp,
	)

	documents := make([]*schema.Document, 0, len(results))
	for _, result := range results {
		if result.Err != nil {
			return nil, fmt.Errorf("[milvus retriever] search result has error: %w", result.Err)
		}
		if result.IDs == nil || result.Fields == nil {
			return nil, fmt.Errorf("[milvus retriever] search result has no ids or fields")
		}
		document, err := DocumentConverter(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("[milvus retriever] failed to convert search result to schema.Document: %w", err)
		}
		documents = append(documents, document...)
	}
	return documents, nil
}

// defaultDocumentConverter returns the default document converter
func DocumentConverter(ctx context.Context, doc client.SearchResult) ([]*schema.Document, error) {
	var err error
	result := make([]*schema.Document, doc.IDs.Len(), doc.IDs.Len())
	for i := range result {
		result[i] = &schema.Document{
			MetaData: make(map[string]any),
		}
	}
	for _, field := range doc.Fields {
		switch field.Name() {
		case "id":
			for i, document := range result {
				document.ID, err = doc.IDs.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get id: %w", err)
				}
			}
		case "content":
			for i, document := range result {
				document.Content, err = field.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get content: %w", err)
				}
			}
		case "metadata":
			for i, document := range result {
				b, err := field.Get(i)
				bytes, ok := b.([]byte)
				if !ok {
					return nil, fmt.Errorf("failed to get metadata: %w", err)
				}
				if err := sonic.Unmarshal(bytes, &document.MetaData); err != nil {
					return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
				}
			}
		default:
			for i, document := range result {
				document.MetaData[field.Name()], err = field.GetAsString(i)
			}
		}
	}
	return result, nil
}
