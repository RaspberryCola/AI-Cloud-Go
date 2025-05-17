# AI-Cloud-Go 基于Golang开发的云盘知识库系统

[English](README_EN) | [中文](README)

## 项目简介
AI-Cloud-Go 是一个基于 Go 语言开发的云盘知识库系统，提供文件存储、用户管理、知识库管理、模型管理等功能，采用现代化的技术栈和架构设计。系统支持多种存储后端，并集成了向量数据库以支持智能检索功能。

前端界面展示：<svg height="16" width="16" viewBox="0 0 16 16" fill="currentColor" style="display: inline-block; vertical-align: middle;">
<path fill-rule="evenodd" d="M8 0C3.58 0 0 0 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
</svg> [AI-Cloud-Frontend](https://github.com/RaspberryCola/AI-Cloud-Frontend)


## 主要功能
**已经实现：**

- [x] 用户模块：支持用户注册、登录、认证
- [x] 多存储后端：支持本地存储、MinIO、阿里云OSS等多种存储方式
- [x] 文件模块：支持文件上传、下载、管理
- [x] 知识库模块：支持创建和管理知识库，支持导入云盘文件或上传新文件
- [x] 模型模块：支持创建和管理自定义LLM模型和Embedding模型
- [x] Agent模块：支持创建和管理Agent
  - [x] 支持自定义LLM、知识库、MCP
  - [x] 对话界面、历史对话


**未来优化**

- [ ] 知识库模块：
  - [ ] 多文件上传
  - [ ] 优化解析状态处理
  - [ ] 支持Rerank

- [ ] Agent模块：
  - [ ] 自定义Tool（HTTP工具）
  - [ ] 优化LLM的参数配置
  - [ ] 跨知识库检索的Rerank实现
- [ ] 模型管理：
  - [ ] 添加常用模型预设：OpenAI，DeepSeek，火山引擎等
  - [ ] 支持Rerank模型

## 技术栈
- 后端框架：Gin
- 数据库：MySQL
- 向量数据库：Milvus
- 对象存储：MinIO/阿里云OSS
- 认证：JWT
- LLM框架：Eino
- 其他：跨域中间件、自定义中间件等

## 目录结构
```
.
├── cmd/                    # 主程序入口
├── config/                 # 配置文件
├── docs/                   # 文档目录
│   ├── CONFIG_README.md    # 配置指南
│   └── QUICKSTART.md       # 快速启动指南
├── internal/               # 内部包
│   ├── component/          # 大模型相关服务
│   ├── controller/         # 控制器层
│   ├── service/            # 业务逻辑层
│   ├── dao/                # 数据访问层
│   ├── middleware/         # 中间件
│   ├── router/             # 路由配置
│   ├── database/           # 数据库（MySQL/Milvus...）
│   ├── model/              # 数据模型
│   ├── storage/            # 后端存储实现（Minio/OSS...）
│   └── utils/              # 工具函数
├── pkgs/                   # 公共包
├── docker-compose.yml      # Docker配置文件
├── go.mod                  # Go 模块文件
└── go.sum                  # 依赖版本锁定文件
```

## 使用说明
1. 克隆项目
```bash
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git
cd AI-Cloud-Go
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境
- 确保已安装并启动 MySQL
- 确保已安装并启动 Milvus（如需使用向量检索功能）
- 配置存储后端（本地存储/MinIO/阿里云OSS）

可以通过Docker快速配置：
```bash
docker-compose up -d
```
4. 修改配置信息
- 修改 `config/config.yaml` 中的相关配置 

5. 运行项目
```bash
go run cmd/main.go
```

# AI-Cloud-Go Docker 完整环境配置

> 更多详情请看 [docs 文件夹](/docs/)

## 前置条件
- 已安装 Docker 和 Docker Compose
- 基本了解 Docker 的使用

## 服务组件
本配置包含以下服务：
- **MySQL**: 数据库服务器，配置 `ai_cloud` 数据库
- **MinIO**: 对象存储服务，兼容 S3 协议，配置 `ai-cloud` 存储桶
- **Milvus**: 向量数据库，用于存储和检索文档向量

## 启动步骤

1. 环境配置
   - 修改 `config/config.yaml` 文件，配置各服务连接信息和LLM的API密钥信息
   ```yaml
   llm:
     api_key: "your-llm-api-key"
     model: "deepseek-chat"
     base_url: "https://api.deepseek.com/v1"
   ```
   ⚠️语言模型配置后续将会移动到统一的模型服务管理中
2. 启动 Docker 容器
   ```bash
   docker-compose up -d
   ```

3. 检查服务状态
   ```bash
   docker ps
   ```

4. 运行项目
   ```bash
   go run cmd/main.go
   ```

5. 停止容器
   ```bash
   docker-compose down
   ```

## 配置详情

### MySQL
- 主机: localhost
- 端口: 3306
- 用户名: root
- 密码: 123456
- 数据库: ai_cloud

### MinIO (对象存储)
- 端点: localhost:9000
- 管理控制台: http://localhost:9001
- 访问密钥: minioadmin
- 密钥: minioadmin
- 存储桶: ai-cloud

### Milvus (向量数据库)
- 配置路径: `milvus.address` 在 `config.yaml` 中
- 默认地址: localhost:19530
- 管理界面: 需要额外安装Attu (Milvus官方GUI工具)
  - 在端口9091只提供监控信息: http://localhost:9091/webui (Milvus 2.5.0+版本)
  - 完整管理界面需安装Attu: `docker run -p 8000:3000 -e MILVUS_URL=localhost:19530 zilliz/attu:latest`
  - 访问Attu: http://localhost:8000

### 环境配置
项目使用 `config/config.yaml` 文件配置服务连接和第三方 AI 模型的访问，主要包括：

1.**语言模型配置**
   - 用于知识库问答和智能处理
   - 默认使用 DeepSeek 的 deepseek-chat 模型

2.**Milvus配置**
   - 用于向量存储和检索
   - 配置项: `milvus.address`

## 故障排除

### 常见问题

1. **程序启动卡住**
   - 检查 Milvus 是否正常启动，查看日志 `docker logs milvus-standalone`
   - 检查 `config/config.yaml` 文件中的 Milvus 地址配置是否正确
   - 检查 `config/config.yaml` 文件是否配置了正确的 API 密钥
   - 确保 MySQL 中已创建 ai_cloud 数据库

2. **MinIO 连接问题**
   - 检查 MinIO 服务是否正常运行
   - 验证 config.yaml 中的 MinIO 配置是否与实际运行环境一致

3. **向量数据库操作失败**
   - 检查 Milvus 服务状态
   - 确认向量维度设置是否与模型输出一致
   - 确认 config.yaml 中的 milvus.address 配置是否正确

4. **Milvus 管理界面访问失败**
   - Milvus 2.5以上版本在端口9091提供简易WebUI: http://localhost:9091/webui
   - 完整的管理界面需要安装Attu工具: `docker run -p 8000:3000 -e MILVUS_URL=localhost:19530 zilliz/attu:latest`

## 开发调试

### API 测试
项目启动后，可通过以下端点进行测试：
- 用户注册: POST http://localhost:8080/api/users/register
- 用户登录: POST http://localhost:8080/api/users/login
- 文件上传: POST http://localhost:8080/api/files/upload
- 更多 API 详见代码中的路由配置

### 服务地址
- 应用后端: http://localhost:8080
- MinIO 控制台: http://localhost:9001
- Milvus 管理界面: http://localhost:9091
- Ollama 服务: http://localhost:11434 (如果启用)

## 注意事项
- 如果您已经在本地运行了 MySQL，可以直接使用 `mysql -u root -p < init.sql` 创建 ai_cloud 数据库
- 首次启动时，Milvus 会自动创建必要的集合和索引，可能需要一些时间
- 使用生产环境时，请替换配置文件中的示例 API 密钥为您自己的有效密钥
- 配置 config.yaml 时确保数据库和存储服务的连接信息与 Docker 环境一致
- Go主程序使用端口8080，Ollama服务使用端口11434，避免端口冲突