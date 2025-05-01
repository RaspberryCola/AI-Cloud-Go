package milvus

import "ai-cloud/pkgs/consts"

var (
	defaultSearchFields = []string{
		consts.FieldNameID,
		consts.FieldNameContent,
		consts.FieldNameKBID,
		consts.FieldNameDocumentID,
		consts.FieldNameMetadata,
	}
)
