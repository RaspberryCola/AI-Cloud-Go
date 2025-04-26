package model

import (
	"time"
)

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID               string    `gorm:"primaryKey;type:char(36)"` // UUID
	Name             string    `gorm:"not null"`                 // 知识库名称
	Description      string    // 知识库描述
	UserID           uint      `gorm:"index"`    // 创建者ID
	EmbedModelID     string    `gorm:"index"`    // 关联的embedding模型id
	MilvusCollection string    `gorm:"not null"` //对应的milvus collection名称
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}

// Document 知识库文档
type Document struct {
	ID              string    `gorm:"primaryKey;type:char(36)"` // UUID
	UserID          uint      `gorm:"index"`                    // 所属的用户
	KnowledgeBaseID string    `gorm:"index"`                    // 所属知识库ID
	FileID          string    `gorm:"index"`                    // 关联的文件ID
	Title           string    // 文档标题
	DocType         string    // 文档类型(pdf/txt/md)
	Status          int       // 处理状态(0:待处理,1:处理中,2:已完成,3:失败)
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

// 存储到milvus中
type Chunk struct {
	ID           string    `json:"id"`
	Content      string    `json:"content"`       // chunk内容
	KBID         string    `json:"kb_id"`         // 知识库ID（知识库级别的检索）
	DocumentID   string    `json:"document_id"`   // 文档ID
	DocumentName string    `json:"document_name"` // 文档名称
	Index        int       `json:"index"`         // 第几个chunk
	Embeddings   []float32 `json:"embeddings"`    // chunk向量
	Score        float32   `json:"score"`         // 返回分数信息
}

type AddFileRequest struct {
	FileID string `json:"file_id"`
	KBID   string `json:"kb_id"`
}

type BatchDeleteDocsReq struct {
	KBID   string   `json:"kbid"`
	DocIDs []string `json:"doc_ids" binding:"required"`
}

type CreateKBRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	EmbedModelID string `json:"embed_model_id" binding:"required"`
}

type RetrieveRequest struct {
	KBID  string `json:"kb_id"`
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}
