# Deploy

该目录用于前后端联合部署。

## 目录约定

请按以下结构组织仓库：

```text
ai-cloud/
├── frontend/   # AI-Cloud-Frontend
└── backend/    # AI-Cloud-Go
```

其中当前目录应为：

```text
ai-cloud/backend/deploy
```

## 快速开始

```bash
cp .env.example .env
make up
```

## 常用命令

```bash
make up
make down
make logs
make config
```

## 说明

- `docker-compose.full.yml` 会同时启动前端、后端、MySQL、MinIO、Milvus
- 前端构建上下文要求同级目录存在 `frontend/`
- 后端构建上下文为上级目录 `backend/`
- 部署前请先修改 `.env` 中的密钥、数据库、对象存储和 LLM 配置
