package utils

import (
	"fmt"
	"strings"
)

func ValidateSortParameter(sort string, allowedFields []string) error {
	clauses := strings.Split(sort, ",")

	for _, clause := range clauses {
		parts := strings.Split(clause, ":")
		field, order := parts[0], parts[1]
		if !contains(allowedFields, field) {
			return fmt.Errorf("无效的排序字段：%s", field)
		}
		if order != "asc" && order != "desc" {
			return fmt.Errorf("无效的排序方向：%s", order)
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
