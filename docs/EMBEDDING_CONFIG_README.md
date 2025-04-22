# 嵌入服务配置说明

## 嵌入服务架构变更

我们已经移除了本地 FastAPI 嵌入服务（embedding-service），现在嵌入功能仅支持以下两种方式：

1. 远程 API 嵌入服务（如 OpenAI API）
2. 本地 Ollama 嵌入服务

这种简化的架构降低了系统复杂度，减少了依赖，同时保留了最常用的两种嵌入方式。

## 配置文件

所有嵌入相关配置都集中在 `config/config.yaml` 文件中：

```yaml
# 嵌入模型配置
embedding:
  service: "ollama" # remote 或 ollama

  # 远程嵌入模型配置（OpenAI API，当 service=remote 时使用）
  remote:
    api_key: "sk-example-embedding-key"
    model: "text-embedding-3-large"
    base_url: "https://api.openai.com/v1"

  # Ollama嵌入模型配置（当 service=ollama 时使用）
  ollama:
    url: "http://localhost:11434"
    model: "mxbai-embed-large"
```

## 配置说明

### 远程 API 服务

使用 OpenAI 或其他兼容 API 的远程嵌入服务：

```yaml
embedding:
  service: "remote"
  remote:
    api_key: "your-api-key"
    model: "text-embedding-3-large"
    base_url: "https://api.openai.com/v1"
```

* `api_key`: 您的 API 密钥
* `model`: 使用的嵌入模型名称
* `base_url`: API 基础 URL，可以替换为其他兼容 OpenAI API 的服务提供商

### Ollama 本地服务

使用 Ollama 在本地运行嵌入模型：

```yaml
embedding:
  service: "ollama"
  ollama:
    url: "http://localhost:11434"
    model: "mxbai-embed-large"
```

* `url`: Ollama 服务的 URL，默认为 http://localhost:11434
* `model`: 使用的 Ollama 模型名称，默认为 mxbai-embed-large

## Ollama 设置指南

要使用 Ollama 嵌入服务，请按照以下步骤操作：

1. 从 [Ollama 官网](https://ollama.com/download) 下载并安装 Ollama
2. 运行以下命令拉取嵌入模型：
   ```bash
   ollama pull mxbai-embed-large
   ```
3. 启动 Ollama 服务：
   ```bash
   OLLAMA_HOST="0.0.0.0" OLLAMA_ORIGINS="*" OLLAMA_KEEP_ALIVE="24h" ollama serve
   ```
4. 在配置文件中设置 `embedding.service` 为 `"ollama"`

## 性能对比

| 服务类型 | 向量维度 | 本地资源占用 | 网络依赖 | API密钥 | 适用场景 |
|---------|---------|------------|---------|---------|---------|
| 远程 API | 1536 | 无 | 需要稳定网络 | 需要 | 生产环境，精确向量化 |
| Ollama | 1024 | 约1.2GB | 不需要 | 不需要 | 生产/开发，高质量离线替代 | 