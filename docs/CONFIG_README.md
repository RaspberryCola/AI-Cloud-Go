# AI-Cloud-Go 配置指南

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

# Milvus向量数据库配置
milvus:
  address: "localhost:19530"  # Milvus服务地址
  index_type: "IVF_FLAT"  # 索引类型 (IVF_FLAT, IVF_SQ8, HNSW)
  metric_type: "COSINE"  # 距离计算方式 (COSINE, L2, IP)
  nlist: 128  # IVF索引聚类数量
  # 搜索参数
  nprobe: 16  # 搜索时检查的聚类数量，值越大结果越精确但越慢
  # 字段最大长度配置
  id_max_length: "64"  # ID字段最大长度
  content_max_length: "65535"  # 内容字段最大长度
  doc_id_max_length: "64"  # 文档ID字段最大长度
  doc_name_max_length: "256"  # 文档名称字段最大长度
  kb_id_max_length: "64"  # 知识库ID字段最大长度

rag:
  chunk_size: 1500
  overlap_size: 500

cors:
  # CORS配置...

# 语言模型配置（后续移除，通过统一的模块管理）
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
    milvusAddress := cfg.Milvus.Address
    
    // ...
}
```

## Milvus向量数据库配置

配置Milvus服务连接地址和向量集合参数：

```yaml
# Milvus向量数据库配置
milvus:
  address: "localhost:19530"  # Milvus服务地址
  index_type: "IVF_FLAT"  # 索引类型 (IVF_FLAT, IVF_SQ8, HNSW)
  metric_type: "COSINE"  # 距离计算方式 (COSINE, L2, IP)
  nlist: 128  # IVF索引聚类数量
  # 搜索参数
  nprobe: 16  # 搜索时检查的聚类数量，值越大结果越精确但越慢
  # 字段最大长度配置
  id_max_length: "64"  # ID字段最大长度
  content_max_length: "65535"  # 内容字段最大长度
  doc_id_max_length: "64"  # 文档ID字段最大长度
  doc_name_max_length: "256"  # 文档名称字段最大长度
  kb_id_max_length: "64"  # 知识库ID字段最大长度
```

此配置在初始化Milvus客户端和创建集合时使用：

```go
// 初始化Milvus客户端
milvusClient, err := client.NewClient(ctx, client.Config{
    Address: config.GetConfig().Milvus.Address,
})

// 使用配置创建集合
milvusConfig := config.GetConfig().Milvus
address := milvusConfig.Address
// ...
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
| `LLM_API_KEY` | `llm.api_key` |
| `LLM_MODEL` | `llm.model` |
| `LLM_BASE_URL` | `llm.base_url` |
| `MILVUS_ADDRESS` | `milvus.address` |

## 注意事项

1. 配置文件中的敏感信息（如API密钥）不应提交到版本控制系统
2. 可以考虑使用环境变量覆盖配置文件中的敏感信息
3. 为不同环境（开发、测试、生产）准备不同的配置文件 