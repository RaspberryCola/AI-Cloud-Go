package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (fc *FileController) PageList(ctx *gin.Context) {
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

func (fc *FileController) Delete(ctx *gin.Context) {
	fileID := ctx.Query("file_id")
	if fileID == "" {
		response.ParamError(ctx, errcode.ParamValidateError, "参数错误")
		return
	}
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户未认证")
		return
	}

	if err := fc.fileService.DeleteFileOrFolder(userID, fileID); err != nil {
		response.InternalError(ctx, errcode.FileDeleteFailed, "删除失败")
		return
	}

	response.SuccessWithMessage(ctx, "删除成功", nil)
}

func (fc *FileController) CreateFolder(ctx *gin.Context) {
	//var req models.CreateFolderRequest
	//if err := ctx.ShouldBind(&req); err != nil {
	//	ctx.JSON(http.StatusBadRequest, common.Error(100, "参数错误"))
	//	return
	//}
	req := model.CreateFolderReq{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	if req.ParentID != nil && *req.ParentID == "" {
		req.ParentID = nil
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	err = fc.fileService.CreateFolder(userID, req.Name, req.ParentID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "文件夹创建失败")
		return
	}
	response.SuccessWithMessage(ctx, "创建成功", nil)
}

// BatchMove 批量移动文件/文件夹
func (fc *FileController) BatchMove(ctx *gin.Context) {
	var req model.BatchMoveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	// 获取当前用户ID
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 执行批量移动
	if err := fc.fileService.BatchMoveFiles(userID, req.FileIDs, req.TargetParentID); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "移动失败")
		return
	}

	response.SuccessWithMessage(ctx, "移动成功", nil)
}

func (fc *FileController) Search(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户校验失败")
		return
	}

	key := ctx.Query("key")

	page, pageSize, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		response.ParamError(ctx, errcode.ParamValidateError, "分页参数错误")
		return
	}

	sort := ctx.DefaultQuery("sort", "name:asc")
	if err := utils.ValidateSortParameter(sort, []string{"name", "upadated_at"}); err != nil {
		response.ParamError(ctx, errcode.ParamValidateError, "排序参数错误")
	}

	total, files, err := fc.fileService.SearchList(userID, key, page, pageSize, sort)
	if err != nil {
		response.InternalError(ctx, errcode.FileSearchFailed, "搜索文件失败")
	}

	response.PageSuccess(ctx, files, total)

}

func (fc *FileController) Rename(ctx *gin.Context) {
	var req model.RenameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	if err = fc.fileService.Rename(userID, req.FileID, req.NewName); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, fmt.Sprintf("重命名失败 %s", err))
		return
	}

	response.SuccessWithMessage(ctx, "重命名成功", nil)

}
