package milvus

// defaultSchema is the default schema for milvus by eino
type defaultSchema struct {
	ID           string    `json:"id" milvus:"name:id"`
	Content      string    `json:"content" milvus:"name:content"`
	DocumentID   string    `json:"document_id" milvus:"document_id"`
	DocumentName string    `json:"document_name" milvus:"name:document_name"`
	KBID         string    `json:"kb_id" milvus:"name:kb_id"`
	ChunkIndex   int       `json:"chunk_index" milvus:"name:chunk_index"`
	Vector       []float32 `json:"vector" milvus:"name:vector"`
	//Metadata []byte `json:"metadata" milvus:"name:metadata"`
}
