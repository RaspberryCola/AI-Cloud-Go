# AI-Cloud-Go

AI-Cloud-Go is a Go-based backend for LLM applications. It provides user management, file storage, knowledge base management, model management, agent configuration, chat, and conversation history, with Milvus-based retrieval and multiple object storage backends.

[English](README_EN) | [中文](README)

## Features

- User registration, login, and authentication
- File upload, download, and management
- Knowledge base creation, document import, retrieval, and chat
- Custom LLM and embedding model management
- Agent configuration, debug chat, and conversation history
- Support for local storage, MinIO, and Alibaba Cloud OSS

## Tech Stack

- Go 1.23.4
- Gin
- GORM
- MySQL
- Milvus
- MinIO / OSS
- JWT
- Eino

## Quick Start

### Requirements

- Go 1.23.4+
- Docker / Docker Compose

### Install Dependencies

```bash
go mod download
```

### Configuration

The backend supports two configuration sources:

- `config/config.yaml`
- environment variable overrides

See [config/config.yaml.example](./config/config.yaml.example) for the template.

### Start Dependencies

To quickly start the default dependency stack:

```bash
docker-compose up -d
```

This starts:

- MySQL
- MinIO
- Milvus

### Start the Server

```bash
go run cmd/main.go
```

Default server address: `http://localhost:8080`

## Deployment

### Standalone Backend Deployment

Use this when the frontend is deployed separately or when you only need the API service.

Option A, run from source:

```bash
go run cmd/main.go
```

Option B, run with Docker:

```bash
docker build -t ai-cloud-go .
docker run -p 8080:8080 --env-file .env ai-cloud-go
```

### Full Stack Deployment

Use this when you want to deploy frontend and backend together.

```bash
mkdir ai-cloud
cd ai-cloud
git clone https://github.com/RaspberryCola/AI-Cloud-Frontend.git frontend
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git backend
cd backend/deploy
cp .env.example .env
make up
```

See [deploy/README.md](./deploy/README.md) for full deployment details.

## Project Structure

```text
.
├── cmd/              # Application entry
├── config/           # Configuration
├── deploy/           # Full-stack deployment entry
├── docs/             # Detailed docs
├── internal/         # Core application code
├── pkgs/             # Shared packages
├── docker-compose.yml
├── go.mod
└── go.sum
```

## More Docs

- [deploy/README.md](./deploy/README.md)
- [docs/CONFIG_README.md](./docs/CONFIG_README.md)
- [docs/QUICKSTART.md](./docs/QUICKSTART.md)
