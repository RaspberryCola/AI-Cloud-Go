package model

type ChatResponse struct {
	Response   string  `json:"response"`
	References []Chunk `json:"references"`
}

type ChatRequest struct {
	Query string   `json:"query"`
	KBs   []string `json:"kbs"`
}
