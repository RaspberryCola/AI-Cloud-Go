# AI-Cloud-Go 快速启动指南

这份指南将帮助您快速设置并运行AI-Cloud-Go系统，包括各种支持的嵌入服务配置。

## 前置条件

- Go 1.16+
- Docker和Docker Compose
- Git
- 确保以下端口未被占用: 8080, 3306, 9000, 9001, 19530, 9091, 11434

## 步骤一：获取代码

```bash
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git
cd AI-Cloud-Go
```

## 步骤二：设置环境

### 配置系统参数

1. 确保项目根目录存在`config`文件夹，如果不存在请创建
2. 在`config`文件夹中创建或修改`config.yaml`文件:

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "3306"
  user: "root"
  password: "123456"
  name: "ai_cloud"

storage:
  type: "minio"  # local, oss, minio
  minio:
    endpoint: "localhost:9000"
    bucket: "ai-cloud"
    access_key_id: "minioadmin"
    access_key_secret: "minioadmin"
    use_ssl: false
    region: ""

# Milvus向量数据库配置
milvus:
  address: "localhost:19530"
  index_type: "IVF_FLAT"
  metric_type: "COSINE"
  nlist: 128
  # 搜索参数
  nprobe: 16
  # 字段最大长度配置
  id_max_length: "64"
  content_max_length: "65535"
  doc_id_max_length: "64"
  doc_name_max_length: "256"
  kb_id_max_length: "64"

# 语言模型配置(后续会移除到统一的模型管理中）
llm:
  api_key: "your-llm-api-key" # 替换为您的语言模型API密钥
  model: "deepseek-chat"
  base_url: "https://api.deepseek.com/v1"
  max_tokens: 4096
  temperature: 0.7

rag:
  chunk_size: 1500
  overlap_size: 500
```

## 步骤三：启动基础服务

### 启动所有Docker服务

确保项目根目录包含`docker-compose.yml`文件，然后运行:

```bash
# 启动MySQL, MinIO, Milvus服务
docker-compose up -d
```

或者，如果您只想启动特定服务：

```bash
# 仅启动MySQL和MinIO
docker-compose up -d mysql-init minio minio-init

# 仅启动Milvus相关服务
docker-compose up -d etcd minio-for-milvus standalone
```

### 检查服务状态

```bash
# 列出所有容器
docker ps

# 检查MySQL是否正常初始化
docker logs mysql-init

# 检查MinIO是否正常启动
curl http://localhost:9000

# 检查Milvus是否正常启动
docker logs milvus-standalone
```

确保所有服务都正常运行，没有错误信息。

## 步骤四：设置Ollama (如果使用Ollama作为嵌入服务)

如果您选择使用Ollama作为嵌入服务：

1. 从[Ollama官网](https://ollama.com/download)下载并安装Ollama

2. 拉取嵌入模型：
   ```bash
   ollama pull mxbai-embed-large
   ```

3. 启动Ollama服务：
   ```bash
   OLLAMA_HOST="0.0.0.0" OLLAMA_ORIGINS="*" ollama serve
   ```
   
   您也可以创建一个启动脚本`start-ollama.sh`:
   ```bash
   #!/bin/bash
   OLLAMA_HOST="0.0.0.0" OLLAMA_ORIGINS="*" OLLAMA_KEEP_ALIVE="24h" ollama serve
   ```
   
   然后运行`chmod +x start-ollama.sh`和`./start-ollama.sh`

4. 在前端模型服务中添加模型：

5. 验证Ollama服务是否正常运行：
   ```bash
   curl http://localhost:11434/api/tags
   ```
   应返回包含`mxbai-embed-large`的模型列表。

## 步骤五：确认Milvus配置

Milvus的连接地址和向量集合参数在配置文件中指定：

```yaml
# Milvus向量数据库配置
milvus:
  address: "localhost:19530"
  index_type: "IVF_FLAT"
  metric_type: "COSINE"
  nlist: 128
  # 搜索参数
  nprobe: 16
  # 字段最大长度配置
  id_max_length: "64"
  content_max_length: "65535"
  doc_id_max_length: "64"
  doc_name_max_length: "256"
  kb_id_max_length: "64"
```

如果您使用的是自定义的Milvus部署或远程Milvus服务，请相应地修改地址。您也可以根据需要调整以下参数：

- `index_type`: 索引类型，支持IVF_FLAT、IVF_SQ8、HNSW等
- `metric_type`: 距离计算方式，支持COSINE、L2、IP等
- `nlist`: IVF索引的聚类数量
- `nprobe`: 搜索时检查的聚类数量，值越大结果越精确但查询越慢
- `*_max_length`: 各字段的最大长度设置，特别是处理大文档时可能需要调整content_max_length

验证Milvus是否正常运行：

```bash
# 检查Milvus容器状态
docker ps | grep milvus

# 查看Milvus日志
docker logs milvus-standalone
```

## 步骤六：启动AI-Cloud-Go

### 下载依赖

```bash
go mod download
```

### 运行应用

```bash
go run cmd/main.go
```

应用将在http://localhost:8080运行。您应该看到类似以下的输出：

```
[GIN-debug] Listening and serving HTTP on :8080
```

## 步骤七：验证安装

### 检查API服务

```bash
# 检查健康状态
curl http://localhost:8080/api/health

# 使用Swagger查看API文档
# 在浏览器中访问: http://localhost:8080/swagger/index.html
```

### 测试嵌入服务

根据您选择的嵌入服务，验证其正常工作：

```bash
# 如果使用OpenAI
curl -H "Authorization: Bearer 您的API密钥" https://api.openai.com/v1/embeddings -d '{"model":"text-embedding-3-large", "input":"测试文本"}'

# 如果使用Ollama
curl -X POST http://localhost:11434/api/embed -d '{"model":"mxbai-embed-large", "input":"测试文本"}'
```

### 注册和登录

要使用系统的大多数功能，您需要先注册并登录以获取JWT令牌：

1. 注册用户:
   ```bash
   curl -X POST http://localhost:8080/api/users/register -H "Content-Type: application/json" -d '{"username":"testuser", "password":"testpassword", "email":"test@example.com"}'
   ```

2. 登录并获取Token:
   ```bash
   curl -X POST http://localhost:8080/api/users/login -H "Content-Type: application/json" -d '{"username":"testuser", "password":"testpassword"}'
   ```
   
   复制返回的`token`值，供后续API调用使用。

## 使用知识库功能

使用您在登录时获得的JWT令牌：

1. 创建知识库：
   ```bash
   curl -X POST http://localhost:8080/api/kb/create \
     -H "Authorization: Bearer 您的JWT令牌" \
     -H "Content-Type: application/json" \
     -d '{"name":"测试知识库","description":"这是一个测试知识库"}'
   ```

2. 上传文档到知识库（使用multipart/form-data）：
   ```bash
   curl -X POST http://localhost:8080/api/kb/addNew \
     -H "Authorization: Bearer 您的JWT令牌" \
     -F "kb_id=知识库ID" \
     -F "file=@/path/to/your/document.pdf"
   ```

3. 查询知识库：
   ```bash
   curl -X POST http://localhost:8080/api/kb/retrieve \
     -H "Authorization: Bearer 您的JWT令牌" \
     -H "Content-Type: application/json" \
     -d '{"kb_id":"知识库ID","query":"您的问题","top_k":3}'
   ```

## 故障排除

### 初始化问题
- 如果启动时显示找不到配置文件，请确保`config/config.yaml`文件存在并格式正确
- 如果显示找不到`init.sql`，请确认项目根目录中有此文件

### MySQL连接问题
- 错误：`dial tcp 127.0.0.1:3306: connect: connection refused`
- 解决：确保MySQL容器已启动，运行`docker ps | grep mysql`

### Milvus连接问题
- 错误：`无法连接到Milvus`
- 解决：
  1. 确保Milvus容器正在运行：`docker ps | grep milvus`
  2. 检查config.yaml中的milvus.address配置是否正确
  3. 如果使用自定义Milvus部署，确保端口映射正确

