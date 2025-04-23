package config

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}

// MinioConfig Minio配置
type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	Region          string `mapstructure:"region"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `mapstructure:"type"` // local/oss/minio
	Local LocalConfig `mapstructure:"local"`
	OSS   OSSConfig   `mapstructure:"oss"`
	Minio MinioConfig `mapstructure:"minio"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BaseDir string `mapstructure:"base_dir"` // 本地存储根目录（如 /data/storage）
}

// OSSConfig OSS配置
type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"` // 使用字符串表示时间，便于配置
}

// RAGConfig RAG配置
type RAGConfig struct {
	ChunkSize   int `mapstructure:"chunk_size"`
	OverlapSize int `mapstructure:"overlap_size"`
}

// EmbeddingConfig 嵌入模型配置
type EmbeddingConfig struct {
	// 使用哪种嵌入模型: remote 或 ollama
	Service string `mapstructure:"service"`

	// 远程嵌入模型配置（如OpenAI API）
	Remote RemoteEmbeddingConfig `mapstructure:"remote"`

	// Ollama嵌入模型配置
	Ollama OllamaEmbeddingConfig `mapstructure:"ollama"`
}

// RemoteEmbeddingConfig 远程嵌入模型配置
type RemoteEmbeddingConfig struct {
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	BaseURL   string `mapstructure:"base_url"`
	Dimension int    `mapstructure:"dimension"`
}

// OllamaEmbeddingConfig Ollama嵌入模型配置
type OllamaEmbeddingConfig struct {
	URL       string `mapstructure:"url"`
	Model     string `mapstructure:"model"`
	Dimension int    `mapstructure:"dimension"`
}

// LLMConfig 语言模型配置
type LLMConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	BaseURL     string  `mapstructure:"base_url"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float32 `mapstructure:"temperature"`
}

// AppConfig 应用配置
type AppConfig struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Storage   StorageConfig   `mapstructure:"storage"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RAG       RAGConfig       `mapstructure:"rag"`
	Embedding EmbeddingConfig `mapstructure:"embedding"`
	LLM       LLMConfig       `mapstructure:"llm"`
}
