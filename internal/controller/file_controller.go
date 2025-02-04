package controller

import (
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FileController struct {
	fileService service.FileService
}

func NewFileController(fileService service.FileService) *FileController {
	return &FileController{fileService: fileService}
}

func (fc *FileController) Upload(ctx *gin.Context) {
	// 1. 获取用户ID
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 2. 解析表单文件
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "上传失败")
		return
	}

	// 获取文件内容
	file, err := fileHeader.Open()
	if err != nil {

		response.ParamError(ctx, errcode.FileParseFailed, "上传失败")
		return
	}
	defer file.Close()

	// 4. 获取父目录ID（可选参数）
	parentID := ctx.PostForm("parent_id") // 空字符串表示根目录

	// 调用 Service 层处理文件上传
	err = fc.fileService.UploadFile(userID, fileHeader, file, parentID)
	if err != nil {
		response.InternalError(ctx, errcode.FileUploadFailed, "上传失败")
		return
	}
	//// 获取文件的url？作用是？
	//url, err := fc.fileService.GetFileURL(newFile.StorageKey)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, common.Error(100, "获取文件URL失败"))
	//	return
	//}
	response.SuccessWithMessage(ctx, "文件上传成功", nil)

}

func (fc *FileController) List(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}
	// 获取父目录，处理根目录情况
	parentID := ctx.Query("parent_id")
	var parentIDPtr *string
	if parentID != "" {
		parentIDPtr = &parentID
	}

	page, pageSize, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "分页参数错误")
		return
	}

	sort := ctx.DefaultQuery("sort", "name:asc")
	if err := utils.ValidateSortParameter(sort, []string{"name", "updated_at"}); err != nil {
		response.ParamError(ctx, errcode.ParamValidateError, "排序参数错误")
		return
	}

	total, files, err := fc.fileService.PageList(userID, parentIDPtr, page, pageSize, sort)
	if err != nil {
		response.InternalError(ctx, errcode.FileListFailed, "获取文件列表失败")
		return
	}

	response.PageSuccess(ctx, files, total)
}

func (fc *FileController) Download(ctx *gin.Context) {
	fileID := ctx.Query("file_id")

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	fileMeta, fileData, err := fc.fileService.DownloadFile(fileID)

	if err != nil {
		response.InternalError(ctx, errcode.FileNotFound, "文件不存在")
		return
	}

	if userID != fileMeta.UserID {
		response.UnauthorizedError(ctx, errcode.ForbiddenError, "权限不足")
		return
	}
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileMeta.Name))
	ctx.Header("Content-Type", fileMeta.MIMEType)
	ctx.Header("Content-Length", strconv.FormatInt(fileMeta.Size, 10))
	ctx.Data(http.StatusOK, fileMeta.MIMEType, fileData)
}
