package dao

import (
	"ai-cloud/config"
	"ai-cloud/internal/model"
	"ai-cloud/pkgs/consts"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MilvusDao interface {
	// 创建集合
	CreateCollection(ctx context.Context, collectionName string, dimension int) error

	// 保存文本块到向量数据库
	SaveChunks(ctx context.Context, collectionName string, chunks []model.Chunk, dimension int) error

	// 在知识库中搜索相似向量，返回相似度排序后的文本块
	Search(kbID, collectionName string, vector []float32, topK int) ([]model.Chunk, error)

	// 删除指定文档的所有文本块
	DeleteChunks(docIDs []string, collectionName string) error
}

type milvusDao struct {
	mv client.Client // Milvus客户端
}

func NewMilvusDao(milvus client.Client) MilvusDao {
	return &milvusDao{mv: milvus}
}

// 根据collectionNam和dimension信息创建集合
func (m *milvusDao) CreateCollection(ctx context.Context, collectionName string, dimension int) error {
	//  检查集合是否存在
	exists, err := m.mv.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合存在失败: %w", err)
	}

	if exists {
		return nil
	}

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
				Name:     consts.FieldNameDocumentName,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.DocNameMaxLength,
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
				Name:     consts.FieldNameChunkIndex,
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     consts.FieldNameVector,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": strconv.Itoa(dimension),
				},
			},
		},
	}

	// 创建集合
	if err := m.mv.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("创建集合失败: %w", err)
	}

	// 创建索引
	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	if err := m.mv.CreateIndex(ctx, collectionName, "vector", idx, false); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 加载集合到内存
	if err := m.mv.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("加载集合失败: %w", err)
	}

	return nil
}

// SaveChunks 保存文本块到Milvus向量数据库
// 该方法会过滤掉无效的文本块，然后将有效的文本块插入到向量数据库中
// 如果插入失败，会自动重试指定次数
func (m *milvusDao) SaveChunks(ctx context.Context, collectionName string, chunks []model.Chunk, dimension int) error {
	if len(chunks) == 0 {
		return fmt.Errorf("num_rows should be greater than 0: invalid parameter[expected=invalid num_rows][actual=0")
	}
	fmt.Printf("SaveChunks: 准备插入%d个文本块\n", len(chunks))

	// 准备有效的数据
	preparedData, err := m.prepareChunkData(chunks, collectionName, dimension)
	if err != nil {
		return err
	}
	// 设置目标集合名称
	preparedData.collectionName = collectionName
	// 创建数据列
	columns := m.createDataColumns(preparedData)

	// 插入数据
	return m.insertDataWithRetry(ctx, preparedData.collectionName, columns, 3)
}

// chunkData 文本块数据结构，用于存储预处理后的向量数据
type chunkData struct {
	collectionName string      // 目标集合名称
	vectorDim      int         // 向量维度
	ids            []string    // 文本块ID列表
	contents       []string    // 文本块内容列表
	documentIDs    []string    // 对应的文档ID列表
	documentNames  []string    // 对应的文档名称列表
	kbIDs          []string    // 对应的知识库ID列表
	chunkIndices   []int32     // 文本块在文档中的索引位置
	vectors        [][]float32 // 文本块的向量表示
}

// prepareChunkData 验证和准备文本块数据
// 该方法会过滤掉无效的文本块（空内容或空向量），并确保文档名不超过最大长度限制
// 返回处理后的数据结构和可能的错误
func (m *milvusDao) prepareChunkData(chunks []model.Chunk, collectionName string, dimension int) (*chunkData, error) {

	data := &chunkData{
		collectionName: collectionName,
		vectorDim:      dimension,
	}

	// 遍历验证并准备数据
	for i, chunk := range chunks {
		// 验证chunk数据
		if len(chunk.Content) == 0 {
			fmt.Printf("prepareChunkData warn: No.%d vector is null, it will be passed\n", i)
			continue
		}
		if len(chunk.Embeddings) == 0 {
			fmt.Printf("prepareChunkData warn: No.%d vector is null, it will be passed\n", i)
			continue
		}

		// 确保文档名不超过限制长度
		docName := chunk.DocumentName
		if len(docName) > 250 {
			docName = docName[:250]
		}

		// 添加有效数据
		data.ids = append(data.ids, chunk.ID)
		data.contents = append(data.contents, chunk.Content)
		data.documentIDs = append(data.documentIDs, chunk.DocumentID)
		data.documentNames = append(data.documentNames, docName)
		data.kbIDs = append(data.kbIDs, chunk.KBID)
		data.chunkIndices = append(data.chunkIndices, int32(chunk.Index))
		data.vectors = append(data.vectors, chunk.Embeddings)
	}

	// 确保有有效数据要插入
	if len(data.ids) == 0 {
		return nil, fmt.Errorf("过滤无效数据后，没有有效的文本块可以插入")
	}

	return data, nil
}

// createDataColumns 创建Milvus数据列
// 将预处理的数据转换为Milvus插入操作所需的列格式
// 返回包含所有数据列的切片
func (m *milvusDao) createDataColumns(data *chunkData) []entity.Column {
	idColumn := entity.NewColumnVarChar(consts.FieldNameID, data.ids)
	contentColumn := entity.NewColumnVarChar(consts.FieldNameContent, data.contents)
	documentIDColumn := entity.NewColumnVarChar(consts.FieldNameDocumentID, data.documentIDs)
	documentNameColumn := entity.NewColumnVarChar(consts.FieldNameDocumentName, data.documentNames)
	kbIDColumn := entity.NewColumnVarChar(consts.FieldNameKBID, data.kbIDs)
	indexColumn := entity.NewColumnInt32(consts.FieldNameChunkIndex, data.chunkIndices)
	vectorColumn := entity.NewColumnFloatVector(consts.FieldNameVector, data.vectorDim, data.vectors)

	return []entity.Column{
		idColumn,
		contentColumn,
		documentIDColumn,
		documentNameColumn,
		kbIDColumn,
		indexColumn,
		vectorColumn,
	}
}

// insertDataWithRetry 尝试插入数据，失败时自动重试
func (m *milvusDao) insertDataWithRetry(ctx context.Context, collectionName string, columns []entity.Column, maxRetries int) error {
	var result *multierror.Error
	baseDelay := 100 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		fmt.Printf("insertDataWithRetry debug, try to insert data: (%d/%d)...\n", i+1, maxRetries)
		_, err := m.mv.Insert(ctx, collectionName, "", columns...)
		if err == nil {
			fmt.Printf("insertDataWithRetry Success!\n")
			return nil
		}

		result = multierror.Append(result, fmt.Errorf("insertDataWithRetry %d/%d failed: %w", i+1, maxRetries, err))

		// 指数退避
		delay := baseDelay * time.Duration(1<<uint(i))
		time.Sleep(delay)
	}
	return result.ErrorOrNil()
}

// DeleteChunks 删除指定文档ID列表对应的所有文本块
// 使用IN操作符构建删除表达式，一次性删除多个文档的所有块
func (m *milvusDao) DeleteChunks(docIDs []string, collectionName string) error {
	// 构建删除表达式，使用 IN 操作符
	expr := fmt.Sprintf("%s in [\"%s\"]", consts.FieldNameDocumentID, strings.Join(docIDs, "\",\""))
	// 删除
	if err := m.mv.Delete(context.Background(), collectionName, "", expr); err != nil {
		return fmt.Errorf("删除向量数据失败：%w", err)
	}
	return nil
}

// Search 在知识库中搜索相似向量
func (m *milvusDao) Search(kbID, collectionName string, vector []float32, topK int) ([]model.Chunk, error) {
	// 构建搜索参数
	sp, _ := entity.NewIndexIvfFlatSearchParam(config.GetConfig().Milvus.Nprobe)
	expr := fmt.Sprintf("%s == \"%s\"", consts.FieldNameKBID, kbID)

	// 执行搜索
	searchResult, err := m.mv.Search(
		context.Background(),
		collectionName,      // 集合名称：指定要搜索的Milvus集合
		[]string{},          // 分区名称：空表示搜索所有分区
		expr,                // 过滤表达式：限制搜索范围，这里只搜索指定知识库ID的文档
		consts.SearchFields, // 输出字段：指定返回结果中包含哪些字段
		[]entity.Vector{entity.FloatVector(vector)}, // 查询向量：将输入向量转换为Milvus向量格式
		consts.FieldNameVector,                      // 向量字段名：指定在哪个字段上执行向量搜索
		config.GetConfig().Milvus.GetMetricType(),   // 度量类型：如何计算向量相似度（如余弦相似度、欧几里得距离等）
		topK, // 返回数量：返回的最相似结果数量
		sp,   // 搜索参数：索引特定的搜索参数，如nprobe（探测聚类数）
	)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}
	return m.parseSearchResults(searchResult)
}

// parseSearchResults 解析搜索结果，将Milvus返回结果转换为模型数据
func (m *milvusDao) parseSearchResults(searchResult []client.SearchResult) ([]model.Chunk, error) {
	var chunks []model.Chunk
	log.Printf("SearchResult长度：%v\n", len(searchResult))
	for _, res := range searchResult {
		log.Printf("IDs: %s\n", res.IDs)
		log.Printf("Fields: %s\n", res.Fields)
		log.Printf("Scores: %v\n", res.Scores)
	}

	for _, result := range searchResult {
		idCol, ok := result.IDs.(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for ID column: %T", result.IDs)
		}

		contentCol, ok := result.Fields.GetColumn(consts.FieldNameContent).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for content column")
		}

		docIDCol, ok := result.Fields.GetColumn(consts.FieldNameDocumentID).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for document ID column")
		}

		docNameCol, ok := result.Fields.GetColumn(consts.FieldNameDocumentName).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for document Name column")
		}

		kbIDCol, ok := result.Fields.GetColumn(consts.FieldNameKBID).(*entity.ColumnVarChar)
		if !ok {
			return nil, fmt.Errorf("unexpected type for KB ID column")
		}

		indexCol, ok := result.Fields.GetColumn(consts.FieldNameChunkIndex).(*entity.ColumnInt32)
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
