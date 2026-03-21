package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var AppConfigInstance *AppConfig

// InitConfig 初始化配置
func InitConfig() {
	// 初始化 AppConfigInstance
	AppConfigInstance = defaultConfig()

	// 加载配置
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			log.Fatalf("Error reading config file: %v", err)
		}
		log.Printf("config file not found, using defaults and environment overrides")
	} else {
		// 监听配置变化
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			cfg := defaultConfig()
			if err := viper.Unmarshal(cfg); err != nil {
				log.Printf("loadConfig failed, unmarshal config err: %v", err)
				return
			}
			applyEnvOverrides(cfg)
			AppConfigInstance = cfg
		})
	}

	// 解析配置
	if err := viper.Unmarshal(AppConfigInstance); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
	applyEnvOverrides(AppConfigInstance)
}

// GetConfig 获取配置
func GetConfig() *AppConfig {
	return AppConfigInstance
}

func defaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "123456",
			Name:     "ai_cloud",
		},
		JWT: JWTConfig{
			Secret:          "your-jwt-secret",
			ExpirationHours: 24,
		},
		Storage: StorageConfig{
			Type: "minio",
			Local: LocalConfig{
				BaseDir: "./storage_data",
			},
			OSS: OSSConfig{
				Endpoint:        "your-oss-endpoint",
				Bucket:          "your-oss-bucket",
				AccessKeyID:     "your-access-key-id",
				AccessKeySecret: "your-access-key-secret",
			},
			Minio: MinioConfig{
				Endpoint:        "localhost:9000",
				Bucket:          "ai-cloud",
				AccessKeyID:     "minioadmin",
				AccessKeySecret: "minioadmin",
				UseSSL:          false,
				Region:          "",
			},
		},
		Milvus: MilvusConfig{
			Address:          "localhost:19530",
			IndexType:        "IVF_FLAT",
			MetricType:       "COSINE",
			Nlist:            128,
			Nprobe:           16,
			IDMaxLength:      "64",
			ContentMaxLength: "65535",
			DocIDMaxLength:   "64",
			DocNameMaxLength: "256",
			KbIDMaxLength:    "64",
		},
		RAG: RAGConfig{
			ChunkSize:   1500,
			OverlapSize: 500,
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           "12h",
		},
		LLM: LLMConfig{
			APIKey:      "your-llm-api-key",
			Model:       "deepseek-chat",
			BaseURL:     "https://api.deepseek.com/v1",
			MaxTokens:   10240,
			Temperature: 0.7,
		},
	}
}

func applyEnvOverrides(cfg *AppConfig) {
	overrideString("SERVER_PORT", &cfg.Server.Port)

	overrideString("DB_HOST", &cfg.Database.Host)
	overrideString("DB_PORT", &cfg.Database.Port)
	overrideString("DB_USER", &cfg.Database.User)
	overrideString("DB_PASSWORD", &cfg.Database.Password)
	overrideString("DB_NAME", &cfg.Database.Name)

	overrideString("JWT_SECRET", &cfg.JWT.Secret)
	overrideInt("JWT_EXPIRATION_HOURS", &cfg.JWT.ExpirationHours)

	overrideString("STORAGE_TYPE", &cfg.Storage.Type)
	overrideString("STORAGE_LOCAL_BASE_DIR", &cfg.Storage.Local.BaseDir)
	overrideString("STORAGE_OSS_ENDPOINT", &cfg.Storage.OSS.Endpoint)
	overrideString("STORAGE_OSS_BUCKET", &cfg.Storage.OSS.Bucket)
	overrideString("STORAGE_OSS_ACCESS_KEY_ID", &cfg.Storage.OSS.AccessKeyID)
	overrideString("STORAGE_OSS_ACCESS_KEY_SECRET", &cfg.Storage.OSS.AccessKeySecret)
	overrideString("STORAGE_MINIO_ENDPOINT", &cfg.Storage.Minio.Endpoint)
	overrideString("STORAGE_MINIO_BUCKET", &cfg.Storage.Minio.Bucket)
	overrideString("STORAGE_MINIO_ACCESS_KEY_ID", &cfg.Storage.Minio.AccessKeyID)
	overrideString("STORAGE_MINIO_ACCESS_KEY_SECRET", &cfg.Storage.Minio.AccessKeySecret)
	overrideBool("STORAGE_MINIO_USE_SSL", &cfg.Storage.Minio.UseSSL)
	overrideString("STORAGE_MINIO_REGION", &cfg.Storage.Minio.Region)

	overrideString("MILVUS_ADDRESS", &cfg.Milvus.Address)
	overrideString("MILVUS_COLLECTION_NAME", &cfg.Milvus.CollectionName)
	overrideInt("MILVUS_VECTOR_DIMENSION", &cfg.Milvus.VectorDimension)
	overrideString("MILVUS_INDEX_TYPE", &cfg.Milvus.IndexType)
	overrideString("MILVUS_METRIC_TYPE", &cfg.Milvus.MetricType)
	overrideInt("MILVUS_NLIST", &cfg.Milvus.Nlist)
	overrideInt("MILVUS_NPROBE", &cfg.Milvus.Nprobe)
	overrideString("MILVUS_ID_MAX_LENGTH", &cfg.Milvus.IDMaxLength)
	overrideString("MILVUS_CONTENT_MAX_LENGTH", &cfg.Milvus.ContentMaxLength)
	overrideString("MILVUS_DOC_ID_MAX_LENGTH", &cfg.Milvus.DocIDMaxLength)
	overrideString("MILVUS_DOC_NAME_MAX_LENGTH", &cfg.Milvus.DocNameMaxLength)
	overrideString("MILVUS_KB_ID_MAX_LENGTH", &cfg.Milvus.KbIDMaxLength)

	overrideInt("RAG_CHUNK_SIZE", &cfg.RAG.ChunkSize)
	overrideInt("RAG_OVERLAP_SIZE", &cfg.RAG.OverlapSize)

	overrideCSV("CORS_ALLOW_ORIGINS", &cfg.CORS.AllowOrigins)
	overrideCSV("CORS_ALLOW_METHODS", &cfg.CORS.AllowMethods)
	overrideCSV("CORS_ALLOW_HEADERS", &cfg.CORS.AllowHeaders)
	overrideCSV("CORS_EXPOSE_HEADERS", &cfg.CORS.ExposeHeaders)
	overrideBool("CORS_ALLOW_CREDENTIALS", &cfg.CORS.AllowCredentials)
	overrideString("CORS_MAX_AGE", &cfg.CORS.MaxAge)

	overrideString("LLM_API_KEY", &cfg.LLM.APIKey)
	overrideString("LLM_MODEL", &cfg.LLM.Model)
	overrideString("LLM_BASE_URL", &cfg.LLM.BaseURL)
	overrideInt("LLM_MAX_TOKENS", &cfg.LLM.MaxTokens)
	overrideFloat32("LLM_TEMPERATURE", &cfg.LLM.Temperature)
}

func overrideString(key string, target *string) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		*target = value
	}
}

func overrideInt(key string, target *int) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("invalid integer for %s: %v", key, err)
			return
		}
		*target = parsed
	}
}

func overrideBool(key string, target *bool) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("invalid boolean for %s: %v", key, err)
			return
		}
		*target = parsed
	}
}

func overrideFloat32(key string, target *float32) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		parsed, err := strconv.ParseFloat(value, 32)
		if err != nil {
			log.Printf("invalid float for %s: %v", key, err)
			return
		}
		*target = float32(parsed)
	}
}

func overrideCSV(key string, target *[]string) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		parts := strings.Split(value, ",")
		items := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				items = append(items, trimmed)
			}
		}
		if len(items) > 0 {
			*target = items
		}
	}
}
