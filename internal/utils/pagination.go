package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ParsePaginationParams(ctx *gin.Context) (page int, pageSize int, err error) {
	pageStr := ctx.DefaultQuery("page", "1")
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0, fmt.Errorf("无效的页码参数")

	}

	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return 0, 0, fmt.Errorf("无效的页面大小参数")
	}

	return page, pageSize, nil
}
