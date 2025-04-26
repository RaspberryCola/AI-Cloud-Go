package controller

import (
	"ai-cloud/internal/model"
	"ai-cloud/internal/service"
	"ai-cloud/internal/utils"
	"ai-cloud/pkgs/errcode"
	"ai-cloud/pkgs/response"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ModelController struct {
	svc service.ModelService
}

func NewModelController(svc service.ModelService) *ModelController {
	return &ModelController{svc: svc}
}

func (c *ModelController) CreateModel(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "获取用户失败")
		return
	}

	var req model.CreateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误:"+err.Error())
		return
	}

	m := &model.Model{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		Type:      req.Type,
		ShowName:  req.ShowName,
		Server:    req.Server,
		BaseURL:   req.BaseURL,
		ModelName: req.ModelName,
		APIKey:    req.APIKey,
		// embedding
		Dimension: req.Dimension,
		// llm
		MaxOutputLength: req.MaxOutputLength,
		Function:        req.Function,
		// common
		MaxTokens: req.MaxTokens,
	}

	if err := c.svc.CreateModel(ctx.Request.Context(), m); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, err.Error())
		fmt.Println("创建模型失败:", err)
		return
	}

	response.SuccessWithMessage(ctx, "创建模型成功", nil)
}

// TODO:修改返回格式
func (c *ModelController) UpdateModel(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "获取用户失败")
		return
	}

	var req model.UpdateModelRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误:"+err.Error())
		return
	}

	m := &model.Model{
		ID:        req.ID,
		UserID:    userID,
		ShowName:  req.ShowName,
		Server:    req.Server,
		BaseURL:   req.BaseURL,
		ModelName: req.ModelName,
		APIKey:    req.APIKey,
		// embedding
		Dimension: req.Dimension,
		// llm
		MaxOutputLength: req.MaxOutputLength,
		Function:        req.Function,
		// common
		MaxTokens: req.MaxTokens,
	}

	if err := c.svc.UpdateModel(ctx.Request.Context(), m); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "更新模型失败："+err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "更新模型成功", nil)
}

func (c *ModelController) DeleteModel(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "用户验证失败")
		return
	}
	kbID := ctx.Query("model_id")

	if err := c.svc.DeleteModel(ctx.Request.Context(), userID, kbID); err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "删除模型失败："+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "删除模型成功", nil)
}

func (c *ModelController) GetModel(ctx *gin.Context) {
	// 获取用户ID并验证
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "获取用户失败")
		return
	}

	modelID := ctx.Query("model_id")

	m, err := c.svc.GetModel(ctx.Request.Context(), userID, modelID)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取模型失败："+err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "获取模型成功", m)
}

func (c *ModelController) PageModels(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "获取用户失败")
		return
	}

	var req model.PageModelRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.ParamError(ctx, errcode.ParamBindError, "参数错误："+err.Error())
		return
	}

	models, count, err := c.svc.PageModels(ctx.Request.Context(), userID, req.Type, req.Page, req.Size)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取模型列表失败："+err.Error())
		return
	}

	response.PageSuccess(ctx, models, count)
}

func (c *ModelController) ListModels(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		response.UnauthorizedError(ctx, errcode.UnauthorizedError, "获取用户失败")
		return
	}

	modelType := ctx.Query("type")
	models, err := c.svc.ListModels(ctx.Request.Context(), userID, modelType)
	if err != nil {
		response.InternalError(ctx, errcode.InternalServerError, "获取模型列表失败："+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "获取模型列表成功", models)
}
