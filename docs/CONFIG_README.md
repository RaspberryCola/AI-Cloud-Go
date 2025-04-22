# AI-Cloud-Go 配置指南

## 配置迁移说明

我们已将项目的所有配置从环境变量(`.env`文件)迁移到YAML配置文件(`config/config.yaml`)中，以统一配置管理方式，提高可维护性。

### 主要变更

1. 将嵌入模型(Embedding)相关配置从环境变量迁移到YAML配置
2. 将语言模型(LLM)相关配置从环境变量迁移到YAML配置
3. 更新了相关服务实现，使用配置文件而非环境变量

### 配置文件结构

配置文件(`config/config.yaml`)结构如下：

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "3306"
  user: "root"
  password: "123456"
  name: "ai_cloud"

jwt:
  secret: "your-jwt-secret"
  expiration_hours: 24

storage:
  type: "minio" # local, oss, minio
  local:
    base_dir: "./storage_data"
  oss:
    endpoint: "oss-endpoint"
    bucket: "bucket-name"
    access_key_id: ""
    access_key_secret: ""
  minio:
    endpoint: "localhost:9000"
    bucket: "ai-cloud"
    access_key_id: "minioadmin"
    access_key_secret: "minioadmin"
    use_ssl: false
    region: ""

rag:
  chunk_size: 1500
  overlap_size: 500

cors:
  # CORS配置...

# 嵌入模型配置
embedding:
  service: "ollama" # remote, local, ollama

  # 远程嵌入模型配置（OpenAI API，当 service=remote 时使用）
  remote:
    api_key: "your-api-key"
    model: "text-embedding-3-large"
    base_url: "https://api.openai.com/v1"

  # 本地FastAPI嵌入模型配置（当 service=local 时使用）
  local:
    url: "http://embedding-api:8000"
    url_host: "http://localhost:8008"
    model: "paraphrase-multilingual-MiniLM-L12-v2"
    dimension: 384

  # Ollama嵌入模型配置（当 service=ollama 时使用）
  ollama:
    url: "http://localhost:11434"
    model: "mxbai-embed-large"

# 语言模型配置
llm:
  api_key: "your-llm-api-key"
  model: "deepseek-chat"
  base_url: "https://api.deepseek.com/v1"
  max_tokens: 4096
  temperature: 0.7
```

## 配置使用

在Go代码中，可以通过以下方式访问配置：

```go
import "ai-cloud/config"

func main() {
    // 初始化配置（在应用启动时调用一次）
    config.InitConfig()
    
    // 获取配置实例
    cfg := config.GetConfig()
    
    // 访问配置项
    port := cfg.Server.Port
    embeddingService := cfg.Embedding.Service
    llmModel := cfg.LLM.Model
    
    // ...
}
```

## 嵌入服务配置

### 远程API服务 (OpenAI)

使用 OpenAI API 进行文本向量嵌入：

```yaml
embedding:
  service: "remote"
  remote:
    api_key: "your-api-key"
    model: "text-embedding-3-large"
    base_url: "https://api.openai.com/v1"
```

### 本地FastAPI服务

使用本地部署的FastAPI服务进行文本向量嵌入：

```yaml
embedding:
  service: "local"
  local:
    url: "http://embedding-api:8000"  # Docker网络内部地址
    url_host: "http://localhost:8008" # 宿主机访问地址
    model: "paraphrase-multilingual-MiniLM-L12-v2"
    dimension: 384
```

### Ollama本地服务

使用Ollama本地服务进行文本向量嵌入：

```yaml
embedding:
  service: "ollama"
  ollama:
    url: "http://localhost:11434"
    model: "mxbai-embed-large"
```

## 语言模型配置

配置LLM服务：

```yaml
llm:
  api_key: "your-api-key"
  model: "deepseek-chat"  # 或其他支持的模型
  base_url: "https://api.deepseek.com/v1"
  max_tokens: 4096
  temperature: 0.7
```

## 从环境变量迁移到配置文件

如果您之前使用`.env`文件配置项目，请按照以下对应关系迁移到`config.yaml`：

| 环境变量 | 配置文件路径 |
|---------|------------|
| `EMBEDDING_SERVICE` | `embedding.service` |
| `EMBEDDING_API_KEY` | `embedding.remote.api_key` |
| `EMBEDDING_MODEL` | `embedding.remote.model` |
| `EMBEDDING_BASE_URL` | `embedding.remote.base_url` |
| `LOCAL_EMBEDDING_URL` | `embedding.local.url` |
| `LOCAL_EMBEDDING_URL_HOST` | `embedding.local.url_host` |
| `LOCAL_EMBEDDING_MODEL` | `embedding.local.model` |
| `LOCAL_EMBEDDING_DIM` | `embedding.local.dimension` |
| `OLLAMA_URL` | `embedding.ollama.url` |
| `OLLAMA_EMBEDDING_MODEL` | `embedding.ollama.model` |
| `LLM_API_KEY` | `llm.api_key` |
| `LLM_MODEL` | `llm.model` |
| `LLM_BASE_URL` | `llm.base_url` |

## 注意事项

1. 配置文件中的敏感信息（如API密钥）不应提交到版本控制系统
2. 可以考虑使用环境变量覆盖配置文件中的敏感信息
3. 为不同环境（开发、测试、生产）准备不同的配置文件 