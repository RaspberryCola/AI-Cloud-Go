package milvus

import (
	"ai-cloud/config"
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
	"strings"
)

type MilvusRetrieverConfig struct {
	Client         client.Client      // Required
	Embedding      embedding.Embedder // Required
	Collection     string             // Required
	KBIDs          []string           // Required 至少要查询一个知识库
	SearchFields   []string           // Optional defaultSearchFields
	TopK           int                // Optional default is 5
	ScoreThreshold float64            // Optional default is 0
}

type MilvusRetriever struct {
	config MilvusRetrieverConfig
}

func NewMilvusRetriever(ctx context.Context, conf *MilvusRetrieverConfig) (*MilvusRetriever, error) {
	// 检查必要配置，设置默认值
	if err := conf.check(); err != nil {
		return nil, fmt.Errorf("[NewMilvusRetriever] check config failed : %w", err)
	}
	// 检查Collection是否存在
	exists, err := conf.Client.HasCollection(ctx, conf.Collection)
	if err != nil {
		return nil, fmt.Errorf("[NewMilvusRetriever] check milvus collection failed : %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("[NewMilvusRetirever] collection %s not exists", conf.Collection)
	}

	// 检查是否load，没load的话load
	collection, err := conf.Client.DescribeCollection(ctx, conf.Collection)
	if err != nil {
		return nil, fmt.Errorf("[NewRetriever] failed to describe collection: %w", err)
	}

	if !collection.Loaded {
		err = conf.Client.LoadCollection(ctx, conf.Collection, false)
		if err != nil {
			return nil, fmt.Errorf("[NewMilvusRetriever] failed to load collection: %w", err)
		}
	}

	return &MilvusRetriever{
		config: *conf,
	}, nil
}

var FieldNameVector = "vector"

func (m *MilvusRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	// retrieve的时候指定参数
	co := retriever.GetCommonOptions(&retriever.Options{
		SubIndex:       nil,
		TopK:           &m.config.TopK,
		ScoreThreshold: &m.config.ScoreThreshold,
		Embedding:      m.config.Embedding,
	}, opts...)

	emb := co.Embedding
	vectors, err := emb.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("[MilvusRetriver.Retrieve] embedding has error: %w", err)
	}
	// 检查结果数量是否正确
	if len(vectors) != 1 {
		return nil, fmt.Errorf("[MilvusRetriver.Retrieve] invalid return length of vector, got=%d, expected=1", len(vectors))
	}

	vector := utils.ConvertFloat64ToFloat32Embedding(vectors[0])

	// 构造查询条件和参数
	kbIDs := m.config.KBIDs
	var expr string
	if len(kbIDs) > 0 {
		quotedIDs := make([]string, len(kbIDs))
		for i, id := range kbIDs {
			quotedIDs[i] = fmt.Sprintf(`"%s"`, id)
		}
		expr = fmt.Sprintf("%s in [%s]", consts.FieldNameKBID, strings.Join(quotedIDs, ","))
	} else {
		expr = "0 == 1"
	}

	var results []client.SearchResult
	sp, _ := entity.NewIndexIvfFlatSearchParam(config.GetConfig().Milvus.Nprobe)
	metricType := config.GetConfig().Milvus.GetMetricType()
	results, err = m.config.Client.Search(
		ctx,
		m.config.Collection,   // 集合名称：指定要搜索的Milvus集合
		[]string{},            // 分区名称：空表示搜索所有分区
		expr,                  // 过滤表达式：限制搜索范围，这里只搜索指定知识库ID的文档
		m.config.SearchFields, // 输出字段：指定返回结果中包含哪些字段
		[]entity.Vector{entity.FloatVector(vector)}, // 查询向量：将输入向量转换为Milvus向量格式
		consts.FieldNameVector,                      // 向量字段名：指定在哪个字段上执行向量搜索（对应Index）
		metricType,                                  // 度量类型：如何计算向量相似度（如余弦相似度、欧几里得距离等）
		m.config.TopK,                               // 返回数量：返回的最相似结果数量
		sp,                                          // 搜索参数：索引特定的搜索参数，如nprobe（探测聚类数）
	)

	documents := make([]*schema.Document, 0, len(results))
	for _, result := range results {
		if result.Err != nil {
			return nil, fmt.Errorf("[MilvusRetriver.Retrieve] search result has error: %w", result.Err)
		}
		if result.IDs == nil || result.Fields == nil {
			return nil, fmt.Errorf("[MilvusRetriver.Retrieve] search result has no ids or fields")
		}
		document, err := DocumentConverter(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("[MilvusRetriver.Retrieve] failed to convert search result to schema.Document: %w", err)
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

	importantMetaFields := map[string]bool{
		consts.FieldNameDocumentID: true,
		consts.FieldNameKBID:       true,
	}

	for _, field := range doc.Fields {
		switch field.Name() {
		case consts.FieldNameID:
			for i, document := range result {
				document.ID, err = doc.IDs.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get id: %w", err)
				}
			}
		case consts.FieldNameContent:
			for i, document := range result {
				document.Content, err = field.GetAsString(i)
				if err != nil {
					return nil, fmt.Errorf("failed to get content: %w", err)
				}
			}
		case consts.FieldNameMetadata:
			for i := range result {
				val, _ := field.Get(i)
				bytes, ok := val.([]byte)
				if !ok {
					return nil, fmt.Errorf("metadata field is not []byte")
				}
				var meta map[string]any
				if err := sonic.Unmarshal(bytes, &meta); err != nil {
					return nil, fmt.Errorf("unmarshal metadata failed: %w", err)
				}
				for k, v := range meta {
					result[i].MetaData[k] = v
				}
			}
		default:
			if importantMetaFields[field.Name()] {
				for i := range result {
					val, err := field.GetAsString(i)
					if err != nil {
						return nil, fmt.Errorf("get field %s failed: %w", field.Name(), err)
					}
					result[i].MetaData[field.Name()] = val
				}
			}
		}
	}
	return result, nil
}

func (m *MilvusRetriever) GetType() string {
	return "Milvus"
}

// 检查必要配置
func (m *MilvusRetrieverConfig) check() error {
	if m.Client == nil {
		return fmt.Errorf("[NewMilvusRetriever] milvus client is nil")
	}
	if m.Embedding == nil {
		return fmt.Errorf("[NewMilvusRetriever] embedding is nil")
	}
	if m.Collection == "" {
		return fmt.Errorf("[NewMilvusRetriever] collection is empty")
	}
	if m.SearchFields == nil {
		m.SearchFields = defaultSearchFields // 默认搜索字段
	}
	if m.TopK == 0 {
		m.TopK = 5 // 默认返回结果数量
	}
	if m.ScoreThreshold == 0 {
		m.ScoreThreshold = 0 // 默认相似度阈值
	}

	return nil
}
