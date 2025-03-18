package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"github.com/gin-gonic/gin"
)

type KBController struct {
	kbService   service.KBService
	fileService service.FileService
}

func NewKBCotroller(kbService service.KBService, fileService service.FileService) *KBController {
	return &KBController{kbService: kbService, fileService: fileService}
}

func (kc *KBController) Create(ctx *gin.Context) {
	// 1. 获取用户ID
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}
	if err := kc.kbService.CreateDB(req.Name, req.Description, userID); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "创建失败")
		return
	}

	response.SuccessWithMessage(ctx, "创建知识库成功", nil)
}

func (kc *KBController) PageList(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	page, pageSize, err := utils.ParsePaginationParams(ctx)
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "分页参数错误")
		return
	}

	total, kbs, err := kc.kbService.PageList(userID, page, pageSize)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取知识库列表失败")
		return
	}

	response.PageSuccess(ctx, kbs, total)
}

func (kc *KBController) AddFile(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}
	req := model.AddFileRequest{}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}
	file, err := kc.fileService.GetFileByID(req.FileID)

	// 添加文件到知识库
	doc, err := kc.kbService.CreateDocument(userID, req.KBID, file)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "添加文件到知识库失败")
		return
	}
	if err = kc.kbService.ProcessDocument(doc); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "添加文件到知识库成功", nil)
}
