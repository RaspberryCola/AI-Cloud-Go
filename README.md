# AI-Cloud-Go

AI-Cloud-Go 是一个基于 Go 的 LLM 应用后端，提供用户管理、文件存储、知识库管理、模型管理、Agent 配置与对话能力，并集成 Milvus 向量检索与多种对象存储后端。

[English](README_EN) | [中文](README)

## 功能

- 用户注册、登录、认证
- 文件上传、下载、管理
- 知识库创建、文档导入、检索与问答
- 自定义 LLM / Embedding 模型管理
- Agent 配置、调试对话、历史会话
- 支持本地存储、MinIO、阿里云 OSS

## 技术栈

- Go 1.23.4
- Gin
- GORM
- MySQL
- Milvus
- MinIO / OSS
- JWT
- Eino

## 快速开始

### 环境要求

- Go 1.23.4+
- Docker / Docker Compose

### 安装依赖

```bash
go mod download
```

### 配置

后端支持两种配置方式：

- `config/config.yaml`
- 环境变量覆盖

配置模板见 [config/config.yaml.example](./config/config.yaml.example)。

### 启动依赖

如果你希望快速启动默认依赖：

```bash
docker-compose up -d
```

默认会启动：

- MySQL
- MinIO
- Milvus

### 启动服务

```bash
go run cmd/main.go
```

默认服务地址：`http://localhost:8080`

## 部署

### 单独部署后端

适用于前端单独部署或只提供 API 服务的场景。

方式一，源码运行：

```bash
go run cmd/main.go
```

方式二，Docker 运行：

```bash
docker build -t ai-cloud-go .
docker run -p 8080:8080 --env-file .env ai-cloud-go
```

### 联合部署前后端

适用于前后端一体化部署的场景。

```bash
mkdir ai-cloud
cd ai-cloud
git clone https://github.com/RaspberryCola/AI-Cloud-Frontend.git frontend
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git backend
cd backend/deploy
cp .env.example .env
make up
```

详细部署说明见 [deploy/README.md](./deploy/README.md)。

## 目录结构

```text
.
├── cmd/              # 程序入口
├── config/           # 配置
├── deploy/           # 联合部署入口
├── docs/             # 详细文档
├── internal/         # 核心业务代码
├── pkgs/             # 公共包
├── docker-compose.yml
├── go.mod
└── go.sum
```

## 更多文档

- [deploy/README.md](./deploy/README.md)
- [docs/CONFIG_README.md](./docs/CONFIG_README.md)
- [docs/QUICKSTART.md](./docs/QUICKSTART.md)
