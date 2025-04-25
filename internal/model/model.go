package model

import "time"

type Model struct {
	// 基础信息
	ID      string `gorm:"primaryKey;type:char(36)"`
	Type    string `gorm:"not null"` // embedding/llm
	Name    string `gorm:"not null"` // 显示名称
	Server  string `gorm:"not null"` // openai/ollama/huggingface/local
	BaseURL string `gorm:"not null"` // API基础地址
	Model   string `gorm:"not null"` // 模型标识符，例如 deepseek-chat，text-embedding-v3
	APIKey  string // 访问密钥

	// Embedding模型字段
	Dimension int // 向量维度(embedding必填)

	// LLM模型字段
	MaxOutputLength int  `gorm:"default:4096"`
	Function        bool `gorm:"default:false"`

	// 通用字段
	MaxTokens int       `gorm:"default:1024"` // 限制最大的输入长度
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type CreateModelRequest struct {
	// 基础信息
	Type    string `json:"type" binding:"required,oneof=embedding llm"`
	Name    string `json:"name" binding:"required"`
	Server  string `json:"server" binding:"required"`
	BaseURL string `json:"base_url" binding:"required,url"`
	Model   string `json:"model" binding:"required"`
	APIKey  string `json:"api_key"`

	// Embedding
	Dimension int `json:"dimension"`

	// LLM
	MaxOutputLength int  `json:"max_output_length"`
	Function        bool `json:"function"`

	// 通用字段
	MaxTokens int `json:"max_tokens"`
}
