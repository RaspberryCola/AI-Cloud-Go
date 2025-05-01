package model

import "github.com/cloudwego/eino/schema"

type ChatResponse struct {
	Response   string             `json:"response"`
	References []*schema.Document `json:"references"`
}

type ChatRequest struct {
	Query string   `json:"query"`
	KBs   []string `json:"kbs"`
}

// ChatStreamResponse OpenAI 兼容的流式响应格式
type ChatStreamResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []ChatStreamChoice `json:"choices"`
}

type ChatStreamChoice struct {
	Delta        ChatStreamDelta `json:"delta"`
	Index        int             `json:"index"`
	FinishReason *string         `json:"finish_reason"`
}

type ChatStreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}
