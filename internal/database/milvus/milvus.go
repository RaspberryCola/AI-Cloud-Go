package milvus

import (
	"ai-cloud/config"
	"context"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"sync"
)

var (
	once     sync.Once
	instance client.Client
	err      error
)

// InitMilvus 初始化
func InitMilvus(ctx context.Context) (client.Client, error) {

	once.Do(func() {
		instance, err = client.NewClient(ctx, client.Config{
			Address: config.GetConfig().Milvus.Address,
		})

	})
	return instance, err
}

func GetMilvusClient() client.Client {
	return instance
}
