package consts

// 集合相关常量定义
const (
	// CollectionNameTextChunks 文本块集合名称
	CollectionNameTextChunks = "text_chunks"
)

// 字段名称常量定义
const (
	// FieldNameID ID字段名
	FieldNameID = "id"
	// FieldNameContent 内容字段名
	FieldNameContent = "content"
	// FieldNameDocumentID 文档ID字段名
	FieldNameDocumentID = "document_id"
	// FieldNameDocumentName 文档名称字段名
	FieldNameDocumentName = "document_name"
	// FieldNameKBID 知识库ID字段名
	FieldNameKBID = "kb_id"
	// FieldNameChunkIndex 块索引字段名
	FieldNameChunkIndex = "chunk_index"
	// FieldNameVector 向量字段名
	FieldNameVector = "vector"
	// FiledNameMetadata meta信息
	FieldNameMetadata = "metadata"
)

// 查询相关字段
var (
	// SearchFields 搜索结果返回的字段
	SearchFields = []string{
		FieldNameID,
		FieldNameContent,
		FieldNameDocumentID,
		FieldNameDocumentName,
		FieldNameKBID,
		FieldNameChunkIndex,
	}
)
