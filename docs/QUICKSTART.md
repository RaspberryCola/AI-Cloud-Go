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

# 嵌入模型配置
embedding:
  service: "remote" # remote 或 ollama
  
  # 远程嵌入模型配置（OpenAI API，当 service=remote 时使用）
  remote:
    api_key: "your-api-key" # 替换为您的API密钥
    model: "text-embedding-3-large"
    base_url: "https://api.openai.com/v1"
  
  # Ollama嵌入模型配置（当 service=ollama 时使用）
  ollama:
    url: "http://localhost:11434"
    model: "mxbai-embed-large"

# 语言模型配置
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

4. 在`config/config.yaml`中设置：
   ```yaml
   embedding:
     service: "ollama"
     ollama:
       url: "http://localhost:11434"
       model: "mxbai-embed-large"
   ```

5. 验证Ollama服务是否正常运行：
   ```bash
   curl http://localhost:11434/api/tags
   ```
   应返回包含`mxbai-embed-large`的模型列表。

## 步骤五：启动AI-Cloud-Go

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

## 步骤六：验证安装

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
- 错误：`存储向量到 Milvus 失败: 插入数据失败: the length of document_name exceeds max length`
- 解决：文档名称过长，已在新版本中自动处理，如果仍有问题，请升级代码

### 向量数组为空
- 错误：`num_rows should be greater than 0: invalid parameter`
- 解决：检查文档内容是否可提取文本，确保嵌入服务正常工作

### 嵌入服务问题
- 如果使用OpenAI：检查API密钥和网络连接
- 如果使用Ollama：确保模型已下载并且服务运行正常
  - 运行`ollama list`查看已安装的模型
  - 运行`curl http://localhost:11434/api/tags`检查Ollama服务

### 权限问题
- 如果遇到API权限错误，确保您已登录并在请求头中包含有效的JWT令牌
- 错误格式：`{"error":"无效的token"}`或`{"error":"请先登录"}`

## 下一步

- 查看[完整文档](./README.md)了解更多细节
- 阅读[嵌入服务配置文档](./EMBEDDING_CONFIG_README.md)了解嵌入服务架构
- 阅读[配置指南](./CONFIG_README.md)了解完整配置选项
- 探索[API文档](http://localhost:8080/swagger/index.html)了解所有可用端点 