# AI-Cloud-Go 基于Golang开发的云盘知识库系统

## 项目简介
AI-Cloud-Go 是一个基于 Go 语言开发的智能云存储系统，提供文件存储、用户管理、知识库管理等功能，采用现代化的技术栈和架构设计。系统支持多种存储后端，并集成了向量数据库以支持智能检索功能。

## 主要特性
- 用户系统：支持用户注册、登录、认证
- 文件管理：支持文件上传、下载、管理
- 知识库管理：支持创建和管理知识库，文档智能处理
- 多存储后端：支持本地存储、MinIO、阿里云OSS等多种存储方式
- 向量检索：集成Milvus向量数据库，支持智能文档检索
- JWT 认证：使用 JWT 进行用户身份验证
- RESTful API：提供标准的 RESTful 接口
- 跨域支持：内置 CORS 跨域支持

## 技术栈
- 后端框架：Gin
- 数据库：MySQL
- 向量数据库：Milvus
- 对象存储：MinIO/阿里云OSS
- 认证：JWT (JSON Web Token)
- 其他：跨域中间件、自定义中间件等

## 目录结构
```
.
├── cmd/                    # 主程序入口
├── config/                 # 配置文件
├── internal/              # 内部包
│   ├── controller/        # 控制器层
│   ├── service/          # 业务逻辑层
│   ├── dao/              # 数据访问层
│   ├── middleware/       # 中间件
│   ├── router/           # 路由配置
│   ├── database/         # 数据库配置
│   ├── model/           # 数据模型
│   ├── storage/         # 存储实现
│   └── utils/           # 工具函数
├── pkgs/                  # 公共包
├── storage_data/         # 存储数据
├── go.mod                # Go 模块文件
└── go.sum                # 依赖版本锁定文件
```

## 安装说明
1. 克隆项目
```bash
git clone https://github.com/RaspberryCola/AI-Cloud-Go.git
cd AI-Cloud-Go
```

2. 安装依赖
```bash
go mod download
```

3. 配置
- 确保已安装并启动 MySQL
- 确保已安装并启动 Milvus（如需使用向量检索功能）
- 配置存储后端（本地存储/MinIO/阿里云OSS）
- 修改 `config/config.yaml` 中的相关配置

4. 运行项目
```bash
go run cmd/main.go
```