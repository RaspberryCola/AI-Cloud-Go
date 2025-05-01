package milvus

// defaultSchema
type defaultSchema struct {
	ID         string    `json:"id" milvus:"name:id"`
	Content    string    `json:"content" milvus:"name:content"`
	DocumentID string    `json:"document_id" milvus:"name:document_id"`
	KBID       string    `json:"kb_id" milvus:"name:kb_id"`
	Vector     []float32 `json:"vector" milvus:"name:vector"`
	Metadata   []byte    `json:"metadata" milvus:"name:metadata"` // 存放例如DocumentName，Index等信息
}
