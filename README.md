# AI-Cloud-Go — A Golang-Based Cloud Drive & Knowledge Base System

[English](README.md) | [中文](README_CN.md)

## Overview
AI-Cloud-Go is a cloud drive and knowledge base system built using the Go programming language. It offers features such as file storage, user management, knowledge base management, and model management, leveraging modern technology stacks and architecture design. The system supports multiple storage backends and integrates with a vector database to enable intelligent search capabilities.

Frontend repository: <svg height="16" width="16" viewBox="0 0 16 16" fill="currentColor" style="display: inline-block; vertical-align: middle;">
<path fill-rule="evenodd" d="M8 0C3.58 0 0 0 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path>
</svg> [AI-Cloud-Frontend](https://github.com/RaspberryCola/AI-Cloud-Frontend)

## Key Features

### Completed:
- [x] User system: Supports registration, login, and authentication
- [x] Cloud file system: Supports file upload, download, and management
- [x] Knowledge base management: Create and manage knowledge bases, import files from cloud or upload new ones
- [x] Model management: Manage custom LLM and Embedding models
- [x] Multiple storage backends: Supports local storage, MinIO, Alibaba Cloud OSS, etc.
- [x] Vector search: Integrated with Milvus vector database for smart document retrieval

### In Development:
- [ ] Agent functionality

## Tech Stack
- Backend Framework: Gin
- Database: MySQL
- Vector Database: Milvus
- Object Storage: MinIO / Alibaba Cloud OSS
- Authentication: JWT (JSON Web Token)
- LLM Framework: Eino
- Others: CORS middleware, custom middleware, etc.

## Directory Structure
```
.
├── cmd/                    # Main application entry point
├── config/                 # Configuration files
├── docs/                   # Documentation
│   ├── CONFIG_README.md    # Configuration guide
│   └── QUICKSTART.md       # Quick start guide
├── internal/               # Internal packages
│   ├── component/          # Large model related services
│   ├── controller/         # Controller layer
│   ├── service/            # Business logic layer
│   ├── dao/                # Data access layer
│   ├── middleware/         # Middleware components
│   ├── router/             # Routing configuration
│   ├── database/           # Database (MySQL/Milvus...)
│   ├── model/              # Data models
│   ├── storage/            # Storage backend implementations (MinIO/OSS...)
│   └── utils/              # Utility functions
├── pkgs/                   # Shared packages
├── docker-compose.yml      # Docker configuration
├── go.mod                  # Go module file
└── go.sum                  # Dependency version lock
```

## Usage Instructions

1. Clone the project:
```bash
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git
cd AI-Cloud-Go
```

2. Install dependencies:
```bash
go mod download
```

3. Environment setup:
- Ensure MySQL is installed and running
- Ensure Milvus is installed if using vector search
- Configure storage backend (local/MinIO/Alibaba Cloud OSS)

You can use Docker Compose for quick setup:
```bash
docker-compose up -d
```

4. Modify configuration:
- Update `config/config.yaml` accordingly

5. Run the project:
```bash
go run cmd/main.go
```

# AI-Cloud-Go — Docker Full Environment Setup

> For more details, see the [docs folder](/docs/)

## Prerequisites
- Docker and Docker Compose installed
- Basic understanding of Docker usage

## Services Included
This setup includes the following services:
- **MySQL**: Database server, configured with the `ai_cloud` database
- **MinIO**: Object storage service compatible with S3 protocol, configured with bucket `ai-cloud`
- **Milvus**: Vector database for storing and retrieving document vectors

## Setup Steps

1. **Environment Configuration**
   - Edit `config/config.yaml` to configure service connection info and LLM API keys
   ```yaml
   llm:
     api_key: "your-llm-api-key"
     model: "deepseek-chat"
     base_url: "https://api.deepseek.com/v1"
   ```
   ⚠️ LLM configuration will be moved into unified model management in the future

2. **Start Docker Containers**
   ```bash
   docker-compose up -d
   ```

3. **Check Service Status**
   ```bash
   docker ps
   ```

4. **Run the Application**
   ```bash
   go run cmd/main.go
   ```

5. **Stop Containers**
   ```bash
   docker-compose down
   ```

## Configuration Details

### MySQL
- Host: localhost
- Port: 3306
- Username: root
- Password: 123456
- Database: ai_cloud

### MinIO (Object Storage)
- Endpoint: localhost:9000
- Management Console: http://localhost:9001
- Access Key: minioadmin
- Secret Key: minioadmin
- Bucket: ai-cloud

### Milvus (Vector Database)
- Config path: `milvus.address` in `config.yaml`
- Default address: localhost:19530
- Admin UI: Requires installing Attu (official Milvus GUI tool)
   - Lightweight web UI on port 9091: http://localhost:9091/webui (for Milvus 2.5.0+)
   - Full admin interface via Attu:
     ```bash
     docker run -p 8000:3000 -e MILVUS_URL=localhost:19530 zilliz/attu:latest
     ```
   - Visit Attu at: http://localhost:8000

### Environment Configuration
The system uses `config/config.yaml` to configure service connections and AI model access, including:

1. **Language Model Configuration**
   - Used for Q&A and intelligent processing
   - Default: DeepSeek's `deepseek-chat` model

2. **Milvus Configuration**
   - For vector storage and retrieval
   - Config: `milvus.address`

## Troubleshooting

### Common Issues

1. **Application Hangs on Startup**
   - Check if Milvus started correctly: `docker logs milvus-standalone`
   - Confirm Milvus address in `config.yaml` is correct
   - Ensure valid API key is set in config.yaml
   - Make sure the `ai_cloud` database exists in MySQL

2. **MinIO Connection Problems**
   - Verify MinIO is running
   - Confirm config.yaml matches actual MinIO environment settings

3. **Vector DB Operations Fail**
   - Check Milvus service status
   - Ensure vector dimension matches model output
   - Confirm `milvus.address` is properly set

4. **Cannot Access Milvus Web UI**
   - Milvus 2.5+ provides basic UI at http://localhost:9091/webui
   - For full UI, install Attu with:
     ```bash
     docker run -p 8000:3000 -e MILVUS_URL=localhost:19530 zilliz/attu:latest
     ```
   - Access at: http://localhost:8000

## Development and Debugging

### API Testing
After starting the app, you can test APIs like:
- User Registration: POST http://localhost:8080/api/users/register
- User Login: POST http://localhost:8080/api/users/login
- File Upload: POST http://localhost:8080/api/files/upload
- See code for more endpoints

### Service URLs
- App Backend: http://localhost:8080
- MinIO Console: http://localhost:9001
- Milvus Admin UI: http://localhost:9091
- Ollama Service: http://localhost:11434 (if enabled)

## Notes
- If you already have MySQL running locally, create the db via:
  ```bash
  mysql -u root -p < init.sql
  ```
- Milvus automatically creates collections/indexes on first launch (may take time)
- Replace sample API keys in config.yaml with your own credentials before production
- Ensure config.yaml reflects Docker service addresses
- Go server uses port 8080; Ollama uses 11434 – avoid conflicts
