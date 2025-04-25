package database

import (
	"ai-cloud/config"
	"ai-cloud/pkgs/consts"
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// InitMilvus 初始化
func InitMilvus(ctx context.Context) (client.Client, error) {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: config.GetConfig().Milvus.Address,
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

// initTextChunksCollection 初始化集合
func initTextChunksCollection(ctx context.Context, milvusClinet client.Client) error {
	// 从配置中获取集合名称
	milvusConfig := config.GetConfig().Milvus
	collectionName := milvusConfig.CollectionName

	// 检查是否存在
	exists, err := milvusClinet.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查集合存在失败: %w", err)
	}
	if exists {
		return nil
	}

	// 创建集合
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "存储文档分块和向量",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       consts.FieldNameID,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					"max_length": milvusConfig.IDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameContent,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.ContentMaxLength,
				},
			},
			{
				Name:     consts.FieldNameDocumentID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.DocIDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameDocumentName,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.DocNameMaxLength,
				},
			},
			{
				Name:     consts.FieldNameKBID,
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": milvusConfig.KbIDMaxLength,
				},
			},
			{
				Name:     consts.FieldNameChunkIndex,
				DataType: entity.FieldTypeInt32,
			},
			{
				Name:     consts.FieldNameVector,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", milvusConfig.VectorDimension),
				},
			},
		},
	}

	// 创建集合
	if err := milvusClinet.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("initTextChunksCollection failed, CreateCollection err: %+v", err)
	}

	// 构建索引
	idx, indexErr := milvusConfig.GetMilvusIndex()
	if indexErr != nil {
		return fmt.Errorf("initTextChunksCollection failed, GetMilvusIndex err: %+v", err)
	}

	// 创建索引
	if err := milvusClinet.CreateIndex(ctx, collectionName, consts.FieldNameVector, idx, false); err != nil {
		return fmt.Errorf("initTextChunksCollection failed, CreateIndex err: %+v", err)
	}

	// 加载集合到内存
	if err := milvusClinet.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("initTextChunksCollection failed, LoadCollection err: %+v", err)
	}
	return nil
}

//
//// 新增函数：根据模型信息创建collection
//func CreateCollection(ctx context.Context, milvusClient client.Client, collectionName string, dimension int) error {
//	milvusConfig := config.GetConfig().Milvus
//
//	exists, err := milvusClient.HasCollection(ctx, collectionName)
//	if err != nil {
//		return fmt.Errorf("检查集合存在失败: %w", err)
//	}
//
//	if exists {
//		return nil
//	}
//
//	// 创建集合
//	schema := &entity.Schema{
//		CollectionName: collectionName,
//		Description:    "存储文档分块和向量",
//		AutoID:         false,
//		Fields: []*entity.Field{
//			{
//				Name:       FieldNameID,
//				DataType:   entity.FieldTypeVarChar,
//				PrimaryKey: true,
//				AutoID:     false,
//				TypeParams: map[string]string{
//					"max_length": milvusConfig.IDMaxLength,
//				},
//			},
//			{
//				Name:     FieldNameContent,
//				DataType: entity.FieldTypeVarChar,
//				TypeParams: map[string]string{
//					"max_length": milvusConfig.ContentMaxLength,
//				},
//			},
//			{
//				Name:     FieldNameDocumentID,
//				DataType: entity.FieldTypeVarChar,
//				TypeParams: map[string]string{
//					"max_length": milvusConfig.DocIDMaxLength,
//				},
//			},
//			{
//				Name:     FieldNameDocumentName,
//				DataType: entity.FieldTypeVarChar,
//				TypeParams: map[string]string{
//					"max_length": milvusConfig.DocNameMaxLength,
//				},
//			},
//			{
//				Name:     FieldNameKBID,
//				DataType: entity.FieldTypeVarChar,
//				TypeParams: map[string]string{
//					"max_length": milvusConfig.KbIDMaxLength,
//				},
//			},
//			{
//				Name:     FieldNameChunkIndex,
//				DataType: entity.FieldTypeInt32,
//			},
//			{
//				Name:     FieldNameVector,
//				DataType: entity.FieldTypeFloatVector,
//				TypeParams: map[string]string{
//					"dim": strconv.Itoa(dimension),
//				},
//			},
//		},
//	}
//
//	if err := milvusClient.CreateCollection(ctx, schema, 1); err != nil {
//		return fmt.Errorf("创建集合失败: %w", err)
//	}
//
//	// 创建索引
//	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
//	if err != nil {
//		return fmt.Errorf("创建索引失败: %w", err)
//	}
//
//	if err := milvusClient.CreateIndex(ctx, collectionName, "vector", idx, false); err != nil {
//		return fmt.Errorf("创建索引失败: %w", err)
//	}
//
//	// 加载集合到内存
//	if err := milvusClient.LoadCollection(ctx, collectionName, false); err != nil {
//		return fmt.Errorf("加载集合失败: %w", err)
//	}
//
//	return nil
//}
