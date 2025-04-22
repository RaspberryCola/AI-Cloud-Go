package database

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func InitMilvus(ctx context.Context) (client.Client, error) {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: "localhost:19530",
	})
	if err != nil {
		return nil, fmt.Errorf("无法连接到Milvus: %w", err)
	}
	// 初始化文本chunks集合
	if err := initTextChunksCollection(ctx, milvusClient); err != nil {
		return nil, fmt.Errorf("初始化集合失败: %w", err)
	}
	return milvusClient, nil
}

func initTextChunksCollection(ctx context.Context, milvusClinet client.Client) error {
	collectionName := "text_chunks"

	exists, err := milvusClinet.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合存在失败: %w", err)
	}

	if exists {
		return nil
	}

	//if exists {
	//	// 如果集合已存在，先删除它
	//	if err := milvusClinet.DropCollection(ctx, collectionName); err != nil {
	//		return fmt.Errorf("删除现有集合失败: %w", err)
	//	}
	//	fmt.Println("已删除现有的text_chunks集合，将重新创建")
	//}

	// 创建集合
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "存储文档分块和向量",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					"max_length": "64",
				},
			},
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "65535",
				},
			},
			{
				Name:     "document_id",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "64",
				},
			},
			{
				Name:     "document_name",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "256",
				},
			},
			{
				Name:     "kb_id",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "64",
				},
			},
			{
				Name:     "chunk_index",
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": "1024",
				},
			},
		},
	}

	if err := milvusClinet.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("创建集合失败: %w", err)
	}

	// 创建索引
	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	if err := milvusClinet.CreateIndex(ctx, collectionName, "vector", idx, false); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	// 加载集合到内存
	if err := milvusClinet.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("加载集合失败: %w", err)
	}

	return nil
}
