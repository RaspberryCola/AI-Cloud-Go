package milvus

// defaultSchema is the default schema for milvus by eino
type defaultSchema struct {
	ID           string    `json:"id" milvus:"name:id"`
	Content      string    `json:"content" milvus:"name:content"`
	DocumentID   string    `json:"document_id" milvus:"name:document_id"`
	DocumentName string    `json:"document_name" milvus:"name:document_name"`
	KBID         string    `json:"kb_id" milvus:"name:kb_id"`
	ChunkIndex   int32     `json:"chunk_index" milvus:"name:chunk_index"`
	Vector       []float32 `json:"vector" milvus:"name:vector"`
	//Metadata []byte `json:"metadata" milvus:"name:metadata"`
}

//
//func getDefaultFields() []*entity.Field {
//	// 获取 Milvus 配置
//	milvusConfig := config.GetConfig().Milvus
//
//	return []*entity.Field{
//		{
//			Name:       consts.FieldNameID,
//			DataType:   entity.FieldTypeVarChar,
//			PrimaryKey: true,
//			AutoID:     false,
//			TypeParams: map[string]string{
//				"max_length": milvusConfig.IDMaxLength,
//			},
//		},
//		{
//			Name:     consts.FieldNameContent,
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": milvusConfig.ContentMaxLength,
//			},
//		},
//		{
//			Name:     consts.FieldNameDocumentID,
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": milvusConfig.DocIDMaxLength,
//			},
//		},
//		{
//			Name:     consts.FieldNameDocumentName,
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": milvusConfig.DocNameMaxLength,
//			},
//		},
//		{
//			Name:     consts.FieldNameKBID,
//			DataType: entity.FieldTypeVarChar,
//			TypeParams: map[string]string{
//				"max_length": milvusConfig.KbIDMaxLength,
//			},
//		},
//		{
//			Name:     consts.FieldNameChunkIndex,
//			DataType: entity.FieldTypeInt32,
//		},
//		{
//			Name:     consts.FieldNameVector,
//			DataType: entity.FieldTypeFloatVector,
//			TypeParams: map[string]string{
//				"dim": strconv.Itoa(dimension),
//			},
//		},
//	},
//}
