package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

type KBController struct {
	kbService   service.KBService
	fileService service.FileService
}

func NewKBController(kbService service.KBService, fileService service.FileService) *KBController {
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

// 删除知识库
func (kc *KBController) Delete(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	kbID := ctx.Query("kb_id")

	// 删除知识库
	if err := kc.kbService.DeleteKB(userID, kbID); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "删除知识库成功", nil)
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

func (kc *KBController) DocPage(ctx *gin.Context) {
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
	kbID := ctx.Query("kb_id")
	total, docs, err := kc.kbService.DocList(userID, kbID, page, pageSize)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取列表失败")
		fmt.Printf(err.Error())
		return
	}
	response.PageSuccess(ctx, docs, total)
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
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取文件信息失败")
		return
	}

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

	folderID, err := kc.fileService.InitKnowledgeDir(userID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "初始化知识库目录失败"+err.Error())
		return
	}
	fileID, err := kc.fileService.UploadFile(userID, fileHeader, file, folderID)
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

func (kc *KBController) Retrieve(ctx *gin.Context) {
	// 1. 获取用户ID
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 2. 解析请求参数
	var req struct {
		KBID  string `json:"kb_id"`
		Query string `json:"query"`
		TopK  int    `json:"top_k"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	// 3. 调用服务层检索
	chunks, err := kc.kbService.Retrieve(userID, req.KBID, req.Query, req.TopK)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		return
	}

	// 4. 返回结果
	response.Success(ctx, chunks)
}

func (kc *KBController) Chat(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 2. 解析请求参数
	var req model.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	// 3. 调用服务层处理
	resp, err := kc.kbService.RAGQuery(userID, req.Query, req.KBs)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		return
	}

	// 4. 返回结果
	response.Success(ctx, resp)

}

func (kc *KBController) ChatStream(ctx *gin.Context) {
	// 1. 获取用户ID
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 2. 解析请求参数
	var req model.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误")
		return
	}

	// 3. 设置响应头
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	// 4. 调用服务层获取流式响应
	responseChan, err := kc.kbService.RAGQueryStream(ctx.Request.Context(), userID, req.Query, req.KBs)
	if err != nil {
		ctx.SSEvent("error", err.Error())
		return
	}

	// 5. 发送流式响应
	for response := range responseChan {
		data, _ := json.Marshal(response)
		ctx.Writer.Write([]byte("data: " + string(data) + "\n\n"))
		ctx.Writer.Flush()
	}
}

func (kc *KBController) GetKBDetail(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}

	// 获取知识库ID
	kbID := ctx.Query("kb_id")
	if kbID == "" {
		response.ParamError(ctx, errcode.ParamBindError, "知识库ID不能为空")
		return
	}

	// 获取知识库详情
	kb, err := kc.kbService.GetKBDetail(userID, kbID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取知识库详情失败")
		return
	}

	response.Success(ctx, kb)
}
