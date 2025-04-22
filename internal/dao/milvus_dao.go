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
	"time"
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

	// 添加日志，检查传入的chunks数量
	fmt.Printf("SaveChunks: 准备插入%d个文本块\n", len(chunks))
	if len(chunks) == 0 {
		return fmt.Errorf("num_rows should be greater than 0: invalid parameter[expected=invalid num_rows][actual=0")
	}

	// 准备插入数据
	var ids []string
	var contents []string
	var documentIDs []string
	var documentNames []string
	var kbIDs []string
	var chunkIndices []int32
	var vectors [][]float32

	for i, chunk := range chunks {
		// 验证chunk数据
		if len(chunk.Content) == 0 {
			fmt.Printf("警告: 第%d个文本块内容为空，跳过\n", i)
			continue
		}
		if len(chunk.Embeddings) == 0 {
			fmt.Printf("警告: 第%d个文本块向量为空，跳过\n", i)
			continue
		}

		// 确保文档名不超过限制长度
		docName := chunk.DocumentName
		if len(docName) > 250 {
			docName = docName[:250]
		}

		ids = append(ids, chunk.ID)
		contents = append(contents, chunk.Content)
		documentIDs = append(documentIDs, chunk.DocumentID)
		documentNames = append(documentNames, docName)
		kbIDs = append(kbIDs, chunk.KBID)
		chunkIndices = append(chunkIndices, int32(chunk.Index))
		vectors = append(vectors, chunk.Embeddings)
	}

	// 重新检查，确保有数据要插入
	if len(ids) == 0 {
		return fmt.Errorf("过滤无效数据后，没有有效的文本块可以插入")
	}

	// 创建插入的数据列
	idColumn := entity.NewColumnVarChar("id", ids)
	contentColumn := entity.NewColumnVarChar("content", contents)
	documentIDColumn := entity.NewColumnVarChar("document_id", documentIDs)
	documentNameColumn := entity.NewColumnVarChar("document_name", documentNames)
	kbIDColumn := entity.NewColumnVarChar("kb_id", kbIDs)
	indexColumn := entity.NewColumnInt32("chunk_index", chunkIndices)
	vectorColumn := entity.NewColumnFloatVector("vector", 1024, vectors)

	// 插入数据，最多重试3次
	var lastErr error
	for i := 0; i < 3; i++ {
		fmt.Printf("尝试插入数据到Milvus (%d/3)...\n", i+1)
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
		if err == nil {
			fmt.Printf("成功插入%d条数据到Milvus\n", len(ids))
			return nil
		}

		lastErr = err
		fmt.Printf("插入失败 (%d/3): %v\n", i+1, err)
		time.Sleep(1 * time.Second) // 等待1秒后重试
	}

	return fmt.Errorf("插入数据失败，已重试3次: %w", lastErr)
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
