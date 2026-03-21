# devtest

该目录用于本地开发联调，不属于正式自动化测试。

原则：

- 不会被 `go test ./...` 自动执行
- 允许依赖本地或外部服务
- 敏感信息必须通过环境变量注入，不能写死在源码中

建议环境变量：

- `DEVTEST_LLM_BASE_URL`
- `DEVTEST_LLM_API_KEY`
- `DEVTEST_LLM_MODEL`
- `DEVTEST_MCP_SSE_URL`
- `DEVTEST_MILVUS_ADDRESS`
- `DEVTEST_EMBEDDING_BASE_URL`
- `DEVTEST_EMBEDDING_MODEL`
- `DEVTEST_COLLECTION`
- `DEVTEST_KB_ID`

这些文件目前主要作为开发示例和联调脚本保留。如果后续需要纳入 CI，请迁回正式测试目录并移除外部环境依赖。
