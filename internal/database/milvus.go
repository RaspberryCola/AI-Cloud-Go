package database

import (
	"ai-cloud/config"
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

// InitMilvus 初始化
func InitMilvus(ctx context.Context) (client.Client, error) {
	// 初始化 Milvus 客户端
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: config.GetConfig().Milvus.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("无法连接到Milvus: %w", err)
	}

	return milvusClient, nil
}
