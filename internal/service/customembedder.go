package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	openaiEmbed "github.com/cloudwego/eino-ext/components/embedding/openai"
)

// OllamaEmbedder 实现自定义的Ollama嵌入服务
type OllamaEmbedder struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
	Dimensions int
}

// NewOllamaEmbedder 创建一个新的Ollama嵌入服务实例
func NewOllamaEmbedder(ollamaURL, model string, dimensions int) *OllamaEmbedder {
	return &OllamaEmbedder{
		BaseURL: ollamaURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // 设置较长的超时时间
		},
		Dimensions: dimensions,
	}
}

// EmbedStrings 将文本字符串转换为嵌入向量
func (oe *OllamaEmbedder) EmbedStrings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// 准备请求数据
	reqBody := map[string]interface{}{
		"model": oe.Model,
		"input": texts[0], // Ollama只能处理单个文本
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		oe.BaseURL+"/api/embed",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := oe.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API错误, 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var ollamaResp struct {
		Embeddings    [][]float64 `json:"embeddings"`
		Model         string      `json:"model"`
		TotalDuration int64       `json:"total_duration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("解析响应失败: %w, 响应: %s", err, string(bodyBytes))
	}

	// 检查是否有向量
	if len(ollamaResp.Embeddings) == 0 {
		return nil, fmt.Errorf("获取到空的embeddings数组")
	}

	fmt.Printf("成功获取到向量，模型: %s, 长度: %d, 处理时间: %.2f毫秒\n",
		ollamaResp.Model,
		len(ollamaResp.Embeddings[0]),
		float64(ollamaResp.TotalDuration)/1000000)

	return ollamaResp.Embeddings, nil
}

// CreateOpenAILikeEmbedder 创建一个使用openaiEmbed.Embedder的包装，但实际调用Ollama API
func CreateOpenAILikeEmbedder(ctx context.Context, ollamaURL, ollamaModel string, dimension int) *openaiEmbed.Embedder {
	// 创建标准OpenAI embedder作为返回值
	embedder, _ := openaiEmbed.NewEmbedder(ctx, &openaiEmbed.EmbeddingConfig{
		BaseURL:    ollamaURL, // 注意: 这是不正确的设置，实际会在kb_service.go中特殊处理
		Model:      ollamaModel,
		Timeout:    60 * time.Second,
		Dimensions: &dimension,
	})

	// 由于无法安全地替换embedder.EmbedStrings方法
	// 我们在kb_service.go中将不直接使用该方法，而是使用自定义的拦截方式

	return embedder
}

// 直接使用此函数进行Ollama向量化，绕过openai embedder
func OllamaEmbedStrings(ollamaURL, ollamaModel string, ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	// 创建单次使用的HTTP客户端
	ollamaClient := &http.Client{
		Timeout: 60 * time.Second, // 更长超时
	}

	// 准备请求数据 (Ollama API格式)
	reqBody := map[string]interface{}{
		"model": ollamaModel,
		"input": texts[0], // Ollama每次只能处理一个文本
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化Ollama请求失败: %w", err)
	}

	// 创建HTTP请求，使用正确的API端点
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		ollamaURL+"/api/embed",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("创建Ollama请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := ollamaClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送Ollama请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API错误, 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var ollamaResp struct {
		Embeddings    [][]float64 `json:"embeddings"`
		Model         string      `json:"model"`
		TotalDuration int64       `json:"total_duration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		// 尝试读取错误消息
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("解析Ollama响应失败: %w, 响应内容: %s", err, string(bodyBytes))
	}

	// 检查是否有向量
	if len(ollamaResp.Embeddings) == 0 {
		return nil, fmt.Errorf("Ollama返回了空的embeddings数组")
	}

	fmt.Printf("成功从Ollama获取到向量，长度: %d\n", len(ollamaResp.Embeddings[0]))

	return ollamaResp.Embeddings, nil
}
