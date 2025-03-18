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

func (kc *KBController) AddExistFile(ctx *gin.Context) {
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
	// 处理解析文档
	if err = kc.kbService.ProcessDocument(doc); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "添加文件到知识库成功", nil)
}

// 上传新的文件到知识库
func (kc *KBController) AddNewFile(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 获取知识库ID
	kbID := ctx.PostForm("kb_id")
	if kbID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "知识库ID不能为空")
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "文件上传失败")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "文件打开失败")
		return
	}
	defer file.Close()

	folerID, err := kc.fileService.InitKnowledgeDir(userID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "初始化知识库目录失败"+err.Error())
		return
	}
	fileID, err := kc.fileService.UploadFile(userID, fileHeader, file, folerID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "文件上传失败")
		return
	}

	// 将文件添加到知识库中
	f, err := kc.fileService.GetFileByID(fileID)
	if err != nil || f == nil { // 添加对 nil 的检查
		response.InternalError(ctx, errcode.InternalServerError, "获取文件信息失败")
		return
	}
	doc, err := kc.kbService.CreateDocument(userID, kbID, f)
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
