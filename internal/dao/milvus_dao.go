package dao

import (
	"ai-cloud/internal/model"
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"log"
	"sort"
	"strings"
)

type MilvusDao interface {
	SaveChunks(chunks []model.Chunk) error
	Search(kbID string, vector []float32, topK int) ([]model.Chunk, error)
	DeleteChunks(docIDs []string) error
}

type milvusDao struct {
	mv client.Client
}

func NewMilvusDao(milvus client.Client) MilvusDao {
	return &milvusDao{mv: milvus}
}

func (m *milvusDao) SaveChunks(chunks []model.Chunk) error {
	ctx := context.Background()
	collectionName := "text_chunks"

	// 准备插入数据
	var ids []string
	var contents []string
	var documentIDs []string
	var documentNames []string
	var kbIDs []string
	var chunkIndices []int32
	var vectors [][]float32

	for _, chunk := range chunks {
		ids = append(ids, chunk.ID)
		contents = append(contents, chunk.Content)
		documentIDs = append(documentIDs, chunk.DocumentID)
		documentNames = append(documentNames, chunk.DocumentName)
		kbIDs = append(kbIDs, chunk.KBID)
		chunkIndices = append(chunkIndices, int32(chunk.Index))
		vectors = append(vectors, chunk.Embeddings)
	}

	// 创建插入的数据列
	idColumn := entity.NewColumnVarChar("id", ids)
	contentColumn := entity.NewColumnVarChar("content", contents)
	documentIDColumn := entity.NewColumnVarChar("document_id", documentIDs)
	documentNameColumn := entity.NewColumnVarChar("document_name", documentNames)
	kbIDColumn := entity.NewColumnVarChar("kb_id", kbIDs)
	indexColumn := entity.NewColumnInt32("chunk_index", chunkIndices)
	vectorColumn := entity.NewColumnFloatVector("vector", 1024, vectors)

	// 插入数据
	_, err := m.mv.Insert(
		ctx,
		collectionName,
		"",
		idColumn,
		contentColumn,
		documentIDColumn,
		documentNameColumn,
		kbIDColumn,
		indexColumn,
		vectorColumn,
	)
	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	return nil
}

func (m *milvusDao) DeleteChunks(docIDs []string) error {
	ctx := context.Background()
	collectionName := "text_chunks"

	// 构建删除表达式，使用 IN 操作符
	expr := fmt.Sprintf("document_id in [\"%s\"]", strings.Join(docIDs, "\",\""))

	err := m.mv.Delete(ctx, collectionName, "", expr)
	if err != nil {
		return fmt.Errorf("删除向量数据失败：%w", err)
	}

	return nil
}

func (m *milvusDao) parseSearchResults(searchResult []client.SearchResult) ([]model.Chunk, error) {
	var chunks []model.Chunk
	log.Println("SearchResult长度：%v", len(searchResult))
	for _, res := range searchResult {
		log.Println("IDs: %s", res.IDs)
		log.Println("Fields: %s", res.Fields)
		log.Printf("Scores: %v", res.Scores)
	}

	for _, result := range searchResult {
		idCol, ok := result.IDs.(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for ID column: %T", result.IDs)
		}

		contentCol, ok := result.Fields.GetColumn("content").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for content column")
		}

		docIDCol, ok := result.Fields.GetColumn("document_id").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for document ID column")
		}

		docNameCol, ok := result.Fields.GetColumn("document_name").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for document Name column")
		}

		kbIDCol, ok := result.Fields.GetColumn("kb_id").(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for KB ID column")
		}

		indexCol, ok := result.Fields.GetColumn("chunk_index").(*entity.ColumnInt32)
		if !ok {
			return nil, fmt.Errorf("unexpected type for index column")
		}

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

			docName, err := docNameCol.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取文档名称失败：%w", err)
			}

			kbID, err := kbIDCol.GetAsString(i)
			if err != nil {
				return nil, fmt.Errorf("获取知识库ID失败: %w", err)
			}

			index := indexCol.Data()[i]

			score := result.Scores[i]

			chunks = append(chunks, model.Chunk{
				ID:           id,
				Content:      content,
				KBID:         kbID,
				DocumentID:   docID,
				DocumentName: docName,
				Index:        int(index),
				Score:        score,
			})
		}
	}

	// 按Score从高到低排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})

	return chunks, nil
}

func (m *milvusDao) Search(kbID string, vector []float32, topK int) ([]model.Chunk, error) {
	ctx := context.Background()
	collectionName := "text_chunks"

	// 构建搜索参数
	sp, _ := entity.NewIndexIvfFlatSearchParam(16)
	expr := fmt.Sprintf("kb_id == \"%s\"", kbID)

	// 执行搜索
	searchResult, err := m.mv.Search(
		ctx,
		collectionName,
		[]string{},
		expr,
		[]string{"id", "content", "document_id", "document_name", "kb_id", "chunk_index"},
		[]entity.Vector{entity.FloatVector(vector)},
		"vector",
		entity.COSINE,
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	return m.parseSearchResults(searchResult)
}
